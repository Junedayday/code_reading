---
title: Go语言微服务框架 - 2.实现加载静态配置文件
date: 2021-08-21 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

我们的基础RPC服务已经正常运行，我们再来看下一个特性：配置文件的加载。

首先，我们要正确地认识到配置文件的重要性：**在程序交付后，变更代码的成本很大；相对而言，变更配置文件的成本就比较小**。但有的同学又走了另一个极端，也就是将大量的逻辑放入到配置文件中，导致**配置文件膨胀**，本质上就是将部分本应在代码中维护的内容转移到了配置文件，导致配置文件也很难维护。

今天，我们先将重点放到加载配置文件库的技术选型，顺便分享一些常见的问题。

<!-- more -->

## 一个基础的加载配置文件示例

在`Go`语言中，用官方库就能快速实现配置文件的加载，下面就是一个简单的代码实现：

```go
b, err := ioutil.ReadFile("config.json")
if err != nil {
  panic(err)
}
var config MyConfig
err = json.Unmarshal(b, &config)
if err != nil {
  panic(err)
}
```

关键的实现分为两块：

1. 读取文件中的数据
2. 将数据解析到Go程序的对象中，作为可识别的数据结构，这里指定了数据类型为`json`

## v0.2.0：实现加载静态配置文件

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.2.0

### 目标

从配置文件中解析数据到程序中，并具备更高的可读性和扩展性。

### 关键技术点

1. 命令行参数与配置文件的差异
2. github.com/spf13/viper的介绍
3. 使用viper库的推荐方式

### 目录构造

