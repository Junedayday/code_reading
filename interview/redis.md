## 高性能

### 数据结构

#### 全局哈希表

变慢原因：

1. 哈希冲突 - 链式哈希
2. rehash - 哈希表2分配更大空间 -> 映射与拷贝 -> 释放哈希表1
   1. 弊端：线程阻塞
   2. 改进：渐进式rehash
      1. 来一个请求，迁移索引0中的entries到哈希表2
      2. 下一个请求，迁移索引1中的entries到哈希表2
      3. 如果没有请求，定期，如100ms

#### 数据类型

1. String - 简单动态字符串
   1. int - long类型，ptr指向int
      1. 8Byte元数据
      2. 8Byte的INT
   2. raw - 字符串且大于39字节，prt指向SDS（简单动态字符串）
      1. 8Byte元数据
      2. 8Byte的ptr
      3. SDS
   3. embstr - 字符串且小于等于39字节，redisObject紧跟sdshdr
      1. 8Byte元数据
      2. 8Byte的ptr
      3. SDS
2. List
   1. ziplist - 所有字符串元素长度都小于64字节，并且保存的元素数量小于512个
   2. linklist - 其余
3. Hash
   1. ziplist - 所有字符串元素长度都小于64字节，并且保存的元素数量小于512个
   2. hash - 其余
4. Set
   1. intset - 对象保存的所有元素都是整数值，保存的元素数量不超过512个
   2. hash - 其余
5. Sorted Set
   1. ziplist - 所有元素成员的长度都小于64字节，保存的元素数量小于128个
   2. skiplist - 其余

> 扩展：Bitmap、HyperLogLog 和 GEO

底层

1. SDS
   1. len 4Byte
   2. alloc 4Byte
   3. buf 具体大小 + \0结束符1Byte
2. ziplist
   1. zlbytes列表长度
   2. zltail尾部偏移
   3. zllen元素数量
   4. entries列表
      1. prev_len前一个entry长度，1Byte/5Byte
      2. len自身长度，4Byte
      3. encoding，1Byte
      4. content，实际数据
   5. zlend结束标志符
3. 全局哈希表的每项dictEntry（很占用空间）
   1. key指针 8Byte
   2. value指针 8Byte
   3. next指针 8Byte
   4. jemalloc分配时，会用2的幂次数

ziplist和intset意义 - 节约内存，缓存友好

#### 范围操作

SCAN渐进式遍历：SSCAN/HSCAN/ZSCAN

- SCAN cursor [MATCH pattern] [COUNT count]
- 返回new-cursor + key信息
- new-cursor = 0 时结束

### 单线程

#### 网络IO和键值对读写采用单线程：

- 内存中的高效数据结构
- 网络IO采用**多路复用**机制
  - sokcet - 主动套接字
  - listen - 监听套接字
  - accept - 已连接套接字
  - select/epoll提供了基于事件的回调机制
- 意义： 避免多线程开发的并发控制问题

#### Redis IO瓶颈：

1. 耗时操作
   1. 大key
   2. 复杂度高的命令+数据量大
   3. key集中过期
   4. 内存上限后淘汰
   5. AOF always刷盘
   6. 主从全量同步生成RDB
2. 并发：单线程读写客户端，无法利用多核（Redis6.0后多线程读写）

### AOF

AOF - Append Only File，写后日志，记录具体命令

#### 风险

1. 写入日志前重启，丢失
2. 写入操作阻塞主线程

#### 方案

1. 高性能：None
2. 高可靠：Always
3. 平衡：Everysec

#### 文件过大 - AOF重写

- 实现：直接根据最新数据库里最新的数据生成新日志
- 主线程后台fork出bgrewriteaof子进程，并内存拷贝
- 两处日志：AOF日志+子进程的AOF重写日志

#### bgrewriteaof性能风险

- fork子进程时，内存页表越大，fork阻塞时间越久
- fork后是Copy On Write的，先是共享内存，会逐渐分离，所以Big Key或者Huge Page会影响到业务

### RDB

RDB - Redis Database 数据快照

#### 快照范围

