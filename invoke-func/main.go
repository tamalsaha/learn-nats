package main

import (
	"context"
	"fmt"
	"reflect"
)

type Request struct{}

func run(ctx context.Context, data Request) error {
	return nil
}

func main() {
	ctx := context.Background()
	data := &Request{}
	var fn any = run
	parms := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(data).Elem(),
	}
	results := reflect.ValueOf(fn).Call(parms)
	fnErr, _ := results[0].Interface().(error)
	fmt.Println(fnErr)
}
