---
title: gRPC源码分析(三)：从Github文档了解gRPC的项目细节
date: 2021-02-20 19:34:47
categories: 
- 源码阅读
tags:
- gRPC
---

## [官方Git总览](https://github.com/grpc)

我们先看看GRPC这个项目的总览，主要分三种：

- 基于C实现，包括了 C++, Python, Ruby, Objective-C, PHP, C#
- 其余语言实现的，最主要是go，java，node
- proposal，即grpc的RFC，关于实现、讨论的文档汇总

从这里可以看出，gRPC虽然是支持多语言，但原生的实现并不多。如果想在一些小众语言里引入gRPC，还是有很大风险的，有兴趣的可以搜索下TiDB在探索rust的gRPC的经验分享。



## [gRPC-Go](https://github.com/grpc/grpc-go)

作为一名Go语言开发者，我自然选择从最熟悉的语言入手。同时，值得注意的是，grpc-go是除了`C家族系列`以外使用量最大的repo，加上Go语言优秀的可读性，是一个很好的入门gRPC的阅读材料。

进入项目，整个README.md文档也不长。通常情况下，如果你能啃完这个文档及相关链接，你对这个开源项目就已经超过99%的人了。

对Repo的相关注意事项，大家逐行阅读即可，整体比较简单，我简单列举下关键点：

