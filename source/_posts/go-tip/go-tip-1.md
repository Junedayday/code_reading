---
title: Go语言技巧 - 1.【惊艳亮相】如何写出一个优雅的main函数
date: 2021-03-07 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![Go-Study](https://i.loli.net/2021/05/05/2bmr98tG3xDneL5.jpg)

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

它的功能很简单：提供一个监听在`8080`端口的服务器，处理`URL`为`/hello`的请求，并打印出hello。

可以用一个简单的curl请求来打印结果：

```shell
curl http://localhost:8080/hello
```

也可以用对应的`kill`杀死了对应的进程：

```shell
kill -9 {pid}
```

但有一个问题：

**如果程序因为代码问题而意外退出（例如panic），无法和kill这种人为强制杀死的情况进行区分**

<!-- more -->

## 引入signal

`kill`工具是`Linux`系统中，往进程发送一个信号。所以，我们的关键是去实现 **捕获信号** 的功能。

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

至此，我们的主函数已经能区分正常的信号退出了。



## 优雅退出的需求

服务端程序经常会处理各种各样的逻辑，如操作数据库、读写文件、RPC调用等。根据其对 **原子性** 的要求，我将处理逻辑区分为两种：

- 一种是**无严格数据质量**要求的，即程序直接崩溃也没有问题，比如一个普通查询；
- 另一种是有 **原子性** 要求的，即不希望运行到一半就退出，例如写文件、修改数据等，**最好是程序提供一定的缓冲时间**，等待这部分的逻辑处理完，优雅地退出。

在复杂系统中，为了保证数据质量，**优雅退出** 是一个必要特性。

```go
func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)

	// 模拟并发进行的处理业务逻辑
	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				// 我们希望程序能等当前这个周期休眠完，再优雅退出
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



### 华丽的解决方案 - channel嵌套channel

这个解决方案不太容易想到（看过Rob Pike的演讲视频除外，可在go官网看到）。

这个方案的核心结构为`chan chan`。

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

> 这个方案很酷，建议大家多思考思考，尤其是channel中传递的数据为error时，就能有更多信息了



### 标准解决方案 - 引入上下文context

`go`语言里的上下文`context`不仅仅可以传递数值，也可以控制子`goroutine`的生命周期，很自然地有了如下解决方案。

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



> 有兴趣的朋友可以空闲时想一个问题：社区里有人认为context是一个很不好的实现：
>
> context意思为上下文，最初的设计意为传递数值，也就是一个 **数据流** ；
>
> 而go中的context又延伸出了 控制goroutine生命周期的功能，也就成了 **控制流** 。
>
> 这么看下来，其实context就有 角色不清晰 的味道了。
>
> 但不可否认，context已经在go语言中大量被采用，这个问题可以作为大家自己设计模块时的参考。



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
> 千万不要去想着在主函数层面去控制每个`http handler`，也就是每个`http`请求的优雅退出，这样很难控制代码的复杂度。对于每个`http`请求的控制，应该交给`http server`这个框架去实现。
>
> 所以，在主函数中，其实需要优雅退出的选项其实很有限。



## 延伸思考

本次我们讲的是`main`函数控制其`goroutine`的优雅退出，其实我们延伸开来，就是 **父Goroutine怎么保证子Goroutine优雅退出** 这个问题。

虽然有解决方案，但我这是想泼一盆冷水，希望大家想想一个问题：**既然这个子Goroutine是有价值的，不想轻易丢失，那么为什么不放到主Goroutine中呢？** 其实，很多时候，我们都在 **滥用Goroutine** 。我希望大家更多地抛开语言特性，从整体思考以下三个问题：

1. **明确调用链路** - 梳理整个调用流程，区分关键和非关键的步骤，以及在对应步骤上发生错误时的处理方法
2. **用MQ解耦服务** - 跨服务的调用如果比较费时，大部分时候更建议采用消息队列解耦
3. **面向错误编程** - 关键业务的`Goroutine` 里代码要考虑所有可能发生错误的点，保证程序退出或`panic/recover`也不要出现 **脏数据**。



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

