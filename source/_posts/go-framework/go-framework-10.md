---
title: Go语言微服务框架 - 10.接口文档-openapiv2的在线文档方案
date: 2021-11-01 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

随着项目的迭代，一个服务会开放出越来越多的接口供第三方调用。

虽然`protobuf`已经是通用性很广的IDL文件了，但对于未接触过这块的程序员来说，还是有很大的学习成本。在综合可读性和维护性之后，我个人比较倾向于使用oepnapiv2的方案，提供在线接口文档。

接下来，我们一起来看看这部分的实现。

<!-- more -->

## v0.7.0：接口文档-openapiv2的在线文档方案

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.7.0

### 目标

项目提供在线接口文档，供第三方快速地了解接口细节。

### 关键技术点

1. 了解buf的openapiv2的插件
2. 用swagger工具合并文档
3. 利用swagger相关容器提供在线文档

> swagger 是 openapiv2 的一种具体实现，在下文可等同于一个概念。
>
> 可以参考swagger官网了解详情：https://swagger.io/specification/v2/

### 目录构造

```
--- micro_web_service            项目目录
	|-- gen                            从idl文件夹中生成的文件，不可手动修改
	   |-- idl                             对应idl文件夹
	      |-- demo                             对应idl/demo服务，包括基础结构、HTTP接口、gRPC接口
	    	|-- order                            对应idl/order服务，同上
     |-- swagger.json                    新增：openapiv2的接口文档
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
	|-- swagger.sh                     新增：生成openapiv2的相关脚本
```

## 1.了解buf的openapiv2的插件

从[gRPC-Gateway的文档](https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_openapi_output/#other-plugin-options)中，我们可以找到对应的buf插件使用方式：在`buf.gen.yaml`文件中，我们添加如下插件内容：

```yaml
version: v1
plugins:
  - name: openapiv2
    out: gen/openapiv2
```

运行`buf generate`后，在`gen/openapiv2`目录下会根据我们在`idl`文件中的目录结构，生成多个接口文档。

## 2.用swagger工具合并文档

用buf标准的openapiv2插件会生成多份swagger文档，管理多个文件对使用方来说并不方便。最佳的使用体验，就是能将多个文档合并起来，用一个API文档统一交付。

这里，我们借助goswagger工具，合并文档。工具具体的安装方式可参考链接：https://goswagger.io/install.html。

安装后，运行如下命令，生成到文件 gen/swagger.json：

```shell
# 合并swagger文档
swagger mixin gen/openapiv2/idl/*/*.json -o gen/swagger.json
# 删除原始文档
rm -rf gen/openapiv2
```

## 3.利用swagger相关容器提供在线文档

在统一了swagger文件后，在线接口文档的实现方案有很多，例如swagger官网就可以提供简单的渲染。

这里，我用了个人比较常用的docker镜像redoc为例，搭建一个在线接口文档平台。

> 该镜像更多的使用方式可参考：https://hub.docker.com/r/redocly/redoc/ 

运行如下命令，即将swagger.json加载到镜像中：

```shell
docker run --name swagger -it --rm -d -p 80:80 -v $(pwd)/gen/swagger.json:/usr/share/nginx/html/swagger.json -e SPEC_URL=swagger.json redocly/redoc
```

我们在本地打开浏览器，输入 http://127.0.0.1:80/ 就能看到文档。

> 扩展点 - 公共的文档服务器：
>
> 我们往往更希望把文档放在一个公共的服务器上，可以简单地利用这两个关键技术实现：
>
> 1. https://hub.docker.com/r/redocly/redoc/ 中的watch方案，即watch某个目录下的文件，根据文件变化实时更新接口
> 2. 利用scp命令，将本地swagger.json上传到远端服务器
>
> 更复杂点的方案，可以考虑结合git流程来实现。

## 总结

至此，我们实现了一个关键性的功能：**代码即接口文档**，保证了接口文档随着代码更新的实时性。

同时，希望大家能够认识到接口文档的价值，最好能做到**接口文档即代码**，也就是将相关程序的逻辑尽可能地通过接口文档表达清楚。哪怕前期接口文档问题很多，只要我们不断迭代，后续总能趋于稳定，降低维护接口的成本。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

