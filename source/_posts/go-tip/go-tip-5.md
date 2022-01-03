---
title: Go语言技巧 - 5.【初探Go Module】Go语言的版本管理
date: 2021-07-03 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## Go Mod的官方说明

Go语言自从推出了`go mod`作为版本管理工具后，结束Go语言版本管理工具的纷争，实现了大一统。

相信有很多人都对这个版本管理的机制都有基础的概念、但并不深入。而官方把最核心的实现，都放在这一篇 https://golang.org/ref/mod 文档中。

今天，我们一起来读读这一篇文章。

<!-- more -->

## 快速入门 - 5篇介绍Go Mod系列的官方博客

新手直接阅读这篇文章的门槛有点高，我建议可以先看看下面这五篇较为通俗的官方博客，能帮助我们了解一些背景知识。

### 1. using-go-modules 

https://blog.golang.org/using-go-modules 

1. `go.mod`放在项目的根目录，抛弃原来的`GOPATH`
2. 用`MAJOR.MINOR.PATCH`格式管理版本，详细可参考[Semantic Versioning 2.0.0](https://semver.org/)
3. 用`go.sum`保证依赖文件被完整下载（如果公司搭建私有库就会出现校验问题，需要关闭GOSUM）
4. 项目内部的库，不再是相对路径，而用的是 `go mod模块名` + `相对路径` ，定义更加清晰

> 第四点的价值很大，可读性大大提高，我们可以将一整个path来辅助命名，如 market/order，前者market可以帮助后者order的含义做一定补充

### 2. migrating-to-go-modules

https://blog.golang.org/migrating-to-go-modules

将项目迁移到go mod，主要讲的是对接老的版本管理系统，如`godeps`。

### 3. publishing-go-modules

https://blog.golang.org/publishing-go-modules

版本命名规则：推荐不稳定版本用`v0.x.x`开始，稳定后改成`v1.x.x`。

**pseudo-version** 直译为假的版本，一般是直接依赖branch，而不是按规范依赖tag

### 4. v2-go-modules

https://blog.golang.org/v2-go-modules

注意主版本名为`v2`之后的，项目里对应的要新增一个`v2`或者更高版本号的目录。

> 我个人感觉升级到v2这种方式不太友好，需要我们再维护一整个目录

### 5. module-compatibility

https://blog.golang.org/module-compatibility

1. 用不定参数`...`的特性提高函数的入参扩展性
2. 引入函数式的参数，也就是将具体的执行逻辑作为参数，传递进来
3. 引入`Option types`模式，更详细地可以参考我的[文章](https://junedayday.github.io/2021/02/20/go-patterns/go-patterns-5/)
4. 用接口`interface`分离调用和实现
5. 如果结构体`struct`明确不能对比，就用一个`doNotCompare`的`field`来告诉调用者

```go
type doNotCompare [0]func()

type Point struct {
        doNotCompare
        X int
        Y int
}
```

> 第五点非常有意思，也就是要求我们在前期设计对象时，就要考虑清楚这个对象后续的特性，比如示例中的是否可以对比



## 从go.mod的文件格式讲起

进入正题，我们来一起看看`go.mod`，它的定义简单示例如下：

```
module example.com/my/thing

go 1.12

require example.com/other/thing v1.0.2
require example.com/new/thing/v2 v2.3.4
exclude example.com/old/thing v1.2.3
replace example.com/bad/thing v1.4.5 => example.com/good/thing v1.4.5
retract [v1.9.0, v1.9.5]
```

IDE会帮助你格式化，记住以下关键词即可（重点为前三个）

1. **module** - go mod init指令定义的**库名**
2. **go** - **要求go语言的最低版本**，会影响到后面依赖库的下载
3. **require** - **必备库**，也就是代码中直接import的部分
4. **replace** - **替换库**，在重构时挺好用（比如某个开源组件有问题，内部fork了一版，直接replace即可）
5. **retract** **撤回版本**，告诉调用本库的项目，部分版本有严重问题、不要引用

> go mod 底层实现依赖 - **MVS** 最小版本选择。
>
> 这个特性很有意思，后续单独来讲讲这块，一开始就不深入到细节了



## 加深理解 - incompatible和indirect

在我们整理`go.mod`文件时，经常能看到两个奇怪的字符`indirect`和`incompatible`。我们来详细地分析一下。

### incompatible - 兼容v2及以上的版本号

上面我们已经讲过，如果一个库的tag为v1以上，如v2，就必须得创建一个v2的目录。例如

```
require example.com/new/thing/v2 v2.3.4
```

这就要求我们在项目`example.com/new/thing`下新建v2目录，再存放代码。但是，很多库往往只是升级个主版本号，并不会去新建目录、还需要迁移代码。为了兼容这个情况，就会引入`+incompatible`。例如

```
require example.com/new/thing v2.3.4+incompatible
```

### indirect - 未在go.mod里定义、但间接调用的库

我们先聊一个简单的场景：**当前项目为A，调用了项目B，B又调用了C。对A进行编译，需要B和C的相关代码**。

在完全规范的项目中：

- **条件1** - A的go.mod里包含B
- **条件2** - B的go.mod里包含C

在编译A时，会在go.mod找到B的信息，所以B是`require`字段；而C的信息已经被维护在B的go.mod里了，不需要在A的go.mod里维护。

而什么样的情况会发生indirect呢？它对应的是 **条件2** 缺失的场景

1. B**没有启用Go Module**，采用的是老项目管理方式
2. B的**go.mod部分缺失**，未填写模块C

> 最常见的部分缺失场景是：项目虽然有go.mod，但实际编译不走Go Module，而是如vendor目录等方式

用一句话总结，**A库无法根据B库的`go.mod`找到C库**。



## 常见命令介绍

相关的指令有很多，我重点分两块来说：



先是**高频使用**的命令：

### 用go mod init初始化项目

初始化项目，保证module名称与git路径一致。

例如 `go mod init github.com/example/a`

### 用go get下载指定依赖库与版本

常见的flags

- **-d** 只更新`go.mod`中的依赖，轻量级
- **-u** 更新指定库与依赖它的库，全量

例如`go get -d github.com/example/b`

### 根据go.mod下载依赖库go mod download/vendor

其中download是下载到Go Module的缓存中，而vendor是下载到vendor依赖路径。官方推荐前者。

我经常会去手动编辑`go.mod`文件，然后用这个指令刷新一下依赖库

### 整理依赖go mod tidy

整理并更新go mod的依赖信息，保证当前的`go.mod`为最新。



然后是**排查依赖库问题**用到的：

### 查看库的支持版本go list

- `go list -m all` 查看本项目的所有依赖库与版本
- `go list -m -versions {module名}` 查看module支持的版本号
- `go list -m -json {module名}@{版本号}` 用json格式查看指定module版本号的信息，如创建时间

### 查看当前库的依赖关系go mod graph

查看所有go mod的依赖，一般在查依赖关系时用到

### 查看指定库是怎么被依赖的go mod why

查指定库是怎么被依赖的

###  查看二进制文件的依赖信息go version -m

查看指定（go文件编译的）二进制文件的版本信息



## 设置GOPRXOY

大部分人使用`go.mod`的最大问题是无法下载代码库，也就是代理的设置，网上也有很多教程，我这边给三个我常用的：

1. 阿里云：GOPROXY=https://mirrors.aliyun.com/goproxy,direct
2. 七牛云：GOPROXY=https://goproxy.cn,direct
3. 全球代理：GOPROXY=https://goproxy.io,direct

> 公司私有库需要私有代理。



## 小结

本讲的内容到这里就告一段落了，相信通过这篇文章，大家已经能应对绝大部分Go Module的场景。

下一讲，我会重点讲Go Module最核心的 **Minimal version selection (MVS)** 机制。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

