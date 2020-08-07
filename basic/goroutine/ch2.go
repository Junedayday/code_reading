package main

import (
	"fmt"
	"time"
)

// 示例1
type Ball struct {
	hits int
}

func passBall() {
	table := make(chan *Ball)
	go player("ping", table)
	go player("pong", table)

	// Tip: 核心逻辑：往channel里放入数据，作为启动信号；从channel读出数据，作为关闭信号
	table <- new(Ball)
	time.Sleep(time.Second)
	<-table
}

func player(name string, table chan *Ball) {
	for {
		// Tip: 刚进goroutine时，先阻塞在这里
		ball := <-table
		ball.hits++
		fmt.Println(name, ball.hits)
		time.Sleep(100 * time.Millisecond)
		// Tip: 运行到这里时，另一个goroutine在收数据，所以能准确送达
		table <- ball
	}
}

// 示例2
func passBallWithClose() {
	// Tip 虽然可以通过GC自动回收channel资源，但我们仍应该注意这点
	table := make(chan *Ball)
	go playerWithClose("ping", table)
	go playerWithClose("pong", table)

	table <- new(Ball)
	time.Sleep(time.Second)
	<-table
	close(table)
}

func playerWithClose(name string, table chan *Ball) {
	for {
		ball, ok := <-table
		if !ok {
			break
		}
		ball.hits++
		fmt.Println(name, ball.hits)
		time.Sleep(100 * time.Millisecond)
		table <- ball
	}
}

// 示例3
type sub struct {
	// Tip 把chan error看作一个整体，作为关闭的通道
	closing chan chan error
	updates chan string
}

func (s *sub) Close() error {
	// Tip 核心逻辑：两层通知，第一层作为准备关闭的通知，第二层作为关闭结果的返回
	errc := make(chan error)
	// Tip 第一步：要关闭时，先传一个chan error过去，通知要关闭了
	s.closing <- errc
	// Tip 第三步：从chan error中读取错误，阻塞等待
	return <-errc
}

func (s *sub) loop() {
	var err error
	for {
		select {
		case errc := <-s.closing:
			// Tip 第二步：收到关闭后，进行处理，处理后把error传回去
			errc <- err
			close(s.updates)
			return
		}
	}
}
