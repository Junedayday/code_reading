## 基础概念

### 术语

- Client
  - Producer
  - Consumer
    - Consumer Group - P2P
      - Consumer Instance
    - Rebalance
    - Consumer Offset
- Broker
- Replica
  - Leader Replica
  - Follower Replica
- Topic
  - Partitioning
    - Record
      - Offset
- Consumer Group
  - Consumer Instance

### 定位

- 消息引擎系统
- 分布式流处理平台
  - 实现端到端的正确性 - 精确一次
    - 所有的数据流转和计算都在 Kafka 内部完成
  - kafka 对于流式计算的定位

### 部署方案

- 操作系统
- 磁盘
  - 机械磁盘即可
- 磁盘容量
  - 数据压缩比，如0.75
  - 备份数
- 带宽
  - 千兆网络按700M计算，常规使用1/3



## 原理

### 分区策略

- 轮询策略
- 随机策略
- 按消息键保序策略 - Key-ordering
  - 分区下消息有序性

### 压缩

- 生产端压缩
- Broker压缩

### 消息不丢失

只对“已提交”的消息（committed message）做有限度的持久化保证

- 发送丢失
  - producer是异步的
  - fire and forget
  - 使用有回调通知的API
- 消费丢失
  - 先消费，后更新位移
  - 多线程异步处理消费消息，Consumer 程序不要开启自动提交位移，而是要应用程序手动提交位移

### 幂等/事务型Producer

- 幂等 Producer
  - 空间去换时间的优化思路，即在 Broker 端多保存一些字段
  - 只能保证单分区上的幂等性
- 事务型 Producer
  - read_committed

### 消费组Rebalance

- 理想情况下，Consumer 实例的数量应该等于该 Group 订阅主题的分区总数
- 消费者组的5 种状态
  - Empty
    - 没有成员
    - 但可能有历史偏移
  - Dead
    - 没有成员
    - 元数据已经被协调者移除
  - PreparingRebalance
    - 准备重平衡
    - 所有成员都要重新请求加入到消费组
  - CompletingRebalance
    - 所有成员已经加入
    - 等待分配方案
  - Stable
    - 已正常消费数据
- Rebalance本质上是一种协议
  - 规定了一个 Consumer Group 下的所有 Consumer 如何达成一致，来分配订阅 Topic 的每个分区
- 时机
  - 组成员数量发生变化
  - 订阅主题数量发生变化
  - 订阅主题的分区数发生变化
- Consumer 实例已挂从而要退组
  - 定期地向 Coordinator 发送心跳请求
    - 心跳超时
    - 心跳频率
  - Consumer 端应用程序两次调用 poll 方法的最大时间间隔
    - 超时没消费完，就会离开消费组

### 位移主题Offsets Topic

- key保存的信息
  - Group ID
  - 主题名
  - 分区号
  - 其它
- 提交位移的方式
  - 自动提交位移
  - 手动提交位移
- 专门的后台线程定期地巡检待 Compact 的主题
  - 看看是否存在满足条件的可删除数据
  - 这个后台线程叫 Log Cleaner

### 多线程消费组

- 每个线程维护专属的 KafkaConsumer 实例，负责完整的消息获取、消息处理流程
- 使用单或多线程获取消息，同时创建多个消费线程执行消息处理逻辑

### 副本 - Replica

- 好处
  - 提供数据冗余 - Kafka只有第一点的优点
  - 提供高伸缩性
  - 改善数据局部性
- 本质 - 一个只能追加写消息的提交日志
- Follower 异步拉取
- Follower 不对外提供服务，好处：
  - Read-your-writes
  - Monotonic Reads
- In-sync Replicas（ISR）
  - 与 Leader 同步的副本
  - 包括Leader
- Unclean 领导者选举
  - 把所有不在 ISR 中的存活副本都称为非同步副本
  - C 或 A

### Kafka处理请求

- Reactor
  - 1个Dispatcher
  - 工作线程池

### 控制器

- 主题管理（创建、删除、增加分区）
- 分区重分配
- Preferred 领导者选举
- 集群成员管理（新增 Broker、Broker 主动关闭、Broker 宕机）
- 提供给其余Broker数据服务

### 高水位和Leader Epoch

- 水位
  - 在时刻 T，任意创建时间（Event Time）为 T’，且 T’≤T 的所有事件都已经到达或被观测到
  - 水位是一个单调增加且表征最早未完成工作（oldest work not yet completed）的时间戳。
- Kafka的水位是消息位移，也加高水位
- 2个作用
  - 定义消息可见性，即用来标识分区下的哪些消息是可以被消费者消费的
  - 帮助 Kafka 完成副本同步
- LEO - Log End Offset
- Leader Epoch - 大致可以认为是 Leader 版本
  - Epoch。一个单调增加的版本号。每当副本领导权发生变更时，都会增加该版本号。小版本号的 Leader 被认为是过期 Leader，不能再行使 Leader 权力。
  - 起始位移（Start Offset）。Leader 副本在该 Epoch 值上写入的首条消息的位移

### Kafka Stream

- 5 个步骤 - 原子性
  - 读取最新处理的消息位移
  - 读取消息数据
  - 执行处理逻辑
  - 将处理结果写回到 Kafka
  - 保存位置信息

## 高级

### Exactly Once

- 通常策略
  - 下游系统具有幂等性
  - Kafka的At Least Once
- 事务机制的幂等性
  - 幂等性发送
    - Producer端
      - Producer ID
      - Sequence Number
    - Broker端维护一个序号
      - PID
      - Topic
      - Partition
  - 事务性保证
    - Transaction ID，内部用Producer ID
    - 获取一个单调递增的epoch，区分新老Producer