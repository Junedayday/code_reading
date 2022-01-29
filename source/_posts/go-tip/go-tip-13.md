---
title: Go语言技巧 - 13.【浅析微服务框架】Go-Kit概览
date: 2022-01-23 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## Go-Kit概况

截止到本文发布时，`Go-Kit`在github上的star数为22.2k，超过了我们已经一起看过的`Go-Micro`与`Kratos`。

`Go-Kit`不同于前两者，它更像是一种Go语言的工具集，而不是一种统一化的框架。

<!-- more -->

## 概览

官网 - https://gokit.io/

Github - https://github.com/go-kit/kit

**Go kit is a collection of Go (golang) packages (libraries) that help you build robust, reliable, maintainable microservices.** 官方的定义是一种功能集合collection。

## 三个分层

 官方将`Go-Kit`核心分成了三层，分别为:

1. Transport layer
2. Endpoint layer
3. Service layer

### Transport

**The transport domain is bound to concrete transports like HTTP or gRPC.**

通信协议相关的传输层。

这一层具有很多通信协议相关的功能。比如在`HTTP`协议中，根据返回数据的具体格式，在`HTTP`头中设置`Content-Type`。这部分的功能具有很强的重复性，如果设计良好，往往不需要太多的coding。

### Endpoint

**An endpoint is like an action/handler on a controller; it’s where safety and antifragile logic lives.**

RPC函数的控制层入口。

从一个请求来说，通过`Transport`上一定的路由关系后，就会落到具体的`Endpoint`层。一个具体的`Endpoint`有两个关键的、类型确定的参数：请求与响应。

定义中还讲到了安全与反脆弱性，可以理解为在`Endpoint`层要处理panic等异常信息，不要让`Endpoint`与`Service`层的问题影响到整个服务的稳定性。

### Services

**Services are where all of the business logic is implemented.**

业务逻辑的实现层。

整个框架的对`Service`的定义很宽松，给业务开发很大的空间。如何组织，可以参考其余框架的实践。

## 官方示例

`Go-Kit`的详细信息并不多，我们就从一个官方的示例入手，来更好地了解`Go-Kit`，链接如下：

https://github.com/go-kit/examples/blob/master/stringsvc1/main.go

### 1. 定义服务接口

```go
// 两个功能接口
type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
}
```

### 2.主函数+Transport层

```go
func main() {
	svc := stringService{}

  // Uppercase方法的定义，包括request和response的编解码
	uppercaseHandler := httptransport.NewServer(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		encodeResponse,
	)

  // Count方法的定义，包括request和response的编解码
	countHandler := httptransport.NewServer(
		makeCountEndpoint(svc),
		decodeCountRequest,
		encodeResponse,
	)

  // 将上面两个endpoint注册上去
	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 3. Endpoint层

我们以其中一个`Endpoint`为例：

```go
func makeCountEndpoint(svc StringService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
    // 类型判断
		req := request.(countRequest)
		v := svc.Count(req.S)
		return countResponse{v}, nil
	}
}
```

### 4.Service层

```go
type stringService struct{}

func (stringService) Count(s string) int {
	return len(s)
}
```

## 思考

从整体的调用链路来说，`Go-Kit`的使用方式还是比较简单的。我依旧尝试着从中挑两个改进点和大家聊聊：

### 1.Transport层的定义过于冗余

编解码方式往往可以被包含在协议中。

以`HTTP`为例，可以根据Content-Type去解析数据，而不是在编码中自行定义。这点是`Go-Kit`微服务框架为了兼容各种编解码方式，而引入的额外工作量，我个人反倒是建议可以在这块做一些强限制，提高编写代码的便利性。

### 2.Endpoint层的请求与响应数据类型为interface{}

定义如：`func(_ context.Context, request interface{}) (interface{}, error)`。

对框架来说可以做到统一，但对使用者来说很容易带来不好的体验，比如说：

- 请求要做一次数据转换，这个预期的结构体是什么样的？如果转换不成功又该如何处理？
- 响应通过`interface{}`返回，但如果我们要提取内部某个字段做埋点或监控，又需要做一次转换。

## 小结

我们不妨以`Kratos`中的protobuf提供的`gRPC-Gateway`方案进行对比，其实`Transport`+`Endpoint`层完全可以通过protoc等代码生成工具实现。但是，`Go-Kit`为了兼容各类RPC框架，无法在这一块利用代码生成等技术继续提效，而只能通过人工组合。

这一方面体现了集合类框架的价值 - 兼容性强，但也带来了一个最大的弊端 - 不如体系化的框架那么方便。

官网有一段[关于依赖注入](https://gokit.io/faq/#dependency-injection-mdash-why-is-func-main-always-so-big)的内容，这块的思想与`Kratos`框架提供的`wire`有异曲同工之妙。但在细节的实现上会有一些差异，我们会在后面聊到这块~

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

