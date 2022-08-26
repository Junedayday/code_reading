---
title: etcd源码分析 - 4.【打通核心流程】processInternalRaftRequestOnce四个细节
date: 2022-07-12 12:00:00
categories: 
- 技术框架
tags:
- Go-etcd
---

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd.jpg)

在上一讲，我们继续梳理了`PUT`请求到`EtcdServer`这一层的逻辑，并大概阅读了其中的关键函数`processInternalRaftRequestOnce`。

这个方法里面有不少细节，我们今天就选择其中有价值的四点来看看。

<!-- more -->

### 1. entry索引 - appliedIndex与committedIndex

在etcd中，我们将每个客户端的操作（如PUT）抽象为一个日志项（entry）。如果这个操作生效，etcd就将这个entry项同步给其它etcd server，作为数据同步。

操作有顺序之分，于是服务端就保存了一个长entry数组，用一个关键的索引index来进行区分entry数组（即一个分界的标志），对entry状态进行分类：

- entry处于状态A - 小于等于索引的entry项
- entry处于状态B - 大于索引的entry项

> 一般状态A和B都是互补的，即是一种二分类状态。

而由于分布式的特性，entry不能立刻完成执行的，于是这里就区分出了两种状态，它们复用一个entry数组：

- 已应用 - applied
- 已提交 - committed

对应索引`appliedIndex`与`committedIndex`：

```go
// 函数用atomic保证原子性
ai := s.getAppliedIndex()
ci := s.getCommittedIndex()
// 两者的差值，表示已应用但是未提交的entry数，不能太多
if ci > ai+maxGapBetweenApplyAndCommitIndex {
  return nil, ErrTooManyRequests
}
```

entry数组中的索引的一致性非常重要，尤其是在并发的场景下。而示例中的原子操作，其实是一种乐观锁的实现。

> 更多的细节就涉及到分布式相关了，这里就不展开。

### 2.id生成器 - idutil.Generator

`Generator`数据结构不复杂，它的设计详情都放在了备注里，我们可以自行阅读：

```go
// Generator generates unique identifiers based on counters, timestamps, and
// a node member ID.
//
// The initial id is in this format:
// High order 2 bytes are from memberID, next 5 bytes are from timestamp,
// and low order one byte is a counter.
// | prefix   | suffix              |
// | 2 bytes  | 5 bytes   | 1 byte  |
// | memberID | timestamp | cnt     |
```

在很多分布式系统中，都需要有一套唯一id生成器。etcd的这个方案相对简单，就是 成员id+时间戳 的组合方案。

> 关于分布式唯一id，更全面的设计可以参考Snowflake，如 https://segmentfault.com/a/1190000020899379 

### 3.认证模块 - auth.AuthStore

```go
authInfo, err := s.AuthInfoFromCtx(ctx)
```

认证功能在成熟软件中非常常见。在etcd，被独立到了`etcd/auth`模块中。这个模块的内部调用不复杂，功能就是从`context`中提取出 **用户名+版本信息**。

这个提取过程中值得注意的是，`AuthStore`是从`grpc`的`metadata`提取出想要的认证信息，而`metadata`类似于`HTTP1`协议中的header，是一种用KV形式保存和提取数据的结构。

> 串联一下我们之前的思路，etcd通过grpc-gateway将HTTP1转化成了gRPC，那么就有一个 HTTP header到grpc metadata的映射过程，有兴趣的可以去研究一下。

总体来说，etcd的认证模块做得很简单，也方便其接入service-mesh。

### 4.多协程小工具 - wait.Wait

`wait.Wait`是一个很精巧的小工具，使用起来非常简单：

```go
// 示例代码
ch := s.w.Register(id)
s.w.Trigger(id, nil)
```

我们可以在`etcd/pkg/wait`目录下看到它的具体实现，我提取了重点

```go
// 通过id，来等待和触发对应的事件。
// 注意使用的顺序：先等待，再触发。
type Wait interface {
  // 等待，即注册一个id
	Register(id uint64) <-chan interface{}
	// 触发，用一个id
	Trigger(id uint64, x interface{})
	IsRegistered(id uint64) bool
}

// 实现：读写锁+map数据结构
type list struct {
	l sync.RWMutex
	m map[uint64]chan interface{}
}

// 注册一个id
func (w *list) Register(id uint64) <-chan interface{} {
	w.l.Lock()
	defer w.l.Unlock()
	ch := w.m[id]
	if ch == nil {
    // go官方建议带buffer的channel尽量设置大小为1
		ch = make(chan interface{}, 1)
		w.m[id] = ch
	} else {
    // 不允许重复
		log.Panicf("dup id %x", id)
	}
	return ch
}

// 触发id的channel
func (w *list) Trigger(id uint64, x interface{}) {
	w.l.Lock()
	ch := w.m[id]
	delete(w.m, id)
  // 取出ch后直接Unlock（可以思考一下与defer的区别）
	w.l.Unlock()
  // 如果触发的id不存在map里，就直接跳过这个判断
	if ch != nil {
		ch <- x
		close(ch)
	}
}
```

了解`Wait`的实现之后，我们就知道在正常情况下，`Register`和`Trigger`必须一一对应。

但是，我们再往下看`processInternalRaftRequestOnce`这部分代码，发现了一个异常点：

```go
select {
  // 异常：没有找到Trigger，难道忘了？
	case x := <-ch:
		return x.(*applyResult), nil
  // 正常：用Trigger退出
	case <-cctx.Done():
		proposalsFailed.Inc()
		s.w.Trigger(id, nil) 
		return nil, s.parseProposeCtxErr(cctx.Err(), start)
  // 正常：整个server停止，此时不用关心单个Trigger了
	case <-s.done:
		return nil, ErrStopped
}
```

这里，我们可以做个简单的猜测：在另一个goroutine中，这个etcd server进行了一个操作，包括下面两步：

1. 往`ch`这个channel里发送了一个`*applyResult`结构的消息
2. 对wait进行了Trigger操作

## 小结

今天我们进一步阅读了`processInternalRaftRequestOnce`中的四个细节，加强了etcd server对请求处理的印象。

etcd作为一款优秀的开源项目，其模块设计比较精巧，而阅读源码的同学也要掌握一个技巧：**适当控制阅读深度**。比如，在阅读`PUT`请求时，第一阶段阅读到`EtcdServer`的`processInternalRaftRequestOnce`这层即可：

- 如果继续深入看`raftNode`等实现，很容易导致你的整体思路变成过程性的调用，学习不成体系
- 这时，回过头来巩固一下当前学习的部分，通过串联细节来加深印象，会对你梳理整体更有帮助

> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

