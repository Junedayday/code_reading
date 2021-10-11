---
title: Go语言微服务框架 - 3.日志库的选型与引入
date: 2021-08-25 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

日志是一个框架的重要组成部分，那今天我们一起来看看这部分。

衡量日志库有多个指标，我们今天重点关注两点：**简单易用** 与 **高性能**。简单易用是一个日志库能被广泛使用的必要条件，而**高性能**则是企业级的日志库非常重要的衡量点，也能在源码层面对我们有一定的启发。

<!-- more -->

## v0.3.0：日志库的选型与引入

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.3.0

### 目标

选择一个开源的日志组件引入，支持常规的日志打印。

### 关键技术点

3. 三款开源日志库的横向对比
2. zap日志库的关键实现
3. 关于日志参数的解析

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
	   |-- config                          配置相关的文件夹
	      |-- viper.go                         viper的相关加载逻辑
     |-- zlog                            新增：封装日志的文件夹
        |-- zap.go                           新增：zap封装的代码实现
	|-- buf.gen.yaml                   buf生成代码的定义
	|-- buf.yaml                       buf工具安装所需的工具
	|-- gen.sh                         buf生成的shell脚本
	|-- go.mod                         Go Module文件
	|-- main.go                        项目启动的main函数
```

## 1.三款开源日志库的横向对比

- glog: https://github.com/golang/glog
- logrus: https://github.com/sirupsen/logrus
- zap: https://github.com/uber-go/zap

如果用一次词语分别进行概括三者的特性，我分别会用：**glog - 代码极简，logrus - 功能全面， zap - 高性能**。经过反复思考，这个框架会选择zap库作为日志引擎的基本组件，主要考量如下：

1. **高性能** - 性能是一个日志库很重要的属性，它往往由前期的设计决定，很难通过后面的优化大幅度提高，所以zap的高性能很难被替代；
2. **方便封装** - zap设计简单，容易进行二次封装（glog更简洁，相应地就需要更多的封装代码）
3. **大厂背书** - zap库被很多大公司引用，作为内部的日志库的底层，再二次开发
4. **源码学习** - zap库对性能追求极高，可以作为高性能Go语言代码的分析样例

## 2.zap日志库的关键实现

### 最简化的调用

zap日志库的调用很简单，只需要两行代码就能实现初始化：

```go
logger, _ := zap.NewProduction()
defer logger.Sync()
```

但这样的zap代码存在两个明显弊端：

- 默认输出到控制台上
- 无法保存到指定目录的文件

### 核心的日志文件实现

我们增加了一定的特性，代码如下：

```go
var (
	// Logger为zap提供的原始日志，但使用起来比较烦，有强类型约束
	logger *zap.Logger
	// SugaredLogger为zap提供的一个通用性更好的日志组件，作为本项目的核心日志组件
	Sugar *zap.SugaredLogger
)

func Init(logPath string) {
	// 日志暂时只开放一个配置 - 配置文件路径，有需要可以后续开放
	hook := lumberjack.Logger{
		Filename: logPath,
	}
	w := zapcore.AddSync(&hook)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		zap.InfoLevel,
	)

	logger = zap.New(core, zap.AddCaller())
	Sugar = logger.Sugar()
	return
}

// 命名和原生的Zap Log尽量一致，方便理解
func Sync() {
	logger.Sync()
}
```

那我们是如何解决上面两个问题的呢？

1. 利用`go.uber.org/zap/zapcore`中的开放配置
2. 借用了`github.com/natefinch/lumberjack`这个常用的日志切分库，也顺带实现了日志路径的支持

### main函数的调用

```go
var logFilePath = flag.String("l", "log/service.log", "log file path")
flag.Parse()

zlog.Init(*logFilePath)
defer zlog.Sync()
```

至此，我们的日志功能已经基本打通。

## 3.关于日志参数的解析

日志参数常见的方式分2种，一个是来自`flag`的解析，另一个是来自配置文件。

随着我们功能的拓展，日志库肯定会支持越来越复杂的场景。那这个时候用`flag`解析的扩展性就会很差，所以，我更推荐在微服务的框架中，**用配置文件的方式去加载日志的相关配置**。但这种方式会带来一个常见的现象：

程序代码的实现为：先加载配置文件，后加载日志，导致配置文件出错时，无法通过日志来排查，需要用控制台或者进程管理工具协助定位问题。

后续，随着框架的迭代，我会开放出更多的日志参数，目前只放出了一个日志路径的参数作为示例。

## 后续的两点核心需求

至此，我们添加的代码量并不多，也算成功地实现了一个日志打印的功能。但在实际的工程中，日志模块还需要实现两个比较大的功能：

1. 支持Go程序Panic/Error Wrapping风格的**多行打印与采集**
2. 支持分布式TraceId的打印，用来排查**微服务调用链路**

这两块的内容会结合具体的相关相关技术，会在后续专题中专门分享，请大家重点关注。

## 总结

`zap`库的代码是一个很棒的实现，我会在接下来的**Go语言技巧系列**中详细分析，欢迎大家进行关注。

至此，我们的框架逐渐成型，接下来我将对`GORM`做一个简单的讲解，引入到框架中。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

