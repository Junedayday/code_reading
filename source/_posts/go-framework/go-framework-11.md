---
title: Go语言微服务框架 - 11.接口的参数校验功能-buf中引入PGV
date: 2021-11-11 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

随着API在线文档的发布，服务的接口将会被开放给各种各样的调用方。

大量开发接口的朋友会经常遇到**接口参数校验**的问题。举个例子，我们希望将某个字段是必填的，如`name`，我们经常会需要做两步：

1. 在程序中加一个**判断逻辑**，当这个字段为空时返回错误给调用方
2. 在接口文档中加上**注释**，告诉调用方这个参数必填

一旦某项工作被拆分为两步，就很容易出现**不一致性**：对应到参数检查，我们会经常遇到文档和具体实现不一致，从而导致双方研发的沟通成本增加。那么，今天我将引入一个方案，实现两者的一致性。

> 为了缩小讨论范围，我们将 **参数校验** 限定为简单规则。
>
> 而复合条件的检查（逻辑组合等），不在本次的讨论范围内，主要考虑到2点：
>
> 1. 要生成跨语言的方案，技术上比较难实现
> 2. 复合条件往往是一种业务逻辑的检查，放在接口层面不合适

<!-- more -->

## v0.7.1：接口的参数校验功能

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.7.1

### 目标

在线接口文档提供参数校验的逻辑，并自动生成相关代码。

### 关键技术点

1. 参数校验的技术选型
2. 在buf中引入PGV
3. 在框架中引入参数检查
4. buf格式检查

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
	   |-- model                           model层，定义对象的接口方法，具体实现在dao层
	   |-- mysql                           MySQL连接
	   |-- server                          服务器的实现，对idl中定义服务的具体实现
	   |-- service                         service层，作为领域实现的核心部分
     |-- zlog                            封装zap日志的代码实现
  |-- pkg                            开放给第三方的工具库
     |-- gormer                          gormer二进制工具，用于生成Gorm相关Dao层代码
	|-- buf.gen.yaml                   修改：buf生成代码的定义，新增参数校验逻辑
	|-- buf.yaml                       buf工具安装所需的工具，从v1beta升到v1
	|-- gen.sh                         生成代码的脚本：buf+gormer
	|-- go.mod                         Go Module文件
	|-- gormer.yaml                    将gormer中的参数移动到这里
	|-- main.go                        项目启动的main函数
	|-- swagger.sh                     生成openapiv2的相关脚本
```

## 1.参数校验的技术选型

从搜索引擎可知，protobuf的主流参数校验采用两者：

1. go-proto-validators https://github.com/mwitkow/go-proto-validators
2. protoc-gen-validate https://github.com/envoyproxy/protoc-gen-validate

这里，我们最终选用的是protoc-gen-validate（PGV），决定性的理由有两个：

1. buf的官方文档更倾向于PGV - https://docs.buf.build/lint/rules/#custom-options
2. PGV由envoy背书，长期来看更具维护性

## 2.在buf中引入PGV

protoc-gen-validate（PGV）作为一款插件，它已经被集成在了buf工具中。这次，我们就从其调用的顺序，来理解一下buf里的重要文件：

### 2.1 核心文件 - buf.yaml

具体引用路径可以在buf库 - https://buf.build/ 搜索找到，然后在文件中里添加一个依赖项：

```yaml
deps:
  - buf.build/envoyproxy/protoc-gen-validate
```

### 2.2 生成的定义文件 - buf.gen.yaml

这个文件定义了我们要生成什么样的代码，具体增加如下：

```yaml
plugins:
  - name: validate
    out: gen
    opt:
      - paths=source_relative
      - lang=go
```

其中，要注意opt选项要增加一个参数`lang=go`，类似的，我们也可以生成其余语言的代码。

### 2.3 proto定义文件

我们以分页参数为例，添加2条规则，即要求页码、每页数量均大于0。

```protobuf
import "validate/validate.proto";

