## GFS

### 三个原则

- 简单
- 根据硬件特性来进行设计取舍
  - 重视磁盘顺序读写
- 根据实际应用的特性，放宽了数据一致性
  - Append - At Least Once

### 单 Master 架构 - 三种身份

1. 目录服务
   1. 两种服务器
      1. master主控节点
      2. chunkserver存储数据的节点
         1. 64MB chunk，唯一标识handle编号
         2. 三份副本（replica）：1个primary+2个secondary
   2. 三种主要的元数据
      1. 文件和 chunk 的命名空间信息
      2. 全路径文件名到多个 chunk handle 的映射关系
      3. chunk handle 到 chunkserver 的映射关系
   3. 客户端读取
      1. 客户端发出两部分信息 - 文件名+对应chunk(根据数据段offset/length计算)
      2. master把对应的chunkserver给客户端
      3. 客户端去chunkserver获取数据
2. Backup Master
   1. 同步复制
3. Shadow Maste
   1. 只读
   2. 异步复制
   3. 小概率不一致，相对高可用

### Master的可用性

- 主从实现
  - 所有数据保存在内存
  - Checkpoints + Replay操作日志
  - 彻底挂了，使用Backkup Master
- 切换主机
  - 监控程序修改DNS

### 兼顾硬件性能 - 网络瓶颈

- GFS 的数据写入（多个chunkserver）
  - 客户端从 master 要写入的数据，在哪些 chunkserver 上
  - 获取replica的信息
  - 客户端把要写的数据发给所有的 replica，chunkserver会把数据放在LRU的缓冲区里
  - 发送一个写请求给到主副本，主副本进行排序，然后落到chunk里
  - 主副本让次副本以同样的顺序写入数据
  - 次副本回复同步完成
  - 主副本告诉客户端已经完成
- 分离控制流和数据流
  - 控制流 - 数据在那里
  - 数据流 - 写什么数据
- 流水线式网络传输
  - 客户端发给最近的chunkserver（不一定是主副本）
  - chunkserver流水线式传输
- 文件复制
  - Snapshot指令
  - 客户端 -> 主副本 -> 次副本 -> chunkserver自己执行

### 放宽数据一致性

- 一致性
  - 多客户端从主、次副本读到的数据一样
- 确定性
  - 完整读到数据
- GFS的一致性
  - 存储了主副本的 chunkserver，来管理同一个 chunk 下数据写入的操作顺序
- GFS的非确定性
  - 随机的数据写入极有可能要跨越多个 chunk
  - 并发写的数据串了
  - 没有原子性或事务性
- Record Appends
  - 随机写入不是GFS的主要写入模式
  - 追加逻辑 - 最后一个 chunk 所在的主副本服务器
    - chunk 写得下，那就追加写，然后次副本也追加写
    - chunk写不下，填满后，让次副本填满，再告诉客户端去下一个chunk的chunkserver上追加剩余的
    - 请求会在chunk server上排队，所以不会发生覆盖
    - chunk能存下要追加的数据量，限制一次最大追加16MB（chunk本身为64MB），控制碎片和写入效率
  - 写入失败会导致脏数据，但保证至少一次
  - 保障不了数据追加的顺序
  - 解决问题
    - 写入失败 - 客户端对写入的数据去添加校验和
    - 顺序性 - 对每个写入记录生成唯一ID+时间戳，根据ID排序和去重
- 高并发和高性能、简单



## MapReduce

### 编程模型 - Template Method Pattern

- map
- shuffle
- reduce

### 协同

- 调度系统Scheduler
  - 先找到 GFS 上的对应路径，把数据进行分片（Split）- 64MB
  - 启动多个 MapReduce 程序的复刻（fork）进程
  - master 进程 把 map和reduce任务分配给有限的 worker进程
  - 被分配到 map 任务的 worker 会读取某一个分片，缓存到内存里
  - 会定期地写到 map 任务所在机器的本地硬盘，并根据分区函数分成多个区域
  - 文件存放位置会穿给master，master再分配给诶reduce的worker上
  - reduce 任务的 worker通过RPC从磁盘上读取文件，进行shuffle，如排序
  - reduce任务把结果输出到文件里
  - master 会唤醒启动 MapReduce 任务的用户程序，继续后面的逻辑

![map-reduce](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/map_reduce.webp)