1. 建议阅读官网文档（恭喜你，上次我们已经读完了官方文档）
2. 在项目中的引入，建议用go mod
3. 优先支持3个Go语言最新发布的版本
4. FAQ中的常见问题，主要关注[package下载问题](https://github.com/grpc/grpc-go#io-timeout-errors)和[如何开启追踪日志](https://github.com/grpc/grpc-go#how-to-turn-on-logging)

通读完成，我们再深入看看[文档细节](https://github.com/grpc/grpc-go#documentation)，Example这块我们在官网的测试中已经看过，我们的接下来重点是godoc和具体细节的文档。



## [go doc](https://godoc.org/google.golang.org/grpc)

#### DefaultBackoffConfig

注意，这个变量被弃用，被挪到 `ConnectParams`里了([详情链接](https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md))。那这个所谓的连接参数是什么用呢？代码不长，我们选择几个比较重要的内容来阅读下，原链接可以[点击这里](https://github.com/grpc/grpc-go/blob/v1.29.x/internal/backoff/backoff.go#L54)。

```go
// Backoff returns the amount of time to wait before the next retry given the
// number of retries.
// 根据retries返回等待时间，可以认为是一种退避策略
func (bc Exponential) Backoff(retries int) time.Duration {
	if retries == 0 {
    // 之前没有retries过，就返回BaseDelay
		return bc.Config.BaseDelay
	}
	backoff, max := float64(bc.Config.BaseDelay), float64(bc.Config.MaxDelay)
  // 等待时间不能超过max，等待时间 = BaseDelay * Multiplier的retries次方
  // Multiplier默认1.6，并不是官方http包中的2
	for backoff < max && retries > 0 {
		backoff *= bc.Config.Multiplier
		retries--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
  // 乘以一个随机因子，数值为(1-Jitter,1+Jitter)，默认为(0.8,1.2)，防止同一时刻有大量请求发出，引起锁的问题
	backoff *= 1 + bc.Config.Jitter*(grpcrand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}
```



#### EnableTracing

用来设置是否开启 trace，追踪日志



#### Code

gRPC的错误码，原代码见[链接](https://github.com/grpc/grpc-go/blob/v1.29.x/codes/codes.go#L29)，我们大概了解其原因即可：

- **OK** 正常
- **Canceled** 客户端取消
- **Unknown** 未知
- **InvalidArgument** 未知参数
- **DeadlineExceeded** 超时
- **NotFound** 未找到资源
- **AlreadyExists** 资源已经创建
- **PermissionDenied** 权限不足
- **ResourceExhausted** 资源耗尽
- **FailedPrecondition** 前置条件不满足
- **Aborted** 异常退出
- **OutOfRange** 超出范围
- **Unimplemented** 未实现方法
- **Internal** 内部问题
- **Unavailable** 不可用状态
- **DataLoss** 数据丢失
- **Unauthenticated** 未认证

读完上面的内容，发现跟HTTP/1.1的Status Code非常相似。



#### CallOption

调用在客户端 `Invoke` 方法中，包括before发送前，after为接收后。

官方提供了几个常用的CallOption，按场景调用。



#### ClientConn

抽象的客户端连接。

值得注意的是，conns是一个map，所以实际可能有多个tcp连接。



#### CodeC

定义了Marshal和Unmarshal的接口，在grpc底层实现是proto，详细可见 [codec](https://github.com/grpc/grpc-go/blob/v1.29.x/encoding/proto/proto.go#L39)



#### Compressor

压缩相关的定义



#### MetaData

元数据，也就是key-value，可以类比到http的header



#### DialOption

客户端新建连接时的选项，按场景调用。



#### ServerOption

服务端监听时的选项，按场景调用。



## 文档

[文档链接](https://github.com/grpc/grpc-go/tree/master/Documentation)

#### [benchmark](https://github.com/grpc/grpc-go/blob/master/Documentation/benchmark.md)

性能测试，有兴趣的可以细看gRPC是从哪几个维度做RPC性能测试的。



#### [Compression](https://github.com/grpc/grpc-go/blob/master/Documentation/compression.md)

可用[encoding.RegisterCompressor](https://github.com/grpc/grpc-go/blob/v1.29.x/encoding/encoding.go#L66)实现自定义的压缩方法。

注意，压缩算法应用于客户端和服务端两侧。



#### [Concurrency](https://github.com/grpc/grpc-go/blob/master/Documentation/concurrency.md)

支持并发，从三个角度分析：

- `ClientConn`支持多个Goroutine
- `Steams`中，`SendMsg`/`RecvMsg`可分别在两个Goroutine中运行，但任何一个方法运行在多个Goroutine上是不安全的
- `Server`每个客户端的invoke会对应一个Server端的Goroutine



#### [Encoding](https://github.com/grpc/grpc-go/blob/master/Documentation/encoding.md)

类似Compression，可用[encoding.RegisterCodec](https://github.com/grpc/grpc-go/blob/v1.29.x/encoding/encoding.go#L105)实现自定义的序列化方法。



#### [go mock](https://github.com/grpc/grpc-go/blob/master/Documentation/gomock-example.md)

用mock生成测试代码，详细可细看。



#### [Authentication](https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-auth-support.md)

认证的相关选项，包括 TLS/OAuth2/GCE/JWT ，一般用前两者即可。



#### [Metadata](https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md)

介绍了Metadata的使用，类比于HTTP/1.1的Header。



#### [Keepalive](https://github.com/grpc/grpc-go/blob/master/Documentation/keepalive.md)

长连接的参数分为3类：

- ClientParameters 客户端侧参数，主要用来探活
- SeverParameters 服务端参数，控制连接时间
- EnforcementPolicy 服务端加强型参数



#### [log level](zhttps://github.com/grpc/grpc-go/blob/master/Documentation/log_levels.md)

四个级别的log level，针对不同场景：

- `Info` 用于debug问题
- `Warning` 排查非关键性的问题
- `Error` gRPC调用出现无法返回到客户端的问题
- `Fatal`  导致程序无法恢复的致命问题



#### [proxy](https://github.com/grpc/grpc-go/blob/master/Documentation/proxy.md)

使用默认的HTTP或HTTPS代理。



#### [rpc error](https://github.com/grpc/grpc-go/blob/master/Documentation/rpc-errors.md)

结合官方提供的错误码，用 `status.New` 或者 `status.Error` 创建错误。



#### [server reflection](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md)

服务端方法映射，跟着教程走即可。

值得一提的是，采用c++中的grpc_cli模块，可以查看指定端口暴露出来的服务详情。



#### [versioning](https://github.com/grpc/grpc-go/blob/master/Documentation/versioning.md)

版本演进，一般情况下每6周一个小版本，紧急修复会打补丁号。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili：https://space.bilibili.com/293775192
>
> 公众号：golangcoding

