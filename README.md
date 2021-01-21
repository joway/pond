# Pond

Generic Object Pool for Golang.

## Get Started

```go
ctx := context.Background()
cfg := pond.NewDefaultConfig()
cfg.MinIdle = 1
cfg.ObjectCreateFactory = func(ctx context.Context) (interface{}, error) {
	return &conn{addr: "127.0.0.1"}, nil
}
cfg.ObjectValidateFactory = func(ctx context.Context, object interface{}) bool {
	c := object.(*conn)
	return c.addr != ""
}
cfg.ObjectDestroyFactory = func(ctx context.Context, object interface{}) error {
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

## Benchmark

Bench with [go-commons-pool](https://github.com/jolestar/go-commons-pool):

```text
BenchmarkPool-8                          3116902               358 ns/op              71 B/op          2 allocs/op
BenchmarkPoolWithConcurrent-8            3683365               326 ns/op               0 B/op          0 allocs/op
BenchmarkCommonsPool-8                    1828080               669 ns/op             103 B/op          3 allocs/op
BenchmarkCommonsPoolWithConcurrent-8      1715344               703 ns/op              32 B/op          1 allocs/op
```
