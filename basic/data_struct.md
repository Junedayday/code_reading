## 数据结构进阶

### [map](data/map.go)

1. `map` 读取某个值时 - 返回结果可以为 `value,bool` 或者 `value`。注意后者，在key不存在时，会返回value对应类型的默认值
2. `map` 的 `range` 方法需要注意 - `key,value` 或者 `key`。注意后者，可以和`slice`的使用结合起来
3. `map` 的底层相关的实现 - 串联 初始化、赋值、扩容、读取、删除 这五个常见实现的背后知识点，详细参考示例代码链接与源码

### [map示例](data/map_code.go)

1. `map` 的 `range` 操作 - key、value 都是值复制
2. `map` 如何保证按key的某个顺序遍历？ - 分两次遍历，第一次取出所有的key并排序；第二次按排序后的key去遍历(这时你可以思考封装map和slice到一个结构体中)？
3. `map` 的使用上，有什么要注意的？ - 遍历时，尽量只修改或删除当前key，操作非当前的key会带来不可预知的结果
4. 从 `map` 的设计上，我们可以学到 - Go语言对map底层的hmap做了很多层面的优化与封装，也屏蔽了很多实现的细节，适用于绝大多数的场景；而少部分有极高性能要求的场景，就需要深入到hmap中的相关细节。

### [slice](data/slice.go)

1. 熟悉 `slice` 的底层数据结构 -  实际存储数据的`array`，当前长度`len`与容量`cap`
2. `slice的扩容机制` - 不严格来说，当长度小于1024时，cap翻倍；大于1024时，增加1/4
3. `slice` 有很多特性与 `map` 一致 - 记住一点，代码中操作的`slice`和`map`只是上层的，实际存储数据的是`array`与`hmap`

### [interface](data/interface.go)

1. interface的两种类型 - `数据结构的interface`，侧重于类型；`面向对象中接口定义的interface`，侧重于方法的声明
2. 了解interface的底层定义 - `eface`和`iface`，都分为两个部分：`类型`与`数据`
3. `iface`底层对类型匹配进行了优化 - `map`+`mutex`组合
