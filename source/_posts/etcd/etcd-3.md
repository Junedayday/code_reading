---
title: etcd源码分析 - 3.【打通核心流程】PUT键值对的执行链路
date: 2022-07-04 12:00:00
categories: 
- 技术框架
tags:
- Go-etcd
---

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd.jpg)

在上一讲，我们一起看了etcd server是怎么匹配到对应的处理函数的，如果忘记了请回顾一下。

今天，我们再进一步，看看`PUT`操作接下来是怎么执行的。

<!-- more -->

## HTTP1部分

### request_KV_Put_0

整个函数主要分为两步：

1. 解析请求到`etcdserverpb.PutRequest`数据结构；
2. `client`执行`PUT`操作；

关于解析部分，我们暂时不用关心如何反序列化的（反序列化是一种可替换的插件，常见的如json/protobuffer/xml），重点看看它的数据结构：

```go
type PutRequest struct {
	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Lease int64 `protobuf:"varint,3,opt,name=lease,proto3" json:"lease,omitempty"`
	PrevKv bool `protobuf:"varint,4,opt,name=prev_kv,json=prevKv,proto3" json:"prev_kv,omitempty"`
	IgnoreValue bool `protobuf:"varint,5,opt,name=ignore_value,json=ignoreValue,proto3" json:"ignore_value,omitempty"`
	IgnoreLease bool `protobuf:"varint,6,opt,name=ignore_lease,json=ignoreLease,proto3" json:"ignore_lease,omitempty"`
}
```

从我们执行的`etcdctl put mykey "this is awesome"`为例，不难猜到：

- Key - mykey
- Value - this is awesome

接下来，我们去看看client是如何执行`PUT`的。

### etcdserverpb.kVClient

`request_KV_Put_0`函数中的client是一个接口`KVClient`，包括Range/Put/DeleteRange/Txn/Compact五种操作。

> 这里提一下，很多开源库将接口与其实现，用大小写来区分，来强制要求外部模块依赖其接口：
>
> 比如KVClient作为接口，而kVClient作为其实现是小写的，所以外部模块无法直接使用kVClient这个数据结构。

它的实现可以很容易地翻阅代码找到，是`etcdserverpb.kVClient`。我们去看看对应的`PUT`方法。

```go
func (c *kVClient) Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error) {
	out := new(PutResponse)
	err := grpc.Invoke(ctx, "/etcdserverpb.KV/Put", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
```

这里，我们就找到了HTTP调用gRPC的影子，也就是这个`Invoke`方法。

## gRPC部分

### proto文件

关于gRPC的调用部分，我比较推荐从最原始的`proto`文件开始阅读，主要包括2个文件：

- `etcd/etcdserver/etcdserverpb/rpc.proto` 原始文件
- `etcd/etcdserver/etcdserverpb/rpc.pb.go` 生成文件

从下面的定义可以看到HTTP1采用了`POST`方法，对应URL为`/v3/kv/put`：

```proto
rpc Put(PutRequest) returns (PutResponse) {
  option (google.api.http) = {
    post: "/v3/kv/put"
    body: "*"
  };
}
```

### etcdserverpb.RegisterKVServer

我们要注意，proto文件及其生成的go代码只是定义了server的接口，具体的实现需要开发者自行编码实现，通过注册函数`RegisterKVServer`将两者串联起来。

查找该函数的调用，分为三个，各有用途：

1. `grpc.go` - server的调用处
2. `grpc_proxy.go` - proxy代理模式，忽略
3. `mockserver.go` - mock服务，忽略

跳转到1对应的代码处，我们看到了注册函数`pb.RegisterKVServer(grpcServer, NewQuotaKVServer(s))`。

### NewQuotaKVServer

进一步跳转，来到了`NewKVServer`函数中。

### NewKVServer

这个函数新建了一个`kvServer`对象，它实现接口`etcdserverpb.KVServer`。我们再看对应的`PUT`方法。

### (*kvServer) Put

`Put`方法代码很少：

```go
func (s *kvServer) Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error) {
	if err := checkPutRequest(r); err != nil {
		return nil, err
	}

	resp, err := s.kv.Put(ctx, r)
	if err != nil {
		return nil, togRPCError(err)
	}

	s.hdr.fill(resp.Header)
	return resp, nil
}
```

而这里的`s.kv`，其定义为接口`etcdserver.RaftKV`，定义了如下五个方法：

```go
type RaftKV interface {
  // 范围操作
	Range(ctx context.Context, r *pb.RangeRequest) (*pb.RangeResponse, error)
  // KV操作
	Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error)
  // 删除范围
	DeleteRange(ctx context.Context, r *pb.DeleteRangeRequest) (*pb.DeleteRangeResponse, error)
  // 事务
	Txn(ctx context.Context, r *pb.TxnRequest) (*pb.TxnResponse, error)
  // 压缩
	Compact(ctx context.Context, r *pb.CompactionRequest) (*pb.CompactionResponse, error)
}
```

etcd server集群之间采用的是RAFT协议，而`RaftKV`则是实现的关键。查找RaftKV的具体实现`EtcdServer`，我们就找到了如下代码：

### (*EtcdServer) Put

```go
func (s *EtcdServer) Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error) {
	ctx = context.WithValue(ctx, traceutil.StartTimeKey, time.Now())
	resp, err := s.raftRequest(ctx, pb.InternalRaftRequest{Put: r})
	if err != nil {
		return nil, err
	}
	return resp.(*pb.PutResponse), nil
}
```

值得注意的是，这里将多种请求命令（如PUT/RANGE），都封装到了一个结构体`InternalRaftRequest`中。

我们继续跳转。

### (*EtcdServer) raftRequest

```go
func (s *EtcdServer) raftRequest(ctx context.Context, r pb.InternalRaftRequest) (proto.Message, error) {
	return s.raftRequestOnce(ctx, r)
}
```

一般来说，带`Once`关键字的函数，强调只执行一次，简单的可以用`sync.Once`函数实现，复杂的会结合`sync`和`atomic`进行针对性的设计。

我们再进一步跳转。

### (*EtcdServer) processInternalRaftRequestOnce

这部分的代码我做了个精简，如下：

```go
// 发起RAFT提案Propose（分布式共识算法的术语，不清楚的同学有个初步印象即可）
err = s.r.Propose(cctx, data)

// 监控的metrics，表示提案处于Pending计数+1，退出则-1
proposalsPending.Inc()
defer proposalsPending.Dec()

// 处理结果异步返回，分为三个情况
select {
  // 正常返回结果
	case x := <-ch:
		return x.(*applyResult), nil
  // 超时等异常处理
	case <-cctx.Done():
		proposalsFailed.Inc()
		s.w.Trigger(id, nil) // GC wait
		return nil, s.parseProposeCtxErr(cctx.Err(), start)
  // 被正常关闭
	case <-s.done:
		return nil, ErrStopped
}
```

## raftNode部分

### (raftNode)Propose

如果我们对`Propose`方法感兴趣，就需要深入学习`raftNode`这一大块了，它是对RAFT协议的整体封装。

在`etcd`里，`raftNode`是一个比较独立的模块，我们会在后续模块专门分析。

## 小结

通过本篇的代码阅读，我们经历了 HTTP1 -> gRPC -> raftNode 三层，对整个`PUT`调用链有了一个基本印象。

![](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/etcd-1-main-Page-3.drawio%20(1).png)

我在图中特别标注了一些关键的接口与实现。

> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

