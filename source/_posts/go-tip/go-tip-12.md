---
title: Go语言技巧 - 12.【Go实体框架】Facebook开源ent概览
date: 2021-12-31 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## Ent概览

**Simple, yet powerful ORM for modeling and querying data.**

`Ent`作为一款由`Facebook`开源的库，官方定义为`An entity framework for Go`。从整个微服务框架来看，它更准确的定位应是 **数据模型层的工具库**。了解`Ent`这款企业级工具的大致实现，不仅有助于我们在技术选型时拓宽视野，也能帮助我们能更好地认识数据模型层。

<!-- more -->

## 三大特性

### Schema As Code

> Simple API for modeling any database schema as Go objects. 

从定义来看这个特性非常棒 - `Ent` 可以将各种异构数据库映射到Go语言的结构体。

但在实际的开发中，如果你对各类数据库有深入的理解，就会清楚地知道这个特性在对数据库特性有一定要求时，框架层面就很难满足了。

### Easily Traverse Any Graph

> Run queries, aggregations and traverse any graph structure easily.

强调对图结构的 **查询、聚合和遍历**。这里的图数据库和传统的关系型数据库差别不小，有兴趣的朋友可搜索**图数据库**的相关概念。

### Statically Typed And Explicit API

> 100% statically typed and explicit API using code generation.

利用代码生成的能力，保证静态类型和显示声明的API。

### 特性总结

三大特性，分别从 **支持的数据库能力集**、**针对图形数据处理能力** 和 **代码生成的输出形式**，描述了`Ent`框架的优点。

这里，我会更聚焦于第二点中的关键词：**图形数据**。让我们带着对三个特性的初印象，开始了解相关官方示例。

## Ent实践

> Ent工具的使用方式并不是本篇的重点，具体的操作方法我会放在链接里，文中只给出关键性的内容

### 1.创建实体

链接 - https://entgo.io/docs/getting-started/#create-your-first-entity 

```go
u, err := client.User.
        Create().
        SetAge(30).
        SetName("a8m").
        Save(ctx)
```

代码和`GORM`非常类似，但不支持复杂结构体的传入，面对大量参数时比较麻烦。

### 2.查询实体

链接 - https://entgo.io/docs/getting-started/#query-your-entities

```go
u, err := client.User.
        Query().
        Where(user.Name("a8m")).
        // `Only` fails if no user found,
        // or more than 1 user returned.
        Only(ctx)
```

基本同上，表达方式还是很明确的。但对于`Only`这种新引入的关键词，对新人来说有学习成本。

### 3.Edge相关

- https://entgo.io/docs/getting-started/#add-your-first-edge-relation
- https://entgo.io/docs/getting-started/#add-your-first-inverse-edge-backref
- https://entgo.io/docs/getting-started/#create-your-second-edge

我们以一个复杂Edge为例：

```go
cars, err := client.Group.
        Query().
        Where(group.Name("GitHub")). // (Group(Name=GitHub),)
        QueryUsers().                // (User(Name=Ariel, Age=30),)
        QueryCars().                 // (Car(Model=Tesla, RegisteredAt=<Time>), Car(Model=Mazda, RegisteredAt=<Time>),)
        All(ctx)
```

从表达式上来看，就是查询Group、然后关联查询User、最后再查到Car。

首先，我们要认识到 - **抛开背后的实现，这种表达方式很简洁**。

如果底层是`MySQL`，这里至少关联了三张实体表（JOIN），很容易引起性能问题。这个问题也就是上面所说的，**框架屏蔽了异构数据库**而导致的。

## 参考资料

Github - https://github.com/ent/ent

官网 - https://entgo.io/

## 思考

通过相关资料和简单实操，我对于`Ent`框架的定位是 - **一个面向图数据库的ORM框架**。相信随着图数据库的逐渐成熟，`Ent`会更具价值。但考虑到以下两点：

1. 图数据库的成熟周期还需要一段时间，当前的维护成本高；
2. 在非图数据库上使用`Ent`，对开发者的要求很高，既要了解`Ent`对不同数据库的底层实现，又要懂数据库原理。

> 举个例子，ent的部分Edge特性需要依赖数据库的外键，但如今主流数据库的实践，倡导去外键，而是将相关逻辑转移到程序代码里。

所以，我不建议将`Ent`引入到项目中。关于`Ent`更多的细节需要大家自行阅读和实践。

这里，我抛出一个自己的理解：**从编程语言框架层面，不应过度基础设施的复杂度。从异构数据库来说，它们的特性、维护方式、设计模式都各不相同，应寻找每种数据库对应的工具库，而不应期望毕其功于一役。**

换一句话，如果期望一个工具库能适配十种数据库，那么换一种角度，这十种数据库更应该被封装成一种数据库。**通用性如果能沉淀在基础设施上，价值远大于在工具库上做适配。**

## 小结

`Ent`能在Facebook等公司与Kratos框架上沉淀，证明了它具备实际工程落地的能力，但对使用者的要求很高，很难具备普适性。

前文为了表达个人想法，我在论述观点时会相对态度鲜明，但并非对`Ent`持有否定态度。相反地，从具体的实现细节来看，`Ent`给了我不少启发，尤其是强调静态类型，能看出它对性能的追求。

> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
> ![二维码](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/my_wechat.jpeg)
>
> 

