package pond

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var loopSize = 500

type contextKeyName struct{}

type testObject struct {
	name string
}

func (o testObject) Close() error {
	return nil
}

var testObjectCreateFactory ObjectCreateFactory = func(ctx context.Context) (interface{}, error) {
	var name string
	cval := ctx.Value(contextKeyName{})
	if cval != nil {
		name = cval.(string)
	}
	return &testObject{name: name}, nil
}

func TestBasicPool(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig(testObjectCreateFactory)
	p, _ := New(cfg)
	for i := 0; i < loopSize; i++ {
		name := strconv.Itoa(i)
		obj, err := p.BorrowObject(context.WithValue(ctx, contextKeyName{}, name))
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, obj.(*testObject).name, "0")
		err = p.ReturnObject(ctx, obj)
		assert.NoError(t, err)
	}
}

func TestPoolAutoEvict(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig(testObjectCreateFactory)
	cfg.MaxSize = 0 //no limit
	cfg.MaxIdle = 0
	cfg.MinIdleTime = time.Millisecond * 100
	cfg.EvictInterval = time.Millisecond * 100
	p, _ := New(cfg)
	objs := make([]*testObject, 0)
	for i := 0; i < loopSize; i++ {
		name := strconv.Itoa(i)
		//without return, always create new object
		obj, err := p.BorrowObject(context.WithValue(ctx, contextKeyName{}, name))
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		testObj := obj.(*testObject)
		assert.Equal(t, testObj.name, name)
		objs = append(objs, testObj)
	}
	assert.Equal(t, 0, p.IdleSize())
	assert.Equal(t, loopSize, p.manager.Size())

	//return one
	err := p.ReturnObject(ctx, objs[0])
	assert.Equal(t, loopSize, p.manager.Size())
	assert.NoError(t, err)
	assert.Equal(t, 1, p.IdleSize())
	//evict one
	time.Sleep(cfg.EvictInterval * 2)
	assert.Equal(t, cfg.MaxIdle, p.IdleSize())
	assert.Equal(t, loopSize-1, p.manager.Size())

	//invalidate one
	err = p.InvalidateObject(ctx, objs[1])
	assert.NoError(t, err)
	assert.Equal(t, 0, p.IdleSize())
	time.Sleep(time.Millisecond * 200)
	assert.Equal(t, 0, p.IdleSize())
	assert.Equal(t, loopSize-2, p.Size())

	//return all
	for i := 2; i < len(objs); i++ {
		err := p.ReturnObject(ctx, objs[i])
		assert.NoError(t, err)
	}
	assert.Equal(t, loopSize-2, p.IdleSize())
	time.Sleep(time.Millisecond * 200)
	assert.Equal(t, 0, p.IdleSize())
	assert.Equal(t, 0, p.manager.Size())
}

