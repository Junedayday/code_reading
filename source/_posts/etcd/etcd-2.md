---
title: etcd源码分析 - 2.【打通核心流程】PUT键值对匹配处理函数
date: 2022-06-28 12:00:00
categories: 
- 技术框架
tags:
- Go-etcd
---

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd.jpg)

在阅读了etcd server的启动流程后，我们对很多关键性函数的入口都有了初步印象。

那么，接下来我们一起看看对键值对的修改，在etcd server内部是怎么流转的。

<!-- more -->

## PUT键值对的HTTP请求

用`etcdctl`这个指令，我们可以快速地用命令`etcdctl put key value`发送PUT键值对的请求。

但`etcdctl`对请求做了封装，我们要了解原始的HTTP请求格式，才能方便地阅读相关代码。相关的途径有很多，比如抓包、读源码等，这里为了可阅读性，我给出一个`curl`请求。

```shell
curl -L http://localhost:2379/v3/kv/put -X POST -d '{"key":"mykey","value":"myvalue"}'
```

主要关注如下三点：

1. Method - `POST`
2. URL -` /v3/kv/put`
3. Body - `{"key":"mykey","value":"myvalue"}`

> 这个请求是v3版本的，而v2版本的差异比较大，暂不细谈。

## Mux的路由匹配

## 背景知识介绍

为了更好地介绍下面的内容，我先介绍mux下的2个概念。

- `pattern`指的是一种URL的匹配模式，最常见的如全量匹配、前缀匹配、正则匹配。当一个请求进来时，它会有自己的一个URL，去匹配`mux`中预先定义的多个`pattern`，找到一个最合适的。这是一种**URL路由规则的实现**。
- 当请求匹配到一个`pattern`后，就会执行它预定义的`handler`，也就是一个处理函数，返回结果。

所以， `pattern`负责匹配，而`handler`负责执行。在不同语境下，它们的专业术语有所差异，大家自行对应即可。

### http mux的创建

我们要找HTTP1.X的路由匹配逻辑，就回到了上一节最后看到的代码中：

```go
// 创建路由匹配规则
httpmux := sctx.createMux(gwmux, handler)

// 新建http server对象
srvhttp := &http.Server{
  Handler:  createAccessController(sctx.lg, s, httpmux),
  ErrorLog: logger, // do not log user error
}
// 这个cumx.HTTP1是检查协议是否满足HTTP1
httpl := m.Match(cmux.HTTP1())
// 运行server
go func() { errHandler(srvhttp.Serve(httpl)) }()
```

### （*serveCtx)createMux

本函数不长，但很容易让读源码的同学陷入误区，我们一起来看看。这块代码主要分为三段：

```go
func (sctx *serveCtx) createMux(gwmux *gw.ServeMux, handler http.Handler) *http.ServeMux {
	httpmux := http.NewServeMux()
  // 1.注册handler
	for path, h := range sctx.userHandlers {
		httpmux.Handle(path, h)
	}

  // 2.注册grpcGateway mux中的handler到/v3/路径下
	if gwmux != nil {
		httpmux.Handle(
			"/v3/",
			wsproxy.WebsocketProxy(
				gwmux,
				wsproxy.WithRequestMutator(
					// Default to the POST method for streams
					func(_ *http.Request, outgoing *http.Request) *http.Request {
						outgoing.Method = "POST"
						return outgoing
					},
				),
				wsproxy.WithMaxRespBodyBufferSize(0x7fffffff),
			),
		)
	}
  // 3.注册根路径下的handler
	if handler != nil {
		httpmux.Handle("/", handler)
	}
	return httpmux
}
```

第一点，可以通过简单的代码阅读，看到是对`pprof`和`debug`这些通用功能的URL功能注册，也是一些用户自定义的`handler`注册，这就很好地对应到`sctx.userHandlers`这个变量的命名了。

第三点很快就能被排除，它注册的是对根路径下的handler。我们阅读代码，找到handler最原始的生成处，就能看到它是对version、metrcis这类handler的注册。

所以，我们的重点就放在了`gwmux`这个对象上。阅读它的创建过程，就得跳转到上层函数。

### (*serveCtx)registerGateway

在函数中，我们可以看到它注册了一个类型为`registerHandlerFunc`的handlers列表，包括如下内容：

```go
handlers := []registerHandlerFunc{
		etcdservergw.RegisterKVHandler, // KV键值对的处理
		etcdservergw.RegisterWatchHandler, // Watch监听
		etcdservergw.RegisterLeaseHandler, // Lease租约
		etcdservergw.RegisterClusterHandler, // 集群
		etcdservergw.RegisterMaintenanceHandler, // 维护相关
		etcdservergw.RegisterAuthHandler, // 认证
		v3lockgw.RegisterLockHandler, // 锁
		v3electiongw.RegisterElectionHandler, // 选举
}
for _, h := range handlers {
  if err := h(ctx, gwmux, conn); err != nil {
    return nil, err
  }
}
```

我们聚焦到PUT请求的处理，它自然走的是`etcdservergw.RegisterKVHandler`这个入口。

### RegisterKVHandler

本函数位于`etcd/etcdserver/etcdserverpb/gw/rpc.pb.gw.go`。它其实是用protobuf自动生成的，其中用到了`grpc-gateway`这个关键性技术，它的作用是将HTTP1的请求转换成gRPC，实现一个server可以同时支持HTTP1与gRPC，并且只写一份gRPC处理的代码即可。

> 有兴趣地可以去看看 https://github.com/grpc-ecosystem/grpc-gateway 项目。
>
> 大致调用链路为： HTTP1 -> gRPC -> 自己实现的handler

### RegisterKVHandlerClient

该函数是由proto文件生成的，这里我忽略了关于context的处理，提取关键性的内容：

```go
mux.Handle("POST", pattern_KV_Put_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
  // 反序列化请求和序列化响应
  inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
  // 执行PUT请求
  resp, md, err := request_KV_Put_0(rctx, inboundMarshaler, client, req, pathParams)
  // 返回结果
  forward_KV_Put_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)
})
```

序列化与反序列化存在多种选择，我们暂不深入，先来看看处理这部分的工作：

首先是如何匹配请求，也就是`http://localhost:2379/v3/kv/put`，对应如下：

```go
var pattern_KV_Put_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"v3", "kv", "put"}, ""))
```

而最核心的处理，也就是解析PUT请求的函数`request_KV_Put_0`与返回处理结果的函数`forward_KV_Put_0`，我们放到下一讲再来看。

## 小结

今天我们看了`PUT`请求在etcd server中通过`mux`的匹配逻辑，思路参考下图。

在阅读代码期间，我们接触到了grpc-gateway这个技术方案，有兴趣的朋友可以参考我的[另一篇文章](https://junedayday.github.io/2021/08/19/go-framework/go-framework-1/)。

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd-2-mux.drawio.png)

> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

