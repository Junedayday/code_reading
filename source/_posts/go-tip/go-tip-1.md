---
title: Go语言技巧 - 1.【惊艳亮相】如何写出一个优雅的main函数
date: 2021-03-07 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![Go-Study](https://i.loli.net/2021/02/28/BnVH86E5owhsaFd.jpg)

## 一个简单的main函数

我们先来看看一个最简单的`http服务端`的实现

```go
// http服务
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	http.ListenAndServe(":8080", mux)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello")
}
```

它的功能很简单，我们就是一个监听在8080端口，会处理`URL`为`/hello`的请求，并打印出hello。

可以用一个简单的curl请求来打印结果：

```shell
curl http://localhost:8080/hello
```

也可以用对应的`kill`杀死了对应的进程：

```shell
kill -9 {pid}
```

但有一个问题：

**如果程序因为代码问题而退出（例如panic），无法和kill这种人为强制杀死的情况进行区分**



## 引入signal

`kill`的其是`Linux`系统中，往进程发送一个信号。所以，我们的关键是去实现 **捕获信号** 的功能。

```go
// http服务
func main() {
	// 创建一个 sig 的 channel，捕获系统的信号，传递到sig中
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	// http服务改造成异步
	go http.ListenAndServe(":8080", mux)

	// 程序阻塞在这里，除非收到了interrupt或者kill信号
	fmt.Println(<-sig)
}
```

至此，我们的主函数已经能区分正常的信号了。



## 优雅退出的需求

服务端程序经常会处理各种各样的逻辑，如操作数据库、读写文件、RPC调用等。根据其对 **原子性** 的要求，我将它的工作区分为两种：

- 一种是无严格要求的，即程序直接崩溃也没有问题，比如一个普通查询；
- 另一种是有 **原子性** 要求的，即不希望运行到一半就退出，例如写文件、修改数据等，**最好是程序提供一定的缓冲时间**，等待这部分的逻辑处理完，优雅地退出。

在复杂系统中，为了保证数据质量，**优雅退出** 是一个必要特性。

```go
func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)

	// 这里类似于http的handler，有一些业务逻辑，在不断地并发处理
	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				// 我们希望程序能让当前这个周期休眠完，再优雅退出
				time.Sleep(time.Duration(i) * time.Second)
			}
		}(i)
	}

	fmt.Println(<-sig)
}
```

这里是一个简单的示例，开启了10个`goroutine`并发处理，那么这时捕获信号后，这10个协程就立刻停止。而**优雅退出**，则是希望能执行完当前的`Sleep`再退出。



## 一对一的解决方案

我们先简化问题：主函数对应的是一个需要优雅关闭的协程。

```go
func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)

	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()

	fmt.Println(<-sig)
}
```

整体操作如下：

- 父`goroutine`通知子`goroutine`准备优雅地关闭
- 子`goroutine`通知父`goroutine`已经关闭完成

我们回忆下在`goroutine`传递消息的几个方案（排除共享的全局变量这种方式）。



### 最直观的解决方案 - 2个channel

既然我们要在父子goroutine中传递消息，最直接的想法是启用2个 `channel` 用来通信，对应到代码：

- 父`goroutine`通知子`goroutine`准备优雅地关闭，也就是`stopCh`

- 子`goroutine`通知父`goroutine`已经关闭完成，也就是`finishedCh`

  具体代码实现如下

```go
func main() {
	sig := make(chan os.Signal)
	stopCh := make(chan struct{})
	finishedCh := make(chan struct{})
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)

	go func(stopCh, finishedCh chan struct{}) {
		for {
			select {
			case <-stopCh:
				fmt.Println("stopped")
				finishedCh <- struct{}{}
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}(stopCh, finishedCh)

	<-sig
	stopCh <- struct{}{}
	<-finishedCh
	fmt.Println("finished")
}
```



### channel嵌套channel

这个解决方案不太容易想到（可以参考Rob Pike的演讲视频）。它的核心结构体为`chan chan`。

示例代码如下：

```go
func main() {
	sig := make(chan os.Signal)
	stopCh := make(chan chan struct{})
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)

	go func(stopChh chan chan struct{}) {
		for {
			select {
			case ch := <-stopCh:
				// 结束后，通过ch通知主goroutine
				fmt.Println("stopped")
				ch <- struct{}{}
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}(stopCh)

	<-sig
	// ch作为一个channel，传递给子goroutine，待其结束后从中返回
	ch := make(chan struct{})
	stopCh <- ch
	<-ch
	fmt.Println("finished")
}
```



### 引入上下文context

`go`语言里的上下文`context`不仅仅可以传递数值，也可以控制子`goroutine`的生命周期。我们很自然地有了如下解决方案。

实例代码如下：

```go
func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)
	ctx, cancel := context.WithCancel(context.Background())
	finishedCh := make(chan struct{})

	go func(ctx context.Context, finishedCh chan struct{}) {
		for {
			select {
			case <-ctx.Done():
				// 结束后，通过ch通知主goroutine
				fmt.Println("stopped")
				finishedCh <- struct{}{}
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}(ctx, finishedCh)

	<-sig
	cancel()
	<-finishedCh
	fmt.Println("finished")
}
```



## 一对多的解决方案

一对多的解决方案可以复用 **一对一解决方案** 中的思想。我这边也给出另外一个 `context` + `sync.WaitGroup` 的解决方案。

```go
func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)
	ctx, cancel := context.WithCancel(context.Background())
	num := 10

	// 用wg来控制多个子goroutine的生命周期
	wg := sync.WaitGroup{}
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func(ctx context.Context) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					fmt.Println("stopped")
					return
				default:
					time.Sleep(time.Duration(i) * time.Second)
				}
			}
		}(ctx)
	}

	<-sig
	cancel()
	// 等待所有的子goroutine都优雅退出
	wg.Wait()
	fmt.Println("finished")
}
```

大家要注意一下，在追求 **优雅退出** 时要注意 **控制细粒度** 。

> 比如一个`http`服务器，我们要控制整个`http server`的优雅退出。
>
> 千万不要去想着在主函数层面去控制每个`handler`，也就是每个`http`请求的优雅退出，这样很难保证代码的复杂度。对于每个`http`请求的控制，应该交给`http server`这个框架去实现。
>
> 所以，在主函数中，其实需要优雅退出的选项其实很有限。



## 总结

`main`函数是`go`程序的入口，如果在这里写出一段优雅的代码，很容易给阅读自己源码的朋友留下良好的印象。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

