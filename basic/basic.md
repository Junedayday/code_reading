# 面试必备基础知识

## Goroutine相关

### [WaitGroup](goroutine/wg.go)

1. 理解WaitGroup的实现 - 核心是`CAS`的使用
2. Add与Done应该放在哪？ - `Add放在Goroutine外，Done放在Goroutine中`，逻辑复杂时建议用defer保证调用
3. WaitGroup适合什么样的场景？ - `并发的Goroutine执行的逻辑相同时`，否则代码并不简洁，可以采用其它方式

### [Context](goroutine/ctx.go)

1. Context上下文 - 结合Linux操作系统的`CPU上下文切换/子进程与父进程`进行理解
2. 如何优雅地使用context - 与`select`配合使用，管理协程的生命周期
3. Context的底层实现是什么？ - `mutex`与`channel`的结合，前者用于初始部分参数，后者用于通信

### [Channel](goroutine/ch.go)

1. channel用于Goroutine间通信时的注意点 - 合理设置channel的size大小 / 正确地`关闭channel`
2. 合理地运用channel的发送与接收 - 运用函数传入参数的定义，限制 `<- chan` 和 `chan <-`
3. channel的底层实现 - `环形队列`+`发送、接收的waiter通知`，结合Goroutine的调度思考
4. 理解并运用channel的阻塞逻辑 - 理解channel的每一对 `收与发` 之间的逻辑，巧妙地使用
5. 思考channel嵌套后的实现逻辑 - 理解用 `chan chan` 是怎么实现 `两层通知` 的？

### [sync.Map](goroutine/sync_map.go)

1. sync.Map的核心实现 - 两个map，一个用于写，另一个用于读，这样的设计思想可以类比`缓存与数据库`
2. sync.Map的局限性 - 如果写远高于读，dirty->readOnly 这个类似于 `刷数据` 的频率就比较高，不如直接用 `mutex + map` 的组合
3. sync.Map的设计思想 - 保证高频读的无锁结构、空间换时间

### [sync.Cond](goroutine/sync_cond.go)

1. sync.Cond的核心实现 - 通过一个锁，封装了`notify 通知`的实现，包括了`单个通知`与`广播`这两种方式
2. sync.Cond与channel的异同 - channel应用于`一收一发`的场景，sync.Cond应用于`多收一发`的场景
3. sync.Cond的使用探索 - 多从专业社区收集意见 https://github.com/golang/go/issues/21165

### [sync.Pool](goroutine/sync_pool.go)

1. sync.Pool的核心作用 - 读源码，`缓存稍后会频繁使用的对象`+`减轻GC压力`
2. sync.Pool的Put与Get - Put的顺序为`local private-> local shared`，Get的顺序为 `local private -> local shared -> remote shared`
3. 思考sync.Pool应用的核心场景 - `高频使用且生命周期短的对象，且初始化始终一致`，如fmt
4. 探索Go1.13引入`victim`的作用 - 了解`victim cache`的机制

### [atomic](goroutine/atomic.go)

1. `atomic` 适用的场景 - 简单、简单、简单！不要将atomic用在复杂的业务逻辑中
2. `atomic.Value` 与 `mutex` - 学习用两者解决问题的思路
3. 了解 `data race` 机制 - atomic可以有效地减少数据竞争

## 数据结构进阶

### [map](data/map.go)

1. `map` 读取某个值时 - 返回结果可以为 `value,bool` 或者 `value`。注意后者，在key不存在时，会返回value对应类型的默认值
2. `map` 的 `range` 方法需要注意 - `key,value` 或者 `key`。注意后者，可以和`slice`的使用结合起来
3. `map` 的底层相关的实现 - 串联 初始化、赋值、扩容、读取、删除 这五个常见实现的背后知识点，详细参考示例代码链接与源码

### [map示例](data/map_code.go)

1. `map` 的 `range` 操作 - key、value 都是值复制
2. `map` 如何保证按key的某个顺序遍历？ - 分两次遍历，第一次取出所有的key并排序；第二次按排序后的key去遍历(这时你可以思考封装map和slice到一个结构体中)？
3. `map` 的使用上，有什么要注意的？ - 遍历时，尽量只修改或删除当前key，操作非当前的key会带来不可预知的结果
4. 从 `map` 的设计上，我们可以学到 - Go语言对map底层的hmap做了很多层面的优化与封装，也屏蔽了很多实现的细节，适用于绝大多数的场景；而少部分有极高性能要求的场景，就需要深入到hmap中的相关细节。

### [slice](data/slice.go)

1. 熟悉 `slice` 的底层数据结构 -  实际存储数据的`array`，当前长度`len`与容量`cap`
2. `slice的扩容机制` - 不严格来说，当长度小于1024时，cap翻倍；大于1024时，增加1/4
3. `slice` 有很多特性与 `map` 一致 - 记住一点，代码中操作的`slice`和`map`只是上层的，实际存储数据的是`array`与`hmap`
