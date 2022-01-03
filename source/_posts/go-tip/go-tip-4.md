---
title: Go语言技巧 - 4.【错误的三种处理】探索不同代码风格背后的哲学
date: 2021-06-27 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## 背景介绍

通过前面两讲，我们对错误的认知已经超过很多人了。让我们继续去看看常见项目中对错误的处理方式，探索背后的深意。

在介绍具体的处理方式前，我们先来模拟一个场景：我们要去动物园进行一次游玩，主要行为有

- 进入动物园
- 参观熊猫
- 参观老虎
- 离开动物园

<!-- more -->

## 第一种风格 - 经典Go语言的处理模式

```go
// 一次旅游
type ZooTour1 interface {
	Enter() error // 进入
	VisitPanda(panda *Panda) error // 看熊猫
	VisitTiger(tiger *Tiger) error // 看老虎
	Leave() error // 离开
}

func Tour1(t ZooTour1, panda *Panda, tiger *Tiger) error {
	if err := t.Enter(); err != nil {
		return errors.WithMessage(err, "Enter failed")
	}

	if err := t.VisitPanda(panda); err != nil {
		return errors.WithMessagef(err, "VisitPanda failed, panda is %v", panda)
	}

	if err := t.VisitTiger(tiger); err != nil {
		return errors.WithMessagef(err, "VisitTiger failed, tiger is %v", tiger)
	}

	if err := t.Leave(); err != nil {
		return errors.WithMessage(err, "Leave failed")
	}

	return nil
}
```

这个处理风格非常经典。我们先不深入讨论，看完下一种后再做对比。



## 第二种风格 - 类似Try-Catch的代码风格

```go
type ZooTour2 interface {
	Enter()
	VisitPanda(panda *Panda)
	VisitTiger(tiger *Tiger)
	Leave()

	Err() error // 统一处理
}

func Tour2(t ZooTour2, panda *Panda, tiger *Tiger) error {
	t.Enter()
	t.VisitPanda(panda)
	t.VisitTiger(tiger)
	t.Leave()

	if err := t.Err(); err != nil {
		return errors.WithMessage(err, "ZooTour failed")
	}
	return nil
}
```

这一整块的代码风格非常类似**Try Catch**，即先写业务逻辑，在最后对错误进行集中处理。

> 标准库中的`bufio.Scanner`就是参考这种方式实现的。

不过，由于Go语言对error的处理没有往外抛的机制，所以需要专门针对error做处理：

> 新手千万不要把panic的机制和错误处理混为一谈。

```go
// ZooTour的具体实现，需要保存一个error
type myZooTour struct {
	err error
}

func (t *myZooTour) VisitPanda(panda *Panda) {
  // 遇到错误就要直接返回，再处理其余逻辑
	if t.err != nil {
		return
	}
	// ...
}
```



## 两种风格的对比

如果分别用一个词来形容前两种风格，我倾向于：

1. **过程式的调用**
2. **集中处理错误**

两种风格无法说清孰优孰劣，但有各自适宜的场景，我们来列举两种：



### 不关注错误的发生，而关注错误发生后的统一处理

内部存在大量的`VisitXXX`的函数，业务不关注发生错误的处理逻辑，而是关注整个流程完成后对error的处理。

例如，调用过程中如果出现了某个动物不在的问题，我们不关心，继续访问下一个，最后统一处理一下，看看有多少动物是不在的，打印一下即可。

这时，第二种处理方式明显会更简洁。

> 一般推荐在工具类中采用这种方式，处理的内容比较直观，不会有太多异常case



### 错误有多种分类，会影响到程序的运行逻辑

例如`VisitPanda(panda *Panda)` 可能产生的错误分2类：

- 不影响主流程：例如发现panda不见了，但还要接着继续参观其余动物

- 影响主流程：例如突然收到动物园闭园的通知，不能参观其余动物了

这时，如果我们采用第二种风格，就得在每个函数内部加上很多特殊的业务逻辑：

