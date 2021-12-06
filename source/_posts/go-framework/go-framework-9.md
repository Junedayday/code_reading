---
title: Go语言微服务框架 - 9.分布式链路追踪-OpenTracing的初步引入
date: 2021-10-26 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

我们从API层到数据库层的链路已经打通，简单的CRUD功能已经可以快速实现。

随着模块的增加，我们会越发感受到系统的复杂性，开始关注系统的可维护性。这时，有个名词会进入我们的视野：**分布式链路追踪**。相关的内容可以参考这我的两篇文章：

- OpenTelemetry https://junedayday.github.io/2021/10/14/readings/go-digest-2/
- Jaeger https://junedayday.github.io/2021/10/20/readings/go-digest-3/

我们接下来直接进入实战。

<!-- more -->

## v0.6.0：分布式链路追踪-OpenTracing的初步引入

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.6.0

### 目标

在项目中引入Jaeger为代表的OpenTracing，用一个traceid串联整个请求的链路。

### 关键技术点

1. trace的初始化
2. 将opentracing的设置到grpc和grpc-gateway中
3. 将traceid引入到log组件中
4. HTTP请求返回traceid

> 前两点我将一笔带过，在 https://junedayday.github.io/2021/10/20/readings/go-digest-3/ 这篇中已有详细的讲解

### 目录构造

```
--- micro_web_service            项目目录
	|-- gen                            从idl文件夹中生成的文件，不可手动修改
	   |-- idl                             对应idl文件夹
	      |-- demo                             对应idl/demo服务，包括基础结构、HTTP接口、gRPC接口
	    	|-- order                            对应idl/order服务，同上
	|-- idl                            原始的idl定义
	   |-- demo                            业务package定义，protobuffer的原始定义
	   |-- order                           业务order定义，同时干
	|-- internal                       项目的内部代码，不对外暴露
	   |-- config                          配置相关的文件夹，viper的相关加载逻辑
	   |-- dao                             Data Access Object层，是model层的实现
	   |-- gormer                          从pkg/gormer中生成的相关代码，不允许更改
	   |-- model                           model层，定义对象的接口方法，具体实现在dao层
	   |-- mysql                           MySQL连接
	   |-- server                          服务器的实现，对idl中定义服务的具体实现
	   |-- service                         service层，作为领域实现的核心部分
     |-- zlog                            封装zap日志的代码实现
  |-- pkg                            开放给第三方的工具库
     |-- gormer                          gormer二进制工具，用于生成Gorm相关Dao层代码
	|-- buf.gen.yaml                   buf生成代码的定义，从v1beta升到v1
	|-- buf.yaml                       buf工具安装所需的工具，从v1beta升到v1
	|-- gen.sh                         生成代码的脚本：buf+gormer
	|-- go.mod                         Go Module文件
	|-- gormer.yaml                    将gormer中的参数移动到这里
	|-- main.go                        项目启动的main函数
```

## 1.trace的初始化

创建了一个jaeger的trace并设置到opentracing包里的全局变量中。

```go
traceCfg := &jaegerconfig.Configuration{
  ServiceName: "MyService",
  Sampler: &jaegerconfig.SamplerConfig{
    Type:  jaeger.SamplerTypeConst,
    Param: 1,
  },
  Reporter: &jaegerconfig.ReporterConfig{
    LocalAgentHostPort: "127.0.0.1:6831",
    LogSpans:           true,
  },
}
tracer, closer, err := traceCfg.NewTracer(jaegerconfig.Logger(jaeger.StdLogger))
if err != nil {
  panic(err)
}
defer closer.Close()
opentracing.SetGlobalTracer(tracer)
```

## 2.将opentracing的设置到grpc和grpc-gateway中

利用了拦截器的特性，类似于middleware。

```go
// grpc-gateway
opts := []grpc.DialOption{
  grpc.WithInsecure(),
  grpc.WithUnaryInterceptor(
    grpc_opentracing.UnaryClientInterceptor(
      grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
    ),
  ),
}

if err := demo.RegisterDemoServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf(":%d", config.Viper.GetInt("server.grpc.port")), opts); err != nil {
  return errors.Wrap(err, "RegisterDemoServiceHandlerFromEndpoint error")
}

// grpc
s := grpc.NewServer(grpc.UnaryInterceptor(grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer()))))

```

## 3.将traceid引入到log组件中

从Opentracing对Go语言的相关介绍可以得知，trace信息被放在go语言的context里。于是，就有了下面这一段提取traceid的代码。

```go
// 为了使用方便，不修改zap源码，这里利用With函数返回一个SugaredLogger
func WithTrace(ctx context.Context) *zap.SugaredLogger {
	var jTraceId jaeger.TraceID
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentCtx := parent.Context()
		if tracer := opentracing.GlobalTracer(); tracer != nil {
			mySpan := tracer.StartSpan("my info", opentracing.ChildOf(parentCtx))
      // 提取出一个jaeger的traceid
			if sc, ok := mySpan.Context().(jaeger.SpanContext); ok {
				jTraceId = sc.TraceID()
			}
			defer mySpan.Finish()
		}
	}

	return Sugar.With(zap.String(jaeger.TraceContextHeaderName, fmt.Sprint(jTraceId)))
}
```

## 4.HTTP请求返回traceid

在拦截器里，解析出trace信息，设置到http的头里。

```go
trace, ok := serverSpan.Context().(jaeger.SpanContext)
if ok {
  w.Header().Set(jaeger.TraceContextHeaderName, fmt.Sprint(trace.TraceID()))
}
```

## 示例

我们模拟一个简单的请求

```shell
curl --request GET 'http://127.0.0.1:8081/v1/orders'
```

从返回的结果来看，可以看到`Uber-Trace-Id`头里有个具体的trace-id，例如5fd1fc3ba1715909。

而在应用代码中，我们添加了一行日志：

```go
func (orderSvc *OrderService) List(ctx context.Context, pageNumber, pageSize int, condition *gormer.OrderOptions) ([]gormer.Order, int64, error) {
	zlog.WithTrace(ctx).Infof("page number is %d", pageNumber)
	// zlog信息
	return orders, count, nil
}
```

具体的打印如下:

```
2021-10-22T17:25:05.591+0800	info	service/order.go:26	page number is 0	{"uber-trace-id": "5fd1fc3ba1715909"}
```

虽然格式还不是那么优美，但traceid信息已经填入到了日志中。

至此，调用方只要提供返回的trace-id，我们就可以在程序日志中查找到相应的日志信息，方便针对性地排查问题。

## 总结

OpenTracing是服务治理非常关键的一环。利用traceid串联一个请求的整个生命周期，能帮助我们快速地排查问题，在实际生产环境上能更快地定位问题。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

