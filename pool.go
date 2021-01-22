package pond

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrPoolClosed                  = errors.New("pool has been closed")
	ErrPoolFulled                  = errors.New("pool is full")
	ErrPoolExhausted               = errors.New("pool is exhausted")
	ErrObjectNotFound              = errors.New("object not found")
	ErrObjectValidateFailed        = errors.New("object validate failed")
	ErrObjectCreateFactoryNotFound = errors.New("the factory of object creating not found")
)

//Pool is a thread-safe pool
type Pool struct {
	manager    *poolManager
	config     Config
	actionLock sync.RWMutex //lock for borrow/return/evict/... actions
	returnedCh chan struct{}

	evictorTicker *time.Ticker
	closed        bool
}

//New create a pool by config
func New(config Config) (*Pool, error) {
	if config.ObjectCreateFactory == nil {
		return nil, ErrObjectCreateFactoryNotFound
	}
	p := &Pool{
		manager:    newPoolManager(),
		config:     config,
		returnedCh: make(chan struct{}, 1),
	}
	if config.AutoEvict {
		p.evictorTicker = time.NewTicker(p.config.EvictInterval)
		go p.StartEvictor()
	}
	return p, nil
}

func (p *Pool) isClosed() bool {
	return p.closed
}

func (p *Pool) isFull() bool {
	if p.config.MaxSize <= 0 {
		return false
	}
	return p.manager.Size() >= p.config.MaxSize
}

func (p *Pool) ActiveSize() int {
	p.actionLock.RLock()
	defer p.actionLock.RUnlock()
	return p.manager.ActiveSize()
}

func (p *Pool) IdleSize() int {
	p.actionLock.RLock()
	defer p.actionLock.RUnlock()
	return p.manager.IdleSize()
}

func (p *Pool) Size() int {
	p.actionLock.RLock()
	defer p.actionLock.RUnlock()
	return p.manager.Size()
}

func (p *Pool) createObject(ctx context.Context) error {
	if p.isFull() {
		return ErrPoolFulled
	}
	//create a new one
	object, err := p.config.ObjectCreateFactory(ctx)
	if err != nil {
		return err
	}
	p.manager.Create(object)
	return nil
}

//BorrowObject promise to return a idle object. It will be blocked when there is no any idle object.
func (p *Pool) BorrowObject(ctx context.Context) (interface{}, error) {
	var object interface{}
	validateCount := 0
	for object == nil {
		p.actionLock.Lock()
		var err error
		object, err = p.borrowObject(ctx)
		p.actionLock.Unlock()
		if err == nil {
			return object, nil
		}

		switch err {
		case ErrObjectNotFound:
			select {
			case <-p.returnedCh:
			//wait for context canceled
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		case ErrObjectValidateFailed:
			validateCount++
			if validateCount > p.config.MaxValidateAttempts {
				return nil, ErrObjectValidateFailed
			}
		default:
			return nil, err
		}
	}

	return object, nil
}

func (p *Pool) borrowObject(ctx context.Context) (interface{}, error) {
	if p.isClosed() {
		return nil, ErrPoolClosed
	}
	//if there is no idle objects
	if p.manager.IdleSize() <= 0 {
		if p.isFull() {
			//if pool is exhausted, and NonBlocking enabled
			if p.config.Nonblocking {
				return nil, ErrPoolExhausted
			}
		} else {
			//if pool is not full, just create a new object
			if err := p.createObject(ctx); err != nil {
				return nil, err
			}
		}
	}

	po := p.manager.Borrow()
	if po == nil {
		return nil, ErrObjectNotFound
	}

	object := po.Object()
	//validate object
	vFactory := p.config.ObjectValidateFactory
	success := true
	if vFactory != nil {
		success = vFactory(ctx, object)
	}
	if !success {
		_ = p.invalidateObject(ctx, object)
		return nil, ErrObjectValidateFailed
	}
	return object, nil
}

//InvalidateObject delete and destroy the active object
func (p *Pool) InvalidateObject(ctx context.Context, object interface{}) error {
	p.actionLock.Lock()
	defer p.actionLock.Unlock()
	return p.invalidateObject(ctx, object)
}

func (p *Pool) invalidateObject(ctx context.Context, object interface{}) error {
	p.manager.Deactivate(object)
	return p.destroyObject(ctx, object)
}

func (p *Pool) ReturnObject(ctx context.Context, object interface{}) error {
	p.actionLock.Lock()
	defer p.actionLock.Unlock()

	if p.isClosed() {
		//if return after closing, just invalidate object
		return p.invalidateObject(ctx, object)
	}

	p.manager.Return(object)

	p.returned()
	return nil
}

func (p *Pool) returned() {
	//returnedCh waiting borrower
	//make sure never blocked
	select {
	case p.returnedCh <- struct{}{}:
	default:
	}
}

func (p *Pool) Evict(ctx context.Context) error {
	p.actionLock.Lock()
	defer p.actionLock.Unlock()

	if p.isClosed() {
		return ErrPoolClosed
	}

	minIdle, maxIdle := p.config.MinIdle, p.config.MaxIdle
	maxSize := p.config.MaxSize
	minIdleTime := p.config.MinIdleTime

	//protect config
	if maxIdle < 0 {
		maxIdle = 0
	}
	if minIdle < 0 {
		minIdle = 0
	}
	if minIdle > maxIdle {
		minIdle = maxIdle
	}

	//evict: pop all idle objects exceed maxIdle
	evicting := p.manager.IdleSize() - maxIdle
	for i := 0; i < evicting; i++ {
		earliest := p.manager.Earliest()
		if earliest.IdleTime() < minIdleTime {
			break
		}

		if !p.evictEarliest(ctx) {
			break
		}
	}

	//warmup: ensure there are at least minIdle objects
	warmup := minIdle - p.manager.IdleSize()
	newSize := p.manager.Size() + warmup
	if newSize > maxSize {
		warmup = maxSize - p.manager.Size()
	}
	for i := 0; i < warmup; i++ {
		if err := p.createObject(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pool) evictEarliest(ctx context.Context) bool {
	popped := p.manager.PopEarliest()
	if popped == nil {
		return false
	}
	_ = p.destroyObject(ctx, popped.Object())
	return true
}

func (p *Pool) destroyObject(ctx context.Context, object interface{}) error {
	if object == nil || p.config.ObjectDestroyFactory == nil {
		return nil
	}
	return p.config.ObjectDestroyFactory(ctx, object)
}

func (p *Pool) StartEvictor() {
	for range p.evictorTicker.C {
		_ = p.Evict(context.Background())
	}
}

func (p *Pool) Close(ctx context.Context) error {
	p.actionLock.Lock()
	defer p.actionLock.Unlock()

	if p.isClosed() {
		return ErrPoolClosed
	}
	p.closed = true

	if p.evictorTicker != nil {
		p.evictorTicker.Stop()
	}

	//destroy all idle objects
	//active object should be destroyed after client returned to make sure pool will not close a active object
	p.manager.RangeIdle(func(object interface{}) {
		_ = p.destroyObject(ctx, object)
	})

	return nil
}
