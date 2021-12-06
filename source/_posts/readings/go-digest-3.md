---
title: 技术阅读摘要 - 3.Jaeger技术分析
date: 2021-10-20 12:00:00
categories: 
- 成长分享
tags:
- Digest
---

![Go-Framework](https://i.loli.net/2021/08/15/QZ3lGpkvgdfXW7R.jpg)

## 概览

通过上一次技术阅读摘要，我们了解了分布式链路追踪这项技术，Jaeger是其主流的实现方案。

今天，我们就一起来看看Jaeger的相关资料，初步掌握这门技术。

<!-- more -->

## 官方资料

- Jaeger官网 https://www.jaegertracing.io/
- Github https://github.com/jaegertracing/jaeger
- Dapper https://research.google/pubs/pub36356/
- OpenZipkin https://zipkin.io/

## 核心概念

Jaeger的官方定义非常简洁 - **Jaeger: open source, end-to-end distributed tracing**。

关键词是**端到端的分布式追踪**。怎么理解这个端到端呢？它更多地是关注分布式系统中的**入和出**。从一个HTTP服务来看，它关注的是请求和响应的具体数据。对应到如今k8s中盛行的sidecar模式，就是一个网络的sidecar，将所有的请求进行标注（如带上traceId）。

## 架构

Jaeger的官方文档上资料很丰富，更新也比较频繁。有些朋友会觉得阅读官方文档非常累，常常通篇阅读后发现抓不到重点、也没有什么印象。这里，我推荐一个页面 - https://www.jaegertracing.io/docs/1.27/architecture/ ，并结合我的理解，方便大家快速理解。

### Span

Span是分布式链路追踪中一个通用的术语，字面翻译为 **带名称的Jaeger逻辑单元**。

这里对逻辑单元的定义比较有争论，在我看来，逻辑单元的定义因具体场景而变化：

- 在单体架构中，需要拆分成多个模块，每个模块定义成一个逻辑单元
- 在一个简单的微服务中，可以将服务定义成一个逻辑单元
- 在一个复杂的微服务中，可能需要根据更细的领域定义成一个逻辑单元

### Trace

一个具体消息在整个分布式系统中的流转。

### 组件

Jaeger提供了2种架构的解决方案。我们先看看通用的部分：

- jaeger-client作为具体语言的内部库，嵌入到应用程序中
- jaeger-agent作为sidecar，部署在容器或机器上，用来从jaeger收集数据，并推送到jaeger collector
- jaeger collector负责将数据保存到数据库或MQ中
- jaeger-query + UI 查询并显示数据

而差异点就在于保存和分析数据的技术方案：

- 简单方案：直接保存到数据库中，用Spark Jobs进行分析
- 高性能方案：用Kafka来削峰填谷，用Flink流式计算提高性能

## Jaeger Go

### Trace的初始化

```go
traceCfg := &jaegerconfig.Configuration{
  // 服务名
  ServiceName: "MyService",
  // 采样参数
  Sampler: &jaegerconfig.SamplerConfig{
    Type:  jaeger.SamplerTypeConst,
    Param: 1,
  },
  // 上报，这里通过jaeger sidecar的端口来上报日志
  Reporter: &jaegerconfig.ReporterConfig{
    LocalAgentHostPort: "127.0.0.1:6831",
    LogSpans:           true,
  },
}
// 初始化tracer
tracer, closer, err := traceCfg.NewTracer(jaegerconfig.Logger(jaeger.StdLogger))
if err != nil {
  panic(err)
}
defer closer.Close()
// 将tracer设置到opentracing的全局变量中
opentracing.SetGlobalTracer(tracer)
```

上面这段逻辑描述了 **创建jaeger的tracer并保存到opentracing的全局变量中**。

这里强调一点：opentracing是一套标准，包括jaeger、zipkin等具体实现。我们可以深入看看`NewTracer`这个函数。它的注释很好地说明了这一点。具体的细节实现，我们暂时无需关注。

```go
// Tracer implements opentracing.Tracer.
type Tracer struct {
}
```

### 技术组件引入Opentracing

通过上面的工作，我们已经在程序中引入了jaeger。但在实际的开发过程中，我们程序内部会有一些组件也需要引入jaeger的链路追踪，来实现更精细化的监控。

以gRPC-Gateway为例，引入Opentracing的链接如下： https://grpc-ecosystem.github.io/grpc-gateway/docs/operations/tracing/#opentracing-support 。这里面的代码可以直接引用，就不细看了。

目前，支持原生的Opentracing的组件越来越多。在引入一个复杂的组件时，我们要先了解清楚是否可以集成Opentracing，降低后续的运维复杂度。

### 提取TraceId信息

整个jaeger的引入并不复杂，就已经能很好地实现链路监控了。但在实际的开发过程中，我们仍有一个非常关键的需求：**如何将一个请求的trace信息，引入到业务代码中，跟踪业务代码的处理过程**。这一点，在debug问题时非常有意义，尤其是面对一些自己不太熟悉的代码。

开发人员面对这个场景，最常用的逻辑就是log，那就意味着我们要将traceid注入到日志中。那么怎么获取traceid呢？下面看一段示例代码：

```go
// 从ctx获取span
if parent := opentracing.SpanFromContext(ctx); parent != nil {
  parentCtx := parent.Context()
  // 获取opentracing中的全局tracer
  if tracer := opentracing.GlobalTracer(); tracer != nil {
    mySpan := tracer.StartSpan("my info", opentracing.ChildOf(parentCtx))
    // 由于前面opentracing中的tracer是jaeger的，所以你这里转化为jaeger.SpanContext
    if sc, ok := mySpan.Context().(jaeger.SpanContext); ok {
      // 这里，就能获取traceid等信息了，可以放在日志里
      _ = sc.TraceID()
    }
    defer mySpan.Finish()
  }
}
```

逻辑就是从go语言的上下文context信息中，用Opentracing里定义的全局tracer，提取出traceId等信息。

## 总结

关于Jaeger内容有很多延伸点，但本文只作入门，点到即可。

如果只能记住一点，我希望大家能认识到：**Jaeger是Opentracing标准的一个实现**。从本文能看到，在标准统一后，具体实现的变更会变得非常简单：例如要将Jaeger替换成Zipkin，只需要初始化tracer处做到替换即可。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

