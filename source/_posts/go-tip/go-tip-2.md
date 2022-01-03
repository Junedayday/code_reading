---
title: Go语言技巧 - 2.【错误处理】谈谈Go Error的前世今生
date: 2021-05-05 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![go-tip](https://cloud-fitter-1305666920.cos.ap-beijing.myqcloud.com/go-study.jpeg)

## 从Go 2 Error Proposal谈起

`Go`对`error`的处理一直都是很大的争议点，这点官方也已多次发文，并在2019年1月推出了一篇Proposal，有兴趣的可以点击链接细细品读。

[官方原文链接](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md)

下面，我会结合Proposal原文，发表一些自己的看法（会带上主观意见），欢迎讨论。

<!-- more -->

## 目标

这篇Proposal有一句话很好地解释了对`error`的期许：

**making errors more informative for both programs and people**

错误不仅是告诉机器怎么做的，也是告诉人发生了什么问题。



## 回顾

先让我们一起简单地回顾一下`error`的现状，来更好地理解这个 **more informative** 指的是什么。

原始的error定义为：

```go
type error interface {
  Error() string
}
```

这里面的包含信息很少：一个Error() 的方法，即用字符串返回对应的错误信息。

最常用的`error`相关方法是2种：

1. 创建`error` - `fmt.Errorf`，它是针对`Error() `方法返回的字符串进行加工，如附带一些参数信息（暂不讨论%w这个wrap错误的实现）
2. 使用`error` - 由于我们将`error`的输出结果定义为字符串，所以使用`error`时，一旦涉及到细节，就只能使用一些`string`的方法了

举个具体的例子：

```go
func main() {
	// 假设 readFile 存在于第三方或公用的库，我们没有权限修改、或者修改它的影响面很大
	_, err := readFile("test")

	// 错误中包含业务逻辑:
	// 1. 文件不存在时，认为是 正常
	// 2. 其余报错时，认为是 异常
	if err != nil {
		if strings.Index(err.Error(), "no such file or directory") >= 0 {
			log.Println("file not exist")
			os.Exit(0)
		}
		log.Println("open file error")
		os.Exit(1)
	}
}

func readFile(fileName string) ([]byte, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read file %s error %v", fileName, err)
	}
	return b, nil
}
```

这里存在3个明显的问题：

1. **破坏性** - `fmt.Errorf` 破坏了原有的error，将它从一个 **具体对象** 转化为 **扁平的** `string`，再填充到了新的`error`中。所以，通过`fmt.Errorf`处理后的error，都只传递了一个`string`的信息
2. **实现僵化** - **"no such file or directory"** 这个错误信息用的是**硬编码**，对第三方`readFile`的内容有强依赖，不灵活
3. **排查问题效率低** - 可以通过日志组件了解到error在`main`函数哪行发生，但无法知道错误从`readFile`中的哪行返回过来的

> 其中第一个破坏性的问题，其实就是破坏了error这个interface背后的具体实现，违背了面向对象的继承原则。



## Handle Errors Only Once

在工程中，为了解决 **排查问题效率低** 这个问题，有一个很常见的做法（以上面的readFile为例）:

```go
func readFile(fileName string) ([]byte, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("read file %s error %v", fileName, err)
		return nil, fmt.Errorf("read file %s error %v", fileName, err)
	}
	return b, nil
}
```

没错，就是 **打印错误并返回**。有大量排查问题经验的同学，对此肯定是深恶痛绝： **一个错误能找到N处打印，看得人眼花缭乱**。

这里违背了一个关键性的原则：**对错误只进行一次处理，处理完之后就不要再往上抛了，而打印错误也是一种处理。**

结合三种具体的场景，我们分析一下：

1. 一个程序模块内，`error`不断往上抛，最上层处理；
2. 一个公共的工具包中，`error`不记录，传给调用方处理；
3. 一个RPC模块的调用中，`error`可以记录，作为`debug`信息，而具体的处理仍应交给调用方。

示例参考文章

- https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
- https://www.orsolabs.com/post/go-errors-and-logs/



## 理论实现

那么，怎么样的`error`才是合适的呢？我们分两个角度来看这个`error`：

1. 对程序来说，`error`要包含**错误细节**：如错误类型、错误码等，方便在模块间传递；
2. 对人来说，`error`要包含**代码信息**：如相关的调用参数、运行信息，方便查问题；



用原文一句话来归纳：**hide implementation details from programs while displaying them for diagnosis**

- Wrap - 隐藏实现，针对代码调用时的堆栈信息
- Is/As - 展示细节，针对底层真正实现的数据结构



## 当前实现

`Go`语言发展多年，已经有了很多关于`error`的处理方法，但大多为过渡方案，我就不一一分析了。

这里我以 github.com/pkg/errors 为例，也是这个**官方Proposal**的重点参考对象，简单地分享一下大致实现思路。

代码量并不多，大家可以自行阅读源码：

### New 产生错误的堆栈信息

```go
func New(message string) error {
	return &fundamental{
		msg:   message,
		stack: callers(),
	}
}

type fundamental struct {
	msg string
	*stack
}
```

**关键点** stack保存了错误产生的堆栈信息，如函数名、代码行



### Wrap 包装错误

```go
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	err = &withMessage{
		cause: err,
		msg:   message,
	}
	return &withStack{
		err,
		callers(),
	}
}
```

**关键点** 将错误包装出一个全新的堆栈。一般只用于对外接口产生错误时，包括标准库、RPC。



### WithMessage 添加普通信息

```go
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause: err,
		msg:   message,
	}
}
```

**关键点** 添加错误信息，增加一个普通的堆栈打印



### Is 解析Sentinel错误、即全局错误变量

```go
func Is(err, target error) bool { return stderrors.Is(err, target) }

func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflectlite.TypeOf(target).Comparable()
	for {
		if isComparable && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		// TODO: consider supporing target.Is(err). This would allow
		// user-definable predicates, but also may allow for coping with sloppy
		// APIs, thereby making it easier to get away with them.
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}
```

**关键点** 反复Unwrap、提取错误，解析并对比错误类型



## As - 提取出具体的错误数据结构

```go
func As(err error, target interface{}) bool { return stderrors.As(err, target) }

func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflectlite.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflectlite.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	if e := typ.Elem(); e.Kind() != reflectlite.Interface && !e.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	targetType := typ.Elem()
	for err != nil {
		if reflectlite.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflectlite.ValueOf(err))
			return true
		}
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		err = Unwrap(err)
	}
	return false
}
```

**关键点** 反复Unwrap、提取错误，提取底层的实现类型



## 小结

`Go`语言对`error`的定义很简单，虽然带来了灵活性，但也导致处理方式泛滥，一如当年的**Go语言的版本管理**。如今的**go mod**版本管理机制已经”一统江湖“，随着大家对`error`这块的不断深入，`Error Handling`也总会达成共识。

接下来，我会结合实际代码样例，写一个具体工程中 **Error Handling** 的操作方法，提供一定的参考。





> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

