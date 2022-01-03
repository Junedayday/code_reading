---
title: Go语言技巧 - 9.【浅析微服务框架】Kratos概览
date: 2021-12-20 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## Kratos框架概况

截止到本文发布时，Kratos在github上的star数达到了15.9k。其中，在2021年7月，也正式推出了v2这个大版本。

本人并不是Kratos的重度使用者，主要会通过官方介绍对它的特性进行剖析。接下来的内容依旧包含大量主观认知，可能会对官方文档有理解上的偏差，欢迎大家与我讨论。

<!-- more -->

## 概览

**Kratos is a web application framework with expressive, elegant syntax. We've already laid the foundation.**

Kratos的官网上的介绍比较朴实，但有两个词值得我们关注 - **富有表现力的**、**优雅的**。一般在微服务框架里，我们看到最多的形容词，往往来自下面两个维度：

- 开发者维度：比如简单易用、组件丰富
- 工程化维度：比如高效、通用性强

但Kratos的切入点是框架层面的能力，尤其是`elegant`这个词，隐含了作者对代码洁癖的追求。接下来，我们具体剖析这个框架。

![Kratos](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/kratos.png)

## 项目结构

 参考 https://go-kratos.dev/docs/intro/layout，我们对关键目录做一下分析。

### api

```
├── api // 下面维护了微服务使用的proto文件以及根据它们所生成的go文件
│  └── helloworld
│      └── v1
│          ├── error_reason.pb.go
│          ├── error_reason.proto
│          ├── error_reason.swagger.json
│          ├── greeter.pb.go
│          ├── greeter.proto
│          ├── greeter.swagger.json
│          ├── greeter_grpc.pb.go
│          └── greeter_http.pb.go
```

从目录结构看到，里面包含了三类文件：

- `*.proto` 原始IDL文件
- `*.go` 利用protoc工具生成的go文件，包括http和grpc的服务相关代码
- `*.swagger.json` 利用工具生成的swagger接口文档

这部分的实现全是基于开源的`protobuf`解决方案，对开发者很友好。这里提一个点：**尽可能地用目录区分原始IDL文件与生成的文件**。我简单列举两个优点：

1. 让开发者更聚焦于原始IDL文件 - 其余文件均是从proto文件自动生成出来的，不应过多关注。
2. 有利于IDL文件的传播 - proto文件可以快速生成其余语言的代码，独立文件夹更有利于扩散给外部调用者。

### cmd

```
├── cmd  // 整个项目启动的入口文件
│  └── server
│      ├── main.go
│      ├── wire.go  // 我们使用wire来维护依赖注入
│      └── wire_gen.go
```

cmd简单来说就是main函数入口。

### internal/biz

```
├── internal  // 该服务所有不对外暴露的代码，通常的业务逻辑都在这下面，使用internal避免错误引用
│   ├── biz   // 业务逻辑的组装层，类似 DDD 的 domain 层，data 类似 DDD 的 repo，而 repo 接口在这里定义，使用依赖倒置的原则。
│  │  ├── README.md
│  │  ├── biz.go
│  │  └── greeter.go
```

internal目录是go语言的一个特性，内部代码不会暴露给外部。

biz被理解为业务逻辑的组装层，如果能正确地理解这个概念，就能把握整个框架的分层设计了。我们从两个关键词来理解这个biz目录的设计：

1. 业务逻辑 - 业务逻辑包括但不限于单个对象的增删改查，会处理很多进阶的内容，例如：
   1. 复合对象操作，如操作对象A后，再操作对象B
   2. 特殊逻辑，如创建A对象失败时，等待10s后再创建
   3. 并发策略，如并发访问对象A和对象B
2. 组装层 - 重点在于组装底层基础的代码，如CRUD，而不是在biz层直接去操作数据库等

整体来说，biz这一层应重点考虑**业务逻辑的信息密度**，让业务开发者的重点放在这一层，把基础实现往下沉。

### internal/data

```
├── internal  // 该服务所有不对外暴露的代码，通常的业务逻辑都在这下面，使用internal避免错误引用
│  ├── data  // 业务数据访问，包含 cache、db 等封装，实现了 biz 的 repo 接口。我们可能会把 data 与 dao 混淆在一起，data 偏重业务的含义，它所要做的是将领域对象重新拿出来，我们去掉了 DDD 的 infra层。
│  │  ├── README.md
│  │  ├── data.go
│  │  └── greeter.go
```

data被理解为缓存与数据库的封装，与底层数据存储相关，一般都是跟着数据库的类型适配。

### internal/server

```
├── internal  // 该服务所有不对外暴露的代码，通常的业务逻辑都在这下面，使用internal避免错误引用
│  ├── server  // http和grpc实例的创建和配置
│  │  ├── grpc.go
│  │  ├── http.go
│  │  └── server.go
```

前面IDL文件（protobuf）生成了RPC方法的接口（interface），这里就是RPC方法的具体实现。

### internal/service

```
├── internal  // 该服务所有不对外暴露的代码，通常的业务逻辑都在这下面，使用internal避免错误引用
│  └── service  // 实现了 api 定义的服务层，类似 DDD 的 application 层，处理 DTO 到 biz 领域实体的转换(DTO -> DO)，同时协同各类 biz 交互，但是不应处理复杂逻辑
│      ├── README.md
│      ├── greeter.go
│      └── service.go
```

service被定义成对数据结构的处理层。

## 架构概览

Kratos里包含了大量组件，很多模块都与前面Go-Micro的有共同之处，我就不再赘述了。而且通过上面的目录层面的划分，重点是api -> server -> service -> biz -> data 的调用逻辑。

这里，我们关注两个重要特性：

1. wire - https://go-kratos.dev/docs/guide/wire
2. ent - https://go-kratos.dev/docs/guide/ent

这两块内容比较多，我会单独出两篇文章进行分享。

## 参考资料

Github - https://github.com/go-kratos/kratos 

文档 - https://go-kratos.dev/docs/

## 思考

Kratos是一个典型的 **基于完善的基建** 而成的Go语言开发框架，可以发现它有3个关键点：

1. RPC层复用protobuf的能力
2. 底层依赖Kubernetes的能力
3. 各类工具复用开源库的能力

很多中大型公司的内部框架都是按照这种思路实现的，只是会封装一些公司通用能力，比如通用的RPC能力。

## 小结

整体来说，Kratos的实现与我推崇的理念基本一致，即复用生态+平台的能力。

在一些细节的技术选型上会存在差异，例如Kratos更注重Bilibili公司的历史沉淀，而我会更关注社区的当前主流实现，并抛开包袱、尽可能地实现自动化。

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

