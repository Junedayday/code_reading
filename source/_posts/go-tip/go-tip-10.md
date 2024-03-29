---
title: Go语言技巧 - 10.【初始化代码生成】Wire工具基础讲解
date: 2021-12-25 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## Wire概览

在讲解Kratos的过程中，我们引入了google推出的wire这个工具。我们先阅读一下官方的定义：

**Wire is a code generation tool that automates connecting components using dependency injection.**

从关键词入手：

- **code generation 代码生成**，一方面说明了有学习成本，需要了解这个工具的原理；另一方面，也说明了它的目标是消除重复性的coding
- **automates connecting components 自动连接组件**，明确了wire工具的目标是将多个对象组合起来
- **dependency injection 依赖注入**，指明了wire实现自动连接组件的思想。依赖注入是一个很强大的功能，我会在下面结合具体的case聊一聊

我们从具体的case着手，学习wire这个工具。

<!-- more -->

## 基本示例

### 常规实现

我简化了官方的示例，给出一个注释后的代码，方便大家阅读：

```go
package main

// Part-1 Message对象
type Message string

func NewMessage() Message {
	return Message("Hi there!")
}

// Part-2 Greeter对象,依赖Message
func NewGreeter(m Message) Greeter {
	return Greeter{Message: m}
}

type Greeter struct {
	Message Message
}

func (g Greeter) Greet() Message {
	return g.Message
}

func main() {
	message := NewMessage()
	greeter := NewGreeter(message)

	greeter.Greet()
}
```

这里的调用很直观，分为3步：

1. 用`NewMessage`创建`Message`对象
2. 通过`NewGreeter`方法，将`Message`对象注入到`Greeter`对象里
3. 调用`Greeter`的方法，其实内部用到了前面注入的`Message`对象

### 依赖注入

依赖注入的详细定义可以参考链接 - https://en.wikipedia.org/wiki/Dependency_injection，我就不赘述了。这里我用具体的case进行对比，方便大家理解：

```go
type Greeter struct {
	Message Message
}

// 依赖注入
func NewGreeter(m Message) Greeter {
	return Greeter{Message: m}
}

func (g Greeter) Greet() Message {
	return g.Message
}

// 非依赖注入
func NewGreeter() Greeter {
	return Greeter{}
}

func (g Greeter) Greet() Message {
	g.Message = NewMessage()
	return g.Message
}
```

看完例子，可能大家对DI已经有个初步的概念了，我这边再重复一下关键点：

1. `Greeter`的方法`Greet()`会依赖内部的`Message`对象，所以我们说 - **Greeter的实现依赖Message**
2. `Message`的初始化分为两种：创建Greeter对象前和调用Greet方法时，前者被称为**依赖注入**，相当于**在初始化时把依赖项注入进去，而不是使用时再创建**。
3. DI，最直接的好处就是可以很方便地调整这个注入项，比如Greet升级成GreetV2，或者单测的MockGreet。

### 使用wire生成代码

我们先安装wire工具：

```shell
go get github.com/google/wire/cmd/wire
```

再编写一个`wire.go`

```go
//+build wireinject

package main

import "github.com/google/wire"

func InitializeGreeter() Greeter {
	wire.Build(NewGreeter, NewMessage)
	return Greeter{}
}
```

运行命令`wire gen`生成wire_gen.go

```go
// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package main

// Injectors from wire.go:

func InitializeGreeter() Greeter {
	message := NewMessage()
	greeter := NewGreeter(message)
	return greeter
}

```

最后，可以在`main`函数里使用

```go
func main() {
	greeter := InitializeGreeter()

	greeter.Greet()
}
```

### wire的大致实现

可以看到，wire这个工具基本能力就体现在`wire.Build(NewGreeter, NewMessage)`里，把这里面的两个初始化函数串联了起来，形成了一个整体的InitializeGreeter。

## 基本扩展

### 带error的处理

我们新增一个方法，初始化结果里增加一个error返回值：

```go
// Part-3 Greeter对象,依赖Message,并且返回error方法
func NewGreeterV2(m Message) (Greeter, error) {
	if m == "" {
		return Greeter{}, errors.New("empty message")
	}
	return Greeter{Message: m}, nil
}
```

然后在`wire.go`里调整函数返回值增加一个error

```go
func InitializeGreeter() (Greeter, error) {
	wire.Build(NewGreeterV2, NewMessage)
	return Greeter{}, nil
}
```

最后，在`wire_gen.go`里生成了带error的新方法

```go
func InitializeGreeterV2() (Greeter, error) {
	message := NewMessage()
	greeter, err := NewGreeterV2(message)
	if err != nil {
		return Greeter{}, err
	}
	return greeter, nil
}
```

### 增加一个入参

我们新增一个方法，增加一个name的入参

```go
// Part-3 Greeter对象,依赖Message和参数name,并且返回error方法
func NewGreeterV3(m Message, name string) (Greeter, error) {
	if name == "" {
		return Greeter{}, errors.New("empty name")
	}
	return Greeter{Message: m}, nil
}
```

`wire.go`里也增加一个`string`类型的入参（变量名可以任意）

```go
func InitializeGreeterV3(greetName string) (Greeter, error) {
	wire.Build(NewGreeterV3, NewMessage)
	return Greeter{}, nil
}
```

最后生成对应的方法

```go
func InitializeGreeterV3(greetName string) (Greeter, error) {
	message := NewMessage()
	greeter, err := NewGreeterV3(message, greetName)
	if err != nil {
		return Greeter{}, err
	}
	return greeter, nil
}
```

## Provider和Injector

Wire里面提了两个关键性的概念，为了方便大家阅读文档时能快速理解，我这里再专门说明下：

- **Provider** - 即各个初始化函数，如`NewXXX`
- **Injector** - 即Initial的函数，将各个Provider注入到wire中，生成一个新的初始化函数

## 参考资料

Github - https://github.com/google/wire

DI - https://en.wikipedia.org/wiki/Dependency_injection

## 思考

`wire`工具的实现逻辑很清晰 - **按一定规则组装多个Provider到Injector中**。

生成的代码 **结构简单而具有规律**，所以用代码生成技术很有价值，既减少了重复性工作，又能引入DI的思想方便程序的扩展。

## 小结

至此，我们对wire的基础用法已经了然于胸，但更多的价值需要深入理解DI这个概念，最好能结合到具体的工程实践上。如果你对这块还没有太深刻的理解，建议结合网上的相关资料了解DI在工程中的价值，会让你使用wire这个工具时更有感触。

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

