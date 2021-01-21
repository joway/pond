package pond

import (
	"context"
	"github.com/jolestar/go-commons-pool/v2"
	"strconv"
	"testing"
	"time"
)

func getPool() (*Pool, error) {
	cfg := NewConfig(testObjectCreateFactory)
	cfg.MaxSize = 8
	cfg.MaxIdle = 8
	cfg.MinIdle = 0
	cfg.MinIdleTime = 30 * time.Minute
	return New(cfg)
}

func getCommonPool() *pool.ObjectPool {
	factory := pool.NewPooledObjectFactorySimple(
		func(ctx context.Context) (interface{}, error) {
			var name string
			cval := ctx.Value(contextKeyName{})
			if cval != nil {
				name = cval.(string)
			}
			return &testObject{
				name: name,
			}, nil
		},
	)
	ctx := context.Background()
	return pool.NewObjectPoolWithDefaultConfig(ctx, factory)
}

func BenchmarkPool(b *testing.B) {
	ctx := context.Background()
	p, _ := getPool()
	for i := 0; i < b.N; i++ {
		name := strconv.Itoa(i)
		obj, err := p.BorrowObject(context.WithValue(ctx, contextKeyName{}, name))
		if err != nil {
			panic(err)
		}
		o := obj.(*testObject)
		if o.name == "" {
			panic("name should not be empty")
		}
		err = p.ReturnObject(ctx, obj)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkPoolWithConcurrent(b *testing.B) {
	ctx := context.Background()
	p, _ := getPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj, err := p.BorrowObject(ctx)
			if err != nil {
				panic(err)
			}
			err = p.ReturnObject(ctx, obj)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkCommonPool(b *testing.B) {
	ctx := context.Background()
	p := getCommonPool()
	for i := 0; i < b.N; i++ {
		name := strconv.Itoa(i)
		obj, err := p.BorrowObject(context.WithValue(ctx, contextKeyName{}, name))
		if err != nil {
			panic(err)
		}

		o := obj.(*testObject)
		if o.name == "" {
			panic("name should not be empty")
		}
		err = p.ReturnObject(ctx, obj)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkCommonPoolWithConcurrent(b *testing.B) {
	ctx := context.Background()
	p := getCommonPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj, err := p.BorrowObject(ctx)
			if err != nil {
				panic(err)
			}
			err = p.ReturnObject(ctx, obj)
			if err != nil {
				panic(err)
			}
		}
	})
}
