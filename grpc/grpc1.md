# gRPC源码分析（一）：gRPC的系统调用过程



## 准备工作

参考[官方文档](https://grpc.io/docs/quickstart/go/)，进行部署并运行成功



## 分析思路：GRPC是怎么实现方法调用的

1. 分析PB生成的对应文件
2. 运行server
3. 运行client



## 1. 分析PB生成的对应文件

### HelloRequest/HelloReply 结构分析

存在三个冗余字段 `XXX_NoUnkeyedLiteral` `XXX_unrecognized` `XXX_sizecache`

这部分主要是兼容proto2的，我们暂时不用细究



### GreeterClient客户端

传入一个 cc grpc.ClientConnInterface 客户端连接

可调用的方法为SayHello，其内部的method为"/helloworld.Greeter/SayHello”，也就是`/{package}.{service}/{method}` ，作为一个唯一的URI



### GreeterServer服务端

需要自己实现一个SayHello的方法

其中有个 UnimplementedGreeterServer 的接口，可以嵌入到对应的server结构体中（有方法未实现时，会返回codes.Unimplemented）



## 2. 运行server

### 定义server

这里pb.UnimplementedGreeterServer被嵌入了server结构，所以即使没有实现SayHello方法，编译也能通过。

但是，我们通常要强制server在编译期就必须实现对应的方法，所以生产中建议不嵌入。



### 实现自己的业务逻辑

```go
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error){
  //
}
```



### 注册TCP监听端口

```go
lis, err := net.Listen("tcp", port)
```

因为gRPC的应用层是基于HTTP2的，所以这里不出意外，监听的是tcp端口



### grpc.NewServer()

1. 入参为选项参数options
2. 自带一组defaultServerOptions，最大发送size、最大接收size、连接超时、发送缓冲、接收缓冲
3. `s.cv = sync.NewCond(&s.mu)` 条件锁，用于关闭连接
4. 全局参数 `EnableTraciing` ，会调用golang.org/x/net/trace 这个包



### pb.RegisterGreeterServer(s, &server{})

对比自己创建的server和pb中定义的server，确定每个方法都已经实现

service放在 `m map[string]*service` 中，所以一个server可以放多个proto定义的服务

内部的method和stream放在 service 中的两个map中



### s.Serve(lis)

1. listener 放到内部的map中
2. for循环，进行tcp连接，这一部分和http源码中的ListenAndServe极其类似
3. 在协程中进行handleRawConn
4. 将tcp连接封装对应的creds认证信息
5. 新建newHTTP2Transport传输层连接
6. 在协程中进行serveStreams，而http1这里为阻塞的
7. 函数HandleStreams中参数为2个函数，前者为处理请求，后者用于trace
8. 进入handleStream，前半段被拆为service，后者为method，通过map查找
9. method在processUnaryRPC处理，stream在processStreamingRPC处理，这两块内部就比较复杂了，涉及到具体的算法，以后有时间细读



## 3. 运行client

### grpc.Dial

新建一个conn连接，这里是一个支持HTTP2.0的客户端，暂不细讲



###  pb.NewGreeterClient(conn)

新建一个client，包装对应的method，方便调用SayHello



### 调用SayHello

```go
r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
```

1. 核心调用的是 Invoke 方法，具体实现要看grpc.ClientConn中
2. grpc.ClientConn中实现了Invoke方法，在call.go文件中，详情都在invoke中
