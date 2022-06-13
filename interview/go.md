## 基础

### 字符串

- rune
  - Unicode 码点
- UTF-8 编码
  - 1~4个码点

### Map

- Buckets
  - 每个bucket大小为8
  - 超过8则会overflow bucket，链表串联
- hashcode
  - 高bit位 - bucket中确定 key的位置
  - 低bit位 - 选定bucket
- 确定key和value的大小
  - runtime.maptype
- 扩容LoadFactor
  - 6.5

### 类型

- Type Definition
  - type A int
  - 可以自定义方法
- Type Alias
  - type A = int
  - 与原方法共享

### receiver

- 需要修改，用*T
- receiver 参数类型的 size 较大时，用*T
- T 类型是否需要实现某个接口

### 接口

- iface - 有方法
- eface - 没有方法
- 装箱（boxing）
  - 接口类型的装箱实际就是创建一个 eface 或 iface 的过程

### 指针

- 指针的解引用 - dereference
  - 通过指针变量读取或修改其指向的内存地址上的变量值

## 并发

### 并行与并发

- 区别
  - 并行（parallelism），指的就是在同一时刻，有两个或两个以上的任务（这里指进程）的代码在处理器上执行
    - 必要条件是具有多个处理器或多核处理器
  - 将程序分成多个可独立执行的部分的结构化程序的设计方法，就是并发设计
  - 并发不是并行，并发关乎结构，并行关乎执行
- CSP（Communicationing Sequential Processes，通信顺序进程）

### Goroutine调度器

- 将 Goroutine 按照一定算法放到不同的操作系统线程中去执行
- G-M 模型
  - 单一全局互斥锁(Sched.Lock) 和集中状态存储的存在，导致所有 Goroutine 相关操作都要上锁
  - Goroutine 传递问题：M 经常在 M 之间传递 Goroutine，这导致调度延迟增大，也增加了额外的性能损耗
  - 每个 M 都做内存缓存，导致内存占用过高，数据局部性较差
  - 由于系统调用（syscall）而形成的频繁的工作线程阻塞和解除阻塞，导致额外的性能损耗
- G-P-M 调度模型
  - P 是一个“逻辑 Proccessor”
  - 每个 G要想真正运行起来，首先需要被分配一个 P，也就是进入到 P 的本地运行队列（local runq）中
  - 抢占式调度
    - 在每个函数或方法的入口处加上了一段额外的代码，让运行时有机会在这段代码中检查是否需要执行抢占调度
    - 问题 - 只在有函数调用的地方才能插入“抢占”代码
  - 对非协作的抢占式调度的支持
    - 基于系统信号的，也就是通过向线程发送信号的方式来抢占正在运行的 G
- G 的抢占调度
  - sysmon 的 M - 监控线程，不需要绑定 P 就可以运行
    - 释放闲置超过 5 分钟的 span 内存
    - 如果超过 2 分钟没有垃圾回收，强制执行
    - 将长时间未处理的 netpoll 结果添加到任务队列
    - 向长时间运行的 G 任务发出抢占调度 - retake
    - 收回因 syscall 长时间阻塞的 P
  - 如果一个 G 任务运行 10ms，sysmon 就会认为它的运行时间太久而发出抢占式调度的请求
  - channel 阻塞或网络 I/O 情况下的调度
    - M 会尝试运行 P 的下一个可运行的 G
    - 这个时候 P 没有可运行的 G 供 M 运行，那么 M 将解绑 P，并进入挂起状态
    - 当 I/O 操作完成或 channel 操作完成，在等待队列中的 G 会被唤醒，标记为可运行（runnable），并被放入到某 P 的队列中，绑定一个 M 后继续执行
  - 系统调用阻塞情况下的调度
    - G 会阻塞
    - 执行这个 G 的 M 也会解绑 P，与 G 一起进入挂起状态
    - 如果此时有空闲的 M，那么 P 就会和它绑定，并继续执行其他 G；如果没有空闲的 M，但仍然有其他 G 要去执行，那么 Go 运行时就会创建一个新 M

### channel

- 类型
  - send-only
  - recv-only
- 关闭channel
  - Comma-ok,ok值为false
  - for range循环结束
  - 发送端负责关闭 channel
- 对一个 nil channel 执行获取操作，这个操作将阻塞

### mutex

- 状态
  - 锁定
  - 唤醒
  - 饥饿
- 性能和公平
  - 正常模式
    - FIFO
    - 和新goroutine竞争锁
    - 新G正在运行，更容易竞争成功
  - 饥饿模式
    - 长尾问题
    - 直接交给队头G，新G进入队尾
    - 触发条件（or）
      - 当一个G等待锁时间超过 1 毫秒
      - 当前队列只剩下一个 G
- 自旋
  - 锁被占用且不饥饿
  - 小于最大自旋次数
  - CPU核大于1
  - 空闲P
  - 当前G所在的P下，local 待运行队列为空
- RWMutex
  - 写锁，会将readerCount设置为负数

### WaitGroup

- 2 个计数器
  - 请求计数器 v
  - 等待计数器 w

## GC