- save 主线程
- bgsave 子进程
  - COW - Copy On Write
  - 子进程读到的是修改前的数据

#### AOF+RDB选择

- 数据不能丢失 - AOF+RDB
- 允许分钟级别丢失 - RDB
- 只用AOF - everysec

### 异步机制

#### 阻塞点

- 客户端
  - 1 - 集合全量查询和聚合操作 - 复杂度O(N)
  - 2 - bigkey 删除操作
  - 3 - 清空数据库
- 磁盘交互
  - 4 - AOF 日志同步写 - 1~2ms
- 主从交互
  - 3 - 从库清空数据库
  - 5 - 加载 RDB 文件
- 切片集群实例交互

#### 异步的子线程

主线程启动后，创建3个子线程：

- AOF 日志写操作 - 对应阻塞点4
- 键值对删除（惰性删除）- 对应阻塞点2、3
- 文件关闭的异步执行

交互机制：主线程通过 **一个链表形式的任务队列** 和子线程进行交互

### CPU优化

#### CPU架构

- 逻辑核 - 超线程
- 物理核
  - 包含2个逻辑核
  - 共享1级+2级缓存
- CPU Socket 插槽
  - 包含多个物理核
  - 共享3级缓存
- NUMA
  - 跨Socket时，访问直连内存速度比远端内存快

#### 绑核

- 绑核（物理核）命令 - taskset
  - Redis实例
  - 网络中断处理程序
  - 两者绑在一个Socket上
  - 注意：CPU 核的编号
- 绑物理核
  - 让主线程、子进程、后台线程共享2个逻辑核，缓解 CPU 资源竞争
- 源码优化
  - 把子进程和后台线程绑到不同的 CPU 核上
  - Redis6.0 支持 CPU 核绑定的配置

### 响应延迟

#### 确认是否变慢

1. 查看响应延迟
2. 基于当前环境下的 Redis 基线性能
   1. –intrinsic-latency 检测时间，120s
   2. 运行时延迟 是 基线性能的 2 倍及以上，就可以认定 Redis 变慢了

#### 优化三大要素

1. Redis 自身的操作特性
   1. 慢查询命令
   2. 过期key操作
      1. 原理 - 如果超过 25% 的 key 过期了，则重复删除的过程，直到过期 key 的比例降至 25% 以下
      2. 方案 - EXPIREAT 和 EXPIRE 的过期时间参数加上一定大小范围内的随机数
2. 文件系统 - AOF
   1. AOF写回策略
      1. everysec - 后台子线程fsync
      2. always - 主线程中fsync
   2. AOF 重写
      1. 重写的压力比较大时，会导致 fsync 被阻塞，主线程也会阻塞
   3. appendfsync = yes
      1. 把写命令写到内存，可能会丢失
      2. 延迟敏感，但允许丢数据
   4. 高性能+高可靠数据保证
      1. 高速的固态硬盘作为 AOF 日志的写入设备
3. 操作系统
   1. swap
      1. 原因 - 物理机器内存不足
      2. 增加机器的内存或者使用 Redis 集群
   2. 内存大页
      1. 修改正在持久化的数据，需要写时复制
      2. 如果大页=2MB，这就需要复制2MB

## 高可靠

数据尽量少丢失（RDB+AOF），服务尽量少中断

### 数据同步

#### 第一次同步

1. 建立连接，协商同步
   1. 请求 - psync 主库runID+复制进度offset（-1）
   2. 响应 - fullresync 主库runID+主库当前进度offset
2. 主库发送RDB文件，从库清空后加载RDB
   1. 主库在`replication buffer`中记录RDB文件生成后的写操作
   2. 使用RDB而不是AOF的原因：一个是文件小，另一个是加载快
3. 主库把`replication buffer`中的写操作发给从库，从库执行

> 一个Redis不要太大，控制在几GB。

#### 基于长连接的命令传播

