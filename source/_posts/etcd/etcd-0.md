---
title: etcd源码分析 - 0.搭建学习etcd的环境
date: 2022-06-14 12:00:00
categories: 
- 技术框架
tags:
- Go-etcd
---

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd.jpg)

之前，我在b站视频简单地讲述了etcd的功能与特性，有兴趣的可以参考[相关视频](https://www.bilibili.com/video/BV155411Y7Pq/)。

但如果要更深入地研究etcd，就需要我们涉及到源码、并结合实践进行学习。那么，接下来，我将基于`v3.4`这个版本，做一期深入的环境搭建。

<!-- more -->

## 环境准备

1. Macbook - 为了方便读代码与编译运行，也可自行搭建Ubuntu等可视化系统
2. Go语言 - v1.17，我选用的是v1.17.11
3. Goland/VSCode
4. etcd源码 - 建议用Github Desktop进行下载

## 基本调试

为了保证etcd可运行，我们先在根目录上运行`go mod tidy`，保证依赖库没有问题。

接着，我们阅读`Makefile`文件，发现其提供了`make build`指令。运行后，在`bin`目录下生成了`etcd`/`etcdctl`/`etcdutl`三个可执行文件，并且打印出了版本信息。

```txt
./bin/etcd --version
etcd Version: 3.4.18
Git SHA: c2c9e7de0
Go Version: go1.17.11
Go OS/Arch: darwin/amd64
./bin/etcdctl version
etcdctl version: 3.4.18
API version: 3.4
```

我们暂时只关注`etcd`与`etcdctl`，可以简单地将两者理解为服务端与客户端。我们分别在两个终端进行操作：

```shell
# 运行etcd server
./bin/etcd
```

```shell
# 写入一个key
./bin/etcdctl put mykey "this is awesome"

# 读取一个key
./bin/etcdctl get mykey
```

如果你能读取到对应的信息，那么就证明整个环境已经很好地运行起来了。

## 从Makefile看Go的编译步骤

在日常开发的过程中，我们对Go程序的编译往往只是一行简单的`go build`，但在大型工程中往往还不够。我们看看etcd做了什么。

### GIT_SHA

```shell
GIT_SHA=$(git rev-parse --short HEAD || echo "GitNotFound")
GO_LDFLAGS="$GO_LDFLAGS -X ${REPO_PATH}/version.GitSHA=${GIT_SHA}"
```

这个参数是取git最新一次的commit的短hash，用来标识源码的版本，比如c2c9e7de0。

然后，将这个相对唯一的值，作为GO_LDFLAGS中的一个参数，打入到go程序中。

### ldflags

在Makefile中的编译里，我们会用到`-ldflags "$GO_LDFLAGS"`这个参数。通过运行`go help build`，可以看到这么一段说明：

```text
 -ldflags '[pattern=]arg list'
                arguments to pass on each go tool link invocation.
```

也就是用key=value对的格式，将想要的信息传递给Go程序。

> ldflags可以记忆为 load flags，即将标记信息加载到程序中。

## 传递ldflags中的参数

`ldflags`传递参数的方式是 `package_path.variable_name=new_value`。

以示例中的build为例，这个值为`go.etcd.io/etcd/version.GitSHA=${GIT_SHA}`，对应到三块：

1. package_path = go.etcd.io/etcd/version
1. variable_name = GitSHA
1. new_value = ${GIT_SHA}

所以，这里所做的就是将`go.etcd.io/etcd/version`这个package下的`GitSHA`变量替换为想要的值。我们去对应的代码里看，发现对应的代码：

```go
var (
	// MinClusterVersion is the min cluster version this etcd binary is compatible with.
	MinClusterVersion = "3.0.0"
	Version           = "3.4.18"
	APIVersion        = "unknown"

	// Git SHA Value will be set during build
	GitSHA = "Not provided (use ./build instead of go build)"
)
```

所以，我们可以通过编译脚本实现代码中变量的替换。

## 小结

etcd的学习环境搭建并不复杂，主要是有一台Mac电脑。接下来，我们将逐步开始一起阅读代码。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

