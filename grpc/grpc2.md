# gRPC源码分析(二)：从官网文档看gRPC的特性



在第一部分，我们学习了gRPC的基本调用过程，这样我们对全局层面有了一定了解。接下来，我们将结合官方文档，继续深入学习、探索下去。



## 1. Authentication 认证的实现

[官方示例](https://grpc.io/docs/guides/auth/#with-server-authentication-ssltls)

示例很简单，客户端和服务端都大致分为两步：

1. 生成对应的认证信息 `creds`
2. 将认证信息作为 `DialOption` 传入信息



认证方法的底层实现并不在我们今天的讨论范围内。这里值得一提的是，由于请求会存在大量的输入参数，这里提供的方法是 `opts ...DialOption`，也就是可变长度的参数，这一点很值得我们思考和学习。



#### 客户端的认证实现

第一步：将认证信息放入连接中

- `grpc.WithTransportCredentials` 中，将`creds` 保存到`copts.TransportCredentials`
- 调用`Dial`，在内部用 `opt.apply(&cc.dopts)`将认证信息传递到结构中
- `credsClone = creds.Clone()` 使用了一份复制，放到了Balancer中，估计是用于负载均衡的，暂时不用考虑



第二步：将认证信息请求中发出

- 首先我们先找到 `Invoke`函数，这里是发送请求的入口（对这一块有疑问的，查看上一篇）
- 分析一下函数 `invoke` ，调用了`newClientStream`，一大段代码都没有用到`copts.TransportCredentials`中的参数，大致猜测是在`clientStream`中
- 接下来这块，只通过阅读代码，要找到对应使用到`copts.TransportCredentials`很麻烦，建议第一次可以先通过反向查找，调用到这个参数的地方
- `newHTTP2Client` => `NewClientTransport` => `createTransport` => `tryAllAddrs` => `resetTransport` => `connect` => `getReadyTransport` =>`pick` => `getTransport` =>`newAttemptLocked` => `newAttemptLocked` => `newClientStream`
- 这时，我们再正向梳理一下其调用逻辑，大致是查找连接情况，对传输层进行初始化。如果你了解认证是基于传输层`Transport`的，那下次正向查找时，会有一条比较明确的方向了



#### 服务端的认证实现

第一步：将认证信息放入Server结构中

- 将`creds`包装成`ServerOption`，传入`NewServer`中
- 类似Client中的操作，被存至 `opts.creds` 里



第二步：在连接中进行认证

- 参考之前一讲的分析，我们进入函数 `handleRawConn`
- 这次，我们的进展很顺利，一下子就看到了关键函数名`useTransportAuthenticator`
- 在这里，调用了`creds`实现的`ServerHandshake`实现了认证。到这里，认证已经完成，不过我们可以再看看，认证信息是怎么传递的
- 接着，认证信息传入了 `newHTTP2Transport`，保存到结构体`http2Server`中的`authInfo`，最后返回了一个Interface `ServerTransport`
- 在进行连接时，调用了`serveStreams`，然后调用了 `http2Server`的`HandleStreams`方法，这时，我们大致可以猜测，auth在这里被用到了
- 往下看，发现有个对header帧的处理`operateHeaders`，在这里被赋值到 `pr.AuthInfo`里，并被保存到s的Context中
- 一般情况下，Context的调用是十分隐蔽的，我们可以通过反向查找，哪里调用了`peer.FromContext`，然而并没有地方应用，那认证的分析，就告一段落了



## 2. 四类gRPC请求的实现

这一块我们暂不深入源码，先了解使用时的特性



#### 2.1 简单RPC

[代码链接](https://grpc.io/docs/tutorials/basic/go/#simple-rpc)

代码逻辑很直观，即处理后返回



#### 2.2 服务端流式RPC

[代码链接](https://grpc.io/docs/tutorials/basic/go/#server-side-streaming-rpc)

代码的关键在于两个函数`inRange`和 `stream.Send`



#### 2.3 客户端流式RPC

[代码链接](https://grpc.io/docs/tutorials/basic/go/#client-side-streaming-rpc)

用一个for循环进行多次发送，`stream.Recv()`实现了从服务端获取数据，当EOF时，才调用`stream.SendAndClose`结束发送



#### 2.4 双向流式RPC

[代码链接](https://grpc.io/docs/tutorials/basic/go/#bidirectional-streaming-rpc)

将 `SendAndClose` 变为 `Send`，其余基本不变。从这里可以看到，正常的关闭都是由服务端发起的。