1. 2.8之前网络断了，就要全量同步
2. 2.8之后网路欧断了，主库会把写操作写入 `replication buffer`，也会写入`repl_backlog_buffer`
   1. 环形缓冲，master_repl_offset 和 slave_repl_offset
   2. 恢复连接后，从库发psync + slave_repl_offset，主库判断全量同步还是增量同步，并发送对应的操作
   3. 缓冲空间 = （写入命令速度 - 主从网络传输速度） * 命令平均大小
   4. repl_backlog_size = 缓冲空间 * 2~4

> replication buffer 是为了主从全量复制，主库上和从库客户端连接的buffer，非共享
>
> repl_backlog_buffer 是为了支持从库增量复制，主库上专门的buffer，共享

### 哨兵机制

#### 哨兵的三个任务

- 监控 - 周期性PING，异常则自动切换主库
  - 哨兵PING主库超时，可能存在误判 - 主观下线
  - 采用哨兵集群，超过一半主观下线 - 客观下线
- 选主
  - 初筛 - 当前连接状态+之前的连接状态
  - 三轮打分
    - 根据slave-priority配置优先级
    - 根据和旧主的同步进度slave_repl_offset
    - 实例id最小
- 通知 - 通知slave对新主执行replicaof，同时客户端连接新主

#### 哨兵集群 - pub/sub机制

- 订阅频道 - `__sentinel__:hello`
- 用INFO命令获取slave列表，进行slave监控
- 执行主从切换 - 哨兵选举
  - 投完票后，会拒绝其余投票，包括自己
  - 失败后会等待 固定+随机波动 的时间
- 主从切换后，哨兵要告诉client新的主库

#### 客户端事件通知 - pub/sub机制

- pub/sub 机制
- Client向哨兵订阅事件

## 高可扩展

### 切片集群

#### Redis Cluster - Hash Slot

- 数量：16384，即2的14次方
- 对key进行CRC16，然后对这个16bit进行 mod  16384
- 不同性能机器如何分配均匀
  - 手动分配16384个
- 重定向机制 - MOVED
  - 未完成迁移 - ASK报错

> 为何不用一张表保存key和实例的关系：
>
> 1. 修改频繁的性能问题
> 2. 额外存储空间

## 技巧

### 二级编码

对应大量的k-v的存储：用ziplist节约空间

1. k前 7 位作为Hash类型的键
2. k后3位作为Hash类型的值的key
   1. 最多有1000个key
   2. hash-max-ziplist-entries设置为1000
3. v作为Hash类型的值的value

### 集合统计

#### 聚合统计

- 并集 - SUNIONSTORE
- 差集 - SDIFFSTORE
- 交集 - SINTERSTORE

选择从库分析，或读取到客户端后分析

#### 排序统计

- List - 频繁更新时，分页出问题
- Sorted Sort 优先

#### 二值状态统计

Bitmap

- SETBIT key offset value
- GETBIT key offset
- BITCOUNT
- BITOP operation destkey key
  - operation - AND 、 OR 、 NOT 、 XOR

#### 基数统计

- SET 存放用户 - 过大
- Hash 存放用户 - 也很大
- HyperLogLog PFADD/PFCOUNT - 数据量大时，空间总是固定的
  - 计算日活、7日活、月活数据

### 时间序列数据

- Hash+Sorted Set
  - Hash - 单个查询
  - Sorted Set - 范围查询
  - 原子性- 命令MULTI+EXEC
  - 客户端聚合
- 扩展模块 RedisTimeSeries
  - loadmodule redistimeseries.so

### 消息队列

-  List 
  - LPUSH+RPOP
  - 消费者需要循环RPOP，改进为阻塞的BRPOP
  - 消费者没有ACK机制，改进BRPOPLPUSH（即备份List）
- PubSub
  - PubSub只把数据发给在线的消费者，所以消费者重启后会丢数据
  - 不支持数据持久化，只是基于内存的多播机制
- Streams
  - XADD：插入消息，保证有序，可以自动生成全局唯一 ID
  - XREAD：用于读取消息，可以按 ID 读取数据；可以设置阻塞block 10000 (ms)
  - XGROUP：创建消费组
  - XREADGROUP：按消费组形式读取消息
  - XPENDING：可以用来查询每个消费组内所有消费者已读取但尚未确认的消息
  - XACK ：用于向消息队列确认消息处理已完成
  - 保证重启后仍读取未处理的消息：自动使用内部队列，也称为 PENDING List

