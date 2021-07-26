---
title: Go语言学习路线 - 5.基础篇:从一个web项目来谈Go语言的技能点
date: 2021-05-13 12:00:00
categories: 
- 成长分享
tags:
- Go-Study
---

![Go-Study](https://i.loli.net/2021/02/28/BnVH86E5owhsaFd.jpg)



## 从一个Web项目开始

经过了 **入门篇** 的学习，大家已经初步了解Go语言的语法，也能写常见的代码了。接下来，我们就从一个Web项目入手，看看一些常见的技能与知识吧。

我们先简单地聊一下这个Web项目的背景：我们要做的是一个简单的web系统 ，有前端同学负责界面的开发，后端不会考虑高并发等复杂情况。

我们先从一个Web请求出发，看看会涉及到哪些模块。

<!-- more -->

## 前端的请求生命周期

用户在web界面上点击了一个按钮，就由前端发起了一个请求。那这个请求的生命周期是怎么样的呢？

通常情况下，后端的工作是**解析前端的数据，处理对应的业务逻辑，返回操作结果**。

这里，离不开三层概念：

- API层：解析来自前端的数据，转化成go的数据结构
- Service层：包含业务逻辑，是这个请求具体要做的事情
- Dao层：数据持久化，也就是更新到数据库等，保证不丢失

> 不同框架有不同的命名方式，但我个人建议只关注这三层即可。

当然，这三层逻辑并不绝对，会因为项目特点有所调整，但整体的**分层思路**是不会变化的。我认为，如果你能真正地理解web的分层，对项目的框架就能掌握得很棒了。

接下来，我们自顶向下逐层聊聊。



## 第一层：API层

通常来说，API层只做三件事：

1. **根据路由规则，调用具体的处理函数** ，常见的RESTful就是由`URL`+`Method`的作为路由规则；
2. **解析文本或二进制数据到Go结构体**，常见的是用`json`反序列化；
3. **调用下一层Service的函数**

抛开第三点暂且不谈，前两者比较容易理解，大家可以使用标准库里的`net/http`和`encoding/json`来完成。具体的代码我就不写了，网上示例非常多。

那么，API层这么简单，有什么学问嘛？这里，我建议大家看看两个开源库：

- [Gin](https://github.com/gin-gonic/gin)
- [Mux](https://github.com/gorilla/mux)

看看上面的示例，对比一下原生的`net/http`库写出来的代码，是否感觉可读性大大提高？没错，API层关键点之一的就是**可读性**。

不过Gin相对于Mux非常重量级，学习起来成本很大；而Mux虽然可读性提高，但在解析`http body`数据这块效果不佳，还是需要逐个手写结构体。

> 所以，在我看来，这两个都并不是最佳方案，我非常建议有条件的项目能够直接引入 **RPC级别的解决方案**，例如gRPC。这块我会拿具体项目、花好几讲来好好说说。

在开发的过程中，我对API层的开发会重点关注这几点：

- 可读性：可以快速地根据命名了解功能，如**RESTful**
- 高度复用：如引入`mux` 中的各种 middleware，比如 **防止panic** 、**用户认证** 、日志打印等
- 尽量薄：不做或少做业务逻辑处理，复杂处理都丢到service层
- 文档化：将接口的相关参数通过文档给到前端或第三方，尽量做到自动化或半自动化

我再强调一下API层的重要性：**API层是程序最关键的入口和出口，能很好地追踪到数据的前后变化情况。** 一个优秀的API层实现，不仅能让我们少写很多重复性代码，也能大幅度地降低我们排查问题的效率。



## 第二层：Service层

Service层可以理解为服务层，是整个项目中最复杂、也是代码比重往往是最多的。它是一个项目最核心的业务价值所在。

Service是最灵活、也是最考验设计能力的，虽说**没有一套固定的模式**，但还是会有一定的**套路**。

我分享一下个人的三个见解：

1. 单元测试覆盖率要尽量高，这是一个**高频迭代与重构**的模块，也是最容易出现问题的部分；
2. 深入实践 **面向对象与DDD** ，最锻炼工程师抽象、解耦等能力的模块；
3. 选择合适的 **设计模式** 可大幅度地提升研发效率；

再提一句，请跃跃欲试的各位冷静一下，**Service层是和业务一起成长的**，前期没必要过度设计。我们把重点放在**单元测试**的编写上即可，适当地选用一些库来提高效率，如开源的`stretchr/testify`，内部的`reflect`等。



## 第三层：Dao层

Dao层常被理解为数据持久化层，但我们可以将它进行一定的延伸：**将RPC调用也当做Dao层**（不妨认为将数据持久化到了另一个服务），来适配微服务架构的场景。

> 严格意义上，RPC调用和普通的Dao差异有不少，但为了收敛话题，我们暂且不细分。

今天，我们不关注分布式场景下的各种数据问题，也不考虑各种存储中间件的特点，而是聚焦于一个问题：**如何将内存中的对象持久化到数据库中**。在编程领域，这部分的工具被称为**ORM**。

以Go语言对接MySQL为例，最常见的为[gorm](https://github.com/go-gorm/gorm)，它能很便捷地将一个Go语言中的结构体，映射到MySQL数据库某个表中的一行数据。

> 请自行对比一下，用go官方的sql库写增删改查，与用gorm写增删改查的工作量差异。

关于Dao层，我认为有部分的实践是比较通用的：

1. **选用官方或社区高频使用的库**，避免后期出现功能缺失或性能瓶颈的问题；
2. **灵活性比易用性更重要**，通过一层浅封装，往往能更适配项目，达到更棒的易用性；
3. **关注数据库的原理、而不是ORM工具的实现方式**，数据库的原理是长期的积累，对技术选型和排查故障很有帮助。

> 至于不同的数据库ORM有不同的最佳实践，一一列举的工作量太大，我会在工程化的过程中选择性地讲解。



## 串联三层

到这里，我们对这三层有了初步的了解，可以总结为**两边薄（API、Dao），中间厚（Service)**。

这里的实践需要大家不断打磨，比如说：

- API与Dao会随着个人编程能力的提升，不断地总结出更好的编程实践；
- 做性能优化时，优先考虑Dao，其次考虑API，这两部分的提效是最明显的；
- 排查问题时，先分析API的出入口，再分析Dao的出入口，实在解决不了再去看Service（此时已经是严重的业务逻辑问题了）；

到最后，相信大家对这三层认知会进一步提升：

- API：服务对外的门面，通过一个接口定义就能了解大致实现原理；
- Service：复杂的业务逻辑，非该服务的核心成员无需关注，而核心成员须重点维护；
- Dao：无论是调用**ORM**还是**SDK**，都视为一种**工具集**，是一个技术人员沉淀通用能力的重点。



## CRUD程序员

很多程序员都戏称自己是一个只会**CRUD**的码农。让我们换个视角，看看CRUD背后有没有一些的技术点。

- API层：遵循**RESTful**的原则，提高可读性（最好能在一行代码中看到，如`mux`）
  - 将操作（CRUD）对应到HTTP的Method
  - 将资源对象对应到HTTP的URL
- Service层：
  - 对于只是简单的修改，Service不用做复杂处理，透传到Dao层即可
  - 如果涉及到多个表的修改，进行事务处理（如mysql的transaction）
  - 在Dao层出现错误时，适当封装错误信息，提高可读性
- Dao层：
  - 选择并熟练运用ORM，快速实现基本的CRUD
  - 对复杂的ORM进行一层浅封装，方便Service层的调用

经过一段时间的磨练，CRUD的工作能大大提效，我们就能抽出更多的时间去学习其余技能了。



## 结束语

Web项目是我们日常开发最常见的项目类型，也是很多面试考察点的基点。

我建议大家从**分层**着手，明确各层职责，**关注API与Dao层的提效工作，做好Service层的质量保障**，更好地掌控全局。而在具体的开源库的使用过程中，**选对比会用更重要**，集中在**API与Dao层**。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)
