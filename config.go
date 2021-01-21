package pond

import (
	"context"
	"time"
)

type ObjectCreateFactory func(ctx context.Context) (interface{}, error)
type ObjectValidateFactory func(ctx context.Context, object interface{}) bool
type ObjectDestroyFactory func(ctx context.Context, object interface{}) error

const (
	DefaultMaxSize             = 10
	DefaultMinIdle             = 0
	DefaultMaxIdle             = 10
	DefaultMinIdleTime         = time.Minute * 5
	DefaultAutoEvict           = true
	DefaultEvictInterval       = time.Second * 30
	DefaultMaxValidateAttempts = 1
)

var (
	DefaultObjectValidateFactory ObjectValidateFactory = func(ctx context.Context, object interface{}) bool {
		return true
	}
	DefaultObjectDestroyFactory ObjectDestroyFactory = func(ctx context.Context, object interface{}) error {
		return nil
	}
)

type Config struct {
	/**
	The capacity of the pool. If MaxSize <= 0, no capacity limit.
	*/
	MaxSize int
	/**
	The minimum size of the idle objects.
	*/
	MinIdle int
	/**
	The maximal size of the idle objects. Idle objects exceeding MaxIdle will be evicted.
	*/
	MaxIdle int
	/**
	The minimum time that idle object should be reserved.
	*/
	MinIdleTime time.Duration
	/**
	Enable auto evict idle objects. When true, pool will create a goroutine to start a evictor.
	*/
	AutoEvict bool
	/**
	The interval between evict.
	*/
	EvictInterval time.Duration
	/**
	The maximal attempts to validate object.
	*/
	MaxValidateAttempts int
	/**
	The factory of creating object.
	*/
	ObjectCreateFactory ObjectCreateFactory
	/**
	The factory of validating object.
	*/
	ObjectValidateFactory ObjectValidateFactory
	/**
	The factory of destroying object.
	*/
	ObjectDestroyFactory ObjectDestroyFactory
}

func NewConfig(objectCreateFactory ObjectCreateFactory) Config {
	cfg := NewDefaultConfig()
	cfg.ObjectCreateFactory = objectCreateFactory
	return cfg
}

func NewDefaultConfig() Config {
	return Config{
		MaxSize:             DefaultMaxSize,
		MinIdle:             DefaultMinIdle,
		MaxIdle:             DefaultMaxIdle,
		MinIdleTime:         DefaultMinIdleTime,
		AutoEvict:           DefaultAutoEvict,
		EvictInterval:       DefaultEvictInterval,
		MaxValidateAttempts: DefaultMaxValidateAttempts,

		ObjectValidateFactory: DefaultObjectValidateFactory,
		ObjectDestroyFactory:  DefaultObjectDestroyFactory,
	}
}