### 内存碎片

- 原因
  - 内因 - 内存分配器的分配策略
    - 默认jemalloc
  - 外因 - 键值对大小不一样和删改操作
- 如何判断
  - mem_fragmentation_ratio 内存碎片率
    - used_memory_rss / used_memory
    - 正常情况1~1.5
    - 小于1表示用到了swap
- 清理 - 重启
- 清理 - 自动清理4.0
  - activedefrag配置为yes
  - active-defrag-ignore-bytes，即内存碎片达到指定大小
  - active-defrag-threshold-lower，即内存碎片空间占比到指定大小
  - active-defrag-cycle-min/active-defrag-cycle-max CPU 时间的比例范围

### 缓冲区溢出

#### 客户端缓冲区 - Redis服务端

- 输入 - 请求命令
  - 原因
    - 写入了 bigkey
    - 服务端处理过慢
  - CLIENT LIST
    - qbuf，表示输入缓冲区已经使用的大小
    - qbuf-free，表示输入缓冲区尚未使用的大小
- 输出 - 返回结果
  - 16KB 的固定缓冲空间+动态增加的缓冲空间
  - 原因
    - 服务器端返回 bigkey 的大量结果
    - 执行了 MONITOR 命令，会持续占用输出缓冲区 - 不要在生产环境执行MONITOR
    - 缓冲区大小设置得不合理
  - 配置
    - 缓冲区大小限制
    - 持续写入量限制
    - 持续写入时间限制
  - 普通客户端 - 都不限制
  - 订阅客户端，例如
    - 输出缓冲区的大小上限为 8MB
    - 连续 60 秒内对输出缓冲区的写入量超过 2MB

#### 主从集群缓存区

- 全量复制 - 复制缓冲区
  - 主节点在向从节点传输 RDB 文件时，客户端的写命令先保存在复制缓冲区中
  - 溢出则会关闭和从节点的复制操作
  - 配置 client-output-buffer-limit，例如
    - 缓冲区大小的上限512MB
    - 连续 60 秒内的写入量超过 128MB 的话
  - 避免方法
    - 控制主节点数据量，如2~4GB
    - 设置 client-output-buffer-limit
- 增量复制 - 复制积压缓冲区
  - repl_backlog_buffer

### db与缓存

- 缓存
  - CPU - LLC
  - 内存 - page cache
  - 磁盘
- 三种缓存模式
  - 只读缓存模式
  - 采用同步直写策略的读写缓存模式 - client更新db
  - 采用异步写回策略的读写缓存模式 - redis更新db


#### 数据一致性

- 定义
  - 缓存有数据，必须和db一致
  - 缓存没有数据，db必须为最新
- 数据不一致的问题
  - 删除缓存值或更新数据库失败，使用重试机制确保成功
  - **延迟双删** - 在删除缓存值、更新数据库的这两步操作中，有其他线程的并发读操作，导致其他线程读取到旧值
    - redis.delKey(X)
    - db.update(X) - 此时可能被另一个线程更新旧值到redis
    - Thread.sleep(N) - 意义不大，高并发时不确定时间
    - redis.delKey(X)

#### 缓存雪崩avalanche - 大量数据同时失效

1. 情况1 - 大量数据同时过期
   1. 过期时间加随机数
   2. 服务降级 - 非核心直接返回，核心去数据库
2. 情况2 - 缓存实例宕机
   1. 业务系统中实现服务熔断或请求限流机制
   2. 缓存集群

#### 缓存击穿breakdown - 热点数据

不设置过期时间

#### 缓存穿透penetration - 不在redis和db

缓存成为摆设，主要两个场景：恶意攻击和误删

- 缓存空值或缺省值
- 布隆过滤器
- 在请求入口进行合法性检测，避免恶意请求

### 替换策略

把缓存容量设置为总数据量的 15% 到 30%，兼顾访问性能和内存空间开销

#### 淘汰策略

