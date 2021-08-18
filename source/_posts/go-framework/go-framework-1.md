---
title: Go语言微服务框架 - 1.搭建gRPC+HTTP的双重网关服务
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

大家好，我是六月天天。如题所述，从今天开始，我将和大家一起逐步完成一个微服务框架。

整个迭代过程会围绕着两个核心思想进行：

1. **关注技术选型背后的思想**。虽然最终某个技术选型的并不是你喜欢的方案（如RPC、日志、数据库等，你可以fork后自行调整），但我们更关注各个技术组件背后的原理与思想；
2. **聚焦于简单，关注可维护性**。技术框架是项目的基础设施，也是排查复杂业务问题的根本，所以框架层的功能会尽量考虑简单易用，可以让我们花更多的心思在业务开发中。许多开源库提供了大量扩展功能，但我们使用时会尽量**克制**，减少学习和排查问题时的成本。

<!-- more -->



## v0.1.0：搭建gRPC+HTTP的双重网关服务

[项目链接](https://github.com/Junedayday/micro_web_service/tree/v0.1.0)

[gRPC-gateway官方Github](https://github.com/grpc-ecosystem/grpc-gateway)

### 目标

完成RPC服务的框架的搭建

### 关键技术点

1. `protobuffer`定义IDL（Interface Define Language 接口定义语言）
2. `buf`工具生成`Go`代码（包括数据结构和RPC相关服务）
3. `Go`项目实现RPC服务（实现RPC接口）

### 目录构造

```
--- micro_web_service            项目目录
	|-- gen                            从idl文件夹中生成的文件，不可手动修改
	   |-- idl                             对应idl文件夹
	      |-- demo                             对应idl/demo服务
	         |-- demo.pb.go                        demo.proto的基础结构
	         |-- demo.pb.gw.go                     demo.proto的HTTP接口，对应gRPC-Gateway
	         |-- demo_grpc.pb.go                   demo.proto的gRPC接口代码
	|-- idl                            原始的idl定义
	   |-- demo                            业务package定义
	      |-- demo.proto                       protobuffer的原始定义
	|-- internal                       项目的内部代码，不对外暴露
	   |-- server                          服务器的实现
	      |-- demo.go                          server中对demo这个服务的接口实现
	      |-- server.go                        server的定义，须实现对应服务的方法
	|-- buf.gen.yaml                   buf生成代码的定义
	|-- buf.yaml                       buf工具安装所需的工具
	|-- gen.sh                         buf生成的shell脚本
	|-- go.mod                         Go Module文件
	|-- main.go                        项目启动的main函数
```

## 1. protobuffer定义IDL

我们先看一下项目中的`demo.proto`文件，重点关注 **rpc Echo(DemoRequest) returns (DemoResponse)** 这个定义。

```proto
message DemoRequest {
   string value = 1;
}

message DemoResponse {
   int32 code = 1;
}

// 样例服务
service DemoService {
  // Echo 样例接口
  rpc Echo(DemoRequest) returns (DemoResponse) {
    option (google.api.http) = {
      post : "/apis/demo"
      body : "*"
    };
  }
}
```

今天我们暂时不对`protobuffer`的语法做扩展讲解，只需要简单地了解下它的请求结构体`DemoRequest`和响应结构体`DemoResponse`。

## 2. buf工具生成Go代码

我们通过运行项目根目录中的`gen.sh`，会在`gen`目录下生成对应的Go语言代码。

这部分是自动化的工作，每次修改`proto`文件后需要运行。

> buf工具的安装请参考README.md，它是protoc的演进版本，不再需要大量flag参数，更加简单易用。
>
> 注意，如果修改了模块名，buf工具第一次初始化建议使用 buf beta mod init 指令



## 3.Go项目实现RPC服务

我们梳理一下整个逻辑，来看看这个`Go`程序是怎么提供RPC服务的。

1. 在`buf.gen.yaml`中定义了生成的2种服务， `go-grpc`和 `grpc-gateway`，分别表示`gRPC`和`HTTP`
2. `demo.proto`通过脚本，在`gen/idl/demo`生成了2个文件，`*_grpc.pb.go`和`*.pb.gw.go`，分别表示`gRPC`和`HTTP`
3. 在`main`函数中注册两个服务，分别为：
   1. gRPC - `demo.RegisterDemoServiceServer(s, &server.Server{})`
   2. HTTP - `demo.RegisterDemoServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)`
4. 在`internal/server/server.go`中，`server.Server`需要实现`proto`中定义的方法，所以我们加入接口定义`demo.UnsafeDemoServiceServer`
5. 在`internal/server/demo.go`中，实现一个`func (s *Server) Echo(ctx context.Context, req *demo.DemoRequest) (*demo.DemoResponse, error)`方法



## 项目运行

我们用简单的命令来运行，并用RPC访问

```shell
# 编译并运行
go build && ./micro_web_service 

# 模拟HTTP请求
curl --location --request POST 'http://127.0.0.1:8081/apis/demo'
# 收到返回值 {"code":0}

# 而gRPC比较麻烦，是私有协议，我们查看一下对应的网络端口，发现正在监听，也就意味着正常运行
netstat -an | grep 9090
tcp4       0      0  127.0.0.1.9090         127.0.0.1.49266        ESTABLISHED
tcp4       0      0  127.0.0.1.49266        127.0.0.1.9090         ESTABLISHED
tcp46      0      0  *.9090                 *.*                    LISTEN 
```



## 项目的私有化

由于本项目只是一个框架，如果你希望修改为个人的项目，主要改动点在两处：

1. `go.mod`里的模块名，以及`Go`代码内部的import
2. `proto`文件中定义的`go_package`

> 建议用编辑工具全量替换



## 新增接口示例

### 添加proto定义

```protobuf
message EmptyMessage {
}

// Empty 空接口
rpc Empty(EmptyMessage) returns (EmptyMessage) {
  option (google.api.http) = {
    post : "/apis/empty"
    body : "*"
  };
}
```

### 生成Go文件

```shell
bash gen.sh
```

### 添加接口定义

这时候，我们会发现`main.go`中有报错，即提示`server.Server`这个对象需要实现`Empty`方法。于是，我们在`internal/server/demo.go`里添加

```go
func (s *Server) Empty(ctx context.Context, req *demo.EmptyMessage) (*demo.EmptyMessage, error) {
	return &demo.EmptyMessage{}, nil
}
```

### 测试新接口

```shell
# 编译并运行
go build && ./micro_web_service 

# 模拟HTTP请求
curl --location --request POST 'http://127.0.0.1:8081/apis/empty'
# 返回 {} 
```

## 总结

`v0.1.0`版本是一个非常简单的web框架，只有样例的RPC接口。

开放`HTTP`接口是为了兼容传统方案，而`gRPC`则提供了高性能、跨语言的通信方案。从整个实现过程来看，我们只编写了一个具体的实现、也就是`Echo`这个方法，就完成了两种通信方式的兼容。

`gRPC-Gateway`方案还有很多很棒的特性，我会在后续逐一介绍并引入。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

