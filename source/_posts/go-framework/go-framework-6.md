---
title: Go语言微服务框架 - 6.用Google风格的API接口打通MySQL操作
date: 2021-09-19 12:00:00
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

随着RPC与MySQL的打通，整个框架已经开始打通了数据的出入口。

接下来，我们就尝试着实现通过RPC请求操作MySQL数据库，打通整个链路，真正地让这个平台实现可用。

<!-- more -->

## v0.5.0：用Google风格的API接口打通MySQL操作

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.5.0

### 目标

从API出发，实现一个数据库表的增删改查。

### 关键技术点

1. Google风格的API定义
2. model与dao的定义
3. service层的实现

> 注意，最近buf工具进行了一次不兼容的升级，从v1beta升级到了v1，可通过如下链接下载 https://github.com/bufbuild/buf/releases

### 目录构造

```
--- micro_web_service            项目目录
	|-- gen                            从idl文件夹中生成的文件，不可手动修改
	   |-- idl                             对应idl文件夹
	      |-- demo                             对应idl/demo服务
	         |-- demo.pb.go                        demo.proto的基础结构
	         |-- demo.pb.gw.go                     demo.proto的HTTP接口，对应gRPC-Gateway
	         |-- demo_grpc.pb.go                   demo.proto的gRPC接口代码
	    	|-- order                            新增：对应idl/order服务
	         |-- order.pb.go                       新增：order.proto的基础结构
	         |-- order.pb.gw.go                    新增：order.proto的HTTP接口，对应gRPC-Gateway
	         |-- order_grpc.pb.go                  新增：order.proto的gRPC接口代码
	|-- idl                            原始的idl定义
	   |-- demo                            业务package定义
	      |-- demo.proto                       protobuffer的原始定义
	   |-- order                           新增：业务order定义
	      |-- order.proto                      新增：protobuffer的原始定义
	|-- internal                       项目的内部代码，不对外暴露
	   |-- config                          配置相关的文件夹
	      |-- viper.go                         viper的相关加载逻辑
	   |-- dao                             Data Access Object层
	      |-- order.go                         更新：Order对象，订单表，实现model层的OrderRepository
	      |-- order_test.go                    Order的单元测试
	   |-- model                           新增：model层，定义对象的接口方法
	      |-- order.go                         新增：OrderRepository接口，具体实现在dao层
	   |-- mysql                           MySQL连接
	      |-- init.go                          初始化连接到MySQL的工作
	   |-- server                          服务器的实现
	      |-- demo.go                          server中对demo这个服务的接口实现
	      |-- server.go                        server的定义，须实现对应服务的方法
	   |-- service                         新增：service层，作为领域实现的核心部分
	      |-- order.go                         新增：Order相关的服务，目前仅简单的CRUD
     |-- zlog                            封装日志的文件夹
        |-- zap.go                           zap封装的代码实现
	|-- buf.gen.yaml                   更新：buf生成代码的定义，从v1beta升到v1
	|-- buf.yaml                       更新：buf工具安装所需的工具，从v1beta升到v1
	|-- gen.sh                         buf生成的shell脚本
	|-- go.mod                         Go Module文件
	|-- main.go                        项目启动的main函数
```

## 1.Google风格的API定义

由于整体的定义比较多，这里就以

```protobuf
message CreateOrderRequest {
  Order order = 1;
}

message UpdateOrderRequest {
  Order order = 1;
  google.protobuf.FieldMask update_mask = 2;
}

message GetOrderRequest {
  string name = 1;
}

// Order服务
service OrderService {
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse) {
    option (google.api.http) = {
      get: "/v1/orders"
    };
  }

  // 这里body中的order表示HTTP的body里的数据填充到CreateOrderRequest结构中的order对象
  rpc CreateOrder(CreateOrderRequest) returns (Order) {
    option (google.api.http) = {
      post: "/v1/orders"
      body: "order"
    };
  }

  rpc UpdateOrder(UpdateOrderRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      patch: "/v1/orders"
      body: "*"
    };
  }

  // 这里{name=*}表示这个字段填充到GetOrderRequest里的name字段
  rpc GetOrder(GetOrderRequest) returns (Order) {
    option (google.api.http) = {
      get: "/v1/orders/{name=*}"
    };
  }

  rpc DeleteBook(DeleteBookRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/v1/books"
    };
  }
}
```

这里，我们重点关注以下几个方法：

