---
title: Go语言技巧 - 14.【浅析微服务框架】go-zero概览
date: 2022-02-02 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## go-zero概况

`go-zero`是当前处于CNCF孵化中的一个`Go`z语言框架项目，在Github上的star数目前达到14.3K。

作为一款起源于国内的项目，`go-zero`的中文资料比较齐全，对国内开发者相对友好。但前景如何，还需要进一步的观察。今天我们一起来了解这个项目。

<!-- more -->

## 概览

官方文档 - https://go-zero.dev/cn/

Github - https://github.com/zeromicro/zero-doc

> go-zero is a web and rpc framework written in Go. It's born to ensure the stability of the busy sites with resilient design. Builtin goctl greatly improves the development productivity.

官方核心将自己定位为一个 **Go语言的web和rpc框架**。其余描述内容的意义不大，如稳定的、可伸缩的，更多依赖的是Paas平台与程序自身的设计。

## 具体实例

有了前面三个框架的基础，了解`go-zero`会相对容易。这次，我将换一个思路讲解，先从官方的示例出发，再回过头来看看这个框架的核心思想。

以下内容，来自`go-zero`提供的中文示例 - https://go-zero.dev/cn/shorturl.html

### 准备工具

主要安装的二进制工具包括以下三个：

- protoc-gen-go
- protoc
- goctl

其中第三个是`go-zero`自研的。不难看出，`go-zero`是强依赖protobuffer生态的。

### API Gateway代码

```go
type (
  shortenReq {
    url string `form:"url"`
  }

  shortenResp {
    shorten string `json:"shorten"`
  }
)

service shorturl-api {
  @server(
    handler: ShortenHandler
  )
  get /shorten(shortenReq) returns(shortenResp)
}
```

这是一套 `go-zero` 特定的语法。虽说这个语法阅读起来很容易理解，里面有Go语言和protobuffer的影子，但就是一个完全独立的一套方案。

值得注意的是，我们如果要在这个语法中引入各类网关层的特性，如限流参数等，会导致这个语法的学习成本越来越高。

### rpc服务代码

核心为一个protobuffer的定义，如下：

```protobuf
syntax = "proto3";

package transform;

message shortenReq {
    string url = 1;
}

message shortenResp {
    string shorten = 1;
}

service transformer {
    rpc shorten(shortenReq) returns(shortenResp);
}
```

接下来工具会生成相应的代码，以及会有相关的配置文件可供修改。

### 关联API Gateway和rpc服务

由于API Gateway与rpc服务是两个独立的进程，所以需要修改对应的配置（端口等信息），进行打通。

核心代码如下：

```go
func (l *ShortenLogic) Shorten(req types.ShortenReq) (types.ShortenResp, error) {
  // 手动代码开始
    resp, err := l.svcCtx.Transformer.Shorten(l.ctx, &transformer.ShortenReq{
        Url: req.Url,
    })
    if err != nil {
      return types.ShortenResp{}, err
    }

    return types.ShortenResp{
      Shorten: resp.Shorten,
    }, nil
    // 手动代码结束
}
```

即新建出一个rpc服务的对象`Transformer`，调用对应的方法`Shorten`。

### rpc服务最终调用到model层

关于model层的生成，也就是对`MySQL`的CRUD方案，我这里就不专门写了。

我们关注一下最终调用到`MySQL`层的代码：

```go
func (l *ShortenLogic) Shorten(in *transform.ShortenReq) (*transform.ShortenResp, error) {
  // 手动代码开始，生成短链接
  key := hash.Md5Hex([]byte(in.Url))[:6]
  _, err := l.svcCtx.Model.Insert(model.Shorturl{
    Shorten: key,
    Url:     in.Url,
  })
  if err != nil {
    return nil, err
  }

  return &transform.ShortenResp{
    Shorten: key,
  }, nil
  // 手动代码结束
}
```

## 架构概览

有了这么一个实例，我们再回过头来看看整体的架构：

![go-zero](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/architecture-en.png)

`go-zero`的架构图包含了五层，但它的核心聚焦于API Gateway与Services这两层。

- Clients层：对接的是API Gateway上的HTTP服务，要考虑能否自动生成跨语言的SDK
- Cache+DB：关注高频+重复性高CRUD代码的自动生成功能，抽象成model
- API Gateway：这部分功能更像是Go语言+Service Mesh的一个结合方案
- Services：业务代码的具体实现，提供了很多常用的工具集

