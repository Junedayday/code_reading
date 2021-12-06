---
title: Go语言微服务框架 - 8.Gormer迭代-定制专属的ORM代码生成工具
date: 2021-10-10 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

在上一篇，我们写一个`gormer`工具库，支持了简单的CRUD。但是，在实际的开发场景中，这部分的功能仍显得非常单薄。

例如，我们对比一下GORM库提供的`gorm.Model`，它在新增、修改时，会自动修改对应的时间，这个可以帮我们减少很多重复性的代码编写。这里，我就针对现有的gormer工具做一个示例性的迭代。

<!-- more -->

## v0.5.2：Gormer迭代-定制更智能的代码生成工具

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.5.2

### 目标

生成一套智能化的Dao层代码，兼容软删除和硬删除。

> 这里提一下软删除的概念，就是指在数据库中用某个字段标记为删除，但这行数据仍存在；而硬删除就是直接删除整条数据。
>
> 软删除虽然增加了一定的复杂度，但带来的收益很大。最直接的好处就是能保留记录，方便查原始记录。

### 关键技术点

1. gormer.yaml的文件
2. 模板文件的修改
3. 核心结构体梳理
4. API调用示例

### 目录构造

为了方便理解，我简化对应的目录结构

```
--- micro_web_service            项目目录
	|-- gen                            从idl文件夹中生成的文件，不可手动修改
	   |-- idl                             对应idl文件夹
	      |-- demo                             对应idl/demo服务，包括基础结构、HTTP接口、gRPC接口
	    	|-- order                            对应idl/order服务，同上
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
     |-- gormer                          修改：gormer二进制工具，用于生成Gorm相关Dao层代码
	|-- buf.gen.yaml                   buf生成代码的定义，从v1beta升到v1
	|-- buf.yaml                       buf工具安装所需的工具，从v1beta升到v1
	|-- gen.sh                         生成代码的脚本：buf+gormer
	|-- go.mod                         Go Module文件
	|-- gormer.yaml                    新增：将gormer中的参数移动到这里
	|-- main.go                        项目启动的main函数
```

## 1.gormer.yaml的文件

这里先给出具体的建表语句，可以清晰地看到orders表6个字段的具体含义：

```sql
CREATE TABLE orders
(
id bigint PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
name varchar(255) COMMENT '名称，建议唯一',
price decimal(15,3) COMMENT '订单价格',
create_time timestamp NULL DEFAULT NULL COMMENT '创建时间',
update_time timestamp NULL DEFAULT NULL COMMENT '更新时间',
delete_status tinyint(3) COMMENT '删除状态，1表示软删除'
) COMMENT='订单信息表';
```

我们回顾一下之前的gormer程序，它采用了`flag`参数解析的方式。但随着复杂度提升，命令行参数包含了大量的内容，很难维护。这时，就建议采用**配置文件**的方式，保证可读性、可维护性。

```yaml
# 数据库相关的信息
database:
  # 数据库连接
  dsn: "root:123456@tcp(127.0.0.1:3306)/demo"
  # 所有要生成到Go结构体中的表
  tables:
    # name-表名
    # goStruct-Go中结构体名
    # createTime-创建时间的数据库字段，必须为时间格式
    # updateTime-更新时间的数据库字段，必须为时间格式
    # softDeleteKey-软删除的数据库字段，必须为整数型，不填则为硬删除
    # softDeleteValue-表示为软删除的对应值
    - name: "orders"
      goStruct: "Order"
      createTime: "create_time"
      updateTime: "update_time"
      softDeleteKey: "delete_status"
      softDeleteValue: 1

# 项目相关的信息
project:
  # 项目的路径
  base: "./"
  # gorm相关核心结构的代码路径
  gorm: "internal/gormer/"
  # dao层CRUD核心结构的代码路径
  dao: "internal/dao/"
  # 项目的go module信息
  go_mod: "github.com/Junedayday/micro_web_service"
```

## 2.模板文件的修改

这里以两个具有代表性的操作为例，一起来看看具体代码。

### 新增

利用了go template的特性，填充了create_time和update_time字段。这里包含两层if语句：

- 第一层：在`gormer.yaml`里必须指定了createTime代码，否则不要生成这段代码
- 第二层：如果外部传进来的字段里没有指定时间，才填充最新的时间；否则以外部传入为准

```go
daoTmplAdd = `func (repo *{{.StructName}}Repo) Add{{.StructName}}({{.StructSmallCamelName}} *gormer.{{.StructName}}) (err error) {
{{if ne .FieldCreateTime "" }}
    if {{.StructSmallCamelName}}.{{.FieldCreateTime}}.IsZero() {
		{{.StructSmallCamelName}}.{{.FieldCreateTime}} = time.Now()
	}
{{end}}
{{if ne .FieldUpdateTime "" }}
    if {{.StructSmallCamelName}}.{{.FieldUpdateTime}}.IsZero() {
		{{.StructSmallCamelName}}.{{.FieldUpdateTime}} = time.Now()
	}
{{end}}
	err = repo.db.
		Table(gormer.{{.StructName}}TableName).
		Create({{.StructSmallCamelName}}).
		Error
	return
}
`

// 生成后
func (repo *OrderRepo) AddOrder(order *gormer.Order) (err error) {

	if order.CreateTime.IsZero() {
		order.CreateTime = time.Now()
	}

	if order.UpdateTime.IsZero() {
		order.UpdateTime = time.Now()
	}

	err = repo.db.
		Table(gormer.OrderTableName).
		Create(order).
		Error
	return
}
```

