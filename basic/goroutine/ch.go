package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

func ch() {
	var ch = make(chan int)

	go func(ch chan int) {
		// Tip: 由于channel没有设置长度，所以是阻塞的，逐个发送
		ch <- 1
		ch <- 2
		ch <- 3
		fmt.Println("send finished")
	}(ch)

	for {
		select {
		case i := <-ch:
			fmt.Println("receive", i)
		case <-time.After(time.Second):
			fmt.Println("time out")
			os.Exit(1)
		}
	}
}

func chLimit() {
	var ch = make(chan int)

	// Tip: channel参数设置为 chan<- 和 <-chan，可以有效地防止误用发送和接收，例如这里的chan<-只能用于发送
	go func(ch chan<- int) {
		ch <- 1
		ch <- 2
		ch <- 3
		fmt.Println("send finished")
	}(ch)

	for {
		select {
		case i := <-ch:
			fmt.Println("receive", i)
		case <-time.After(time.Second):
			fmt.Println("time out")
			os.Exit(1)
		}
	}
}

func chClose() {
	var ch = make(chan int)

	go func(ch chan<- int) {
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)
		fmt.Println("send finished")
	}(ch)

	for {
		select {
		case i, ok := <-ch:
			if ok {
				fmt.Println("receive", i)
			} else {
				fmt.Println("channel close")
				os.Exit(0)
			}
		case <-time.After(time.Second):
			fmt.Println("time out")
			os.Exit(1)
		}
	}
}

func chCloseErr() {
	var ch = make(chan int)

	go func(ch chan<- int) {
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)
		fmt.Println("send finished")
	}(ch)

	for {
		select {
		// Tip: 如果这里不判断，那么i就会一直得到chan类型的默认值，如int为0，永远不会停止
		case i := <-ch:
			fmt.Println("receive", i)
		case <-time.After(time.Second):
			fmt.Println("time out")
			os.Exit(1)
		}
	}
}

func chTask() {
	var doneCh = make(chan struct{})
	var errCh = make(chan error)

	go func(doneCh chan<- struct{}, errCh chan<- error) {
		if time.Now().Unix()%2 == 0 {
			doneCh <- struct{}{}
		} else {
			errCh <- errors.New("unix time is an odd")
		}
	}(doneCh, errCh)

	select {
	// Tip: 这是一个常见的Goroutine处理模式，在这里监听channel结果和错误
	case <-doneCh:
		fmt.Println("done")
	case err := <-errCh:
		fmt.Println("get an error:", err)
	case <-time.After(time.Second):
		fmt.Println("time out")
	}
}

func chBuffer() {
	var ch = make(chan int, 3)

	go func(ch chan int) {
		// Tip: 由于设置了长度，相当于一个消息队列，这里并不会阻塞
		ch <- 1
		ch <- 2
		ch <- 3
		fmt.Println("send finished")
	}(ch)

	for {
		select {
		case i := <-ch:
			fmt.Println("receive", i)
		case <-time.After(time.Second):
			fmt.Println("time out")
			os.Exit(1)
		}
	}
}

func chBufferRange() {
	var ch = make(chan int, 3)

	go func(ch chan int) {
		// Tip: 由于设置了长度，相当于一个消息队列，这里并不会阻塞
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)
		fmt.Println("send finished")
	}(ch)

	for i := range ch {
		fmt.Println("receive", i)
	}
}
