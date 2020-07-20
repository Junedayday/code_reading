# 面试必备基础知识

## Goroutine相关

### [WaitGroup](goroutine/wg.go)

1. 理解WaitGroup的实现 - 核心是`CAS`的使用
2. Add与Done应该放在哪？ - `Add放在Goroutine外，Done放在Goroutine中`，逻辑复杂时建议用defer保证调用
3. WaitGroup适合什么样的场景？ - `并发的Goroutine执行的逻辑相同时`，否则代码并不简洁，可以采用其它方式