- 不淘汰 - noeviction
- 设置了过期时间的数据中进行淘汰
  - volatile-random
  - volatile-ttl 越早过期的越先被删除
  - volatile-lru
  - volatile-lfu
- 在所有数据范围内进行淘汰
  - allkeys-lru
  - allkeys-random
  - allkeys-lfu
- 优先使用： allkeys-lru 或 volatile-lru

### 缓存污染

留存在缓存中的数据，实际不会被再次访问了，但是又占据了缓存空间

#### LRU

扫描式单次查询的数据，会造成缓存污染

避免链表开销，采用了两个近似方法：

- RedisObject里lru字段记录数据的访问时间戳
- 随机采样，根据lru字段筛选

#### LFU

先比较访问次数，再比较访问时间

24bit的lru字段拆分：

1. ldt，前16bit，记录时间戳
2. counter，后8bit，记录访问次数
   1. 上限为255
   2. 计数规则优化：lfu_log_factor参数

### Pika

固态硬盘SSD

#### 大内存 Redis 实例的潜在问题

- RDB文件大
  - fork时间增加
  - 恢复时间增加
- 主从同步
  - 全量同步
  - 加载时间

#### Pika特点

- RocksDB 保存在SSD中，可直接读
- 优势
  - 实例重启快
  - 全量同步的风险低
- 降低数据的访问性能

### 并发访问

#### 原子操作

1. 单命令操作 - 多操作实现为一个操作
   1. INCR
   2. DECR
2. Lua脚本 - 多操作写到脚本

### 分布式锁

#### 单节点Redis

- SETNX lock_key 1
  - 优化 - SET lock_key unique_value NX PX 10000
- DEL lock_key
  - 优化 - Lua脚本
  - 先取值，再比较unique_value是否相等，避免误释放

#### 分布式节点 - Redlock算法

1. 客户端获取当前时间
2. 客户端按顺序依次向 N 个 Redis 实例执行加锁操作
   1. SET EX PX + 客户端唯一标识
3. 一旦客户端完成了和所有 Redis 实例的加锁操作，客户端就要计算整个加锁过程的总耗时
   1. 加锁成功的两个条件
      1. 客户端从超过半数（大于等于 N/2+1）的 Redis 实例上成功获取到了锁
      2. 客户端获取锁的总耗时没有超过锁的有效时间
   2. 重新计算这把锁的有效时间 = 锁的最初有效时间 - 客户端为获取锁的总耗时

### 事务机制

#### Redis的事务

1. MULTI
2. 多个命令
3. EXEC

#### ACID

- 原子性
  - 命令入队时就报错，会放弃事务执行，保证原子性
  - 命令入队时没报错，实际执行时报错，不保证原子性
  - EXEC 命令执行时实例故障，如果开启了 AOF 日志，可以保证原子性
  - DISCARD 只能清空命令队列
- 一致性
  - 命令入队时就报错，保证一致性
  - 命令入队时没报错，实际执行时报错，保证一致性
  - EXEC 命令执行时实例发生故障，RDB和AOF模式都能保证一致性
- 隔离性
  - 并发操作在 EXEC 命令前执行：隔离性的保证要使用 WATCH 机制来实现，否则隔离性无法保证
    - WATCH机制：监控一个或多个键，如果修改了就放弃执行
  - 并发操作在 EXEC 命令后执行：隔离性可以保证
- 持久性
  - RDB和AOF都得不到保证

### 主从同步与故障切换的问题

- 主从数据不一致
  - 保证网络条件
  - 开发外部程序监控复制进度
- 读取到过期数据
  - Redis同时使用到了两种策略
    - 惰性删除 - 再次读时才删除
    - 定期删除 - 采样删除
  - 主库会删除，从库分版本
    - 3.2前从库不会判断过期，直接返回
    - 3.2后从库会返回空值
  - 从库key过期时间=主从同步完成后（可能有延迟）+主库中的过期时间
    - 尽可能使用EXPIREAT
- 主从切换时
  - 哨兵protected-mode设为no，允许外部服务器访问
  - cluster-node-timeout实例响应心跳消息的超时时间：调大，10s~20s

### 脑裂

