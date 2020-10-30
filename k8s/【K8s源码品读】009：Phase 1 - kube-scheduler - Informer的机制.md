# 【K8s源码品读】009：Phase 1 - kube-scheduler - Informer的机制

## 聚焦目标

理解kube-scheduler启动的流程



## 目录

1. [什么是Informer](#informer)
2. [Shared Informer的实现](#Shared Informer)
3. [PodInformer的背后的实现](#PodInformer)



## Informer

什么是`Informer`？这一节，我将先抛开代码，重点讲一下这个Informer，因为它是理解k8s运行机制的核心概念。



我没有在官方文档中找到`Informer`的明确定义，中文直译为`通知器`。从[这个链接](https://medium.com/@muhammet.arslan/write-your-own-kubernetes-controller-with-informers-9920e8ab6f84)中，我们可以看到一个自定义资源的的处理流程。



我简单概况下，`Informer`的核心功能是 **获取并监听(ListAndWatch)对应资源的增删改，触发相应的事件操作(ResourceEventHandler)**



## Shared Informer

```go
/*
client 是连接到 kube-apiserver 的客户端。
我们要理解k8s的设计：
1. etcd是核心的数据存储，对资源的修改会进行持久化
2. 只有kube-apiserver可以访问etcd
所以，kube-scheduler要了解资源的变化情况，只能通过kube-apiserver
*/

// 定义了 Shared Informer，其中这个client是用来连接kube-apiserver的
c.InformerFactory = informers.NewSharedInformerFactory(client, 0)

// 这里解答了为什么叫shared：一个资源会对应多个Informer，会导致效率低下，所以让一个资源对应一个sharedInformer，而一个sharedInformer内部自己维护多个Informer
type sharedInformerFactory struct {
	client           kubernetes.Interface
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	lock             sync.Mutex
	defaultResync    time.Duration
	customResync     map[reflect.Type]time.Duration
  // 这个map就是维护多个Informer的关键实现
	informers map[reflect.Type]cache.SharedIndexInformer
	startedInformers map[reflect.Type]bool
}

// 运行函数
func (f *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()
	for informerType, informer := range f.informers {
		if !f.startedInformers[informerType] {
      // goroutine异步处理
			go informer.Run(stopCh)
      // 标记为已经运行，这样即使下次Start也不会重复运行
			f.startedInformers[informerType] = true
		}
	}
}

// 查找对应的informer
func (f *sharedInformerFactory) InformerFor(obj runtime.Object, newFunc internalinterfaces.NewInformerFunc) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()
	// 找到就直接返回
	informerType := reflect.TypeOf(obj)
	informer, exists := f.informers[informerType]
	if exists {
		return informer
	}

	resyncPeriod, exists := f.customResync[informerType]
	if !exists {
		resyncPeriod = f.defaultResync
	}
	// 没找到就会新建
	informer = newFunc(f.client, resyncPeriod)
	f.informers[informerType] = informer
	return informer
}

// SharedInformerFactory 是 sharedInformerFactory 的接口定义
type SharedInformerFactory interface {
  // 我们这一阶段关注的Pod的Informer，属于核心资源
	Core() core.Interface
}

// core.Interface的定义
type Interface interface {
	// V1 provides access to shared informers for resources in V1.
	V1() v1.Interface
}

// v1.Interface 的定义
type Interface interface {
  // Pod的定义
	Pods() PodInformer
}

// PodInformer 是对应的接口
type PodInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.PodLister
}
// podInformer 是具体的实现
type podInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// 最后，我们可以看到podInformer调用了InformerFor函数进行了添加
func (f *podInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&corev1.Pod{}, f.defaultInformer)
}
```



## PodInformer

```go
// 实例化PodInformer，把对应的List/Watch操作方法传入到实例化函数，生成统一的SharedIndexInformer接口
func NewFilteredPodInformer() cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
    // List和Watch实现从PodInterface里面查询
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1().Pods(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1().Pods(namespace).Watch(context.TODO(), options)
			},
		},
		&corev1.Pod{},
		resyncPeriod,
		indexers,
	)
}

// 我们先看看Pod基本的List和Watch是怎么定义的
// Pod基本的增删改查等操作
type PodInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	...
}
// pods 是PodInterface的实现
type pods struct {
	client rest.Interface
	ns     string
}

// List 和 Watch 是依赖客户端，也就是从kube-apiserver中查询的
func (c *pods) List(ctx context.Context, opts metav1.ListOptions) (result *v1.PodList, err error) {
	err = c.client.Get().
		Namespace(c.ns).
		Resource("pods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}
func (c *pods) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Get().
		Namespace(c.ns).
		Resource("pods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// 在上面，我们看到了异步运行Informer的代码 go informer.Run(stopCh)，我们看看是怎么run的
func (s *sharedIndexInformer) Run(stopCh <-chan struct{}) {
	// 新建controller
	func() {
		s.startedLock.Lock()
		defer s.startedLock.Unlock()

		s.controller = New(cfg)
		s.controller.(*controller).clock = s.clock
		s.started = true
	}()
	// 运行controller
	s.controller.Run(stopCh)
}

// Controller的运行
func (c *controller) Run(stopCh <-chan struct{}) {
	// 
	r := NewReflector(
		c.config.ListerWatcher,
		c.config.ObjectType,
		c.config.Queue,
		c.config.FullResyncPeriod,
	)
	r.ShouldResync = c.config.ShouldResync
	r.clock = c.clock
	if c.config.WatchErrorHandler != nil {
		r.watchErrorHandler = c.config.WatchErrorHandler
	}

	c.reflectorMutex.Lock()
	c.reflector = r
	c.reflectorMutex.Unlock()

	var wg wait.Group

	wg.StartWithChannel(stopCh, r.Run)

	wait.Until(c.processLoop, time.Second, stopCh)
	wg.Wait()
}
```

