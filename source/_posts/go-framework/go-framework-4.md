---
title: Go语言微服务框架 - 4.初识GORM库
categories: 
- 技术框架
tags:
- Go-Framework
---

![Go-Framework](https://i.loli.net/2021/08/15/QfmqMJGaNOgt7LC.jpg)

数据持久化是服务的必要特性，最常见的组件就是关系型数据库`MySQL`。而在`Go`语言里，`GORM`已经成了对接`MySQL`事实上的标准，那么也就不去横向对比其它库了。

`GORM`库是一个很强大、但同时也是一个非常复杂的工具。为了支持复杂的`SQL`语言，它比之前的配置文件加载工具`github.com/spf13/viper`要复杂不少。今天，我们不会全量地引入`GORM`里的所有特性，而是从一个最简单的场景入手，对它的基本特性有所了解。而后续随着框架的完善，我们会逐渐细化功能。

<!-- more -->

## v0.4.0：引入GORM库

项目链接 https://github.com/Junedayday/micro_web_service/tree/v0.4.0

### 目标

利用GORM实现简单的增删改查功能。

### 关键技术点

1. MySQL工具库的必要功能
2. GORM官方示例分析
3. 使用GORM的思考

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
	   |-- dao                             新增：Data Access Object层
	      |-- order.go                         新增：示例的一个Order对象，订单表
	   |-- mysql                           新增：MySQL连接
	      |-- init.go                          新增：初始化连接到MySQL的工作
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

## 1.MySQL工具库的必要功能

对于`MySQL`数据库来说，我们对它的日常操作其实就关注在CRUD上（也就是增删改查）。

除此以外，还有一些是我们需要关注的点：

- **便捷性**：能快速、方便地实现实现功能，而不用写大量重复性的`SQL`
- **透明性**：ORM经过层层封装，最终与MySQL交互的`SQL`语句可供排查问题
- **扩展性**：支持原生的`SQL`，在复杂场景下的ORM框架不如原始的`SQL`好用

这里，我们先聚焦于第一点，后面两块`GORM`框架是支持的。

## 2.GORM官方示例分析

接下来，我们对照着官方文档，来看看有什么样的注意点。

### 创建

中文文档链接 - https://gorm.io/zh_CN/docs/create.html

```go
// 推荐使用方式：定义一个结构体，填充字段
user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
result := db.Create(&user)

// 不推荐：指定要创建的字段名，也就是user中部分生效，很容易产生迷惑
// 更建议新建一个user结构体进行创建
db.Select("Name", "Age", "CreatedAt").Create(&user)

// 批量创建同推荐
var users = []User{{Name: "jinzhu1"}, {Name: "jinzhu2"}, {Name: "jinzhu3"}}
db.Create(&users)

// 不推荐：钩子相关的特性，类似于数据库里的trigger，隐蔽而迷惑，不易维护
func (u *User) BeforeCreate(tx *gorm.DB) (err error){}

// 不推荐：用Map硬编码创建记录，改动成本大
db.Model(&User{}).Create(map[string]interface{}{
  "Name": "jinzhu", "Age": 18,
})

// 争议点：gorm.Model中预定了数据库中的四个字段，是否应该把它引入到模型的定义中
// 我个人不太喜欢将这四个字段强定义为数据库表中的字段名
type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt DeletedAt `gorm:"index"`
}
```

从上面的一些操作可以看到，我推荐的使用方式有2个特点：

1. 尽可能简单，不要出现**魔法变量**，比如常量字符串
2. **不要让框架强约束表结构的设计**，也是为了后续迁移框架、甚至语言时成本更低

### 查询

中文文档链接 - https://gorm.io/zh_CN/docs/query.html

```go
// 不推荐： 我个人不太建议使用First/Last这种和原生SQL定义不一致的语法，扩展性也不好
// 在这种情况下，我更建议采用Find+Order+Limit这样的组合方式，通用性也更强
db.First(&user)
db.Last(&user)

// 推荐：Find支持返回多个记录，是最常用的方法，但需要结合一定的限制
result := db.Find(&users)

// 不推荐：条件查询的字段名采用hard code，体验不好
db.Where("name = ?", "jinzhu").First(&user)
db.Where(map[string]interface{}{"name": "jinzhu", "age": 20}).Find(&users)

// 推荐：结合结构体的方式定义，体验会好很多
// 但是，上面这种方法不支持结构体中Field为默认值的情况，如0，''，false等
// 所以，更推荐采用下面这种方式，虽然会带来一定的hard code，但能指定要查询的结构体名称。
db.Where(&User{Name: "jinzhu", Age: 20}).First(&user)
db.Where(&User{Name: "jinzhu"}, "name", "Age").Find(&users)

// 推荐：指定排序
db.Order("age desc, name").Find(&users)

// 推荐：限制查询范围，结合Find
db.Limit(10).Offset(5).Find(&users)
```

查询还有很多的特性，今天我们暂不细讲。其中，希望大家能重点看一下默认值问题：

我们固然可以通过在定义字段时，排除这些默认值的情况，如定义`int`类型字段时跳过0、从1开始。但在实际的项目中，定义往往很难控制，我们不得不做一定的妥协，这部分hard code的成本也是可以接受的。

**我们不能因为框架里的一些特性，过度地限制其余组件的使用**。

### 更新

中文文档链接 - https://gorm.io/zh_CN/docs/update.html

更新其实是最麻烦的事情，它包括**更新的字段与条件**。我们来看看几个重点的。

```go
// 不推荐：单字段的更新，不常用
db.Model(&User{}).Where("active = ?", true).Update("name", "hello")

// 不推荐：指定主键的多字段更新，但不支持默认类型
db.Model(&user).Updates(User{Name: "hello", Age: 18, Active: false})

// 不推荐：指定主键的多字段的更新，但字段多了硬编码很麻烦
db.Model(&user).Updates(map[string]interface{}{"name": "hello", "age": 18, "active": false})

// 推荐：指定主键的多字段的更新，指定要更新的字段，*为全字段
db.Model(&user).Select("Name", "Age").Updates(User{Name: "new_name", Age: 0})
db.Model(&user).Select("*").Update(User{Name: "jinzhu", Role: "admin", Age: 0})

// 推荐：指定更新方式的多字段的更新
db.Model(User{}).Where("role = ?", "admin").Updates(User{Name: "hello", Age: 18})
```

### 删除

中文文档链接 - https://gorm.io/zh_CN/docs/delete.html

删除我不太建议使用，更推荐用软删除的方式，也就是**更新一个标识是否已经删除字段**。

```go
db.Delete(&email)
```

## 3.使用GORM的思考

`GORM`是一个非常重量级的工具，尤其是`*gorm.DB`提供了大量的类似于Builder模式的方法，用来拼接`SQL`。

整个使用过程，对于一个不熟悉`SQL`语言的同学来说是很痛苦的，需要大量调试问题的时间；而对于熟悉`SQL`的朋友也会很疑惑，例如`First`等这种自定义命名的底层实现。所以，基于`GORM`库做一个简单封装是非常必要的，能大幅度地降低用户的使用和理解的门槛，也是这个微服务框架后续的改善方向之一。

## 对微服务框架的延伸思考

从之前的分析可以看到，我对微服务的框架有一个很重要的要求 - **透明**，比如不要引入大量和原始SQL无关的特性、不要过度依赖ORM而忘记了原生SQL才是我们最重要的技能。

**透明**也是一个框架能实现简单性的重要特质，减少使用方的理解成本，也就能提高研发效能。

从更高的层面来看整个微服务框架，我们会有更深的体会：

1. **为什么Spring Boot那么成功？**主要是Spring Boot的设计理念是比较符合工程化的，而JVM也提供了一套很好的运行时的机制；与此同时，社区提供了大量的Spring Boot组件供开发者调用，自然比较受欢迎。
2. **Go的微服务框架为什么没有统一？**Go的运行时非常轻量级，很难巧妙地像Spring Boot完成框架层面对组件的大一统。Go语言提供的各类组件，很多都是开源社区对传统服务或云原生理念的自我实践，没有绝对的正确与错误。
3. **那如今社区上的那些微服务框架都不值一提吗？**并不是。如果你仔细看这些框架，其实都是对各类Go优秀组件的拼装，只是各有各的想法。我觉得，所谓的Go微服务框架短期内很难统一，但这些组件都会趋于一致。
4. **那你做这个框架的意义是什么呢？**其实我个人并不觉得本框架比现有框架好，我更关注两点：一是分享引入并迭代各个开源组件的过程，让大家更好地理解框架是怎么完善的；第二个是从工程化的角度去思考微服务框架的问题，从会用框架变得理解框架、并改造框架。

## 总结

我们简单地引入了`GORM`并实现了一套简单的增删改查的代码，更多地是讨论一些技术选型的思考，希望能给大家带来启发。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

