---
title: etcd源码分析 - 5.【打通核心流程】EtcdServer消息的处理函数
date: 2022-07-25 12:00:00
categories: 
- 技术框架
tags:
- Go-etcd
---

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd.jpg)

在上一讲，我们梳理了`EtcdServer`的关键函数`processInternalRaftRequestOnce`里的四个细节。

其中，`wait.Wait`组件使用里，我们还遗留了一个细节实现，也就是请求的处理结果是怎么通过channel返回的。

```go
select {
  // 正常消息的返回，也就是我们本章要研究的
	case x := <-ch:
		return x.(*applyResult), nil
	case <-cctx.Done():
		proposalsFailed.Inc()
		s.w.Trigger(id, nil) 
		return nil, s.parseProposeCtxErr(cctx.Err(), start)
	case <-s.done:
		return nil, ErrStopped
}
```

<!-- more -->

## 明确问题与思路

我们回顾上节的问题，我们就是要找到下面两处操作的代码：

1. 往`ch`这个channel里发送了一个`*applyResult`结构的消息
2. 对wait进行了Trigger操作

常见思路分为两种：

1. 顺序思维 - 也就是自顶向下阅读代码，主要是找到调用的入口
2. 逆向思维 - 通过IDE的代码跳转功能，查找关键函数的调用处，再向上找到对应的调用栈

顺序思维不是一种源码阅读的常见行为，毕竟这需要我们非常了解源码的结构；而逆向思维，是我们快速定位到对应代码的最常见手段。

而为了加深对代码的理解，通常会采用 **先逆向、理清代码调用逻辑，后顺序、理解代码层级设计** 这样的两轮阅读。我们就针对今天的case来看看。

## 逆向阅读 - 调用逻辑

### Trigger的调用

我们利用IDE，可以查到所有的Trigger调用代码，共计10处，可以先根据文件名快速理解：

- 7处 `server.go` - 通用server部分
- 1处 `v2_server.go` - 针对v2版本协议
- 2处 `v3_server.go` - 针对v3版本协议（即我们阅读的`processInternalRaftRequestOnce`函数）

于是，我们就跳转到` server.go`，查看这7个调用函数：

- configure - 2个Tigger
  - 配置相关，直接忽略

- apply - 1个Tigger
  - 发送的数据结构不为`*applyResult`，忽略

- applyEntryNormal - 4个Tigger
  - 前两个为V2版本
  - 后两个为V3版本

确定了入口函数为`applyEntryNormal`，我们接下来就是去用IDE查找调用逻辑，不断跳转，查找它的调用栈了。

### 调用栈分析

1. applyEntryNormal
2. apply
3. applyEntries
4. applyAll
5. run
6. start
7. Start
8. StartEtcd

> 序号越小，表示越底层

这一块的代码跳转非常顺利，每一个方法基本都只有一个被调用方，我们可以快速地逐层查找，直到`main()`函数。接着，我们开始顺序阅读代码的过程。

## 顺序阅读 - 代码设计

`start`之前的方法很简单，我们直接从`run`方法开始看。

### (*EtcdServer) run()

`run()`函数可以拆分为两部分，而关键的分界线是`go`语言里经典的`for+select`语法。在一个常驻的进程中，例如服务器，`for+select`是一个非常优雅的实现，里面的每一个case都是一种处理逻辑，类似IO复用：

```go
for {
  select {
    // 正常消息
    case ap := <-s.r.apply():
    // 超时租约
    case leases := <-expiredLeaseC:
    // 错误信号
    case err := <-s.errorc:
    // 定时器
    case <-getSyncC():
    // 停止信号
    case <-s.stop:
  }
}
```

关于这种使用方法，有一个重点需要注意：**每一个case中的处理耗时要尽可能地少（除了退出），这样才能保证程序的性能。**尤其是对常见请求的处理，例如示例中的正常消息，要尽可能地短。

> 缩短单个case的处理耗时有两种思路：性能优化 或 异步化。
>
> 后者看似很简单，比如开启一个goroutine，但很有可能破坏程序数据的一致性，需要慎重。

正常消息的处理代码很短，即两行：

```go
// 执行功能的函数，关键实现为applyAll，即下一层要看的代码
f := func(context.Context) { s.applyAll(&ep, &ap) }
// sched是一个先入先出的调度方法，而Schedule只是把这个执行函数追加进去
// 这部分真正的执行在另一处，即出队列的地方，暂时无需关心
sched.Schedule(f)
```

### （*EtcdServer) applyAll()

通过`applyEntries()`函数，将每一项`entry`应用到`etcd`服务上。

### (*EtcdServer) applyEntries()

通过`apply()`应用entry，这里有3个返回值：

- term - 轮次，这是raft协议相关
- index - 索引
- shouldstop - 是否停止

### (*EtcdServer) apply()

apply将多个entries进行处理，核心代码结构整理如下：

```go
// 逐个处理entries
for i := range es {
  e := es[i]
  switch e.Type {
    // 常规消息
    case raftpb.EntryNormal:
    // 配置变更
    case raftpb.EntryConfChange:
    // 异常情况
    default:
}
```

而常规消息里的三步处理也很容易理解：

```go
// 应用普通entry的地方
s.applyEntryNormal(&e)
// 设置applied的索引位置，表示已经被应用
s.setAppliedIndex(e.Index)
// 设置轮次term信息
s.setTerm(e.Term)
```

### (*EtcdServer) applyEntryNormal()

整个函数比较长，但核心处理逻辑只有如下两块内容：

```go
// 处理raft请求，将结果返回到 *applyResult 中
var ar *applyResult
needResult := s.w.IsRegistered(id)
if needResult || !noSideEffect(&raftReq) {
  if !needResult && raftReq.Txn != nil {
    removeNeedlessRangeReqs(raftReq.Txn)
  }
  ar = s.applyV3.Apply(&raftReq)
}

// 用Trigger触发wait.Wait组件，将 *applyResult 发送出去
s.goAttach(func() {
  a := &pb.AlarmRequest{
    MemberID: uint64(s.ID()),
    Action:   pb.AlarmRequest_ACTIVATE,
    Alarm:    pb.AlarmType_NOSPACE,
  }
  s.raftRequest(s.ctx, pb.InternalRaftRequest{Alarm: a})
  s.w.Trigger(id, ar)
})
```

## 小结

本篇重点是分享一种常见的阅读代码方式：**自底向上**+**自顶向下**。在阅读了EtcdServer处理请求后，将结果通过channel发送出去的整个逻辑，相关代码的调用链路见下图。

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd-4.png)



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