### 删除

删除的逻辑主要区分了一个字段，即是否在`gormer.yaml`里指定了软删除的字段。

- 指定了软删除的字段，则将这个字段更新为设定的值、并且更新updateTime字段；
- 未指定软删除的字段，则直接硬删除对应的记录；

```go
daoTmplDelete = `func (repo *{{.StructName}}Repo) Delete{{.StructName}}(condition *gormer.{{.StructName}}Options) (err error) {
	if condition == nil {
		return errors.New("delete must include where condition")
	}

	err = repo.db.
        Table(gormer.{{.StructName}}TableName).
		Where(condition.{{.StructName}}, condition.Fields).
{{if eq .FieldSoftDeleteKey "" }} Delete(&gormer.{{.StructName}}{}).
{{ else }}  {{if eq .FieldUpdateTime "" }}
				Select("{{.TableSoftDeleteKey}}").
				Updates(&gormer.{{.StructName}}{
					{{.FieldSoftDeleteKey}}:{{.TableSoftDeleteValue}},
				}).
            {{ else }}
                Select("{{.TableSoftDeleteKey}}","{{.TableUpdateTime}}").
				Updates(&gormer.{{.StructName}}{
					{{.FieldSoftDeleteKey}}:{{.TableSoftDeleteValue}},
					{{.FieldUpdateTime}} : time.Now(),
				}).
            {{ end }}
{{ end }}
		Error
	return
}
`

// 生成后
func (repo *OrderRepo) DeleteOrder(condition *gormer.OrderOptions) (err error) {
	if condition == nil {
		return errors.New("delete must include where condition")
	}

	err = repo.db.
		Table(gormer.OrderTableName).
		Where(condition.Order, condition.Fields).
		Select("delete_status", "update_time").
		Updates(&gormer.Order{
			DeleteStatus: 1,
			UpdateTime:   time.Now(),
		}).
		Error
	return
}
```

## 3.核心结构体梳理

我们再一起看看表结构对应到Go结构体的一个关键结构体，这里分成了4个重要的部分：

1. 表名、结构体名
2. 表中的列信息、结构体中的Field字段信息
3. 创建时间、更新时间
4. 软删除的字段

这个数据结构体，其实是将两个数据源进行了关联映射：

- 原始信息：从MySQL表中查询
- 自定义字段信息：从gormer.yaml查询

```go
type StructLevel struct {
	// table -> struct
	TableName            string
	StructName           string
	StructSmallCamelName string

	// table column -> struct field
	Columns []FieldLevel

	// create time
	TableCreateTime string
	FieldCreateTime string
	// update time
	TableUpdateTime string
	FieldUpdateTime string

	// soft delete
	TableSoftDeleteKey   string
	TableSoftDeleteValue int
	FieldSoftDeleteKey   string
}

type FieldLevel struct {
	FieldName string
	FieldType string
	// gorm tag for field
	GormName string
	// comment from create table sql
	Comment string
}
```

## 4.API调用示例

从API调用的角度来看，程序对外接口如下。有兴趣的可以体验下：

```shell
// List
curl --location --request GET 'http://127.0.0.1:8081/v1/orders'

// Create
curl --location --request POST 'http://127.0.0.1:8081/v1/orders' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "order4",
    "price": 100.3
}'

// Update
curl --location --request PATCH 'http://127.0.0.1:8081/v1/orders' \
--header 'Content-Type: application/json' \
--data-raw '{
    "order": {
        "id": "1",
        "name": "order1",
        "price": 110.8
    },
    "update_mask": "price"
}'

// Get
curl --location --request GET 'http://127.0.0.1:8081/v1/orders/order1'

// Delete
curl --location --request DELETE 'http://127.0.0.1:8081/v1/orders/order1'
```

## 延伸思考

修改到这个版本，gormer工具已经达到了基本可用的阶段。我们回顾一下重点功能：**根据数据库表结构，自动化生成dao层的CRUD代码**，并扩展了两特性：

1. 支持创建时间、修改时间的字段，自动填充
2. 支持软删除与硬删除

从更远的角度来看，还有许多MySQL的特性可以添加，尤其是对事务的支持，有兴趣的可以自行探索。限于篇幅与复杂度，目前就迭代到这个版本。

## 总结

Gormer是一个我们根据日常CRUD需求自行实现的工具，是框架实现高度自动化的重要环节。它的核心思想是 - **在重复的日常开发过程中找到可自动化的环节，实现Generate Code**。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