1. List - 查询列表，对应HTTP的GET方法
2. Get - 查询单个对象，对应HTTP的GET方法
3. Create - 创建对象，对应HTTP的POST方法
4. Update - 更新对象，对应HTTP的PATCH方法
5. Delete - 删除对象，对应HTTP的DELETE方法（本次暂未实现，后续添加软删除时加上）

> 关于Google定义的标准方法细节，可以参考[Google Cloud API链接](https://cloud.google.com/apis/design/standard_methods)，了解对资源、字段等命名的逻辑。
>
> 而对于gRPC-Gateway中对于proto3的语法，可以参考[gRPC-Gateway链接](https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/patch_feature/)。
>
> 以上两块内容比较多，建议边实践边学习，不要一开始就钻细节。

## 2.model与dao的定义

为了将模型的定义与数据库的实现分离，我将两者进行了拆分，分别放置在model与dao目录下，定位的简单介绍如下：

- model，数据模型的定义，更关注对业务层的数据格式统一，底层可以对应各种存储形式，如mysql、redis
- dao，真实数据存储的操作，也就是model层的实现，目前实现了一种mysql的操作

### Model层

重点是统一的数据结构定义`Order`，以及关键接口`OrderRepository`的定义。

```go
// Order针对的是 orders 表中的一行数据
// 在这里定义，是为了分离Model与Dao
type Order struct {
	Id    int64
	Name  string
	Price float32
}

// OrderFields 作为一个 数据库Order对象+fields字段的组合
// fields用来指定Order中的哪些字段生效
type OrderFields struct {
	Order  *Order
	Fields []string
}

type OrderRepository interface {
	AddOrder(order *Order) (err error)
	QueryOrders(pageNumber, pageSize int, condition *OrderFields) (orders []Order, err error)
	UpdateOrder(updated, condition *OrderFields) (err error)
}
```

### Dao层

Dao层代码基本与之前一致，重点关注结构体`OrderRepo`，它是Model层`OrderRepository`的一种MySQL实现。

```go
type OrderRepo struct {
	db *gorm.DB
}

// 将gorm.DB作为一个参数，在初始化时赋值：方便测试时，放一个mock的db
func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}
```

## 3.service层的实现

service是核心业务实现，但目前的示例代码比较简单，基本就是透传CRUD。

```go
// 定义Service的实现，注意orderRepo的定义是model层的interface
type OrderService struct {
	orderRepo model.OrderRepository
}

// 创建对象，注意orderRepo的实现为dao层代码
func NewOrderService() *OrderService {
	return &OrderService{
		orderRepo: dao.NewOrderRepo(mysql.GormDB),
	}
}

// 以List为例，透传查询
func (orderSvc *OrderService) List(ctx context.Context, pageNumber, pageSize int, condition *model.OrderFields) ([]model.Order, error) {
	orders, err := orderSvc.orderRepo.QueryOrders(pageNumber, pageSize, condition)
	if err != nil {
		return nil, errors.Wrapf(err, "OrderService List pageNumber %d pageSize %d", pageNumber, pageSize)
	}
	return orders, nil
}
```

## 4.模拟HTTP接口访问

本服务支持gRPC和HTTP访问，但由于gRPC不方便用工具模拟，我们这里就以HTTP对本服务进行访问

```go
// List
curl --request GET 'http://127.0.0.1:8081/v1/orders'

// Create
curl --request POST 'http://127.0.0.1:8081/v1/orders' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "order1",
    "price": 100.3
}'

// Update 这里 order 表示数据，update_mask表示更新的字段是price
curl --request PATCH 'http://127.0.0.1:8081/v1/orders' \
--header 'Content-Type: application/json' \
--data-raw '{
    "order": {
        "id": "1",
        "name": "order1",
        "price": 110.9
    },
    "update_mask": "price"
}'

// Get 查询name=order1的对象
curl --request GET 'http://127.0.0.1:8081/v1/orders/order1'
```

## 关于Google风格的API总结

Google风格的API和目前的主流RESTful标准的API有很多相似点、也存在一定的区别。

我们没有必要去抠API风格的细节实现、一定要与Google风格完全一致。API接口是一个通用协议，不同团队有自己的理解，就像RESTful标准的细节实现都有差异。

作为对外协议，最重要的是可读性，每个人都可以根据实际项目情况，对接口风格做一些适配性调整。这里介绍Google风格，主要是为了扩展大家的视野、拥有更多的技术实现方式。

## 总结

通过这个版本，我们打通了API接口到MySQL数据库操作的全流程，是对整个框架的一次初步整合。接下来，我们会对这一流程进行精雕细琢，使其更具通用性和易用性。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

