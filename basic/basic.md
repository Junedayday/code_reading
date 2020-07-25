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

1. channel用于goroutine间通信时的注意点 - 如何设置channel的size/正确地关闭channel
2. 合理地运用channel的发送与接收 - 运用函数传入参数的定义，限制 `<-` chan 和 `chan <-`
3. channel的底层实现 - `环形队列`+`发送、接收的waiter通知`