---
title: Go语言技巧 - 15.【Go并发编程】自顶向下地写出优雅的Goroutine
date: 2022-02-22 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## 导语

`Go`语言的Goroutine特性广受好评，初学者也能快速地实现并发。但随着不断地学习与深入，有很多开发者都陷入了对`goroutine`、`channel`、`context`、`select`等并发机制的迷惑中。

这里，我将结合一个具体示例，自顶向下地介绍这部分的知识，帮助大家形成体系。具体代码以下面这段为例：

```go
// parent goroutine
func Foo() {
	go SubFoo()
}

// children goroutine
func SubFoo() {
}
```

这里的`Foo()`为**父Goroutine**，内部开启了一个**子Goroutine** - `SubFoo()`。

## Part1 - 父子Goroutine的生命周期管理

### 聚焦核心

**父Goroutine** 与 **子Goroutine** 最重要的交集 - 是两者的生命周期管理。包括三种：

1. **互不影响** - 两者完全独立、各自运行
2. **parent控制children** - 父Goroutine结束时，子Goroutine也能随即结束
3. **children控制parent** - 子Goroutine结束时，父Goroutine也能随即结束

这个生命周期的关系，重点体现的是两个协程之间的控制关系。

> 注意，这时不要过于关注具体的代码实现，如数据传递，容易绕晕。

### 1-互不影响

两个Goroutine互不影响的代码很简单，如同示例。

不过我们要注意一点，如果子goroutine需要context这个入参，尽量新建。更具体的内容我们看下一节。

### 2-parent控制children

下面是一个最常见的用法，也就是利用了context：

```go
// parent goroutine
func Foo() {
	ctx, cancel := context.WithCancel(context.Background())
	// 退出前执行，表示parent执行完了
	defer cancel()

	go SubFoo(ctx)
}

// children goroutine
func SubFoo(ctx context.Context) {
	select {
	case <-ctx.Done():
		// parent完成后，就退出
		return
	}
}
```

当然，context并不是唯一的解法，我们也可以自建一个channel用来通知关闭。但综合考虑整个Go语言的生态，更建议大家尽可能地使用context，这里不扩散了。

> 延伸 - 如果1个parent要终止多个children时，context的这种方式依然适用，而channel就很麻烦了。

### 3-children控制parent

逻辑也比较直观，我们直接看代码：

```go
// parent goroutine
func Foo() {
	var ch = make(chan struct{})
	go SubFoo(ch)

	select {
	// 获取通知并退出
	case <-ch:
		return
	}
}

// children goroutine
func SubFoo(ch chan<- struct{}) {
	// 通知parent的channel
	ch <- struct{}{}
}
```

#### 情况3的延伸

如果1个parent产生了n个children时，又会有以下两种情况：

1. n个children都结束了，才停止parent
2. n个children中有m个结束，就停止parent

其中，前者的最常用的解决方案如下：

```go
// parent goroutine
func Foo() {
	var wg = new(sync.WaitGroup)
	wg.Add(3)

	go SubFoo(wg)
	go SubFoo(wg)
	go SubFoo(wg)

	wg.Wait()
}

// children goroutine
func SubFoo(wg *sync.WaitGroup) {
	defer wg.Done()
}
```

这两种延伸情况有很多种解法，有兴趣的可以自行研究，网上也有不少实现。

### Par1小结

从生命周期入手，我们能在脑海中快速形成代码的基本结构：

1. 互不影响 - 注意context独立
2. parent控制children - 优先用context控制
3. children控制parent - 一对一时用channel，一对多时用sync.WaitGroup等

但在实际的开发场景中，parent和children的处理逻辑会有很多复杂的情况，导致我们很难像示例那样写出优雅的`select`等方法，我们会在下节继续分析，但不会影响这里梳理出的框架。

## Part2 - for+select的核心机制

一次性的select机制的代码比较简单，单次执行后即退出，讨论的意义不大。接下来，我将重点讨论for+select相关的代码实现。