### 三个参与者

- mutator
  - 应用，不停地修改堆对象图里的指向关系
- allocator
  - 内存分配器
  - 调用 runtime.newobject
  - mmap系统调用
  - tcmalloc内存分配器
    - 尽量减少小对象高频创建与销毁时的锁竞争
    - tiny - 四级分配路径
    - small - 三级分配路径
    - large - 走页分配器
- collector
  - 垃圾回收器

### 内存分配

- arenas
  - Go 向操作系统申请内存时的最小单位
  - 每个64MB
  - 分成以 8KB 为单位的 page，由 page allocator管理
- mspan
  - 由一个或多个 page组成
  - 按照 sizeclass 再划分成多个 element
  - 按内部有没有指针分为两种
    - 有指针：scan
    - 没有指针：noscan
  - allocBits结构：分配了element，对应bit置成1

### 并发标记与清扫算法

- 并发
  - 标记和清扫过程能够与应用代码并发执行
  - 缺陷 - 无法解决内存碎片问题
- 两种语义
  - 语义垃圾（semantic garbage）
    - 有些场景也被称为内存泄露
    - 从语法上可达的对象，但从语义上来讲他们是垃圾
    - 垃圾回收器对此无能为力
  - 语法垃圾（syntactic garbage）
    - 从语法上无法到达的对象
- 三色抽象
  - 黑
    - 已经扫描完毕，子节点扫描完毕
    - gcmarkbits = 1，且在队列外
  - 灰
    - 已经扫描完毕，子节点未扫描完毕
    - gcmarkbits = 1, 在队列内
  - 白
    - 未扫描
    - collector 不知道任何相关信息
- 扫描起点
  - 根对象
  - .bss 段，.data 段以及 goroutine 的栈
- 标记过程
  - 广度优先BFS
  - gc mark worker
    - 一边从工作队列（gcw）中弹出对象
    - 一边把它的子对象 push 到工作队列（gcw）中
    - 工作队列满了，则要将一部分元素向全局队列转移
  - 并发
    - atomic.Or8
  - 协助标记
    - 应用分配内存过快
    - 响应延迟产生影响
- 对象丢失问题
  - 强三色不变性（strong tricolor invariant）
    - 禁止黑色对象指向白色对象
  - 弱三色不变性（weak tricolor invariant）
    - 黑色对象可以指向白色对象，但指向的白色对象，必须有能从灰色对象可达的路径
- 写屏障技术 write barrier
  - GC开始前，runtime.writeBarrier.enabled = true
    - 所有的堆上指针修改操作在修改之前便会额外调用 runtime.gcWriteBarrier
  - 常见技术
    - Dijistra Insertion Barrier 插入写屏障
      - 指针修改时，指向的新对象要标灰
      - 可能导致本来该删除的屏障
    - Yuasa Deletion Barrier 删除写屏障
      - 指针修改时，修改前指向的对象要标灰
  - 混合写屏障
    - 起始无需 STW 打快照，直接并发扫描垃圾即可
    - 栈上新对象为黑色，无需重新扫描栈
    - GC 过程全程无 STW
    - 扫描某一个具体的栈的时候，要停止这个G赋值器的工作
- 回收流程
  - sweep.g，主要负责清扫死对象，合并相关的空闲页
  - scvg.g，主要负责向操作系统归还内存

### GC机制

- 主动触发
  - runtime.GC
- 被动触发
  - 超过2分钟没有任何GC，强制触发
  - 使用Pacing，控制内存增长的比例
- 调优
  - 内存复用sync.Pool
  - 限制G数量

## GMP

### GMP定义

- G
  - sched保存上下文
- M
  - 内核级线程
  - 对应真实的 CPU 数
- P
  - 调度 G 和 M 之间
  - 默认为核心数

### 调度流程

1. G创建
2. 如P的本地队列
   1. 满了则去全局队列
3. M从P的本地队列获取G
   1. P的本地队列空了，去全局队列
   2. 如果还没有，则去其余P获取G - work stealing
4. 调度
   1. G因syscall阻塞M，P会和M解绑hand off，寻找空闲或创建M
      1. M 的栈保存在 G 对象上
      2. 将 M 所需要的寄存器保存在G上
      3. M 将G重新丢到 P 的任务队列
   2. G因 channel 或者 network I/O 阻塞时，不会阻塞 M，M 会寻找其他 runnable 的 G，恢复后会重新进入 runnable 进入 P 队列等待执行
5. 基于信号的抢占式调度
   1. 运行时的G是小于等于P数量的
   2. sysmon 监控 - 变动的周期性检查
      1. 释放闲置超过 5 分钟的 span 物理内存
      2. 如果超过2分钟没有垃圾回收，强制执行
      3. 将长时间未处理的 netpoll 添加到全局队列
      4. 向长时间运行的 G 任务发出抢占调度(超过 10ms 的 g，会进行 retake)
      5. 收回因 syscall 长时间阻塞的 P
6. GMP的阻塞
   1. I/O，select
   2. block on syscall
   3. channel
   4. 等待锁
   5. runtime.Gosched