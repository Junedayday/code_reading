package main

import (
	"context"
	"fmt"
	"time"
)

// Tip: 通过 cancel 主动关闭
func ctxCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			fmt.Println(ctx.Err())
		case <-time.After(time.Millisecond * 100):
			fmt.Println("Time out")
		}
	}(ctx)

	cancel()
}

// Tip: 通过超时，自动触发
func ctxTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	// 主动执行cancel，也会让协程收到消息
	defer cancel()
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			fmt.Println(ctx.Err())
		case <-time.After(time.Millisecond * 100):
			fmt.Println("Time out")
		}
	}(ctx)

	time.Sleep(time.Second)
}

// Tip: 通过设置截止时间，触发time out
func ctxDeadline() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Millisecond))
	defer cancel()
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			fmt.Println(ctx.Err())
		case <-time.After(time.Millisecond * 100):
			fmt.Println("Time out")
		}
	}(ctx)

	time.Sleep(time.Second)
}

func ctxValue() {
	ctx := context.WithValue(context.Background(), "user", "junedayday")
	go func(ctx context.Context) {
		v, ok := ctx.Value("user").(string)
		if ok {
			fmt.Println("pass user value", v)
		}
	}(ctx)
	time.Sleep(time.Second)
}
