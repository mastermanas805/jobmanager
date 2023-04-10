package scheduler

import (
	"context"
	"fmt"
	"time"
)

func panic_test(ctx context.Context, cancel context.CancelFunc) {
	slice := []int{1, 2, 3}
	fmt.Println(slice[10])
	fmt.Println("Continue after panic")
}

func infiniteloop_test(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	for i := 0; true; i++ {
		select {
		case <-ctx.Done():
			panic("Context Cancelled")
		default:
			fmt.Println(i)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func helloworld(ctx context.Context, cancel context.CancelFunc) {
	fmt.Println("Hello World")
}

var FuncOptions = map[string]func(ctx context.Context, cancel context.CancelFunc){
	"panic":         panic_test,
	"infinite_loop": infiniteloop_test,
	"hello_world":   helloworld,
}