### 容错Fault Tolerance

- worker 节点的失效（Worker Failure）
  - 定期ping
  - 换服务器
- master 节点的失效（Master Failure）
  - 客户端再次提交任务
  - Checkpoint写入硬盘
- 为数据异常提供了容错机制
  - 忽略错误

### 性能优化

- 减少网络传输的数据
  - 找离这台服务器最近的、有 worker 的服务器
- 减少中间数据量
  - 定义Combiner函数，对数据进行合并

### 易用性

- 提供一个单机运行的 MapReduce 的库，用来debug
- 在 master 里面内嵌了一个 HTTP 服务器，用来查状态
- 提供了一个计数器（counter）的机制，可以统计一些异常数据



## Bigtable

### 基本数据模型

- 每一行是一条数据
  - 主键 - 行键Row Key
  - 稀疏表 - 每一条记录都要把列和值存下来
  - 列可以存储多个版本
  - 列族Column Family - 同一个列族下的数据会在物理上存储在一起

### 数据分区Paritioning

- 动态区间分区
  - 按照连续的行键一段段地分区
  - 多了分裂，少了合并

### 分区管理Master + Chubby

- Tablet Server - 在线服务
  - 实际提供数据读写服务
  - 分配到 10 到 1000 个 Tablets
  - 分裂合并
  - 不负责存储，而是用SSTable格式写到GFS
- Master
  - 负责分配Tablet给Tablet Server
  - 负载均衡
  - 管理表（Table）和列族的 Schema 变更
- Chubby
  - 确保单master
  - 存储 Bigtable 的Bootstrap/Schema
  - 发现Tablet Servers，终止后进行清理
  - ACL访问权限
- 访问方式
  - Client向Chubby查询，返回给客户端TS1 - Root Tablet
  - Client向TS1查询，包括对应表+行键，返回信息TS2 - METADATA Tablet
  - Client向TS2查询对应表+行键，返回信息TS3 - User Table
  - Client去TS3查询，返回结果
- 调度
  - Tablet Server上线后再Chubby下一个获取独占锁exclusive lock，即注册
  - Master监听目录，发现有注册，即可以分配Tablets
  - 分配策略自行实现
  - Tablet Server是否独占着锁，来确定能否提供服务；移除则重新分配
  - 心跳检测
  - Master与Chubby会话过期，会自杀

### 高性能的随机数据写入

- 将硬盘随机写，转化成了顺序写
  - 把Commit Log和MemTable输出到磁盘的 Major Compaction
  - 局部性原理
    - 最近写放在MemTable
    - 最近读放在Cache
    - 行键用BloomFilter
- 实际写入
  - Tablet Server从Chubby获取权限，缓存到本地，对请求进行格式和权限校验
  - 追加写到GFS，崩溃后会replay
  - Tablet Server把数据写到MemTable
  - 到阈值后，冻结老的Immutable MemTable，转化为SSTable，写入到GFS后释放
- MemTable
  - AVL 红黑树 或 Skip List跳表
  - 根据行键
    - 随机数据插入
    - 随机数据读取
    - 有序遍历
- SSTable
  - data block - 行键、列、值以及时间戳
  - meta block - bloomFilter、index
  - Major Compaction：有序链表的多路归并
- 压缩和缓存机制
  - 通过压缩算法对每个块进行压缩
  - 把每个 SSTable 的布隆过滤器直接缓存在 Tablet Server 里
  - 针对单个 SSTable两级的缓存机制
    - 高层Scan Cache，放在Tablet Server
    - 低层Block Cache，把所在的整个块数据都缓存在 Tablet Server 里
- 数据模型
  - 是一系列的内存 + 数据文件 + 日志文件组合下封装出来的一个逻辑视图

## Thrift

### Delta Encoding

- 紧凑编码
  - 4 个 bit存储 和上一个编号的差
  - 4 个 bit 表示类型信息
- ZigZag 编码 +VQL 可变长数值表
  - 每一个字节的高位1bit标识是否需要读入下一个字节
  - 负数变成正数（乘以2再减去1），而正数去乘以 2
    - 负数最后1bit为1
    - 正数最后1bit为0
  - 7 个 bit 表示 -64~63



## Chubby

### 两阶段提交

- 提交请求
  - 预写日志：包括执行日志redo logs和回滚日志undo logs