message ListOrdersRequest {
  int32 page_number = 1 [(validate.rules).int32 = {gt: 0}];
  int32 page_size = 2   [(validate.rules).int32 = {gt: 0}];
}
```

### 2.4 生成相关代码

因为我们引入了一个新的模块，所以先需要更新依赖，用来下载新模块：

```shell
buf mod update
buf generate
```

### 2.5 参数校验的代码

在2.3引入validate的数据结构定义，会生成一个`*.pb.validate.go`文件，我们截取两个关键函数：

```go
func (m *ListOrdersRequest) Validate() error {
	return m.validate(false)
}

func (m *ListOrdersRequest) ValidateAll() error {
	return m.validate(true)
}
```

从命名不难看出，`Validate`是检查到有一个不符合规则就立刻返回，`ValidateAll`是校验完所有的参数后、将不符合的规则一起返回。这两种处理方式的差异主要在于：

1. 耗时：全量检查相对会花费更多的时间
2. 返回的信息量：全量检查的error会包含更多信息

从服务端的视角，更推荐全量检查，将所有字段的检查结果返回给调用方，方便对方一次性修正。

## 3.在框架中引入参数检查

### 3.1 grpc拦截器

grpc提供了一套拦截器Interceptor的机制，类似于http router中的middleware。之前，我们已经引入了一个拦截器，用于打印trace相关的日志。那么这次又新增了一个拦截器，该如何处理呢？

参考grpc的代码，我们可以看到下面两个函数：

```go
func UnaryInterceptor(i UnaryServerInterceptor) ServerOption {
}

func ChainUnaryInterceptor(interceptors ...UnaryServerInterceptor) ServerOption {
}
```

其中前者是单个拦截器，而后者是一种链式拦截器的概念。毫无疑问，我们需要扩充成多个拦截器。

### 3.2 实现参数校验的拦截

```go
// ValidateAll 对应 protoc-gen-validate 生成的 *.pb.validate.go 中的代码
type Validator interface {
	ValidateAll() error
}

func ServerValidationUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if r, ok := req.(Validator); ok {
		if err := r.ValidateAll(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return handler(ctx, req)
}
```

然后在拦截器中引入我们定义的插件：

```go
s := grpc.NewServer(
  grpc.ChainUnaryInterceptor(
    grpc_opentracing.UnaryServerInterceptor(
      grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
    ),
    ServerValidationUnaryInterceptor,
  ),
)
```

### 3.3 具体调用示例

我们尝试着传一个错误的接口参数，看看返回结果：

```json
{
    "code": 3,
    "message": "invalid ListOrdersRequest.PageNumber: value must be greater than 0; invalid ListOrdersRequest.PageSize: value must be greater than 0",
    "details": []
}
```

可以看到，结果中清晰地说明了不合规的两个参数，以及具体的规则，对调用方来说非常直观。

## 4.buf格式检查

随着buf工具的推进，我们引入了越来越多的内容，protobuf文件也新增了很多东西。这时，我们会希望能将protobuf的格式也能有一定的规范化。在buf之前，已经有prototool等工具，buf对此做了集成。

由于buf的lint检查有很多细节，建议酌情选用。以项目中我选择的为例：

```yaml
lint:
  use:
    - DEFAULT
  except:
    - PACKAGE_VERSION_SUFFIX
    - PACKAGE_DIRECTORY_MATCH
  rpc_allow_google_protobuf_empty_requests: true
  rpc_allow_google_protobuf_empty_responses: true
```

包括两块：

- except排除了两个检查项，即要求protobuf的package带上版本后缀、与代码路径匹配
- 允许request和response设置为empty格式

接下来，运行`buf lint`，会提示你需要修正的地方，逐一修改即可（很多是命名上的规范，增加可读性，推荐按插件的建议进行修改）。

## 总结

本次框架的小迭代高度依赖了buf的生态体系，建议有时间的朋友可以再看看buf的文档链接 - https://docs.buf.build/introduction。buf工具的迭代频率比较高，对其新特性仍处于观望状态，目前没有完全按照其Best Practice推进。

回过头来，我们的参数检查方案依然存在一个明显问题：生成的swagger文档中没有对应的参数要求（Issue - https://github.com/grpc-ecosystem/grpc-gateway/issues/1093）。如果这个问题长期无法解决，我也会给出一套自己的解决方案。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

