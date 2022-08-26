---
title: etcd源码分析 - 1.【打通核心流程】etcd server的启动流程
date: 2022-06-20 12:00:00
categories: 
- 技术框架
tags:
- Go-etcd
---

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd.jpg)

`etcd`的源码相对`Kubernetes`少了很多，但学习成本依旧在。

在第一阶段，我将从主流程出发，讲述一个`PUT`指令是怎么将数据更新到`etcd server`中的。今天，我们先来看看server是怎么启动的。

<!-- more -->

## etcd server启动代码

运行`etcd server`的最简化代码为`./bin/etcd`，无需添加任何参数。我们就根据这个命令来阅读代码，看看启动的主逻辑是怎么样的。

### etcdmain.Main

主入口函数中，只要我们能理解`os.Args`它的含义，就能快速地跳过中间代码，找到下一层函数的入口`startEtcdOrProxyV2()`。

### startEtcdOrProxyV2

本函数较长，就比较考验我们的通读能力。在阅读这一块代码时，我一般会用到三个小技巧：

1. 忽略`err != nil`的判断分支，一般它们都是对异常case的处理；
2. 忽略`变量 == 默认值`的判断分支，如`字符串变量 == ""`，这种多为对默认值的处理，如做变量初始化等；
3. 寻找串联上下文的关键性变量，一般都会有明确的命名或注释；

而在这块代码里呢，我们就能找到2个关键性的变量，以及相关的使用处：

```go
// 表示停止动作与错误的两个channel
var stopped <-chan struct{}
var errc <-chan error

// 两种模式：第一种是正常的etcd server，第二种是代理模式
switch which {
case dirMember:
	stopped, errc, err = startEtcd(&cfg.ec)
case dirProxy:
	err = startProxy(cfg)

// 阻塞并监听两个通道的地方
select {
	case lerr := <-errc:
	case <-stopped:
}
```

通过这部分的代码，我们就能定位到下一层的函数入口 - `startEtcd`。

### startEtcd

```go
func startEtcd(cfg *embed.Config) (<-chan struct{}, <-chan error, error) {
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, nil, err
	}
	osutil.RegisterInterruptHandler(e.Close)
	select {
	case <-e.Server.ReadyNotify(): // wait for e.Server to join the cluster
	case <-e.Server.StopNotify(): // publish aborted from 'ErrStopped'
	}
	return e.Server.StopNotify(), e.Err(), nil
}
```

我们可以从三个关键动作，来了解这个函数的功能：

1. 启动etcd，如果失败则通过`error`返回；
2. 启动etcd后，本节点会加入到整个集群中，就绪后则通过channel`e.Server.ReadyNotify()`收到消息；
3. 启动etcd后，如果遇到异常，则会通过channel`e.Server.StopNotify()`收到消息；

另外，`osutil.RegisterInterruptHandler(e.Close)`这个函数注册了etcd异常退出的函数，里面涉及到一些汇编，有兴趣可以深入阅读。

###  embed.StartEtcd

```go
func  StartEtcd(inCfg *Config) (e *Etcd, err error){}
```

我们先简单地通读一下注释，可以了解到：**本函数返回的Etcd并没有保证加入到集群，而是要等待channel通知**。这就印证了上面的猜想。`StartEtcd`函数很长，我先解释两个关键词：

1. peer - 英文翻译为同等地位的人，在当前语义下表示其余同等的etcd server节点，共同组成集群；
2. client - 即客户端，可以理解为发起etcd请求方，如程序；

我们看到一段代码：

```go
// 新建 etcdserver.EtcdServer 对象
if e.Server, err = etcdserver.NewServer(srvcfg); err != nil {
  return e, err
}

// 启动etcdserver
e.Server.Start()

// 连接peer/client，以及提供metrics指标
if err = e.servePeers(); err != nil {
  return e, err
}
if err = e.serveClients(); err != nil {
  return e, err
}
if err = e.serveMetrics(); err != nil {
  return e, err
}
```

进入Start方法，可以看到里面都是一些常驻的daemon程序，如监控版本/KV值，与我们关注的PUT操作的核心流程无关。所以，我们的目标就转移到`serveClients`函数。

### serveClients

本函数的重点在于下面这段。这里有个变量叫`sctx`，就是server context的简写，是在前面`embed.StartEtcd`里初始化的，主要由context、日志、网络信息组成。

```go
for _, sctx := range e.sctxs {
  go func(s *serveCtx) {
    e.errHandler(s.serve(e.Server, &e.cfg.ClientTLSInfo, h, e.errHandler, gopts...))
  }(sctx)
}
```

重点理解下面这个函数：

```go
func (e *Etcd) errHandler(err error) {
  // 第一次select，如果收到停止消息，则退出，否则到第二个select
	select {
	case <-e.stopc:
		return
	default:
	}
  
  // 第二次select，一般情况下长期阻塞在这里：要么收到停止消息，要么将error从e.errc发送出去
	select {
	case <-e.stopc:
	case e.errc <- err:
	}
}
```

### (*serveCtx)serve()

`serve()`函数我们可以快速地通过缩进来阅读：

```go
// 非安全，即HTTP
if sctx.insecure {
}

// 安全，即HTTPS
if sctx.secure {
}
```

而我们关注的HTTP部分，又分为两块 - HTTP2和HTTP1。而每一个server都有一个关键变量：

`mux`多路复用器 - 在web编程的场景下，往往指多个路由规则的匹配，最常见的如将URL映射到一个处理函数；而创建完`mux`后，将它注册到server中运行起来。

## 小结

到这里，我们串联了整个`main`函数运行的相关代码，也建立了etcd server运行的主要逻辑，我也总结到了下面这张图中。

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd-1-main.drawio.png)

> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