从框架分层来看，我个人是不太认同这种划分方式。这张图中的架构，更多地是为了体现出`go-zero`的两层结构而产出的架构图。尤其是API Gateway这个设计，表现出了团队在Service Mesh上能力的不足、而引入的功能。

## Go语言微服务框架的聚焦点

目前为止，我们已经一起看了四个不同的微服务框架，也许有同学会觉得我总是在到处挑刺，那么我理想中的微服务框架是怎么样的呢？我来谈谈：

### 三个分层

- 控制层controller - 处理数据格式的转换，做一些panic处理等通用性较强的工作
- 业务层service - 聚焦于service层的业务逻辑代码编写
- 数据存储层model - 通用性强的数据存储，对接MySQL、Redis等存储

### 五个聚焦点

1. **控制层以上** - 即请求是怎么进入微服务的，不应该由微服务框架关心，而应交由Paas平台层的产品，尤其是Kubernetes和Service Mesh；
2. **控制层** - 以protobuffer定义+gRPC生态为核心，自动生成代码框架，在对应的server层提供大量通用的middleware处理panic、context、logging等能力；
3. **业务层** - 业务层应高度关注代码的可测试性，也就是单元测试尽可能在这里闭环，这就需要下层的Mock能力+DI的代码风格；
4. **数据存储层** - 数据存储层必须结合code generation实现高度的自动化，尽可能地引用主流库、而非自研，而且有可能的话，后续可以提供多种库的切换方案，如MySQL用原生库/gorm/ent/sqlx等方案；
5. **数据存储层以下** - 如怎么对接分布式的数据库、或者怎么对接其余的微服务，这一类的问题，不应由微服务框架自行解决，而应由底层服务提供SDK或库（如服务发现能力），结合到微服务框架里

这么聊下来，可能还是有点抽象，我再结合两个例子谈谈这块：

1. 服务熔断功能
   1. 在Service Mesh层配置熔断条件(如错误码和错误次数)，在达到具体条件后实现熔断，阻断后续的请求到微服务
   2. Go微服务框架应保证按照Service Mesh层地定义的协议格式返回错误码；
2. 对接分布式服务（Client-Client模式）
   1. 在Paas层提供服务发现的SDK，包括两块功能：获取目标服务的地址列表与多种负载均衡策略
   2. Go微服务框架引入这个SDK，填写目标服务名称+负载均衡策略，SDK选择一个最合适的节点并进行请请求，而如何请求由微服务框架之间的通信协议决定，如HTTP或gRPC；

### 我心中Go框架的核心价值

1. Controller层 - 利用`gRPC`的生态生成具体的代码，充分利用middleware(拦截器)的特性实现panic recovery+logging+metrcis等通用的能力；
2. Model层是体现自动化的最核心模块，必须要充分利用代码生成的技术，体现出两个价值：
   1. 能自动化地实现Mocking，为上层Service的单元测试提供基础保障
   2. 降低用户使用成本，尽可能避免学习一套新的语法

> 最后一点有部分人会不好理解，以MySQL ORM框架为例，它们在函数中提供的查询方法名是Find/Query/Search等，但对应到SQL的关键词是SELECT，这就对熟悉MySQL的同学来说需要一种思维转变，有一定的成本。更好的方式，那就是ORM里的函数直接使用SELECT这种关键词。

Controller这层的开发量很有限，基本已经由社区提供能力，更应该关注如何能形成一套标准规范；而框架更多的精力应关注在Model这层，它需要大量的、易用的SDK积累，能便捷地对接各类中间件，才能体现出框架的价值。

## 小结

总体来说，我觉得`go-zero`是一个早期云原生环境下的框架，在目前Service Mesh大规模落地后，有大量的特性造成了冲突。从研发体验、维护成本和稳定性角度来看，这部分功能更应该交给跨语言的Service Mesh来解决，而不是编程语言强相关的框架。目前，国内互联网大公司基本不会采用`go-zero`，这点占据很大因素。

不过，`go-zero`在Model层做了一定的自动化，是一个很值得学习的能力。作为CNCF的项目之一，`go-zero`后续是否会有大的转变、更加贴合整个CNCF的氛围，我个人还是非常期待的。

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

