---
title: Go语言微服务框架 - 13.监控组件Prometheus的引入
date: 2021-12-12 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

作为云原生程序监控的标准组件，Prometheus支持了各类Paas、Saas平台，并提供了一整套采集+存储+展示的解决方案。

今天我们专注于自定义服务中的Prometheus的监控，在框架中引入Prometheus相关的组件。关于更细致的使用方式，我会给出相关的链接，有兴趣进一步学习Prometheus的同学可以边参考资料边实践。

<!-- more -->

## v0.8.0：监控组件Prometheus的引入

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.8.0 

### 目标

引入prometheus组件，提供标准与自定义的metrics。

### 关键技术点

1. metrics接口的开放
2. 示例counter的初始化
3. 示例counter的计数
4. 学习Prometheus监控使用方法

### 目录构造

```
--- micro_web_service            项目目录
	|-- gen                            从idl文件夹中生成的文件，不可手动修改
	   |-- idl                             对应idl文件夹
	      |-- demo                             对应idl/demo服务，包括基础结构、HTTP接口、gRPC接口
	    	|-- order                            对应idl/order服务，同上
     |-- swagger.json                    openapiv2的接口文档
	|-- idl                            原始的idl定义
	   |-- demo                            业务package定义，protobuffer的原始定义
	   |-- order                           业务order定义，同时干
	|-- internal                       项目的内部代码，不对外暴露
	   |-- config                          配置相关的文件夹，viper的相关加载逻辑
	   |-- dao                             Data Access Object层，是model层的实现
	   |-- gormer                          从pkg/gormer中生成的相关代码，不允许更改
	   |-- metrics                         新增：自定义监控指标
	   |-- model                           model层基本定义由gormer自动生成
	   |-- mysql                           MySQL连接，支持日志打印
	   |-- server                          服务器的实现，对idl中定义服务的具体实现
	   |-- service                         service层，作为领域实现的核心部分
     |-- zlog                            封装zap日志的代码实现
  |-- pkg                            开放给第三方的工具库
     |-- gormer                          gormer二进制工具，用于生成Gorm相关Dao层代码
	|-- buf.gen.yaml                   buf生成代码的定义，新增参数校验逻辑
	|-- buf.yaml                       buf工具安装所需的工具，从v1beta升到v1
	|-- format.sh                      新增：格式化代码的脚本
	|-- gen.sh                         生成代码的脚本：buf+gormer
	|-- go.mod                         Go Module文件
	|-- gormer.yaml                    将gormer中的参数移动到这里
	|-- main.go                        项目启动的main函数
	|-- swagger.sh                     生成openapiv2的相关脚本
```

## 1.metrics接口的开放

Prometheus官方推荐的metrics开放方式为http。将它引入到程序中的代码如下面几行，不过有几个点值得注意：

```go
go func() {
  mux := http.NewServeMux()
  mux.Handle("/metrics", promhttp.Handler())
  http.ListenAndServe(fmt.Sprintf(":%d", config.Viper.GetInt("server.prometheus.port")), mux)
}()
```

1. `http.ListenAndServe` 函数是阻塞的，所以需要开一个goroutine。
2. 为了保证Prometheus的指标监控不与应用的http服务冲突，这里采用了端口隔离，也就是另起一个http服务。
3. `Go` 的 `http` 库如果要支持多port的运行，需要引入`mux`的概念；默认会注册到http库中的`DefaultServeMux`。

为了验证我们的metrics已经正常running，我们可以调用一个curl请求查看一下（具体返回结果不细讲）。

```shell
# 示例的metrics起在8083端口
curl --request GET 'http://127.0.0.1:8083/metrics'
```

## 2.示例counter的初始化

我们先以一个最简单的counter累加器为例，实现一个自定的指标监控。

```go
package metrics

import "github.com/prometheus/client_golang/prometheus"

func init() {
	prometheus.MustRegister(OrderList)
}

var OrderList = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "order_list_counter",
		Help: "List Order Count",
	},
	[]string{"service"},
)
```

代码的逻辑比较简单，我们注意以下三个关键点：

1. `OrderLis`t 是一个全局变量，方便使用方调用；
2. `NewCounterVec` 表示这个Counter是一个向量，包括了两块 - opts和labels
   1. opts包括Name和Help，Name是metrics唯一的名称，Help是metrics的帮助信息
   2. labels是用来过滤、聚合功能的关键参数，提前声明有利于存储端进行优化（可类比数据库索引）
3. `prometheus.MustRegister(OrderList)` 是将metrics注册到prometheus的全局变量里，与main函数里的注册对应

## 3.示例counter的计数

从指标的定义可以看到，我们设计的这个metrics是为了统计订单查询接口的次数，于是我们在代码侧引入：

```go
func (s *Server) ListOrders(ctx context.Context, req *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	metrics.OrderList.With(map[string]string{"service": "example"}).Inc()
  // ...
}
```

函数是一个链式调用，包括两块：

1. With，也就是label信息，用一个map[string]string填入，是个通用功能；
2. Inc，即计数+1，这个方法和具体的metrics类型相关。

接着，我们调用两次对应的接口，可以从metrics信息中看到下面的内容：

```
# HELP order_list_counter List Order Count
# TYPE order_list_counter counter
order_list_counter{service="example"} 2
```

除非程序重启，否则这个Counter会不断累加。

## 4.学习Prometheus监控使用方法

Prometheus监控埋点的使用方式比较直观，上手难度不大。如果你希望进一步了解这块，我推荐两个核心的资料：

- Prometheus官网 - https://prometheus.io/docs/introduction/overview/
- Prometheus的Go语言官方库 - https://github.com/prometheus/client_golang

这两份资料是英文的，可能对部分同学来说学成本比较高，可以考虑先去搜索一些中文翻译文档、了解梗概后，再回过头来看这两篇。如果你希望深入了解Prometheus，必须要仔细看这两块内容，保证实时性。

## 总结

对接Prometheus的自定义metrics是一个应用程序很常见的功能，例如业务指标埋点。在埋点的过程中，有一个大误区需要刚接触Prometheus的同学注意：把计算的工作交给Prometheus引擎，而不要放在你开发的程序里。

例如，你希望计算某个订单的成功率，你不应该用一个metrics对应成功率，而应该给出两个指标，即订单总量和成功的订单量（也可以放在一个指标中，用label区分成功与否），交由Prometheus进行计算，方便后续的各种metrics的扩展。

更多Prometheus的实践，需要大家边学习边实践。如果反响热烈，我也会抽几讲谈谈Prometheus。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