### for+select的核心机制

我们看一个来自官方的斐波那契数列的例子：

```go
package main

import "fmt"

func fibonacci(c, quit chan int) {
	x, y := 0, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}

func main() {
	c := make(chan int)
	quit := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(<-c)
		}
		quit <- 0
	}()
	fibonacci(c, quit)
}
```

代码很长，我们聚焦于for+select这块，它实现了两个功能：

1. `c`传递数据
2. `quit`传递停止的信号

这时，如果你花时间去理解这两个channel的传递机制，容易陷入对select理解的误区；而我们应该从更高的维度，去看这两个case中获取到数据后的操作、即case中的执行逻辑，才能更好地理解整块代码。

### 分析select中的case

我们要注意到，在case里代码运行的过程中，整块代码是无法再回到select、去判断各case的（这里不讨论panic，return，os.Exit()等情况）。

以上面的代码为例，如果`x, y = y, x+y`函数的处理耗时，远大于`x`这个通道中塞入数据的速度，那么这个`x`的写入处将长期处于排队的阻塞状态。这时，不适合采用select这种模式。

所以，**select适合IO密集型逻辑，而不适合计算密集型**。也就是说，select中的每个case（包括default），应消耗尽量少的时间，快速回到for循环、继续等待。IO密集型常指文件、网络等操作，它消耗的CPU很少、更多的时间是在等待返回，它能更好地体现出**runtime调度Goroutine的价值**。

> Go 的 select这个关键词，可以结合网络模型中的select进行理解。

### 父子进程中的长逻辑处理

这时，如果我们的父子进程里，就是有那么一长段的业务逻辑，那代码该怎么写呢？我们来看看下面这一段：

```go
// children goroutine
func SubFoo(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-dataCh:
			LongLogic()
		}
	}
}

func LongLogic() {
	// 如1累加到10000000
}
```

由于`LongLogic()`会花费很长的运行时间，所以当外部的context取消了，也就是父Goroutine发出通知可以结束了，这个子Goroutine是无法快速触发到`<-ctx.Done()`的，因为它还在跑`LongLogic()`里的代码。也就是说，子进程生命周期结束的时间点延长到`LongLogic()`之后了。

这个问题的原因在于违背了我们上面讨论的点，即在select的case里包含了计算密集型任务。

> 补充一下，case里包含长逻辑不代表程序一定有问题，但或多或少地不符合for+select+channel的设计理念。

### 两个长逻辑处理

这时，我们再来写个长进程处理，整个代码结构如下：

```go
func SubFoo(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-dataCh:
			LongLogic()
		case <-dataCh2:
			LongLogic()
		}
	}
}
```

这里，`dataCh`和`dataCh2`会产生竞争，也就是两个通道的 **写长期阻塞、读都在等待LongLogic执行完成**。给channel加个buffer可以减轻这个问题，但无法根治，运行一段时间依旧会阻塞。

### 改造思路

有了上面代码的基础，改造思路比较直观了，将`LongLogic`异步化，我们先通过新建协程来简单实现：

```go
// children goroutine
func SubFoo(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-dataCh:
			go LongLogic()
		case <-finishedCh:
			fmt.Println("LongLogic finished")
		}
	}
}

func LongLogic() {
	time.Sleep(time.Minute)
	finishedCh <- struct{}{}
}
```

代码里要注意一个点，如果`LongLogic()`是一段需要CPU密集计算的代码，比如计算1累加到10000，它是没有办法通过channel等其余方式突然中止的。它具备一定的原子性 - **要么不跑，要么跑完，跑的过程中没有外部插手的地方**。

而如果硬要中断`LongLogic()`，那就往往只能杀掉整个进程。

### Part2小结

我们记住for+select代码块设计的核心要领 - IO密集型。Go语言的goroutine特性，更多地是为了解决IO密集型程序的问题所设计的，对计算密集型的任务较其它语言没有太大优势。落到具体实践上，就是**让每个case中代码的运行时间尽可能地短，快速回到for循环里的select去继续监听各个case中的channel**。