`github.com/spf13/viper`中存在一个全局变量`var v *Viper`（[点击查看](https://github.com/spf13/viper/blob/v1.7.0/viper.go#L62)），如果我们调用默认的viper包，其实就是将参数解析到这个全局变量中。

在具体的项目中，更推荐的方式是将这个变量保存到内部项目中，作为一个项目中的全局变量，所以我们会新建一个`viper.New()`。配置参数会被全局调用，为了保证不会发生**循环依赖**，我们就需要一个专门的`package`来保存这个全局变量，这里对应项目中的`internal/config/viper.go`。

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
	   |-- config                          新增：配置相关的文件夹
	      |-- viper.go                         新增：viper的相关加载逻辑
	|-- buf.gen.yaml                   buf生成代码的定义
	|-- buf.yaml                       buf工具安装所需的工具
	|-- gen.sh                         buf生成的shell脚本
	|-- go.mod                         Go Module文件
	|-- main.go                        项目启动的main函数
```

## 

## 1.命令行参数与配置文件的差异

命令行参数类似于`./demo --config=a.yaml --http_port=8080 --grpc_port=8081`，`Go`语言中自带`flag`包可供解析，开源的有`pflag`和相关的扩展工具，如`cobra`。

而配置文件则是**将参数从文件解析到内存中**，一般用读取文件+反序列化工具来使用。

同样是解析参数到程序里，我们该选择哪种方案呢？我们从可读性和可维护性来对比下：

- 可读性：命令行参数是扁平化的，可读性远不如格式化后的配置文件
- 可维护性：配置文件增加了一个维护项，但成本不高

所以，我个人倾向于的方案是：

- 命令行参数：用于维护极少量关键性的参数，如配置文件的路径
- 配置文件：维护绝大多数的参数

在某些极端的场景中，比如提供一个纯二进制文件作为工具，那不得不把所有配置参数都放入到命令行参数中。这并不是我们微服务框架要讨论的场景。所以，接下来我们重点讨论配置文件的加载。

> 关于pflag相关的内容，在后续程序复杂到一定阶段后会引入。

## 2.github.com/spf13/viper的介绍

对比上面的方案，我们来看一个业内使用最广的Go语言配置加载库`github.com/spf13/viper`的实现，对应的代码如下：

```go
viper.SetConfigName("config")        // config file name without file type
viper.SetConfigType("yaml")          // config file type
viper.AddConfigPath("./")            // config file path
if err := viper.ReadInConfig(); err != nil {
  panic(err)
}
```

而在获取配置文件时，又采用key-value形式的语法：

```go
viper.GetInt("server.grpc.port")
```

详细的特性我会在**Go语言技巧系列**里说明，今天我们聚焦于工程侧的宏观特性，来聊聊这个库的优劣势：

### 可选的参数序列化

从`viper`库的源码中([点击跳转](https://github.com/spf13/viper/blob/v1.7.0/viper.go#L328))，我们可以看到它支持多种**本地文件类型**与**远程k-v数据库**：

```go
// SupportedExts are universally supported extensions.
var SupportedExts = []string{"json", "toml", "yaml", "yml", "properties", "props", "prop", "hcl", "dotenv", "env", "ini"}

// SupportedRemoteProviders are universally supported remote providers.
var SupportedRemoteProviders = []string{"etcd", "consul", "firestore"}
```

我们先忽略远程的存储，先看一下最常用的几个序列化的库：

1. JSON: 官方自带的`encoding/json`
2. TOML: 开源的`github.com/pelletier/go-toml`
3. YAML: 官方推荐的`gopkg.in/yaml.v2`
4. INI：官方推荐的`gopkg.in/ini.v1`

在这四种技术选型时，我个人倾向于选择`JSON`和`YAML`。进一步斟酌时，虽然`JSON`的适用范围最广，但当配置文件复杂到一定程度时，`JSON`格式的配置文件很难通过换行来约束，当存在大量的嵌套时，可读性很差。所以，我个人比较推荐使用`YAML`格式的配置文件，一方面通过强制的换行约束，可读性很棒；另一方面云原生相关技术大量使用了`YAML`作为配置文件，尤其是`Kubernetes`中各种定义。

例如，我们将服务的端口改造到配置文件里，就成了：

```yaml
server:
  http:
    port: 8081
  grpc:
    port: 8082
```

对应的Go语言代码为：

```go
viper.GetInt("server.http.port")
viper.GetInt("server.grpc.port")
```

### 可扩展的获取参数方法

`viper`库提供的获取参数方式为`viper.Get{type}("{level1}.{level2}.{level3}...")`的格式。随着配置文件的扩大，也只需新增Get方法即可。

从获取参数的方法来看，它的设计分为3种：

1. 基本类型，直接提供Get{具体类型}的方法，如`GetInt`，`GetString`；
2. 任意类型，提供`Get(key string) interface{} `，自行转化
3. 复杂类型的反序列化，提供`UnmarshalKey`等方法，更方便地获取复杂结构

我个人建议各位只使用第一类的方法，将配置文件这个模块做到最简化。毕竟，**配置文件的复杂化很容易引入各种问题，占用大量的排查故障的时间**。如果你的系统必须引入一套非常复杂的配置，那么我更建议将它独立成一个专门的服务，用来维护这套配置。



## 3.使用viper库的推荐方式

如果你仔细地阅读viper相关的代码，你会发现它有很多超酷的特性。但今天，我想告诉各位：**请克制地使用进阶的特性，最棒的特性往往是简洁的**。

我们对照着官方的README文件中介绍的特性进行讲解。

### 尽量避免手动设置的参数值

[原文链接](https://github.com/spf13/viper#establishing-defaults)

用`viper.SetDefault`函数可以给某些参数设置默认值，如果只是少数的几个参数还是很容易维护的。但如果设置的值过多了，就会给阅读代码的人带来困扰：**这个参数是来自配置文件，还是在代码某处手动设置的？**也就是存在**二义性**，增加了排查问题的复杂度。

### 明确配置文件的来源

[原文链接](https://github.com/spf13/viper#reading-config-files)

```go
viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
viper.AddConfigPath(".")               // optionally look for config in the working directory
```

`viper`支持多个配置文件的路径，这虽然带来了便利性，但如果多个文件路径中都存在配置文件，那究竟以哪个为准？这也是一个**二义性**的问题，所以我个人更建议只设置一个，而这个路径由`flag`传入。

### 静态配置与动态配置的分离

[原文链接](https://github.com/spf13/viper#watching-and-re-reading-config-files)

`viper`提供了接口`viper.WatchConfig()`，可以监听文件的变化，然后做相应的调整。这个特性很酷，我们可以用它实现**热加载**。但这个特性很容易让人产生混淆：例如发生了某个BUG，如何确定当时的配置文件情况？其实，这就需要引入一定的**版本管理**机制。

我更建议采用**静态配置和动态配置分离**的方案，也就是配置文件负责静态的、固定的配置，一旦启动后只加载一次；而动态的配置放在带版本管理的配置中心里，具备热加载的特性。

所以，我不太建议在配置文件这里引入监听文件变化的特性。

## 核心代码示例

### main.go

从`flag`中解析出配置文件路径，传到`config`包中用于解析。

```go
var configFilePath = flag.String("c", "./", "config file path")
flag.Parse()

if err := config.Load(*configFilePath); err != nil {
  panic(err)
}
```

### internal/config/viper.go

加载的代码并不多，尽量保证配置信息的简洁易懂。

```go
// 全局Viper变量
var Viper = viper.New()

func Load(configFilePath string) error {
	Viper.SetConfigName("config")       // config file name without file type
	Viper.SetConfigType("yaml")         // config file type
	Viper.AddConfigPath(configFilePath) // config file path
	return Viper.ReadInConfig()
}
```

### 配置使用方

```go
config.Viper.GetInt("server.grpc.port")
```

## 使用viper库的注意事项

在使用`viper`获取配置时，我们需要手动组装`key`，也就是`"{level1}.{level2}.{level3}..."`这种形式。这时，我们只能对照着原始配置文件逐个填充字段，一不小心填错、就会发生奇怪的问题。而如果采用的是将配置文件解析到结构体的方法，就能很容易地避免这个问题。

考虑到扩展性，官方库推荐的是手动组装key的方式，所以需要大家在认真查看这个`key`是否有效。

## 总结

加载静态配置文件是一个很常见的功能，`viper`提供了一套完整方案，兼具简洁和扩展性；与此同时，我们要学会**克制**，不要看到了`viper`中提供的各种特性、就想着应用到实际项目中，也就是常说的：**手里拿了个锤子，看啥都是钉子**。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