- 提交执行
  - 所有人都同意
  - 不能反悔
- 协调者和参与者
- 故障后
  - 协调者一直重试
  - 参与者超时回滚的话，可能会出现数据不一致
- 保障了一致性（C），但是牺牲了可用性（A）
  - 出现故障，整个服务器其实就阻塞住
- 成本高：要提前准备好回滚事务

### 三阶段提交

- CanCommit - 小开销
  - 大部分不能执行的事务就能放弃掉
  - 减少redo/undo logs的开销
- PreCommit - 大开销
  - 脑裂，存在网络分区，仍然执行事务
- 提交执行
  - 参与者等待协调者超时，会把事务执行完成

### ACID

- A - 事务要么提交，要么回滚
- C - 应用层面的数据一致性
- I - 多事务之间的隔离，不能看到中间状态
  - Read Uncommitted
  - Read Committed
  - Repeatable Read
  - Serializable
- D - 持久性

### Paxos

- 分布式共识里的“可线性化”（Linearizability）
- 关键词
  - 提案（Proposal） - 要写入的操作
  - 提案者（Proposer） - 接受外部请求，要尝试写入数据的服务器节点
  - 接受者（Acceptor） - Quorum投票
- 提案阶段
  - 提案编号：高位Round，低位服务器编号
- Prepare阶段
  - Prepare 请求，带上提案号
  - 返回响应
    - 如已经接收过其它提案，则返回对应提案和内容
    - 如没有，则返回NULL
- Accept 阶段
  - Prepare得到半数响应
  - 请求
    - 提案号
    - 值分两种
      - 提案号最大的值
      - 全是NULL，放上自己的值
  - 返回
    - 接受
    - 拒绝 - 有更大的提案号
    - 无论接受还是拒绝，都返回最新的提案编号 N
  - 有人拒绝，那么提案者就需要放弃这一轮的提案
  - 超过一般的人接受请求，则通过

### Chubby

- 粗粒度的分布式锁服务
- 频率低，变化少
- Master节点 - Proposer
  - 租期 lease
- 三层系统
  - 状态机复制的系统
  - 数据库
  - 锁服务
- 减少负载 - 事件通知
- lock-delay 非正常释放，等待一段时间
- lock-sequencer 锁序号



## Hive

- 数据模型
  - SQL-like 的 HQL
  - 数据的输入输出部分加上类型系统
  - 宽表
- 数据存储
  - HDFS
  - 分区（Partition），如时间 - 将不同分区的文件放在不同的子目录
  - 分桶（Bucket），对某一列Hash作为文件名，方便采样
- Hive 的架构与性能
  - 对外的接口，包括命令行、web界面、JDBC/ODBC
  - 驱动器
    - 编译器 - Logical Plan
    - 优化器 - Physical Plan
    - 执行引擎和一个有向无环图
  - Metastore，存储所有数据表的位置、结构、分区等信息



## Dremel

- 列式存储 - 多文件追加写入
  - 解决方案 - WAL+MemTable+SSTable
- 行列混合存储
  - 对数据进行分区
  - 同区间的列存储在同一个服务器上
- 还原行数据
  - 重复嵌套字段
    - Repetition Level - 嵌套层级
  - 可选字段
    - Definition Level - 没有填充值，还是上层为NULL
- 系统架构
  - 让计算节点和存储节点放在同一台服务器上
  - 进程常驻，做好缓存，确保不需要大量的时间去做冷启动
  - 树状架构，多层聚合
  - 容错机制 - 监测各个叶子服务器的执行进度，只扫描98%或99%的数据获得近似结果
- 多层服务树
  - 根服务器（root server）
  - 中间服务器（intermediate servers）
    - 通过中间服务器来进行“垂直”扩张
    - 把数据归并的工作并行化
  - 叶子服务器（leaf servers）
    - 不直接存储数据



## Spark

### RDD - Resilient Distributed Dataset

- 函数式
- 惰性求值
- 视图

### 宽依赖和检查点

- 宽依赖
  - 一个RDD影响到多个下游节点
- 检查点
  - REPLICATE
  - 容错和性能的平衡

## Megastore

### 互联网数据库

- 可用性
  - 为远距离链接优化过的同步的、容错的日志复制器
- 伸缩性
  - 数据分区成大量的小数据库
  - 每一个都有独立进行同步复制的数据库日志
  - 存放在每个副本的 NoSQL 的数据存储里

