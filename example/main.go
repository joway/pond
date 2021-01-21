package main

import (
	"context"
	"fmt"
	"github.com/joway/pond"
	"log"
)

type conn struct {
	addr string
}

func main() {
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
}
