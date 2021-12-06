---
title: 技术阅读摘要 - 2.OpenTelemetry技术概览
date: 2021-10-14 12:00:00
categories: 
- 成长分享
tags:
- Digest
---

![Go-Framework](https://i.loli.net/2021/08/15/QZ3lGpkvgdfXW7R.jpg)

## 概览

本系列的第二讲，我原先计划聊一下OpenTracing这个技术，但计划赶不上变化，我发现OpenTracing的官网上已经声明:这部分的技术将迁移到OpenTelemetry。

从OpenTelemetry的官方定义来看： **An observability framework for cloud-native software**，它的重点在于两块：

1. 可观察性：通过metrics、logs和traces数据，观察软件的运行情况
2. 云原生：适配云原生理念

OpenTelemetry的图标采用了一个**望远镜**，可见其核心在于可观察性。

<!-- more -->

## 官方资料

- OpenTelemetry https://opentelemetry.io/ 
- OpenTe4lemetry中文文档 https://github.com/open-telemetry/docs-cn 
- OpenTracing https://opentracing.io/ 
- OpenCensus https://opencensus.io/ 

## 核心概念

我们先引用官方对自身的定义：**OpenTelemetry is a set of APIs, SDKs, tooling and integrations that are designed for the creation and management of *telemetry data* such as traces, metrics, and logs.**

这句话指明了OpenTelemetry实现的3个重点数据：traces、metrics、logs。我们从简单到复杂，逐个讲述一下：

### Logs

日志：依赖程序自身的打印。可通过ELK/EFK等工具采集到统一的平台并展示。

### Metrics

指标：程序将运行中关键的一些指标数据保存下来，常通过RPC的方式Pull/Push到统一的平台。

常见的如请求数、请求延迟、请求成功率等，也可进行一定的计算后获得更复杂的复合指标。

### Traces

分布式追踪：遵循Dapper等协议，获取一个请求在整个系统中的调用链路。

常见的如根据一个HTTP请求的requestID，获取其各个RPC、数据库、缓存等关键链路中的详情。

## 技术标准

到今天，OpenTelemetry还没有完全落地，但这不妨碍我们看清未来的发展方向。

**Metrics以Prometheus为标准，Traces以Jaeger为标准，而Logs暂时还没有明确的标准**，但业界基本以ELK或EFK为技术实现。而我们常会把Traces和Logs这两点结合起来，通过在应用程序的打印日志中添加对应的Traces，来更好地排查整个数据链路。

但这样还不够，Opentelemetry期望的是将三者都关联起来，而引入了Context这个概念。熟悉Go语言的同学都清楚，context被定义为上下文，用于程序中传递数据。而Opentelemetry将这个概念进一步扩大，包括了RPC请求、多线程、跨语言、异步调用等各种复杂场景。

OpenTelemetry的推进工作非常困难，但其带来的价值是不言而喻的。今天，我们依旧以Go语言为例，试试窥一斑而见全豹，对这个技术有个基本掌握。

## Go语言示例

### 现状

参考官方在Go Package上的声明，Traces处于稳定状态，Metrics处于Alpha测试版本，而Logs则处于冻结状态。

> 可见日志的优先级放在了Traces和Metrics之后。从最终实现来说，只要确定了Traces和Metrics的具体标准，Logs的实现并没有那么复杂。

### 1. 创建Exporter

OpenTelemetry要求程序中收集到的数据，都通过一定的途径发送给外部，如控制台、外部存储、端口等，所以就有了Exporter这个概念。

这里以一个简单的控制台Exporter为例：

```go
traceExporter, err := stdouttrace.New(
  stdouttrace.WithPrettyPrint(),
)
if err != nil {
  log.Fatalf("failed to initialize stdouttrace export pipeline: %v", err)
}
```

### 2. 创建Trace Provider

Traces这部分的概念比较多，这里先只讲解一个 - span。在分布式系统中，存在上下游的概念、也就是调用和被调用的关系，在分布式追踪系统中就将它们区分为不同的span。

示例代码初始化了Traces Provider，用于Traces相关的功能：

```go
ctx := context.Background()
bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp))

// Handle this error in a sensible manner 
defer func() { _ = tp.Shutdown(ctx) }()
```

> 如果要深入了解分布式追踪技术，建议搜索Dapper论文或网上的相关资料。

### 3. 创建Meter Provider

类似Traces，Metrics也需要一个Provider，但它的名字叫做Meter Provider。

我们看一下代码：

```go
metricExporter, err := stdoutmetric.New(
  stdoutmetric.WithPrettyPrint(),
)

pusher := controller.New(
  processor.New(
    simple.NewWithExactDistribution(),
    metricExporter,
  ),
  controller.WithExporter(metricExporter),
  controller.WithCollectPeriod(5*time.Second),
)

err = pusher.Start(ctx)

// Handle this error in a sensible manner where possible
defer func() { _ = pusher.Stop(ctx) }()
```

抛开初始化部分，其中还包含了2个关键性的内容：

1. 程序指标的计算部分
2. metrics的发送方式采用了push，周期为5s

### 4. 设置全局选项

这部分的内容不多，也很容易理解，但在实际工程中的价值很大：**让调用者更方便！**

```go
otel.SetTracerProvider(tp)
global.SetMeterProvider(pusher.MeterProvider())
propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{},propagation.TraceContext{})
otel.SetTextMapPropagator(propagator)
```

这里面做的事情很简单，就是将我们程序中自己创建的trace provider和meter provider设置到官方包中，也就是替换了官方包中的全局变量。接下来，我们想使用provider时，就统一调用官方包即可，**不再需要引用本地的变量**。

>  并不是所有的场景都适合把变量存放到统一的package下，可以延伸思考下~
>
> 举个例子，github.com/spf13/viper配置库只支持全局单个对象Viper，而我们程序中要创建多个对象，这时就不适用。

### 5. 创建Metric Instruments

```go
// 设置关键属性
lemonsKey := attribute.Key("ex.com/lemons")
anotherKey := attribute.Key("ex.com/another")

commonAttributes := []attribute.KeyValue{lemonsKey.Int(10), attribute.String("A", "1"), attribute.String("B", "2"), attribute.String("C", "3")}

// 创建一个Meter实例
meter := global.Meter("ex.com/basic")

// 异步的Observer：通过函数回调
observerCallback := func(_ context.Context, result metric.Float64ObserverResult) {
  result.Observe(1, commonAttributes...)
}
_ = metric.Must(meter).NewFloat64ValueObserver("ex.com.one", observerCallback,metric.WithDescription("A ValueObserver set to 1.0"))

// 同步的Recorder：创建一个变量，按需使用
valueRecorder := metric.Must(meter).NewFloat64ValueRecorder("ex.com.two")
boundRecorder := valueRecorder.Bind(commonAttributes...)
defer boundRecorder.Unbind()
```

### 6. 综合示例

```go
// 创建一个Tracer
tracer := otel.Tracer("ex.com/basic")

// 创建了一个包含2个member的baggage，并结合到Go里的context
foo, _ := baggage.NewMember("ex.com.foo", "foo1")
bar, _ := baggage.NewMember("ex.com.bar", "bar1")
bag, _ := baggage.New(foo, bar)
ctx = baggage.ContextWithBaggage(ctx, bag)

// 以下为一个具体调用的示例，多层嵌套
func(ctx context.Context) {
  // 根据传入的ctx，创建一个span
  var span trace.Span
  ctx, span = tracer.Start(ctx, "operation")
  defer span.End()

  span.AddEvent("Nice operation!", trace.WithAttributes(attribute.Int("bogons", 100)))
  span.SetAttributes(anotherKey.String("yes"))

  meter.RecordBatch(
    ctx,
    commonAttributes,
    valueRecorder.Measurement(2.0),
  )

  func(ctx context.Context) {
    // 根据传入的ctx，创建一个子span
    var span trace.Span
    ctx, span = tracer.Start(ctx, "Sub operation...")
    defer span.End()

    span.SetAttributes(lemonsKey.String("five"))
    span.AddEvent("Sub span event")
    boundRecorder.Record(ctx, 1.3)
  }(ctx)
}(ctx)
```

### 链接

- 文档 - https://opentelemetry.io/docs/go/
- Go Package - https://pkg.go.dev/go.opentelemetry.io/otel#section-readme
- Github - https://github.com/open-telemetry/opentelemetry-go

## 总结

从现状来看，OpenTelemetry仍处于初期阶段，使用起来并不那么方便。我们应该把重点放在标准上：

从官方文档 - https://opentelemetry.io/docs/go/exporting_data/ 中可以看出，OpenTelemetry有标准的OTLP Exporter，但目前这块更多的是一个标准，而不是一个具体实践。

针对当前已落地的技术，重要参考就是Jaeger和Prometheus。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

