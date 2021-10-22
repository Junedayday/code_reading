---
title: Go语言微服务框架 - 7.Gormer-自动生成代码的初体验
date: 2021-09-27 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

CRUD是贯穿整个程序员日常开发的基本工作，占据了我们绝大多数的coding时间。

作为一名程序员，我们总是希望能有更简单的开发方式来解决重复性的工作问题。在这个小版本中，我将结合自己的工作，来给出一套自动生成代码的完整方案，供大家借鉴。

<!-- more -->

## v0.5.1：Gormer-自动生成代码的初体验

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.5.1

### 目标

自动生成一套可用的Dao层代码，兼容原始版本。

### 关键技术点

1. Go Template技术概览
2. gormer工具核心思路
3. gormer的模板填充

### 目录构造

```
--- micro_web_service            项目目录
	|-- gen                            从idl文件夹中生成的文件，不可手动修改
	   |-- idl                             对应idl文件夹
	      |-- demo                             对应idl/demo服务
	         |-- demo.pb.go                        demo.proto的基础结构
	         |-- demo.pb.gw.go                     demo.proto的HTTP接口，对应gRPC-Gateway
	         |-- demo_grpc.pb.go                   demo.proto的gRPC接口代码
	    	|-- order                            对应idl/order服务
	         |-- order.pb.go                       order.proto的基础结构
	         |-- order.pb.gw.go                    order.proto的HTTP接口，对应gRPC-Gateway
	         |-- order_grpc.pb.go                  order.proto的gRPC接口代码
	|-- idl                            原始的idl定义
	   |-- demo                            业务package定义
	      |-- demo.proto                       protobuffer的原始定义
	   |-- order                           业务order定义
	      |-- order.proto                      protobuffer的原始定义
	|-- internal                       项目的内部代码，不对外暴露
	   |-- config                          配置相关的文件夹
	      |-- viper.go                         viper的相关加载逻辑
	   |-- dao                             Data Access Object层
	      |-- order.go                         Order对象，订单表，实现model层的OrderRepository
	      |-- order_test.go                    Order的单元测试
	   |-- gormer                          新增：从pkg/gormer中生成的相关代码，不允许更改
	      |-- order.go                         新增：gormer从orders表中获取的真实Gorm结构体
	   |-- model                           model层，定义对象的接口方法
	      |-- order.go                         OrderRepository接口，具体实现在dao层
	   |-- mysql                           MySQL连接
	      |-- init.go                          初始化连接到MySQL的工作
	   |-- server                          服务器的实现
	      |-- demo.go                          server中对demo这个服务的接口实现
	      |-- server.go                        server的定义，须实现对应服务的方法
	   |-- service                         service层，作为领域实现的核心部分
	      |-- order.go                         Order相关的服务，目前仅简单的CRUD
     |-- zlog                            封装日志的文件夹
        |-- zap.go                           zap封装的代码实现
  |-- pkg                            新增：开放给第三方的工具库
     |-- gormer                          新增：gormer工具，用于生成Gorm相关Dao层代码
	|-- buf.gen.yaml                   buf生成代码的定义，从v1beta升到v1
	|-- buf.yaml                       buf工具安装所需的工具，从v1beta升到v1
	|-- gen.sh                         更新：生成代码的脚本：buf+gormer
	|-- go.mod                         Go Module文件
	|-- main.go                        项目启动的main函数
```

## 1.Go Template技术概览

Go的标准库提供了Template功能，但网上的介绍很零散，我建议大家可以阅读以下两篇资料：

- 原理性：官方文档 - https://pkg.go.dev/text/template 
- 实践性：Blog - https://blog.gopheracademy.com/advent-2017/using-go-templates/ 

这里，为了方便大家阅读下面的内容，我简要概括下：

1. 结构体中字段填充 `{{ .FieldName }}`
2. 条件语句 `{{if .FieldName}} // action {{ else }} // action 2 {{ end }}`
3. 循环 `{{range .Member}} ... {{end}}`
4. 流水线 `{{ with $x := <^>result-of-some-action<^> }} {{ $x }} {{ end }}`

> 很多资料会很自然地将Go Template和HTML结合起来，但这只是模板的其中一个用法。
>
> HTML的结构用模板化的方式可以减少大量重复性的代码，但这种思路是前后单不分离的，个人不太推荐。



## 2.gormer工具核心思路

在pkg/gormer目录下提供了一个gormer工具，用于自动生成代码，我对主流程进行简单地讲解：

1. 解析各种关键性的参数
2. 连接测试数据库，获取表信息
3. 逐个处理每个表
   1. 读取数据库中的表结构
   2. 根据表结构生成对应的Go语言结构体，放在internal/gormer下
   3. 生成相关的Dao层代码，放在internal/dao下
4. 执行go fmt格式化代码

其中最关键的是3-b与3-c，它们是生成代码的最关键步骤。我们来看一个关键性的结构体：

