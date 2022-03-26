## 基本架构

### SQL查询

- Server 层
  - 1 - 连接器
    - 账号认证，并保存权限信息、用于后面的校验
    - 执行过程中的临时内存是管理在连接对象里面的
  - 2 - 查询缓存
    - key-value
    - 不建议使用：表被更新会全部清理
  - 3 - 分析器
    - 词法分析
    - 语法分析
      - 包括合法性的校验
  - 4 - 优化器
    - 索引选择
    - 各个表的join顺序选择
  - 5 - 执行器
    - 校验用户对该表的权限
    - 会根据表的引擎，使用对应的接口

- 存储引擎层 - 插件式
  - InnoDB
  - MyISAM
  - Memory

### SQL更新

- redo log

  - Write-Ahead Logging

  - InnoDB 引擎特有

  - 物理日志，数据页的修改

  - 固定空间循环写

- binlog

  - 位于Server 层

  - 逻辑日志

  - 追加写入

- 执行顺序

  - 引擎更新数据
  - redo log prepare
    - 此时崩溃 - 事务回滚
  - 执行器生成binlog
    - 此时崩溃 - 事务提交
  - redo log commit

### 事务隔离

- 隔离级别
  - 读未提交（read uncommitted）
    - 脏读，事务B读到事务A未提交的
  - 读提交（read committed）
    - 不可重复读，事务B读到一条记录，是两份数据（因为事务A提交）
  - 可重复读（repeatable read）
    - 幻读，事务B读到一个范围内的记录，是两份数据（因为事务A提交）
      - 专指“新插入的行”
  - 串行化（serializable ）
- 隔离的实现
  - 回滚段
    - 多个回滚操作，对应多个read-view
    - 没有更老的事务后，会删除
- 事务启动瞬间trx_id的三类数据（注意，真实数据的版本，不一定根据trx_id递增的，存在先启动，但后提交的情况）
  - 已提交的事务 - 可见
  - 中间情况
    - 没提交的事务生成 - 不可见
    - 已提交的事务生成 - 可见
  - 未开始的事务 - 不可见
- 当前读（current read）
  - 更新数据都是先读后写的
  - 当前的记录的行锁被其他事务占用的话，就需要进入锁等待



### 索引

- 数据结构
  - 哈希表
    - 等值查询
  - 有序数组
    - 静态存储，不会再修改
  - 二叉树
    - N叉树，如N=1200
- InnoDB 的索引模型 - B+树
  - 主键索引 - 聚簇索引（clustered index）
  - 非主键索引 - 二级索引（secondary index）
- 页分裂
- 覆盖索引
  - 索引下推 - 将查询字段放到索引里
- 最左前缀原则

#### 普通索引和唯一索引

- 查询 k = 1
  - 普通索引 - 查到到第一个匹配就结束
  - 唯一索引 - 查到第一个不匹配才结束
  - B+树上，性能差别很小
- 更新
  - change buffer
    - 缓存更新
    - 持久化：内存+磁盘
    - 应用到数据页 - merge：访问数据页、定期
  - 唯一索引的更新就不能使用 change buffer
    - 需要读取数据，判断是否冲突
  - redo log 主要节省的是随机写磁盘的 IO 消耗（转成顺序写），而 change buffer 主要节省的则是随机读磁盘的 IO 消耗

#### 索引选择

- 采样统计
- 三种方法
  - force index
  - 修改语句，引导 MySQL 使用我们期望的索引
  - 新建一个更合适的索引，来提供给优化器做选择，或删掉误用的索引
- 选择失效
  - 对索引字段做函数操作，破坏有序性
  - 隐式类型转换
  - 隐式字符编码转换，如utf8/utf8mb4

#### 字符串索引

1. 前缀索引 - 子字符串
2. 字符串倒序 - 如身份证号
3. hash字段



### 锁

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
  - 间隙锁+行锁 = 前开后闭 （a, b]
  - 会导致同样的语句锁住更大的范围，这其实是影响了并发度的
- 可重复读的加锁规则
  - 原则 1：加锁的基本单位是 next-key lock
  - 原则 2：查找过程中访问到的对象才会加锁
  - 优化 1：索引上的等值查询，给唯一索引加锁的时候，next-key lock 退化为行锁
  - 优化 2：索引上的等值查询，向右遍历时且最后一个值不满足等值条件的时候，next-key lock 退化为间隙锁
  - 一个 bug：唯一索引上的范围查询会访问到不满足条件的第一个值为止。



## 基本原理

### 数据同步

#### 脏页

#### 数据空洞

- 删除
- 插入，数据页分裂

### 排序

- sort_buffer
  - 内存
  - 外存
- 单行数据太大时 - rowid
  - 需要再回到原表去取数据
- 随机排序
  - order by rand()
    - Using temporary
    - Using filesort
    - 代价比较大
  - 程序内随机
    - 挑选出id后，去数据库选择



## 知识点串联

### 查询慢

- 查询长时间不返回
  - 等 MDL 锁
  - 等 flush - flush语句很快，一般是被另一个操作阻塞了
  - 等行锁
- 查询慢
  - 没有索引
  - 长事务，有很多undo log
    - lock in share mode - 当前读
    - 一致性读，需要执行这些undo log

### 引发性能问题的慢查询

- 索引没有设计好
  - Online DDL
  - 备库 - 关闭binlog，加上索引，主备切换
  - 主库 - 关闭binlog，加上索引
- SQL 语句没写好
  - 如索引字段加函数
- MySQL 选错了索引

### 数据不丢失

- binlog
  - 一个线程一个binlog cache
  - write - 写到page cache里
  - fsync - 持久化到磁盘
  - sync_binlog
    - 0 - 只 write，不 fsync
    - 1  - 每次提交事务都会执行 fsync
    - N(N>1) - 每次提交事务都 write，但累积 N 个事务后才 fsync
- redo log
  - innodb_flush_log_at_trx_commit
    - 0 - 只留在 redo log buffer
    - 1 - 每次事务提交都将 redo log 直接持久化到磁盘
    - 2 - 每次事务提交都只是把 redo log 写到 page cache
  - 后台线程，每隔 1 秒会将page cache 调用fsync持久化到磁盘

### 主备切换

- 可靠性优先流程
  - 判断备库 B 现在的 seconds_behind_master 是否小于某个值（比如 5 秒）继续下一步，否则持续重试这一步
  - 把主库 A 改成只读状态，即把 readonly 设置为 true
  - 判断备库 B 的 seconds_behind_master 的值，直到这个值变成 0 为止
  - 把备库 B 改成可读写状态，也就是把 readonly 设置为 false
  - 把业务请求切到备库 B
- 可用性优先策略
  - 直接切换，并持续同步binlog
- 会导致数据不一致
- GTID

### 备库延迟

- 并行复制
  - 按表分发策略
- 过期读的方案
  - 强制走主库
  - sleep
  - 判断主备无延迟
    - show slave status
  - 配合 semi-sync
    - binlog发给从库 -> ack主库  -> 返回客户端
  - 等主库位点
  - 等 GTID
    - wait_for_executed_gtid_set

### 判断数据库是否正常

- select 1
- 查表判断
  - 如创建health_check表
- 更新判断
- 内部统计

