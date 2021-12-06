---
title: Go语言微服务框架 - 12.ORM层的自动抽象与自定义方法的扩展
date: 2021-12-02 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

随着接口参数校验功能的完善，我们能快速定位到接口层面的参数问题；而应用服务的分层代码，也可以通过log的trace-id发现常见的业务逻辑问题。

但在最底层与数据库的操作，也就是对GORM的使用，经常会因为我们不了解ORM的一些细节，导致对数据的CRUD失败，或者没有达到预期效果。这时，我们希望能在ORM这一层，也有一个通用的解决方案，来协助问题的排查。趁这个机会，我们也对`gormer`这个工具再做一次迭代，生成新的功能。

<!-- more -->

## v0.7.2：ORM层的自动抽象与自定义方法的扩展

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.7.2

### 目标

gormer工具支持interface的抽象与自定义方法的扩展，并具备日志打印功能。

### 关键技术点

1. model层的自动抽象方案
2. dao层的代码实现
3. MySQL的SQL打印
4. 关于gormer工具的迭代

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
	   |-- model                           修改：model层基本定义由gormer自动生成
	   |-- mysql                           修改：MySQL连接，支持日志打印
	   |-- server                          服务器的实现，对idl中定义服务的具体实现
	   |-- service                         service层，作为领域实现的核心部分
     |-- zlog                            封装zap日志的代码实现
  |-- pkg                            开放给第三方的工具库
     |-- gormer                          gormer二进制工具，用于生成Gorm相关Dao层代码
	|-- buf.gen.yaml                   buf生成代码的定义，新增参数校验逻辑
	|-- buf.yaml                       buf工具安装所需的工具，从v1beta升到v1
	|-- gen.sh                         生成代码的脚本：buf+gormer
	|-- go.mod                         Go Module文件
	|-- gormer.yaml                    将gormer中的参数移动到这里
	|-- main.go                        项目启动的main函数
	|-- swagger.sh                     生成openapiv2的相关脚本
```

## 1.model层的自动抽象方案

之前，我们在dao层已经实现了基本的CRUD相关代码，所以实现一个model层的定义很简单。但考虑到扩展性，也就是这个model层不仅仅需要简单的CRUD代码，还可能需要一些类似于`group by`等复杂sql，甚至包括子查询。

这时候，如果考虑全部用`gormer`工具自动生成的方案，那成本会很高，所以更建议分开维护的方案：简单的CRUD用自动代码生成的方式，而复杂SQL调用GORM库自行实现。我们来阅读代码：

```go
// *.go 自动生成的代码，标准方法
type OrderModel interface {
	AddOrder(ctx context.Context, order *gormer.Order) (err error)
	QueryOrders(ctx context.Context, pageNumber, pageSize int, condition *gormer.OrderOptions) (orders []gormer.Order, err error)
	CountOrders(ctx context.Context, condition *gormer.OrderOptions) (count int64, err error)
	UpdateOrder(ctx context.Context, updated, condition *gormer.OrderOptions) (err error)
	DeleteOrder(ctx context.Context, condition *gormer.OrderOptions) (err error)
	
	// Implement Your Method in ext model
	OrderExtModel
}

// *_ext.go 扩展方法
type OrderExtModel interface {
}
```

为了保证自定义的ext代码不被覆盖，在gormer的代码里添加如下代码：

```go
// 如果extFile已经存在，则不要覆盖
if _, err = os.Stat(extFile); err != nil {
  // 创建ext文件的代码
}
```

## 2.dao层的代码实现

dao层的代码基本同model层，分为`*.go`和`*_ext.go`两个。

为了保证dao层实现了model层的代码，我们也增加了一行代码，方便我们在编译期保证实现。

```go
var _ model.OrderModel = NewOrderRepo(nil)
```

## 3.MySQL的SQL打印

在GORM工具中，提供了一个callback的方式，让用户添加自定义的插件。具体可以参考 https://gorm.io/zh_CN/docs/write_plugins.html。主要实现分下面两步：

```go
// 1 - 操作SQL时，将ctx传入其中，用来传递一些通用参数，如traceid
func (repo *OrderRepo) AddOrder(ctx context.Context, order *gormer.Order) (err error) {
	repo.db.WithContext(ctx).
		Table(gormer.OrderTableName).
		Create(order)
	err = repo.db.Error
	return
}

// 2 - 在操作数据库后，注册对应的插件afterLog，用来打印SQL日志
func InitGorm(user, password, addr string, dbname string) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, addr, dbname)
	GormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	
	// 结束后
	_ = GormDB.Callback().Create().After("gorm:after_create").Register(callBackLogName, afterLog)
	_ = GormDB.Callback().Query().After("gorm:after_query").Register(callBackLogName, afterLog)
	_ = GormDB.Callback().Delete().After("gorm:after_delete").Register(callBackLogName, afterLog)
	_ = GormDB.Callback().Update().After("gorm:after_update").Register(callBackLogName, afterLog)
	_ = GormDB.Callback().Row().After("gorm:row").Register(callBackLogName, afterLog)
	_ = GormDB.Callback().Raw().After("gorm:raw").Register(callBackLogName, afterLog)
	return
}

const callBackLogName = "zlog"

func afterLog(db *gorm.DB) {
	err := db.Error
	ctx := db.Statement.Context
	
	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	if err != nil {
		zlog.WithTrace(ctx).Errorf("sql=%s || error=%v", sql, err)
		return
	}
	zlog.WithTrace(ctx).Infof("sql=%s", sql)
}
```

在`afterLog`这里，我们引用了插件，实现了自定义日志组件的打印。

## 4.关于gormer工具的迭代

在这个小版本中，我们又对gormer工具做了一次迭代。从整个框架的维度来看，我们不仅仅是把它作为一种代码生成的工具，而是一种模块化的抽象能力，关注分层能力的建设。从SQL的log打印来看，我们可以区分出前后的差异：

**原先** - 通过调用一个`公共函数`来打印，需要侵入到每个dao层的具体代码

**修改后** - 通过插件注册到组件中，**无需侵入到具体实现的代码**

**无侵入地实现自定义功能**，这个特性对每个工具组件都非常重要，GORM这里就提供了一个很好的实现思路 - 注册插件，自定义hook。

## 总结

本次迭代的意义很大 - **标志着`gormer`这个组件实现了自定义方法的可扩展**（ext文件）。

接下来，我们还会持续地对`gormer`等low code工具持续优化，实现更多的功能。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

