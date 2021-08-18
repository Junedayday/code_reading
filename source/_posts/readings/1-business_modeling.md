---
title: 阅读摘要1 - 业务建模
categories: 
- 成长分享
tags:
- Digest
---

![]()

## 资料推荐

今天推荐的是一个极客时间的专栏：[如何落地业务建模](https://time.geekbang.org/column/article/386594)。

由于示例中的代码和思路都是基于`JAVA`编程语言的，而为了方便`Go`语言的同学理解，我作一下转述，谈谈个人的心得。



## 业务建模的两个关键步骤

1. 定义业务问题
   1. 清晰定义业务问题
   2. 同步给周围人并都接受
2. 解决问题
   1. 结合实际架构实现模型
   2. 建立沟通与反馈渠道



业务问题（业务人员） -> 数据结构（开发人员）

业务问题（业务人员） <-> 领域模型（统一语言） <-> 数据结构（开发人员）



## 示例1：用户User订阅专栏Subscription

### 常规做法

```go
// User 用户相关信息
type User interface {
	// 查询用户订阅的专栏
	// pn:page number 页码
	// ps:page size 每页大小
	QuerySubscription(pn, ps int) ([]Subscription, error)
}

// Subscription 订阅的专栏信息
type Subscription interface {
}
```

这个做法的实现很容易猜到

- `User` 和 `Subscription` 都对应一张表
-  `Subscription`  中有个 `User`  的信息，用于查询用户的订阅



## 问题1：不能体现Subscription与User的关系

从上面可以看到，`SubScription`中有`User`的信息。从业务特点上来分析，只有用户订阅了专栏，才成立一个具体的`Subscription`对象。

> 注意，这里我们讨论的Subscription不是指具体的一门专栏，那它确实不具备用户属性。
>
> 我们讨论的是订阅场景下的专栏，那么这个专栏肯定是有被订阅的对象的，如果有n个对象订阅了这个专栏，就有n条数据。





> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

