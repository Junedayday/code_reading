---
title: Go语言微服务框架 - 5.GORM库的适配sqlmock的单元测试
date: 2021-09-05 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

随着`GORM`库的引入，我们在数据库持久化上已经有了解决方案。但上一篇我们使用的`GORM`过于简单，应用到实际的项目中局限性很大。

与此同时，我们也缺乏一个有效的手段来验证自己编写的相关代码。如果依靠连接到真实的`MySQL`去验证功能，那成本实在太高。那么，这里我们就引入一个经典的`sqlmock`框架，并配合对数据库相关代码的修改，来实现相关代码的可测试性。

<!-- more -->

## v0.4.1：GORM库的适配sqlmock的单元测试

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.4.1

> 由于主要是针对GORM的小改动，所以增加了一个小版本号

### 目标

利用sqlmock工具，并对数据库相关代码进行修改，实现单元测试。

### 关键技术点

1. Order相关代码的改造
2. 引入sqlmock到测试代码
3. 注意点讲解

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
	   |-- config                          配置相关的文件夹
	      |-- viper.go                         viper的相关加载逻辑
	   |-- dao                             Data Access Object层
	      |-- order.go                         更新：OrderO对象，订单表
	      |-- order_test.go                    新增：Order的单元测试
	   |-- mysql                           MySQL连接
	      |-- init.go                          初始化连接到MySQL的工作
	   |-- server                          服务器的实现
	      |-- demo.go                          server中对demo这个服务的接口实现
	      |-- server.go                        server的定义，须实现对应服务的方法
     |-- zlog                            封装日志的文件夹
        |-- zap.go                           zap封装的代码实现
	|-- buf.gen.yaml                   buf生成代码的定义
	|-- buf.yaml                       buf工具安装所需的工具
	|-- gen.sh                         buf生成的shell脚本
	|-- go.mod                         Go Module文件
	|-- main.go                        项目启动的main函数
```

## 1.Order相关代码的改造

我们要对Order相关的代码进行改造，来满足以下两个点：

1. 可测试性，可以脱离对真实数据库连接的依赖
2. 灵活的更新方法，可以支持对指定条件、指定字段的更新

```go
/*
  gorm.io/gorm 指的是gorm V2版本，详细可参考 https://gorm.io/zh_CN/docs/v2_release_note.html
  github.com/jinzhu/gorm 一般指V1版本
*/

type OrderRepo struct {
	db *gorm.DB
}

// 将gorm.DB作为一个参数，在初始化时赋值：方便测试时，放一个mock的db
func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// Order针对的是 orders 表中的一行数据
type Order struct {
	Id    int64
	Name  string
	Price float32
}

// OrderFields 作为一个 数据库Order对象+fields字段的组合
// fields用来指定Order中的哪些字段生效
type OrderFields struct {
	order  *Order
	fields []interface{}
}

func NewOrderFields(order *Order, fields []interface{}) *OrderFields {
	return &OrderFields{
		order:  order,
		fields: fields,
	}
}

func (repo *OrderRepo) AddOrder(order *Order) (err error) {
	err = repo.db.Create(order).Error
	return
}

func (repo *OrderRepo) QueryOrders(pageNumber, pageSize int, condition *OrderFields) (orders []Order, err error) {
	db := repo.db
	// condition非nil的话，追加条件
	if condition != nil {
		// 这里的field指定了order中生效的字段，这些字段会被放在SQL的where条件中
		db = db.Where(condition.order, condition.fields...)
	}
	err = db.
		Limit(pageSize).
		Offset((pageNumber - 1) * pageSize).
		Find(&orders).Error
	return
}