- min-slaves-to-write：主库能进行数据同步的最少从库数量
  - K/2+1
- min-slaves-max-lag：主从库间进行数据复制时，从库给主库发送 ACK 消息的最大延迟（以秒为单位）
  - 10～20s
- 主库连接的从库中至少有  min-slaves-to-write  个从库，和主库进行数据复制时的 ACK 消息延迟不能超过 Tmin-slaves-max-lag秒

### Codis

- 集群分布
  - 1024个slot，手动或自动分配给codis server
  - 路由表由codis proxy缓存
  - key和slot的映射是通过CRC32计算的
- 对比：
  - 数据路由信息
    - Codis的中心保存在zk，proxy在本地缓存
    - Redis Cluster要在所有实例中保存
- codis-server扩容
  - 同步 - 阻塞server
  - 异步
    - 迁移过程中数据会被标记为只读
    - 对bigkey每个元素用一条指令迁移，过程中崩溃会破坏原子性
  - Redis Cluster只支持同步

### 秒杀

秒杀进行中，需要查验和扣减商品库存

- 库存查验面临大量的高并发请求
  - 切片集群，用不同的实例保存不同商品的库存
- 库存扣减又需要和库存查验一起执行，以保证原子性
  - 原子性的 Lua
  - 使用分布式锁，只有拿到锁的客户端才能执行库存查验和库存扣减

### 数据分布优化

- 两类数据倾斜
  - 数据量倾斜
    - bigkey - 应用代码优化
    - Slot 分配不均衡 - Slot迁移
    - Hash Tag 导致倾斜
      - 只对key的{}计算CRC16
      - Hash Tag为了运行在单实例上，从而支持事务操作和范围查询
  - 数据访问倾斜 - 热点数据
    - 热点数据多副本
      - key加随机前缀
      - 热点数据必须只读

### 通信开销

- Gossip 协议
  - 通信消息大小
    - 一个消息 104 字节
    - 传递集群10%实例的状态
    - 16384Bit的Slot信息
  - 通信频率
    - 每秒从本地的实例列表中随机选出 5 个实例，从中找出一个最久没有通信的实例发 PING
    - 每 100ms 一次的频率，扫描本地的实例列表PONG有超时的，发送PING
      - 调大为20 秒或 25 秒

### Redis 6.0

- 多线程
  - 多个 IO 线程来处理网络请求
  - 主线程 - IO线程 - 主线程
- Tracking - 客户端缓存
  - 普通模式 - key发生变化，服务端会给客户端发送 invalidate 消息
  - 广播模式 - 广播所有 key 的失效情况
- 以用户为粒度设置命令操作的访问权限
- RESP 3 通信协议
  - 支持多种数据类型的区分编码
  - 实现：用不同的开头字符区分不同的数据类型

### NVM

- 特点
  - 能持久化保存数据
  - 读写速度和 DRAM 接近
  - 容量大
- Memory 模式
  - 利用 NVM 容量大的特点，实现大容量实例，保存更多数据
- App Direct 模式
  - 直接在持久化内存上进行数据读写
  - 不需要RDB和AOF

### 规范

- 键值对使用规范
  - 规范一：key 的命名规范
    - 业务名前缀+冒号+具体业务数据名
  - 规范二：避免使用 bigkey
    - String 类型的数据大小控制在 10KB 以下
    - 集合类型的元素个数控制在 1 万以下
  - 规范三：使用高效序列化方法和压缩方法
  - 规范四：使用整数对象共享池
    - 内部维护了 0 到 9999 这 1 万个整数对象
- 数据保存规范
  - 规范一：使用 Redis 保存热数据
  - 规范二：不同的业务数据分实例存储
  - 规范三：在数据保存时，要设置过期时间
  - 规范四：控制 Redis 实例的容量
- 命令使用规范
  - 规范一：线上禁用部分命令
    - KEYS
    - FLUSHALL
    - FLUSHDB
    - 用rename-command重命名
  - 规范二：慎用 MONITOR 命令
  - 规范三：慎用全量操作命令
    - SCAN
    - 业务侧拆分
    - 序列化