func TestPoolEvictManually(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig(testObjectCreateFactory)
	cfg.MaxSize = 100
	cfg.MinIdle = 1
	cfg.MaxIdle = 10
	cfg.AutoEvict = false
	cfg.MinIdleTime = time.Millisecond * 100
	p, _ := New(cfg)
	objs := make([]*testObject, 0)
	for i := 0; i < cfg.MaxSize; i++ {
		name := strconv.Itoa(i)
		obj, err := p.BorrowObject(context.WithValue(ctx, contextKeyName{}, name))
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		testObj := obj.(*testObject)
		assert.Equal(t, testObj.name, name)
		objs = append(objs, testObj)
	}
	assert.Equal(t, p.manager.IdleSize(), 0)
	for _, obj := range objs {
		err := p.ReturnObject(ctx, obj)
		assert.NoError(t, err)
	}
	assert.Equal(t, cfg.MaxSize, p.manager.IdleSize())
	//hit MinIdleTime
	p.Evict(ctx)
	assert.Equal(t, cfg.MaxSize, p.manager.IdleSize())
	//wait MinIdleTime
	time.Sleep(cfg.MinIdleTime)
	p.Evict(ctx)
	assert.Equal(t, cfg.MaxIdle, p.manager.IdleSize())

	latestPObj := p.manager.Latest().Object().(*testObject)
	for i := 0; i < loopSize; i++ {
		name := strconv.Itoa(i)
		obj, err := p.BorrowObject(context.WithValue(ctx, contextKeyName{}, name))
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		testObj := obj.(*testObject)
		//always got latest object
		assert.Equal(t, testObj.name, latestPObj.name)
		err = p.ReturnObject(ctx, obj)
		assert.NoError(t, err)
	}
	time.Sleep(p.config.MinIdleTime)
	p.Evict(ctx)
	assert.Equal(t, p.config.MaxIdle, p.manager.IdleSize())
	obj, err := p.BorrowObject(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	testObj := obj.(*testObject)
	assert.Equal(t, testObj.name, latestPObj.name)
}

func TestPoolEvictPolicy(t *testing.T) {
}

func TestPoolConcurrent(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig(testObjectCreateFactory)
	p, _ := New(cfg)
	var wg sync.WaitGroup
	for i := 0; i < loopSize; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			obj, err := p.BorrowObject(context.WithValue(ctx, contextKeyName{}, strconv.Itoa(n)))
			defer func() {
				err := p.ReturnObject(ctx, obj)
				assert.NoError(t, err)
			}()

			time.Sleep(time.Millisecond)
			assert.NoError(t, err)
			assert.NotNil(t, obj)
			assert.True(t, obj.(*testObject).name != "")
		}(i)
	}
	wg.Wait()
}

func TestPoolConcurrentEvict(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig(testObjectCreateFactory)
	cfg.MaxSize = 10
	cfg.AutoEvict = false
	p, _ := New(cfg)
	stopCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(time.Nanosecond):
				p.Evict(ctx)
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < loopSize; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			obj, err := p.BorrowObject(context.WithValue(ctx, contextKeyName{}, strconv.Itoa(n)))
			defer func() {
				err := p.ReturnObject(ctx, obj)
				assert.NoError(t, err)
			}()
			assert.NoError(t, err)
			assert.NotNil(t, obj)
			assert.True(t, obj.(*testObject).name != "")
		}(i)
	}
	wg.Wait()
	close(stopCh)
}

func TestPoolBasicValidate(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig(testObjectCreateFactory)
	validate := false
	cfg.ObjectValidateFactory = func(ctx context.Context, object interface{}) bool {
		return validate
	}
	p, _ := New(cfg)
	defer p.Close(ctx)

	//test validate true
	validate = true
	obj, err := p.BorrowObject(ctx)
	assert.NotNil(t, obj)
	assert.NoError(t, err)

	//test validate false
	validate = false
	obj, err = p.BorrowObject(ctx)
	assert.Nil(t, obj)
	assert.Equal(t, ErrObjectValidateFailed, err)
}

func TestPoolValidateRetry(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig(testObjectCreateFactory)
	attempts := 0
	cfg.ObjectValidateFactory = func(ctx context.Context, object interface{}) bool {
		defer func() { attempts++ }()
		return attempts > 0
	}
	p, _ := New(cfg)
	defer p.Close(ctx)

	//test validate failed at first time and success at second time
	attempts = 0
	obj, err := p.BorrowObject(ctx)
	assert.NotNil(t, obj)
	assert.NoError(t, err)
}

func TestPoolNonblocking(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig(testObjectCreateFactory)
	cfg.MaxSize = 10
	cfg.Nonblocking = true
	p, _ := New(cfg)
	defer p.Close(ctx)

	objs := make([]interface{}, 0)
	for i := 0; i < cfg.MaxSize; i++ {
		obj, err := p.BorrowObject(ctx)
		assert.NoError(t, err)
		objs = append(objs, obj)
	}
	_, err := p.BorrowObject(ctx)
	assert.Equal(t, ErrPoolExhausted, err)

	//return one and borrow two
	p.ReturnObject(ctx, objs[0])
	_, err = p.BorrowObject(ctx)
	assert.NoError(t, err)
	_, err = p.BorrowObject(ctx)
	assert.Equal(t, ErrPoolExhausted, err)
}