### 复制、分区和数据本地化

- 三种数据复制方案
  - 异步主从
  - 同步主从
  - 乐观主从
- Paxos
  - 同步写入数据
  - 不区分主从
  - 导致性能瓶颈

### 架构设计

- 业务上有关联的数据之间保障“可线性化”
- 实体组（Entity Group）
  - 实体以及挂载在这个实体下的一系列实体
- 跨实体组的操作
  - 两阶段事务
  - 异步消息

### 实体组

- 对 Key 进行预 Join（Pre-Joining with Keys）- 实体和实体所挂载的数据在相邻行
- 两种索引
  - 本地索引 - 实体组内
  - 全局索引 - 弱一致性
- 索引优化
  - 索引中存储数据，覆盖索引
  - 为repeated字段建立索引，反向查询
  - 内联索引（Inline Indexes）
    - 父实体（Parent Entity）能够快速访问子实体（Child Entity）的某些字段

### 事务与隔离性

- 时间戳机制
  - current最新版本 - 可线性化（要保证Apply成功）
  - snapshot完全应用的上一个版本（不保证Apply成功）
  - inconsistent不检查一致性
- 事务提交步骤
  - 读（Read） - 时间戳与事务日志位置
  - 应用层的逻辑（Application Logic） - 读写操作收集到log entry中
  - 提交事务（Commit） - Paxos算法，副本一致，并追加日志
  - 应用事务（Apply） - 将实体和索引的修改写入到Bigtable里
  - 清理工作（Clean UP） -  清理数据
- 时间差 - Commit与Apply之间
- 并发写会竞争追加日志的位置，失败重写

### Paxos优化

- 两阶段优化
  - Accept带上下一次Prepare
- 协同服务器（Coordinator Server）
  - 每个数据中心部署
  - 观察到的最新的实体组的集合
  - 最新的副本 - 直接读取
  - 非最新的 - 追赶共识（catch up）
- 数据的快速写 - Leader-Based
  - 确认了下一个事务日志的位置，以及 Leader 是哪个节点
  - Accept Leader
    - 失败则走Paxos的Prepare请求，进行提案
  - Accept
    - 成功后，没有 Accept 的节点发起Invalidate
  - Apply
    - 应用失败则会报conflict error
- 三种副本
  - 完全副本（Full Replica）
  - 见证者副本（Witness Replica）
  - 只读副本（Read-Only Replica）

## Spanner

### 整体架构设计思路

- 让数据副本，尽量放在离会读写它的用户近的数据中心
- Zone - 物理隔离
  - Zonemaster，负责把数据分配给 Spanserver
  - Spanserver，负责把数据提供给客户端
  - Location Proxy，用来让客户端定位哪一个 Spanserver 可以提供自己想要的数据
- Universe Master
- Placement Driver - 调度数据
- 控制策略
  - 明确指定哪些数据中心应该包括哪些数据
  - 申明要求数据距离用户有多远
  - 申明不同数据副本之间距离有多远
  - 申明需要维护多少个副本

### Spanserver

- 数据模型
  - B树
  - WAL
  - 时间戳(key:string, timestamp:int64) ⇒ string
- 同步复制
  - Paxos 算法
  - Leader 是类似于 Chubby 锁里面的“租约”的机制
- 数据写入时，两份日志
  - Paxos 日志
  - Tablet 日志

### 数据调度

- Paxos Group - 一个 Tablet 和它的所有副本
  - 包含多个目录Directory
  - 一小片连续的行键
  - 不同的目录之间，可以是离散的
  - 当一个目录变得太大的时候，Spanner 还会再进行分片存储
- 把那些频繁共同访问的目录，调度到相同的 Paxos 组里
- 可以根据实际数据被访问后的统计数据，来做动态的调度
- movedir 的后台任务
  - 先在后台转移数据，而当所有数据快要转移完的时候，再启动一个事务转移最后剩下的数据，来减少可能的阻塞时间

### 事务性能

- 协调者
  - 让 Paxos 的 Leader 作为协调者
  - 每个 Spanserver 上，会有一个事务管理器，用来支持分布式事务
    - 跨 Paxos Group 的事务 - 互相协调
    - 单Group - 直接从锁表里获取锁
