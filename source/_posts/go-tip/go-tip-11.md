---
title: Go语言技巧 - 11.【初始化代码生成】Wire进阶使用
date: 2021-12-28 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## Wire进阶

通过上一篇的讲解，我们已经掌握`wire`工具的基本用法了。但应用在实际工程中，这些基本功能还是有很多局限性。

在这一篇，我们一起看看Google推出的`wire`的进阶使用方法，并总结出一套实践思路。

<!-- more -->

## 进阶示例

### Set集合

`Set`特性比较直观：组合几个`Provider`

```go
var BasicSet = wire.NewSet(NewGreeter, NewMessage)

func InitializeGreeter() Greeter {
	wire.Build(BasicSet)
	return Greeter{}
}
```

一般应用在初始化对象比较多的情况下，减少`Injector`里的信息。

### 绑定接口

接口这个特性在面向对象编程时非常有意义，我们来看一个具体的示例

```go
// 抽象出一个 Messager 的接口
type Messager interface {
	Name() string
}

// Message 是Messager的一个具体实现
type Message struct{}

func (m *Message) Name() string {
	return "message"
}

func NewMessage() *Message {
	return &Message{}
}

// Greeter的初始化依赖的是Messager接口，而不是Message这个实现
func NewGreeter(m Messager) Greeter {
	return Greeter{Message: m}
}

type Greeter struct {
	Message Messager
}

func (g Greeter) Greet() Messager {
	return g.Message
}
```

不难看出，我们要做的就是在`NewGreeter(m Messager)`初始化时，用`Message`这个具体实现来代替`Messager`接口。这里，我们就在`wire.go`里引入了 **绑定** 这个方法：

```go
// wire.go

var BasicSet = wire.NewSet(
	NewGreeter,
	wire.Bind(new(Messager), new(*Message)),
	NewMessage,
)

func InitializeGreeter() Greeter {
	wire.Build(BasicSet)
	return Greeter{}
}

// wire_gen.go

func InitializeGreeter() Greeter {
	message := NewMessage()
	greeter := NewGreeter(message)
	return greeter
}
```

### 构造结构体

上面的例子里，我们都定义了具体的构造函数，也就是Provider。但实际开发过程中，我们经常会遇到只有一个具体的结构体，而没有定义具体的函数。这时我们可以采用 **构造结构体的特性**。例如，我们定义一个`MyGreeter`：

```go
type MyGreeter struct {
	Msg Message
}

// wire.go

func InitializeMyGreeter() *MyGreeter {
	wire.Build(
		NewMessage,
		wire.Struct(new(MyGreeter), "Msg"),
	)
	return &MyGreeter{}
}

// wire_gen.go

func InitializeMyGreeter() *MyGreeter {
	message := NewMessage()
	myGreeter := &MyGreeter{
		Msg: message,
	}
	return myGreeter
}
```

### 绑定值

```go
type MyGreeter struct {
	X   int
}

// wire.go

func InitializeMyGreeter() *MyGreeter {
	wire.Build(
		wire.Value(&MyGreeter{X: 42}),
	)
	return &MyGreeter{}
}

// wire_gen.go

func InitializeMyGreeter() *MyGreeter {
	myGreeter := _wireMyGreeterValue
	return myGreeter
}

var (
	_wireMyGreeterValue = &MyGreeter{X: 42}
)
```

### 获取结构体中的字段

这块比较简单，就是从一个结构体里提取一个Public的field，作为一个`Provider`，这里给出一个简单的示例。

```go
type Foo struct {
    S string
    N int
    F float64
}

// wire_gen.go

func injectedMessage() string {
    wire.Build(
        provideFoo,
        wire.FieldsOf(new(Foo), "S"))
    return ""
}
```

### 清理函数

清理函数利用了函数变量的特性，将资源释放函数抛出来。

```go
func provideFile(log Logger, path Path) (*os.File, func(), error) {
    f, err := os.Open(string(path))
    if err != nil {
        return nil, nil, err
    }
    cleanup := func() {
        if err := f.Close(); err != nil {
            log.Log(err)
        }
    }
    return f, cleanup, nil
}
```

## 最佳实践

### 1.区别类型

采用类型别名，和标准类型区分开来，如

```go
type MySQLConnectionString string
```

### 2. 可选结构体

当一个`Injector`需要多个`Provider`时，将这些`Provider`集中到一个`Option`的结构体，即组合多个参数，如

```go
type Options struct {
    // Messages is the set of recommended greetings.
    Messages []Message
    // Writer is the location to send greetings. nil goes to stdout.
    Writer io.Writer
}
```

### 3.合理使用Provider Sets

Set集合了多个Provider效率很高，具体实践过程中要根据实际情况出发，参考 https://github.com/google/wire/blob/main/docs/best-practices.md#provider-sets-in-libraries。

总体来说把握一个原则：`In general, prefer small provider sets in a library. ` 即Set尽量小，多多考虑复合。

### 4.Mocking

Mock这块主要是用于测试，官方给出了两个途径：

- Pass mocks to the injector
- Return the mocks from the injector

初看可能不容易理解，我们结合实际代码就能了解

```go
// 途径1 - 即依赖项以参数注入，这样返回的app和正常的app完全一致
func initMockedAppFromArgs(mt timer) *app {
	wire.Build(appSetWithoutMocks)
	return nil
}

// 途径2 - 内部增加mock的具体field，会与app中的对应变量绑定
func initMockedApp() *appWithMocks {
	wire.Build(mockAppSet)
	return nil
}

type appWithMocks struct {
	app app
	mt  *mockTimer
}
```

整体来说，我个人比较推荐使用方案1，它能保证mock对象的使用方式和真实对象完全一致，能屏蔽很多复杂度。在一个复杂系统中，底层的mock对象可以很容易应用到高层。

## 参考资料

Github - https://github.com/google/wire

Blog - https://go.dev/blog/wire 

Package Doc - https://pkg.go.dev/github.com/google/wire

## 思考

通过这一篇，我们能看到`wire`很多进阶的能力，其实还有一部分特性并未在文档中说明，可以参考package doc学习。

我更建议大家可以从单元测试的角度切入，去理解这个工具的实践：

1. **自底向上地考虑wire的实践**：尤其是db、redis这些基础工具，底层的mock会为上层的mock带来巨大便利；
2. **不断抽离和组合对象中的依赖**：依赖小到某个关键变量、大到某个外部服务，也同时注意组合相似度高的依赖项到`Set`；

## 小结

`wire`的核心是依赖注入，对整个框架的可测试性来说是根基，对`Go`语言这类静态编译的语言尤为重要。

`Java`语言有一整套强大的`JVM`引擎，可以在运行时做各种复杂操作；而静态语言在编写时就决定了程序的基本运行方式，从简单性来说很棒 - **所见即所得**，但也说明了不应有复杂的运行时。这时，一个良好的依赖注入机制对`Go`语言尤为重要。

那么，`Wire`引入的DI思想对`Go`语言来说可以称得上是`银弹`，但我们更应该重视基础库的`Mock`能力，毕竟没有底层能力、就没有高层建设。

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

