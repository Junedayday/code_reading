# 【K8s源码品读】004：Phase 1 - kubectl - 发送创建Pod请求的实现细节

## 聚焦目标

理解kubectl是怎么向kube-apiserver发送请求的



## 目录

1. [向kube-apiserver发送请求](#send-request)
2. [RESTful客户端是怎么创建的](#RESTful-client)
3. [Object是怎么生成的](#object)
4. [发送到kube-apiserver](#post)
5. [kubectl第一阶段源码阅读总结](#summary)



## send request

```go
// 在RunCreate函数中，关键的发送函数
obj, err := resource.
				NewHelper(info.Client, info.Mapping).
				DryRun(o.DryRunStrategy == cmdutil.DryRunServer).
				WithFieldManager(o.fieldManager).
				Create(info.Namespace, true, info.Object)

// 进入create函数，查看到
m.createResource(m.RESTClient, m.Resource, namespace, obj, options)

// 对应的实现为
func (m *Helper) createResource(c RESTClient, resource, namespace string, obj runtime.Object, options *metav1.CreateOptions) (runtime.Object, error) {
	return c.Post().
		NamespaceIfScoped(namespace, m.NamespaceScoped).
		Resource(resource).
		VersionedParams(options, metav1.ParameterCodec).
		Body(obj).
		Do(context.TODO()).
		Get()
}

/*

到这里，我们发现了2个关键性的定义:
1. RESTClient 与kube-apiserver交互的RESTful风格的客户端
2. runtime.Object 资源对象的抽象，包括Pod/Deployment/Service等各类资源

*/
```



## RESTful Client

我们先来看看，与kube-apiserver交互的Client是怎么创建的

```go
// 从传入参数来看，数据来源于Info这个结构
r.Visit(func(info *resource.Info, err error) error{})

// 而info来源于前面的Builder，前面部分都是将Builder参数化，核心的生成为Do函数
r := f.NewBuilder().
		Unstructured().
		Schema(schema).
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		FilenameParam(enforceNamespace, &o.FilenameOptions).
		LabelSelectorParam(o.Selector).
		Flatten().
		Do()

// 大致看一下这些函数，我们可以在Unstructured()中看到getClient函数，其实这就是我们要找的函数
func (b *Builder) getClient(gv schema.GroupVersion) (RESTClient, error) 

// 从返回值来看，client包括默认的REST client和配置选项
NewClientWithOptions(client, b.requestTransforms...)

// 这个Client会在kubernetes项目中大量出现，它是与kube-apiserver交互的核心组件，以后再深入。
```



## Object

`Object`这个对象是怎么获取到的呢？因为我们的数据源是来自文件的，那么我们最直观的想法就是`FileVisitor`

```go
func (v *FileVisitor) Visit(fn VisitorFunc) error {
	// 省略读取这块的代码，底层调用的是StreamVisitor的逻辑
	return v.StreamVisitor.Visit(fn)
}

func (v *StreamVisitor) Visit(fn VisitorFunc) error {
	d := yaml.NewYAMLOrJSONDecoder(v.Reader, 4096)
	for {
		// 这里就是返回info的地方
		info, err := v.infoForData(ext.Raw, v.Source)
  }
}

// 再往下一层看，来到mapper层，也就是kubernetes的资源对象映射关系
func (m *mapper) infoForData(data []byte, source string) (*Info, error){
  // 这里就是我们返回Object的地方，其中GVK是Group/Version/Kind的缩写，后续我们会涉及
  obj, gvk, err := m.decoder.Decode(data, nil, nil)
}
```



这时，我们想回头去看，这个mapper是在什么时候被定义的？

```go
// 在Builder初始化中，我们就找到了
func (b *Builder) Unstructured() *Builder {
	b.mapper = &mapper{
		localFn:      b.isLocal,
		restMapperFn: b.restMapperFn,
		clientFn:     b.getClient,
    // 我们查找资源用到的是这个decoder
		decoder:      &metadataValidatingDecoder{unstructured.UnstructuredJSONScheme},
	}
	return b
}

// 逐层往下找，对应的Decode方法的实现，就是对应的数据解析成data：
func (s unstructuredJSONScheme) decode(data []byte) (runtime.Object, error) {
	// 细节暂时忽略
}
```



## Post

了解了`REST Client`和`Object`的大致产生逻辑后，我们再回过头来看发送的方法

```go
// RESTful接口风格中，POST请求对应的就是CREATE方法
c.Post().
		NamespaceIfScoped(namespace, m.NamespaceScoped).
		Resource(resource).
		VersionedParams(options, metav1.ParameterCodec).
		Body(obj).
		Do(context.TODO()). 
		Get() 

// Do方法，发送请求
err := r.request(ctx, func(req *http.Request, resp *http.Response) {
		result = r.transformResponse(resp, req)
	})

// Get方法，获取请求的返回结果，用来打印状态
switch t := out.(type) {
	case *metav1.Status:
		if t.Status != metav1.StatusSuccess {
			return nil, errors.FromObject(t)
		}
	}
```



## Summary

到这里我们对kubectl的功能有了初步的了解，希望大家对以下的关键内容有所掌握：

1. 命令行采用了`cobra`库，主要支持7个大类的命令；
2. 掌握Visitor设计模式，这个是kubectl实现各类资源对象的解析和校验的核心；
3. 初步了解`RESTClient`和`Object`这两个对象，它们是贯穿kubernetes的核心概念；
4. 调用逻辑
   1. cobra匹配子命令
   2. 用Visitor模式构建Builder
   3. 用RESTClient将Object发送到kube-apiserver