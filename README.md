# Pond

![GitHub release](https://img.shields.io/github/tag/joway/pond.svg?label=release)
[![Go Report Card](https://goreportcard.com/badge/github.com/joway/pond)](https://goreportcard.com/report/github.com/joway/pond)
[![codecov](https://codecov.io/gh/joway/pond/branch/master/graph/badge.svg?token=Y1YO11FZKU)](https://codecov.io/gh/joway/pond)
[![CircleCI](https://circleci.com/gh/joway/pond.svg?style=shield)](https://circleci.com/gh/joway/pond)

Generic Object Pool for Golang.

It has been used in production and serve millions of QPS.

## Adopters

- [Hive](https://github.com/joway/hive): A high-efficiency Goroutine Pool.

## Get Started

```go
type conn struct {
    addr string
}

ctx := context.Background()
cfg := pond.NewDefaultConfig()
//required
cfg.ObjectCreateFactory = func (ctx context.Context) (interface{}, error) {
    return &conn{addr: "127.0.0.1"}, nil
}
//optional
cfg.ObjectValidateFactory = func (ctx context.Context, object interface{}) bool {
    c := object.(*conn)
    return c.addr != ""
}
//optional
cfg.ObjectDestroyFactory = func (ctx context.Context, object interface{}) error {
    c := object.(*conn)
    c.addr = ""
    return nil
}

p, err := pond.New(cfg)
if err != nil {
    log.Fatal(err)
}

obj, err := p.BorrowObject(ctx)
if err != nil {
    log.Fatal(err)
}
defer p.ReturnObject(ctx, obj)
fmt.Printf("get conn: %v\n", obj.(*conn).addr)
```

## Configuration

| Option                        | Default        | Description  |
| ------------------------------|:--------------:| :------------|
| MaxSize                       | 10             |The capacity of the pool. If MaxSize <= 0, no capacity limit.|
| MinIdle                       | 0              |The minimum size of the idle objects.|
| MaxIdle                       | 10             |The maximal size of the idle objects. Idle objects exceeding MaxIdle will be evicted.|
| MinIdleTime                   | 5m             |The minimum time that idle object should be reserved.|
| Nonblocking                   | false          |The blocking policy. If true, it will return ErrPoolExhausted when pool is exhausted.|
| AutoEvict                     | true           |Enable auto evict idle objects. When true, pool will create a goroutine to start a evictor.|
| EvictInterval                 | 30s            |The interval between evict.|
| MaxValidateAttempts           | 1              |The maximal attempts to validate object.|
| ObjectCreateFactory           | **required**   |The factory of creating object.|
| ObjectValidateFactory         | none           |The factory of validating object.|
| ObjectDestroyFactory          | none           |The factory of destroying object.|

## Benchmark

Compare with:

- [go-commons-pool](https://github.com/jolestar/go-commons-pool):

```text
BenchmarkPool-8                          3116902               358 ns/op              71 B/op          2 allocs/op
BenchmarkPoolWithConcurrent-8            3683365               326 ns/op               0 B/op          0 allocs/op
BenchmarkCommonsPool-8                    1828080               669 ns/op             103 B/op          3 allocs/op
BenchmarkCommonsPoolWithConcurrent-8      1715344               703 ns/op              32 B/op          1 allocs/op
```