func (repo *OrderRepo) UpdateOrder(updated, condition *OrderFields) (err error) {
	if updated == nil || len(updated.fields) == 0 {
		return errors.New("update must choose certain fields")
	} else if condition == nil {
		return errors.New("update must include where condition")
	}

	err = repo.db.
		Model(&Order{}).
		// 这里的field指定了order中被更新的字段
		Select(updated.fields[0], updated.fields[1:]...).
		// 这里的field指定了被更新的where条件中的字段
		Where(condition.order, condition.fields...).
		Updates(updated.order).
		Error
	return
}
```

## 2.引入sqlmock到测试代码

sqlmock是检查数据库最常用的工具，我们先不管它使用起来的复杂性，先来看看怎么实现对应的测试代码：

```go
// 注意，我们使用的是gorm 2.0，网上很多例子其实是针对1.0的
var (
	DB   *gorm.DB
	mock sqlmock.Sqlmock
)

// TestMain是在当前package下，最先运行的一个函数，常用于初始化
func TestMain(m *testing.M) {
	var (
		db  *sql.DB
		err error
	)

	db, mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	DB, err = gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// m.Run 是真正调用下面各个Test函数的入口
	os.Exit(m.Run())
}

/*
  sqlmock 对语法限制比较大，下面的sql语句必须精确匹配（包括符号和空格）
*/

func TestOrderRepo_AddOrder(t *testing.T) {
	var order = &Order{Name: "order1", Price: 1.1}
	orderRepo := NewOrderRepo(DB)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `orders` (`name`,`price`) VALUES (?,?)").
		WithArgs(order.Name, order.Price).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := orderRepo.AddOrder(order)
	assert.Nil(t, err)
}

func TestOrderRepo_QueryOrders(t *testing.T) {
	var orders = []Order{
		{1, "name1", 1.0},
		{2, "name2", 1.0},
	}
	page, size := 2, 10
	orderRepo := NewOrderRepo(DB)
	condition := NewOrderFields(&Order{Price: 1.0}, []interface{}{"price"})

	mock.ExpectQuery(
		"SELECT * FROM `orders` WHERE `orders`.`price` = ? LIMIT 10 OFFSET 10").
		WithArgs(condition.order.Price).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "price"}).
				AddRow(orders[0].Id, orders[0].Name, orders[0].Price).
				AddRow(orders[1].Id, orders[1].Name, orders[1].Price))

	ret, err := orderRepo.QueryOrders(page, size, condition)
	assert.Nil(t, err)
	assert.Equal(t, orders, ret)
}

func TestOrderRepo_UpdateOrder(t *testing.T) {
	orderRepo := NewOrderRepo(DB)
	// 表示要更新的字段为Order对象中的id,name两个字段
	updated := NewOrderFields(&Order{Id: 1, Name: "test_name"}, []interface{}{"id", "name"})
	// 表示更新的条件为Order对象中的price字段
	condition := NewOrderFields(&Order{Price: 1.0}, []interface{}{"price"})

	mock.ExpectBegin()
	mock.ExpectExec(
		"UPDATE `orders` SET `id`=?,`name`=? WHERE `orders`.`price` = ?").
		WithArgs(updated.order.Id, updated.order.Name, condition.order.Price).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := orderRepo.UpdateOrder(updated, condition)
	assert.Nil(t, err)
}
```

## 3.注意点讲解

虽然添加了注释，我这边依旧讲一下修改的重点：

1. `gorm.DB`作为一个初始化的参数，将其转变成一个依赖注入，使这块代码更具可测试性
2. 查询和更新采用了一个新的结构体`OrderFields`，是用里面的`fields`声明了`order`中哪个字段生效

## GORM框架的进一步扩展

通过这一次对GORM数据库相关代码的迭代，还是可以发现有些不足：

1. 对复杂SQL的支持不足：如group by、子查询等语句
2. 对field这块限制不好，`id`, `name`， `price`，容易发生误填字段的问题
3. 没有串联日志模块

接下来的模块，我会逐渐对2、3两点进行补充，而第1点需要有选择性地实现，我也会结合具体的场景进行分享。

## 总结

通过这一个小版本，我们让`DAO`这个与数据库交互模块的代码更具可读性（从调用侧可以清楚地了解到要做什么）、健壮性（单元测试）和可扩展性（对后续字段的扩展也很容易支持）。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