- 分布式数据库的事务并发问题
  - 中心化的事务 ID 生成器
  - 时间戳
    - 时间戳 + 服务器编号
- 原子钟和置信区间
  - 原子钟+GPS时钟
  - 误差：1 毫秒到 7 毫秒之间
    - 对比分配的时间戳和本地的时间
    - 等待7ms，然后才提交

### 事务的具体实现

- 严格串行化
- TrueTime API
  - 读写事务
    - 客户端发给协调者
    - 协调者Prepare请求的时间戳
    - Paxos组保证时间戳单调递增
    - 协调者生成事务提交的时间戳
      - 大于等于 Prepare 的时间戳
      - 大于当前Leader分配给之前事务的时间戳
      - 晚于协调者自身当前的时间戳，包括误差
  - 快照读事务
    - 时间戳 - Spanner系统提供
  - 普通的快照读
    - 时间戳 - 客户端指定，或者客户端指定时间戳的upper bound

## S4和Storm

### S4流式计算的逻辑模型

- 计算过程抽象为Processing Element
  - 功能（functionality）
  - 事件类型（types of events）
  - 键（keyed attribute）
  - 值（value）
- 有向无环图（DAG）
  - 由PE 组成
  - 起点是无键 PE（Keyless PE），接收外部事件流
    - 解析消息，生成事件类型（Event Type）、事件的 Key、事件的 Value
- 最终结果 - 发布（Publish）
- 通信层模块决定发送到哪个下层节点

### Storm

- 有向无环图 Topology
  - Spouts数据源
  - Tuple元组，传输的最小粒度
  - Streams数据流
  - Bolts逻辑处理
- 数据流的分组（Grouping）
  - 随机分组（Shuffle Grouping）
  - 字段分组（Fields Grouping）
  - 全部分组（All Grouping），广播
  - 全局分组（Global Grouping），统一发送到一个Bolts
  - 无分组（None Grouping）
  - 指向分组（Direct Grouping），指定下游的Bolts
  - 本地或随机分组（Local or Shuffle Grouping），同一个 worker 进程里
- Master+Worker 的系统架构
  - Nimbus - Master进程
  - Supervisor - 管理机器上的Worker进程
  - Worker进程
  - Zookeeper
    - Nimbus写任务分配
    - Supervisor监听
- 容错
  - ZeroMQ
    - Worker间通信
  - AckerBolt
    - Bolt 告诉 AckerBolt
      - 处理完了某一个 Tuple
      - 衍生往下游的哪些 Tuple 也已经发送出去了
    - 利用位运算里的异或（XOR）
      - 64 位的 message id
    - At Least Once

## Kafka

### 大数据系统的基本框架

- 性能问题
  - 分布式文件系统HDFS（GFS）上的单个文件，只适合一个客户端顺序大批量的写入
- Log Collector
  - 定时地向 HDFS 上转存（Dump）文件

### 系统架构

- Producer
- Broker
  - Topic
  - Partition
    - 物理上 - 一组大小基本相同的 Segment 文件
- Consumer
  - 拉数据

### 高可用

- 多副本
  - Follower去Leader拉取
  - 有多少个 Follower 成功拉取数据之后，认为写入成功
- 负载均衡 - 再平衡（Rebalance）
  - 把分区重新按照 Consumer 的数量进行分配，确保下游的负载是平均的
- 没有顺序保障

### Lambda 架构

- 原始日志 - 主数据（Master Data）
- Speed Layer - Storm 进行实时的数据处理
- Batch Layer - 定时运行 MapReduce 程序
- 服务层（Serving Layer）

弊端：双倍的资源

### Kappa 架构

- 去掉Batch Layer
- 在实时处理层，支持了多个视图版本



## Dataflow

### Exactly Once

- Bolt节点中  message-id去重
  - BloomFilter
  - 按照时间窗口切分多个BloomFilter

### 计算节点迁移的容错问题

- 并行度
  - 增加服务器
  - 增加Bolt数量
- 持久化

### 处理消息的时间窗口

- TickTuple
- 处理时间（Processing Time）
- 事件时间（Event Time）
  - 何时发送统计结果到外部数据库

### MillWheel

- 计算 - Computation
  - 输入
  - 输出
  - 计算逻辑
- 键 - Key
  - 三元组 - Key, Value, TimeStamp