```go
// 结构体名称，对应MySQL表级别的信息
type StructLevel struct {
	TableName      string
	Name           string
	SmallCamelName string
	Columns        []FieldLevel
}

// Field字段名称，对应MySQL表里Column
type FieldLevel struct {
	FieldName string
	FieldType string
	GormName  string
}
```

## 3.gormer的模板填充

结合1、2，我们可以开始生成模板的部分，具体的Template代码如下，它会将StructLevel这个结构体中的字段填充到下面内容中，生成go文件。

```go
var gormerTmpl = `
// Table Level Info
const {{.Name}}TableName = "{{.TableName}}"

// Field Level Info
type {{.Name}}Field string
const (
{{range $item := .Columns}}
    {{$.Name}}Field{{$item.FieldName}} {{$.Name}}Field = "{{$item.GormName}}" {{end}}
)

var {{$.Name}}FieldAll = []{{$.Name}}Field{ {{range $k,$item := .Columns}}"{{$item.GormName}}", {{end}}}

// Kernel struct for table for one row
type {{.Name}} struct { {{range $item := .Columns}}
	{{$item.FieldName}}	{{$item.FieldType}}	` + "`" + `gorm:"column:{{$item.GormName}}"` + "`" + ` {{end}}
}

// Kernel struct for table operation
type {{.Name}}Options struct {
    {{.Name}} *{{.Name}}
    Fields []string
}

// Match: case insensitive
var {{$.Name}}FieldMap = map[string]string{
{{range $item := .Columns}}"{{$item.FieldName}}":"{{$item.GormName}}","{{$item.GormName}}":"{{$item.GormName}}",
{{end}}
}

func New{{.Name}}Options(target *{{.Name}}, fields ...{{$.Name}}Field) *{{.Name}}Options{
    options := &{{.Name}}Options{
        {{.Name}}: target,
        Fields: make([]string, len(fields)),
    }
    for index, field := range fields {
        options.Fields[index] = string(field)
    }
    return options
}

func New{{.Name}}OptionsAll(target *{{.Name}}) *{{.Name}}Options{
    return New{{.Name}}Options(target, {{$.Name}}FieldAll...)
}

func New{{.Name}}OptionsRawString(target *{{.Name}}, fields ...string) *{{.Name}}Options{
    options := &{{.Name}}Options{
        {{.Name}}: target,
    }
    for _, field := range fields {
        if f,ok := {{$.Name}}FieldMap[field];ok {
             options.Fields = append(options.Fields, f)
        }
    }
    return options
}
`
```

生成的代码如下：

```go
// Code generated by gormer. DO NOT EDIT.
package gormer

import "time"

// Table Level Info
const OrderTableName = "orders"

// Field Level Info
type OrderField string

const (
	OrderFieldId         OrderField = "id"
	OrderFieldName       OrderField = "name"
	OrderFieldPrice      OrderField = "price"
	OrderFieldCreateTime OrderField = "create_time"
)

var OrderFieldAll = []OrderField{"id", "name", "price", "create_time"}

// Kernel struct for table for one row
type Order struct {
	Id         int64     `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	Price      float64   `gorm:"column:price"`
	CreateTime time.Time `gorm:"column:create_time"`
}

// Kernel struct for table operation
type OrderOptions struct {
	Order  *Order
	Fields []string
}

// Match: case insensitive
var OrderFieldMap = map[string]string{
	"Id": "id", "id": "id",
	"Name": "name", "name": "name",
	"Price": "price", "price": "price",
	"CreateTime": "create_time", "create_time": "create_time",
}

func NewOrderOptions(target *Order, fields ...OrderField) *OrderOptions {
	options := &OrderOptions{
		Order:  target,
		Fields: make([]string, len(fields)),
	}
	for index, field := range fields {
		options.Fields[index] = string(field)
	}
	return options
}

func NewOrderOptionsAll(target *Order) *OrderOptions {
	return NewOrderOptions(target, OrderFieldAll...)
}

func NewOrderOptionsRawString(target *Order, fields ...string) *OrderOptions {
	options := &OrderOptions{
		Order: target,
	}
	for _, field := range fields {
		if f, ok := OrderFieldMap[field]; ok {
			options.Fields = append(options.Fields, f)
		}
	}
	return options
}
```

dao层的代码逻辑类似，我就不重复填写了。

这里，我将代码拆分成了gormer与dao两层，主要是：

- internal/gormer整个目录是不可变的、只能自动生成，对应基础的数据库表结构
- internal/dao层会添加其余的文件，如定制化的sql。

至此，再将引用的相关代码简单修改，就实现了这一整块功能.

## 总结

本章重点介绍了Go Template在高度重复的代码模块中的应用，结合数据库实现了一个高度自动化的工具gormer。

gormer目前实现的功能比较单一，但只要有了初步自动化的思路，我们可以在后续迭代中慢慢优化，让它适应更多的场景。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

