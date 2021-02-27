---
title: Go编程模式 - 2.基础编码下
date: 2021-02-20 18:31:50
categories: 
- 经典品读
tags:
- Go-Programming-Patterns
---

>  注：本文的灵感来源于GOPHER 2020年大会陈皓的分享，原PPT的[链接](https://www2.slideshare.net/haoel/go-programming-patterns?from_action=save)可能并不方便获取，所以我下载了一份[PDF](https://github.com/Junedayday/code_reading/tree/master/doc/Go_Programming_Patterns.pdf)到git仓，方便大家阅读。我将结合自己的实际项目经历，与大家一起细品这份文档。



## 目录

- [时间格式](#Time)
- [性能1](#Performance1)
- [性能2](#Performance2)
- [扩展阅读](#Further)

注：切勿过早优化！



## Time

这部分的内容实战项目中用得不多，大家记住耗子叔总结出来的一个原则即可：

> 尽量用`time.Time`和`time.Duration`，如果必须用string，尽量用`time.RFC3339`



然而现实情况并没有那么理想，实际项目中用得最频繁，还是自定义的`2006-01-02 15:04:05`

```go
time.Now().Format("2006-01-02 15:04:05")
```



## Performance1

### Itoa性能高于Sprint

主要性能差异是由于`Sprint`针对的是复杂的字符串拼接，底层有个buffer，会在它的基础上进行一些字符串的拼接；

而`Itoa`直接通过一些位操作组合出字符串。

```go
// 170 ns/op
func Benchmark_Sprint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprint(rand.Int())
	}
}

// 81.9 ns/op
func Benchmark_Itoa(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = strconv.Itoa(rand.Int())
	}
}
```



### 减少string到byte的转换

主要了解go的`string`到`[]byte`的转换还是比较耗性能的，但大部分情况下无法避免这种转换。

我们注意一种场景即可：从`[]byte`转换为`string`，再转换为`[]byte`。

```go
// 43.9 ns/op
func Benchmark_String2Bytes(b *testing.B) {
	data := "Hello world"
	w := ioutil.Discard
	for i := 0; i < b.N; i++ {
		w.Write([]byte(data))
	}
}

// 3.06 ns/op
func Benchmark_Bytes(b *testing.B) {
	data := []byte("Hello world")
	w := ioutil.Discard
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}
```



### 切片能声明cap的，尽量初始化时声明

了解slice的扩容机制就能很容易地理解。切片越长，影响越大。

```go
var size = 1000

// 4494 ns/op
func Benchmark_NoCap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		data := make([]int, 0)
		for k := 0; k < size; k++ {
			data = append(data, k)
		}
	}
}

// 2086 ns/op
func Benchmark_Cap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		data := make([]int, 0, size)
		for k := 0; k < size; k++ {
			data = append(data, k)
		}
	}
}
```



### 避免用string做大量字符串的拼接

频繁拼接字符串的场景并不多，了解即可。

```go
var strLen = 10000

// 0.0107 ns/op
func Benchmark_StringAdd(b *testing.B) {
	var str string
	for n := 0; n < strLen; n++ {
		str += "x"
	}
}

// 0.000154 ns/op
func Benchmark_StringBuilder(b *testing.B) {
	var builder strings.Builder
	for n := 0; n < strLen; n++ {
		builder.WriteString("x")
	}
}

// 0.000118 ns/op
func Benchmark_BytesBuffer(b *testing.B) {
	var buffer bytes.Buffer
	for n := 0; n < strLen; n++ {
		buffer.WriteString("x")
	}
}
```



## Performance2

### 并行操作用sync.WaitGroup控制



### 热点内存分配用sync.Pool

注意一下，一定要是`热点`，千万不要 **过早优化**



### 倾向于使用lock-free的atomic包

除了常用的`CAS`操作，还有`atomic.Value`的`Store`和`Load`操作，这里简单地放个实例：

```go
func main() {
	v := atomic.Value{}
	type demo struct {
		a int
		b string
	}

	v.Store(&demo{
		a: 1,
		b: "hello",
	})

	data, ok := v.Load().(*demo)
	fmt.Println(data, ok)
	// &{1 hello} true
}
```

复杂场景下，还是建议用`mutex`。



### 对磁盘的大量读写用bufio包

`bufio.NewReader()`和`bufio.NewWriter()`



### 对正则表达式不要重复compile

```go
// 如果匹配的格式不会变化，全局只初始化一次即可
var compiled = regexp.MustCompile(`^[a-z]+[0-9]+$`)

func main() {
	fmt.Println(compiled.MatchString("test123"))
	fmt.Println(compiled.MatchString("test1234"))
}
```



### 用protobuf替换json

go项目内部通信尽量用`protobuf`，但如果是对外提供api，比如web前端，`json`格式更方便。



### map的key尽量用int来代替string

```go
var size = 1000000

// 0.0442 ns/op
func Benchmark_MapInt(b *testing.B) {
	var m = make(map[int]struct{})
	for i := 0; i < size; i++ {
		m[i] = struct{}{}
	}
	b.ResetTimer()
	for n := 0; n < size; n++ {
		_, _ = m[n]
	}
}

// 0.180 ns/op
func Benchmark_MapString(b *testing.B) {
	var m = make(map[string]struct{})
	for i := 0; i < size; i++ {
		m[strconv.Itoa(i)] = struct{}{}
	}
	b.ResetTimer()
	for n := 0; n < size; n++ {
		_, _ = m[strconv.Itoa(n)]
	}
}
```

示例中`strconv.Itoa`函数对性能多少有点影响，但可以看到`string`和`int`的差距是在数量级的。



## Further

PPT中给出了8个扩展阅读，大家根据情况自行阅读。

如果说你的时间只够读一个材料的话，我推荐大家反复品读一下[Effective Go](https://golang.org/doc/effective_go.html)



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili：https://space.bilibili.com/293775192
>
> 公众号：golangcoding