```go
func (t *myZooTour) VisitTiger(tiger *Tiger) {
  // 要针对特定error进行处理
	if t.err != nil && t.err != ErrorPandaMissing {
		return
	}
	// ...
}
```

很有可能出现一个问题：**把Panda相关的error放到了Tiger里**。

所以，**当错误的类型会影响到代码的运行逻辑，更适合第一种方案**。

> 一般情况下，我们的业务代码都是复杂的，这时候更适合写过程性的代码。



## 第三种风格 - 函数式编程

借用1中的接口定义，我们将它改造成函数式的风格：

```go
type MyFunc func(t ZooTour1) error

func NewEnterFunc() MyFunc {
	return func(t ZooTour1) error {
		return t.Enter()
	}
}

func NewVisitPandaFunc(panda *Panda) MyFunc {
	return func(t ZooTour1) error {
		return t.VisitPanda(panda)
	}
}

func NewVisitTigerFunc(tiger *Tiger) MyFunc {
	return func(t ZooTour1) error {
		return t.VisitTiger(tiger)
	}
}

func NewLeaveFunc() MyFunc {
	return func(t ZooTour1) error {
		return t.Leave()
	}
}
```

然后调用代码示例如下：

```go
func Tour3(t ZooTour1, panda *Panda, tiger *Tiger) error {
	var actions = []MyFunc{
		NewEnterFunc(),
		NewVisitPandaFunc(panda),
		NewVisitTigerFunc(tiger),
		NewLeaveFunc(),
	}

	return ContinueOnError(t, actions)
}

func ContinueOnError(t ZooTour1, funcs []MyFunc) error {
  for _, f := range funcs {
    if err := f(t);err != nil {
      // continue
    }
  }
}

func BreakOnError(t ZooTour1, funcs []MyFunc) error {
  for _, f := range funcs {
    if err := f(t);err != nil {
      // break
    }
  }
}
```

值得一提的是

- `ContinueOnError`表示遇到了error只记录下来，但整个流程继续往下跑
- `BreakOnError`表示遇到了error就直接break，不再跑接下来的`MyFunc`



## 方案三背后的思想与延伸

函数式编程最直观的一个特点是 **延迟执行**，也就是在引用`MyFunc`处不运行，在`ContinueOnError`或`BreakOnError`里才是真正执行的地方。

这个延迟执行的特性，在这里还能达到一个很有意思的效果 - **分离关注点**。

### 关注点1 - 数据结构

样例中的`[]MyFunc`是一个切片，可以简单地理解为**串行执行**，也就是`MyFunc`执行完一个，再执行下一个。

我们可以引入更多的数据结构，例如`[][]MyFunc`，那就可以理解为增加了一层：

每一层中的`[]MyFunc`，代表这里面的所有`MyFunc`是平级的，也就可以采用一定的并发模式来加速执行。

### 关注点2 - 执行逻辑

以`ContinueOnError`或`BreakOnError`为例，它们都是对各种`MyFunc`的处理逻辑。我们还可以引入更多的执行逻辑，比如：

- 容忍特定错误的情况
- 对错误发生的数量有容忍上限
- 保证一定的并发模式

### 流水线的模式

以我们常见的开发流水线为例，常见的包括：代码检查、单元测试、编译、CodeReview、自动化部署等。

这时，数据结构可以用来表示**流水线的结构**，执行逻辑可以用来表示**流水线对异常的处置**。

比如说，我们可以编排为一种串行执行的逻辑：

1. 代码检查
2. 单元测试
3. 编译
4. CodeReview
5. 自动化部署

我们想要加速整个流程，可以考虑修改为：

1. 检查
   1. 代码检查
   2. 单元测试
   3. 编译
2. CodeReview
3. 自动化部署



## 结束语

本文介绍了三种对error的处理方式，代码实现相对简单，大家更需要关注背后的适用场景。

其中，第三种方式是一个很有意思的设计模式，可以帮助大家理解函数式编程的价值。







> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

