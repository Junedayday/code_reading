---
title: Go编程模式 - 7-代码生成
date: 2021-02-20 18:32:05
categories: 
- 经典品读
tags:
- Go-Programming-Patterns
---

>  注：本文的灵感来源于GOPHER 2020年大会陈皓的分享，原PPT的[链接](https://www2.slideshare.net/haoel/go-programming-patterns?from_action=save)可能并不方便获取，所以我下载了一份[PDF](https://github.com/Junedayday/code_reading/tree/master/doc/Go_Programming_Patterns.pdf)到git仓，方便大家阅读。我将结合自己的实际项目经历，与大家一起细品这份文档。



## 目录

- [简单脚本准备](#Simple-Script)
  - [模板文件](#Template)
  - [运行脚本](#Shell)
  - [入口文件](#Generate-File)
  - [运行原理](#Generation)
- [类型替换工具genny](#genny)
- [任意文件转Go](#go-bindata)
- [字符串生成工具stringer](#stringer)





## Simple Script

为了让大家快速了解这块，我们从一个最简单的例子入手。

### Template

首先创建一个模板Go文件，即容器模板：container.tmp.go

```go
package PACKAGE_NAME
type GENERIC_NAMEContainer struct {
    s []GENERIC_TYPE
}
func NewGENERIC_NAMEContainer() *GENERIC_NAMEContainer {
    return &GENERIC_NAMEContainer{s: []GENERIC_TYPE{}}
}
func (c *GENERIC_NAMEContainer) Put(val GENERIC_TYPE) {
    c.s = append(c.s, val)
}
func (c *GENERIC_NAMEContainer) Get() GENERIC_TYPE {
    r := c.s[0]
    c.s = c.s[1:]
    return r
}
```

### Shell

生成的shell脚本，gen.sh

```shell
#!/bin/bash

set -e

SRC_FILE=${1}
PACKAGE=${2}
TYPE=${3}
DES=${4}
#uppcase the first char
PREFIX="$(tr '[:lower:]' '[:upper:]' <<< ${TYPE:0:1})${TYPE:1}"

DES_FILE=$(echo ${TYPE}| tr '[:upper:]' '[:lower:]')_${DES}.go

sed 's/PACKAGE_NAME/'"${PACKAGE}"'/g' ${SRC_FILE} | \
    sed 's/GENERIC_TYPE/'"${TYPE}"'/g' | \
    sed 's/GENERIC_NAME/'"${PREFIX}"'/g' > ${DES_FILE}
```

四个参数分别为

- 源文件名
- 包名
- 类型
- 文件后缀名

### Generate File

最后，增加一个创建代码的go文件。

```go
//go:generate ./gen.sh ./template/container.tmp.go gen uint32 container
func generateUint32Example() {
    var u uint32 = 42
    c := NewUint32Container()
    c.Put(u)
    v := c.Get()
    fmt.Printf("generateExample: %d (%T)\n", v, v)
}

//go:generate ./gen.sh ./template/container.tmp.go gen string container
func generateStringExample() {
    var s string = "Hello"
    c := NewStringContainer()
    c.Put(s)
    v := c.Get()
    fmt.Printf("generateExample: %s (%T)\n", v, v)
}
```

### Generation

我们运行一下 `go generate`，就能产生对应的文件。

1. 运行go generate，工具会扫描所有的文件
2. 如果发现注释有带 go:generate的，会自动运行后面的命令
3. 通过命令生成的代码，会在源文件添加提示，告诉他人这是自动生成的代码，不要编辑

因此，我们不仅仅可以用`shell`脚本，也可以用各种二进制工具来生成代码。值得一提的是，像Kubernetes这种重量级的项目，大量地应用了这种特性。后面我也会和大家分享在开发web项目中的应用。



下面，我也来介绍几个个人认为比较有用的工具。



## genny

源项目链接：https://github.com/cheekybits/genny

### Go文件示例

```go
package queue

import "github.com/cheekybits/genny/generic"

// NOTE: this is how easy it is to define a generic type
type Something generic.Type

// SomethingQueue is a queue of Somethings.
type SomethingQueue struct {
  items []Something
}

func NewSomethingQueue() *SomethingQueue {
  return &SomethingQueue{items: make([]Something, 0)}
}
func (q *SomethingQueue) Push(item Something) {
  q.items = append(q.items, item)
}
func (q *SomethingQueue) Pop() Something {
  item := q.items[0]
  q.items = q.items[1:]
  return item
}
```

### 脚本

```shell
cat source.go | genny gen "Something=string"
```

官方示例还是采用的是shell脚本，建议替换到 go:generate 中，这样的代码更统一

### 原理

可以简单地理解成一个类型替换的工具（PS：擅长用sed脚本的朋友也可直接通过shell脚本实现）



## go-bindata

源网站链接：https://github.com/go-bindata/go-bindata

go-bindata的功能是将任意格式的源文件，转化为Go代码，使我们无需再去打开文件读取了。

这个工具多用在静态网页转化为Go代码（不符合前后端分离的实践），所以具体的使用方式我就不细讲了，大家有兴趣的可以自行阅读教程。

但它有两个优点值得我们关注：无需再进行文件读取操作、压缩。



## stringer

stringer是官方提供一个字符串工具，我个人非常推荐大家使用

文档链接：https://pkg.go.dev/golang.org/x/tools/cmd/stringer 

### Go文件

```go
package painkiller

type Pill int

const (
	Placebo Pill = iota
	Aspirin
	Ibuprofen
	Paracetamol
	Acetaminophen = Paracetamol
)
```

### 脚本

```go
//go:generate stringer -type=Pill
```

于是，就会生成对应的方法`func (Pill) String() string`，也就是直接转化成了其命名。

### 价值

Go语言在调用 `fmt` 等相关包时，如果要将某个变量转化为字符串，默认会寻找它的`String()`方法。

这时，**良好的命名** 能体现出其价值。尤其是在错误码的处理上，无需再去查询错误码对应的错误内容，直接可以通过命名了解。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili：https://space.bilibili.com/293775192
>
> 公众号：golangcoding