上面这段代码比较粗糙，在具体工程中会遇到很多问题，比如无限制地开启了大量的`LongLogic()`协程。我们会在下一节继续来看。

## Part3 - 长耗时功能的优化

通过前面两篇的铺垫，我们对 **父子Goroutine的生命周期管理** 与 **for+select的核心机制** 有了基本的了解，把问题聚焦到了耗时较长的处理函数中。

今天，我们再接着看看在具体工程中的优化点。

### 实时处理

我们先回顾上一讲的这段代码：

```go
case <-dataCh:
	go LongLogic()
```

直觉会认为`go LongLogic()`这里会很容易出现性能问题：当`dataCh`的数据写入速度很快时，有大量的`LongLogic()`还未结束、仍在程序内运行，导致CPU超负荷。

但是，如果这些代码编写的逻辑问题确实就是业务逻辑，即：**程序确确实实需要实时处理这么多的数据**，那我们该怎么做呢？

常规思路中引入 **排队机制** 确实是一个方案，但很容易破坏原始需求 - **实时计算处理**，排队机制会导致延迟，这是业务无法接收的。在现实中，扩增资源是最直观的解决方案，最常见是利用Kubernetes平台的Pod水平扩容机制HPA，保证CPU使用率到达一定程度后自动扩容，而不用在程序中加上限制。

从本质上来说，这个问题是对**实时计算资源**的需求。

### 非实时处理 - 程序外优化

在实际工程中，我们其实往往对实时性要求没有那么高，所以排队等限流机制带来的延时可以接受的，也就是准实时。而综合考虑到研发代码质量的不确定性，迭代过程可能中会引入bug导致调用量暴增，这时限流机制能大幅提升程序的健壮性。

在程序外部，我们可以依赖消息队列进行削峰填谷，如：

- 配置消息积压的告警来保证生产能力与消费能力的匹配
- 配置限流参数来保证不要超过消费者程序的处理极限，避免雪崩

这里的消息队列在软件架构中是一个 **分离生产与消费程序** 的设计，有利于两侧程序的健壮性。在计算密集型的场景中，意义尤为重大，只需要针对计算密集型的消费者进行快速地扩缩容。

### 非实时处理 - 程序内优化

上面消息队列方案虽然很棒，但从系统来说引入了一个新的组件，在业务体量小的场景里，有一种杀鸡用牛刀的感觉，对部分没有消息队列的团队来说成本也较高。

那么，我们尝试在程序中做一下优化。首先，我们在上层要做一次抽象，将逻辑收敛到一个独立的package中(示例中为logic)，方便后续优化

```go
func SubFoo(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-dataCh:
			// logic包内部保证
			logic.Run()
		case result := <-logic.Finish():
			fmt.Println("result", result)
		}
	}
}
```

而logic包中的大致框架如下：

```go
package logic

var finishedCh = make(chan struct{})

func Run() {
	// 在这里引入排队机制
	
	go func() {
		// long time process
		
		<-finishedCh
	}()
}

func Finish() <-chan struct{} {
	return finishedCh
}
```

我们也可以在这里加一个`error`返回，在排队满时返回给调用方，由调用方决定怎么处理，如丢弃或重新排队等。排队机制的代码是业务场景决定的，我就不具体写了。

这种解法，可以类比到一个线程池管理。而更上层的for+select维度来看，类似于一个负责调度任务的master+多个负责执行任务的worker。

### Part3小结

我们分别从三个场景分析了耗时较长的处理函数：

- **实时处理** - 结合Paas平台进行资源扩容
- **非实时处理 - 程序外优化** - 引入消息队列
- **非实时处理 - 程序内优化** - 实现一个线程池控制资源

## 总结

本文分享的内容只是Go并发编程的冰山一角，希望能对大家有所启发，也欢迎与我讨论~



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
> ![二维码](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/my_wechat.jpeg)
>
> 