- 低水位 - Low Watermark
  - 最早的时间戳
  - 上游的 Computation 中，时间戳最早的那一个
  - 每一个 Computation 进程把低水位上报给Injector 模块，然后分发
  - 每一种类型的 Computation，都会有一个自己的水位信息
- 定时器 - Timer
  - 系统自己会根据水位信息，触发 Timer 执行，触发输出
- 消息处理过程
  - 消息去重
  - 处理用户实现的业务逻辑代码
  - 状态的变更，都会被一次性提交给后端的存储层
    - 通过对应的API
  - 发送 Acked 消息给上游
  - 发送结果给下游
- Strong Production
  - 对于下游要发送的数据，会先作为 Checkpoint 写下来
  - 重放 Checkpoint 的日志
- 租约
  - 解决僵尸进程
- Weak Production
  - 消息去重和 Checkpoint开销大

### Dataflow

- ParDo - 并行处理
  - DoFn函数，多服务器
- GroupByKey - Shuffle
- AssignWindows
  - 把三元组变成四元组 - key, value, event_time, window
  - 一个事件可以分配给多个时间窗口
- MergeWindows
  - 合并窗口
  - 处理乱序
- 水位方法的问题
  - 在实际的水位标记之后，仍然有新的日志到达
  - 只要有一条日志来晚了，我们的水位就会特别“低”，导致我们迟迟无法输出计算结果
- 解决方案 - Lamdba架构
  - 触发器（Trigger）机制
  - 触发后策略
    - 抛弃（Discarding）策略
    - 累积（Accumulating）策略
    - 累积并撤回（Accumulating & Retracting）策略（未实现）

## Raft

### 状态机复制 State Machine Replication

- 分布式共识
  - 日志追加写入，半数的服务器完成同步复制
  - 有节点挂掉时，系统可以自动切换恢复
  - 系统的数据需要有“一致性”，也就是不能因为网络延时、硬件故障，导致数据错乱，写入的数据读出来变了或者读不出来

### Raft 算法

- 系统设计 - 始终有一个 Leader
  - Leader 选举问题
  - 日志复制问题
  - 安全性问题 - 换主后数据覆盖
- 三个角色
  - Leader
  - Follower
  - Candidate
- Leader 选举
  - Leader定期往Follower发送心跳，Follower发现超时会选举
  - Follower发起选举，给自己投一票，也用RequestVote请求、要求其余Follower为自己投票
  - RequestVote请求带上一个Term - 任期
  - Follower在本地都会保留当前 Leader 是哪一个任期，投票时会任期+1
  - 其他 Follower接收RequestVote后会对比任期，请求的任期更大则投票，否则拒绝
  - 在一个任期里，一台服务器最多给一个 Candidate 投票
- Candidate三种情况
  - 成功 - 超过半数
  - 另一个Candidate成功，收到心跳请求后，对比Term，变成Follower
  - 无人获胜，超时，Term再自增
- 分票情况
  - 让选举的超时时间在一个区间之内随机化
- 日志复制
  - 一条操作日志的追加写
  - 通过 AppendEntries 的 RPC 调用
  - 每一条日志包含三部分信息
    - 日志的索引
    - 日志的具体内容
    - 任期
  - 每一次 Leader 的日志复制，都变成了一次强制所有 Follower 和当前 Leader 日志同步
- 安全性
  - 在选举的 RPC 里，顺便完成 Leader 是否包含所有最新的日志
  - 如果 Follower 本地有更新的数据，会拒绝投票

### 成员变更Membership Change

- 过渡共识（Joint Consensus）
  - 所有的日志追加写入，都会复制到新老配置里所有的服务器上
  - 新老配置里的任何一个服务器，都有可能被选举成 Leader 节点
  - 投票同时满足
    - 旧配置里半数以上通过
    - 新配置里半数以上通过
  - “双写”的迁移策略
- 日志压实（Log Compaction）
  - 创建快照
  - 清理日志

## Borg

- 资源
  - 限制资源（Resource Limit）
  - 保留资源（Resource Reservation）
- 调度
  - 可行性检查（feasible checking）
  - 根据打分高低来选择一台服务器
- 方案
  - 平均分配
  - 单台最大化
  - 采取尽量减少被“搁浅（stranded）”的资源数量
    - 有Task 的程序包
    - 抢占正在运行的Task资源
    - 分布到不同的物理位置
    - 不同优先级混部
