---
title: Go语言技巧 - 7.【GORM实战剖析】基本用法和原理解析
date: 2021-09-15 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![Go-Study](https://i.loli.net/2021/05/05/2bmr98tG3xDneL5.jpg)

## GORM库的官方文档

GORM库作为Go语言最受欢迎的ORM框架，提供了非常丰富的功能，大家可以通过阅读[中文官网](https://gorm.io/zh_CN/docs/index.html)了解详情。

这里，先着重介绍一个背景：**GORM内部会区分v1与v2两个版本**，其中

- v1的包导入路径为 `github.com/jinzhu/gorm`
- v2的包导入路径为 `gorm.io/gorm`

v1与v2对使用者来说体验相差不大，今天就主要针对v2版本进行讲解。

<!-- more -->

## Talk is Cheap. Show me the code.

接下来，我先给出一套个人比较推荐的CRUD代码。

### 创建

[官方链接 - 创建](https://gorm.io/zh_CN/docs/create.html)

```go
user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
// 直接创建
result := db.Create(&user)
// 指定字段创建
db.Select("Name", "Age", "CreatedAt").Create(&user)

// 批量创建
var users = []User{{Name: "jinzhu1"}, {Name: "jinzhu2"}, {Name: "jinzhu3"}}
db.Create(&users)
```

推荐：

1. 通常：直接用**结构体**或**结构体的切片**进行创建；
2. 特殊：加上指定的字段，也就是其余字段不生效，如上面的`Birthday`。

### 查询

[官方链接 - 查询](https://gorm.io/zh_CN/docs/query.html)

```go
// 查询所有对象
var users []User  
result := db.Find(&users)

// 指定查询条件（where）
db.Where(&User{Name: "jinzhu"}, "name", "Age").Find(&users)

// 限制返回数量
db.Limit(10).Offset(5).Find(&users)

// 查询部分字段（即从select * 改造为 select name, age）
db.Select("name", "age").Find(&users)

// 其余扩展
db.Order("age desc, name").Find(&users)
```

推荐：

1. 普通场景：简单查询用`Find+Where`的函数结合实现，结合`Limit+Offset+Order`实现分页等高频功能；
2. 追求性能：可以引入`Select`避免查询所有字段，但会导致返回结果部分字段不存在的奇怪现象，需要权衡；
3. 复杂查询：例如`Join+子查询`等，推荐使用下面的原生SQL，用GORM拼接的体验并不好。

### 更新

[官方链接 - 更新](https://gorm.io/zh_CN/docs/update.html)

```go
// 更新通常包含两块，一个是要更新的字段Select+Updates，另一个是被更新数据的条件Where
db.Model(&user).Where(&User{Name: "jinzhu"}, "name", "Age").Select("Name", "Age").Updates(User{Name: "new_name", Age: 0})
```

> 零值问题：参考https://gorm.io/zh_CN/docs/update.html#%E6%9B%B4%E6%96%B0%E5%A4%9A%E5%88%97 下的注释

推荐：

1. 普通场景：利用`Select+Updates`指定更新字段，利用`Where`指定更新条件；
2. 特殊场景：复杂SQL用原生SQL。

### 删除

[官方链接 - 删除](https://gorm.io/zh_CN/docs/delete.html)

```go
// 删除条件不建议太复杂，所以可以用简单的Where条件来拼接
db.Where("email LIKE ?", "%jinzhu%").Delete(Email{})
```

推荐：

1. 普通场景：利用`Where`限定删除条件，不建议太复杂；
2. 软删除：在实际项目中，不太建议用`硬删除`的方式，而是用`软删除`，即更新一个标记字段。

### 原生SQL

```go
// 原生SQL，推荐在复杂sql场景下使用
db.Raw("SELECT id, name, age FROM users WHERE name = ?", 3).Scan(&result)
```

## 使用GORM的核心思路梳理

### 一个对象 = 一行数据

示例中的一个`User`对象，完整地对应到具体`users`表中的一行数据，让整个框架更加清晰明了。每当数据库增加了一列，就对应地在结构体中加一个字段。这里有两个注意点：

1. 不要在核心结构体`User`中加入非表中的数据，如一些计算的中间值，引起二义性；
2. [gorm.Model](https://gorm.io/zh_CN/docs/models.html#gorm-Model)可以提升编码效率（会减少重复编码），但会限制数据库表中字段的定义，慎用（个人更希望它能开放成一个接口）；

### 选择生效字段 = 核心结构体 + 字段数组

在 **查询** 和 **更新** 接口里，我推荐的使用方法是采用核心结构体`User`+一个fields的数组，前者保存具体的数据、也实现了结构体复用，后者则选择生效的字段。

这种风格代码和Google推荐的[API风格](https://cloud.google.com/apis/design/standard_methods#update)非常像，可读性很棒。

> 这里还遗留了一个问题，就是fields数组里的字符串必须手输，可以考虑结合go generate自动生成这些fields的字符串常量，减少出错的概率。

### 缩短链式调用

GORM的主要风格是[链式调用](https://gorm.io/zh_CN/docs/method_chaining.html)，类似于Builder设计模式、串联堆起一个SQL语句。这种调用方式扩展性很强，但会带来了一个很严重的问题：容易写出一个超长的链式调用，可维护成本大幅度提高。

所以，在我的推荐使用方式里，区分了两种场景：

1. 简单场景 - **核心结构体 + 字段数组**
2. 复杂场景 - **原生SQL**

### 聚焦微服务的场景

作为一个`ORM`工具，GORM要考虑兼容各种SQL语句，内部非常庞大的。但如今更多地是考虑微服务的场景，这就能抛开大量的历史包袱，实现得更加简洁。这里我简单列举三个不太推荐使用的SQL特性：

1. 减少group by - 考虑将聚合字段再单独放在一个表中
2. 抛弃join - 多表关联采用多次查询（先查A表，然后用In语句去B表查）、或做一定的字段冗余（即同时放在A、B两个表里）
3. 抛弃子查询，将相关逻辑放在代码里

当然，真实业务研发过程中无法完全避免复杂SQL，我们只能有意识地减少引入复杂度。

### 避免引入非原生MySQL的特性

GORM除了常规的SQL功能，还提供了一些[高级特性](https://gorm.io/zh_CN/docs/models.html#%E9%AB%98%E7%BA%A7%E9%80%89%E9%A1%B9)、[模型关联](https://gorm.io/zh_CN/docs/belongs_to.html)、[钩子](https://gorm.io/zh_CN/docs/hooks.html)等，非常炫酷。

但我不推荐大家在实际项目中使用这些特性。只有尽可能地保证这个框架简洁，才能保证代码后续的可维护性。

熟悉MySQL历史的朋友都知道，**存储过程**在以前相当一段时间都是很好的工具，但如今都倡导**去存储过程**。GORM的这些特性和存储过程有异曲同工之处：一个将业务逻辑放在了数据库，另一个则放到了ORM框架里，会导致后续的迁移成本变高。

> 这也是我不推荐使用 gorm.Model的重要原因。

## 从查询接口了解GORM的核心实现

### 两个核心文件

在GORM库中，有两个核心的文件，也是我们调用频率最高的函数所在：**chainable_api.go**和 **finisher_api.go**。顾名思义，前者是整个链式调用的中间部分，后者则是最终获取结果的函数。以查询为例：

```go
db.Where(&User{Name: "jinzhu"}, "name", "Age").Find(&users)
```

其中`Where`是chainable，也就是还在拼接SQL条件，`Find`则是触发真正查询的finisher。

如果一开始过于关注chainable调用，很容易陷入构造SQL的细节，所以这块代码建议从finisher入手，深入看看一个SQL的到底是怎么在GORM中拼接并执行的。

### Find的调用链路

#### 1. Find的主要代码

```go
func (db *DB) Find(dest interface{}, conds ...interface{}) (tx *DB) {
	tx = db.getInstance()
  // conds是查询的条件，这里忽略，我们默认已经在前面的Chainable中完成了所有参数的拼接
	if len(conds) > 0 {
		if exprs := tx.Statement.BuildCondition(conds[0], conds[1:]...); len(exprs) > 0 {
			tx.Statement.AddClause(clause.Where{Exprs: exprs})
		}
	}
	tx.Statement.Dest = dest
  // 关键的执行逻辑
	return tx.callbacks.Query().Execute(tx)
}
```

#### 2. tx.callbacks.Query()的实现

```go
func (cs *callbacks) Query() *processor {
  // Query 是从processors的 map 中取出 query
	return cs.processors["query"]
}

// 这个对应的processor是 gorm.DB，也就是执行DB的Execute
func initializeCallbacks(db *DB) *callbacks {
	return &callbacks{
		processors: map[string]*processor{
			"create": {db: db},
			"query":  {db: db},
			"update": {db: db},
			"delete": {db: db},
			"row":    {db: db},
			"raw":    {db: db},
		},
	}
}
```

#### 3. Execute的执行逻辑

抛开一些周边逻辑，我们聚焦于下面的核心逻辑：

```go
func (p *processor) Execute(db *DB) *DB {

	// processor中注册了多个函数，按顺序执行。
  // 核心的查询逻辑也在这里面
	for _, f := range p.fns {
		f(db)
	}

	return db
}
```

而fns又是来自callbacks

```go
func (p *processor) compile() (err error) {
  // 对 callbacks 会做排序
	if p.fns, err = sortCallbacks(p.callbacks); err != nil {
		p.db.Logger.Error(context.Background(), "Got error when compile callbacks, got %v", err)
	}
	return
}
```

#### 4. Callback的注册

```go
func RegisterDefaultCallbacks(db *gorm.DB, config *Config) {
  // 默认注册了create/query/delete/update/raw 五种 callback 大类，这里以query为例
	queryCallback := db.Callback().Query()
	queryCallback.Register("gorm:query", Query)
	queryCallback.Register("gorm:preload", Preload)
	queryCallback.Register("gorm:after_query", AfterQuery)
	if len(config.QueryClauses) == 0 {
		config.QueryClauses = queryClauses
	}
	queryCallback.Clauses = config.QueryClauses
}
```

#### 5. Query函数的实现

```go
func Query(db *gorm.DB) {
	if db.Error == nil {
    // 构建查询的 SQL 语句
		BuildQuerySQL(db)

    // 查询数据
		if !db.DryRun && db.Error == nil {
			rows, err := db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			if err != nil {
				db.AddError(err)
				return
			}
			defer rows.Close()

      // 将结果输出到目标结构体中
			gorm.Scan(rows, db, false)
		}
	}
}
```

#### 6.核心-构建SQL的实现

```go
func BuildQuerySQL(db *gorm.DB) {
  // SQL为空，表示需要自己构建
	if db.Statement.SQL.String() == "" {
		db.Statement.SQL.Grow(100) // 分配初始空间

		if len(db.Statement.Selects) > 0 { 
      // 表示只select某几个字段，而不是select *
		} else if db.Statement.Schema != nil && len(db.Statement.Omits) > 0 {
      // Omit表示忽略特定字段
		} else if db.Statement.Schema != nil && db.Statement.ReflectValue.IsValid() {
      // 查询到指定结构体
		}

		// 对join的处理，涉及到多表关联，暂时忽略
		if len(db.Statement.Joins) != 0 {
		} else {
			db.Statement.AddClauseIfNotExists(clause.From{})
		}

    // 用一个map去重，符合名字中的 IfNotExists 含义
		db.Statement.AddClauseIfNotExists(clauseSelect)

    // 最后拼接出完整 SQL 的地方
		db.Statement.Build(db.Statement.BuildClauses...)
	}
}
```

## 小结

本文旨在介绍GORM的推荐使用方式，并简单阅读对接数据库的相关代码。这里分享我的四个观点：

1. **Builder设计模式** - 在面对复杂场景中，Builder设计模式扩展性很好，可分为两个阶段：存储数据+处理数据；GORM的调用就是采用了chainable+finisher的两段实现，前者保存SQL相关元数据，后者拼接SQL并执行；
2. **负重前行** - GORM是一个负重前行的框架：它不仅支持了所有原生SQL的特性，也增加了很多类似Hook的高级特性，导致这个框架非常庞大。如果团队没有历史包袱，更推荐**节制**地使用GORM特性，适当封装一层；
3. **interface{}问题** - GORM中许多函数入参的数据类型都是`interface{}`，底层又用reflect支持了多种类型，这种实现会导致两个问题：
   1. reflect导致的底层的性能不高（这点还能接受）
   2. interface{}如果传入了不支持的复杂数据类型时，排查问题麻烦，往往要运行程序时才会报错
4. **高频拼接重复SQL** - 在一个程序运行过程中，执行的SQL语句都比较固定，而变化的往往是参数；从GORM的实现来看，每次执行都需要重新拼接一次SQL语句，是有不小的优化空间的，比如引入一定的cache。

希望这四点能对大家的日常工作有所启发~



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

