

# gRPC源码分析(四)：剖析Proto序列化

在前面的分析中，我们已经知道了使用proto序列化的代码在[encoding目录](https://github.com/grpc/grpc-go/tree/v1.29.x/encoding/proto)中，路径中只有三个文件，其中2个还是测试文件，看起来这次的工作量并不大。

首先，针对读源码是先看源代码还是测试代码，因人而异。个人建议在对源码毫无头绪时，先从测试入手，了解大致功能；如果有一定基础，那么也可以直接入手源代码。我认为优秀的Go源码可读性是非常高的，所以一般情况下，我都直接从源文件入手，遇到问题才会去对应的测试里阅读。



## Marshal

Marshal的代码不多，关键在于传入参数的类型，有2个分支路线：

1. [proto.Marshaler类型](https://github.com/grpc/grpc-go/blob/v1.29.x/encoding/proto/proto.go#L68)，实现了`Marshal() ([]byte, error)`方法
2. [proto.Message类型](https://github.com/grpc/grpc-go/blob/v1.29.x/encoding/proto/proto.go#L54)，实现了`Reset()`、`String() string` 和`ProtoMessage()`三个方法



我们回头看看proto生成的[go文件](https://github.com/grpc/grpc-go/blob/v1.29.x/examples/helloworld/helloworld/helloworld.pb.go#L35)，发现对应的是第二个接口。那我们接着看：

1. 调用了protoBufferPool，是一个sync.Pool，是为了加速proto对象的分配
2. 内部采用的是 `marshalAppend`，字面来看就是 序列化并追加，对应了 ` wire-format`这个概念，并不需要将整个结构加载完毕、再进行序列化
3. 接下来调用的是`protoV2.MarshalOptions`，需要关注的是protoV2是另一个package，`protoV2 "google.golang.org/protobuf/proto"`
4. 在正式marshal前，调用`m.ProtoReflect()`方法，根据名字可以猜测是对Message做反射，详细内容不妨后面再看
5. 最后就是正式的marshal了，分两个分支：`out, err = methods.Marshal(in)`和`out.Buf, err = o.marshalMessageSlow(b, m)`。后者是慢速的，一般情况下是不会用到，我们重点关注前者，这时就需要回头看4中的实现了
6. 逐个往前搜索，`接口protoreflect.Message => ` `接口Message` =>`函数MessageV2`  => `函数ProtoMessageV2Of`  => `函数legacyWrapMessage` => `函数MessageOf` => `类型messageReflectWrapper`，终于，在这里找到了目标函数 `ProtoMethods`
7. 因为我们取的是`methods`，所以很快将代码定位到 `makeCoderMethods` => `marshal` => `marshalAppendPointer` ，最后找到一行核心代码 `b, err = f.funcs.marshal(b, fptr, f, opts)`
8. 那这个marshal什么时候被赋值的呢？在步骤7中，我们查看了methods被赋值的地方，其实旁边就有一个函数 `makeReflectFuncs` ，最后定位到了 `/google.golang.org/protobuf/internal/impl/codec_gen.go` 文件中。每种变量的序列化，都是按照特定规则来执行的。



## 实战

那么 protobuf 实际是如何对每种类型进行Encoding的呢？有兴趣的朋友可以点击[这个链接](https://developers.google.com/protocol-buffers/docs/encoding)，阅读原文。这里，我直接拿出一个实例进行讲解。

#### 定义proto

```protobuf
message People {
	bool male = 1;
	int32 age = 2;
	string address = 3;
}
```



#### 生成对应文件后，编写测试用例

```go
func main() {
	people := &pbmsg.People{
		Male:    true,
		Age:     80,
		Address: "China Town",
	}
	b, _ := proto.Marshal(people)
	fmt.Printf("%b\n", b)
}
```



#### 运行生成结果

```shell
[1000 1 10000 1010000 11010 1010 1000011 1101000 1101001 1101110 1100001 100000 1010100 1101111 1110111 1101110]
```



#### 分析第一个字段Bool

首先，Male是一个bool字段，序号为1。

根据Google上的文档，bool是Varint，所以计算

(field_number << 3) | wire_type = (1<<3)|0 = 8，对应第一个字节： `1000`

然后，它的值true对应第二个字节`1`



#### 分析第二个字段Int

同样的，(field_number << 3) | wire_type = (2<<3)|0 = 16，对应第三个字节`10000`

值80对应`1010000`



#### 分析第三个字段String

因为string是不定长的，所以需要一个额外的长度字段

(field_number << 3) | wire_type = (3<<3)|2=26，对应`11010`

接下来是长度字段，我们有10个英文单词，所以长度为10，对应 `1010`

然后就是10个Byte表示"China Town”了



## 结语

本次的分析到这里就暂时告一段落了，阅读protobuf的相关代码还是非常耗时耗力的。其实这块最主要的复杂度在于为了兼容新老版本，采用了大量的Interface实现。Interface带有面向对象特色，在重构代码时很有意义，不过也给阅读代码时，查找方法对应实现时带来了复杂度。