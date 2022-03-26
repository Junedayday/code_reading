

## 自我介绍
## 离职原因 
内部环境问题

## 看机会重点
看重 大环境一定好，内卷不要太严重

## 工作内容
- 业务需求
- 平台建设
- 稳定性

## 技术
- 线程池实现原理
  - dispatcher
  - Workers
- redis缓存
  - 临时缓存
  - 持久缓存
    - RDB+AOF
  - 分布式锁
    - 单节点
      - SET lock_key unique_value NX PX 10000
      - DEL lock_key
        - 优化 - Lua脚本
        - 先取值，再比较unique_value是否相等，避免误释放
    - 多节点 - Redlock算法
      - 客户端获取当前时间
      - 客户端按顺序依次向 N 个 Redis 实例执行加锁操作
        1. SET EX PX + 客户端唯一标识
      - 一旦客户端完成了和所有 Redis 实例的加锁操作，客户端就要计算整个加锁过程的总耗时
        1. 加锁成功的两个条件
           1. 客户端从超过半数（大于等于 N/2+1）的 Redis 实例上成功获取到了锁
           2. 客户端获取锁的总耗时没有超过锁的有效时间
        2. 重新计算这把锁的有效时间 = 锁的最初有效时间 - 客户端为获取锁的总耗时
- rocketmq 
  - 数据不丢如何实现
    - producer 注意同步和异步，保证收到消息返回
    - broker 同步刷盘，存储成功才返回给producer
      - master-slave 同步复制
    - consumer 消费并处理成功后才响应
- mysql
  - 事务隔离级别
    - 读未提交（read uncommitted）
      - 脏读，事务B读到事务A未提交的
    - 读提交（read committed）
      - 不可重复读，事务B读到一条记录，是两份数据（因为事务A提交）
    - 可重复读（repeatable read）
      - 幻读，事务B读到一个范围内的记录，是两份数据（因为事务A提交）
        - 专指“新插入的行”
    - 串行化（serializable ）
  - 覆盖索引
    - 索引下推
  - 锁
    - 全局锁
      - Flush tables with read lock (FTWRL)
    - 表级锁
      - 表锁
        - lock tables … read/write
      - 元数据锁（meta data lock，MDL)
        - 原理
          - 读读不互斥
          - 读写、写写互斥
        - 作用 - 防止DDL和DML并发的冲突
        - AliSQL语法：设定等待时间NOWAIT/WAIT n
      - Online DDL
        - 拿MDL写锁
        - 降级成MDL读锁
        - 真正做DDL - block
        - 升级成MDL写锁 
        - 释放MDL锁
    - 行锁 - 引擎实现
      - 事务的两阶段锁
        - 可能会冲突的、影响并发的锁，尽量往后放
      - 死锁的两个策略
        - 直接进入等待，直到超时innodb_lock_wait_timeout
        - 死锁检测，回滚某一个事务：很消耗CPU资源
          - 控制并发度
    - 间隙锁 (Gap Lock)
      - 解决幻读
      - 前开后开 - （a,b）
    - next-key lock
  - 主从同步
    - 主库binlog
    - dump thread
    - 网络
    - io thread 与主库连接
    - 中转日志 relay log
    - sql_thread 执行

## 高并发

## 稳定性

- 服务治理
  - 隔离
  - 超时
  - 限流
  - 熔断
  - 降级
  - 重试
  - 负载均衡

## 线上问题排查 cpu 100% 