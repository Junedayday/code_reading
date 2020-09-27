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
