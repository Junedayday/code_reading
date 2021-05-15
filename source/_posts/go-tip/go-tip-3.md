---
title: Go语言技巧 - 3.【Error工程化】Go Error的工程化探索
date: 2021-05-07 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![Go-Study](https://i.loli.net/2021/05/05/2bmr98tG3xDneL5.jpg)

## Go Error的工程化探索

在上一篇，我分享了对 [官方Proposal](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md) 的一些见解，偏向于理论层面。

本篇里，我会具体到代码层面，谈谈如何在一个工程化的项目中利用`github.com/pkg/errors`包，完整实现一套的错误处理机制。



## 全局定义的error实现 - MyError

```go
// 全局的 错误号 类型，用于API调用之间传递
type MyErrorCode int

// 全局的 错误号 的具体定义
const (
	ErrorBookNotFoundCode MyErrorCode = iota + 1
	ErrorBookHasBeenBorrowedCode
)

// 内部的错误map，用来对应 错误号和错误信息
var errCodeMap = map[MyErrorCode]string{
	ErrorBookNotFoundCode:        "Book was not found",
	ErrorBookHasBeenBorrowedCode: "Book has been borrowed",
}

// Sentinel Error： 即全局定义的Static错误变量
// 注意，这里的全局error是没有保存堆栈信息的，所以需要在初始调用处使用 errors.Wrap
var (
	ErrorBookNotFound        = NewMyError(ErrorBookNotFoundCode)
	ErrorBookHasBeenBorrowed = NewMyError(ErrorBookHasBeenBorrowedCode)
)

func NewMyError(code MyErrorCode) *MyError {
	return &MyError{
		Code:    code,
		Message: errCodeMap[code],
	}
}

// error的具体实现
type MyError struct {
	// 对外使用 - 错误码
	Code MyErrorCode
	// 对外使用 - 错误信息
	Message string
}

func (e *MyError) Error() string {
	return e.Message
}
```



## 具体示例 - 借书的三种场景

我们来模拟一个场景：

我去图书馆借几本书，会存在三个场景，分别的处理逻辑如下

1. 找到书 - 不需要任何处理
2. 发现书被借走了 - 打印一下即可，不认为是错误
3. 发现图书馆不存在这本书 - 认为是错误，需要打印详细的错误信息

```go
func main() {
	books := []string{
		"Hamlet",
		"Jane Eyre",
		"War and Peace",
	}

	for _, bookName := range books {
		fmt.Printf("%s start\n===\n", bookName)

		err := borrowOne(bookName)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}

		fmt.Printf("===\n%s end\n\n", bookName)
	}
}

func borrowOne(bookName string) error {
	// Step1: 找书
	err := searchBook(bookName)

	// Step2: 处理
	// 特殊业务场景：如果发现书被借走了，下次再来就行了，不需要作为错误处理
	if err != nil {
		// 提取error这个interface底层的错误码，一般在API的返回前才提取
		// As - 获取错误的具体实现
		var myError = new(MyError)
		if errors.As(err, &myError) {
			fmt.Printf("error code is %d, message is %s\n", myError.Code, myError.Message)
		}

		// 特殊逻辑: 对应场景2，指定错误(ErrorBookHasBeenBorrowed)时，打印即可，不返回错误
		// Is - 判断错误是否为指定类型
		if errors.Is(err, ErrorBookHasBeenBorrowed) {
			fmt.Printf("book %s has been borrowed, I will come back later!\n", bookName)
			err = nil
		}
	}

	return err
}

func searchBook(bookName string) error {
	// 下面两个 error 都是不带堆栈信息的，所以初次调用得用Wrap方法
	// 如果已有堆栈信息，应调用WithMessage方法

	// 3 发现图书馆不存在这本书 - 认为是错误，需要打印详细的错误信息
	if len(bookName) > 10 {
		return errors.Wrapf(ErrorBookNotFound, "bookName is %s", bookName)
	} else if len(bookName) > 8 {
		// 2 发现书被借走了 - 打印一下被接走的提示即可，不认为是错误
		return errors.Wrapf(ErrorBookHasBeenBorrowed, "bookName is %s", bookName)
	}
	// 1 找到书 - 不需要任何处理
	return nil
}
```



## 运行结果

### 1. 找到书 - Helmet

```shell
Hamlet start
===
===
Hamlet end
```

没有任何错误信息



### 2. 发现书被借走了 - Jane Eyre

```shell
Jane Eyre start
===
error code is 2, message is Book has been borrowed
book Jane Eyre has been borrowed, I will come back later!
===
Jane Eyre end
```

**打印被借走的提示**，而错误被 `err = nil` 屏蔽。



### 3. 发现图书馆不存在这本书 - War and Peace

```shell
War and Peace start
===
error code is 1, message is Book was not found
Book was not found
bookName is War and Peace
main.searchBook
        /GoProject/godemo/main.go:98
main.borrowOne
        /GoProject/godemo/main.go:71
main.main
        /GoProject/godemo/main.go:60
runtime.main
        /usr/local/go1.13.5/src/runtime/proc.go:203
runtime.goexit
        /usr/local/go1.13.5/src/runtime/asm_amd64.s:1357
===
War and Peace end
```

**打印了错误的详细堆栈**，在IDE中调试非常方便，可以直接跳转到对应代码位置。



## 关键点

1. `MyError` 作为全局 `error` 的底层实现，保存具体的错误码和错误信息；
2. `MyError`向上返回错误时，第一次先用`Wrap`初始化堆栈，后续用`WithMessage`增加堆栈信息；
3. 从`error`中解析具体错误时，用`errors.As`提取出`MyError`，其中的错误码和错误信息可以传入到具体的API接口中；
4. 要判断`error`是否为指定的错误时，用`errors.Is` + `Sentinel Error`的方法，处理一些特定情况下的逻辑；

> Tips：
>
> 1. 不要一直用errors.Wrap来反复包装错误，堆栈信息会爆炸，具体情况可自行测试了解
> 2. 利用go generate可以大量简化初始化Sentinel Error这块重复的工作
> 3. `github.com/pkg/errors`和标准库的`error`完全兼容，可以先替换、后续改造历史遗留的代码
> 4. 一定要注意打印`error`的堆栈需要用`%+v`，而原来的`%v`依旧为普通字符串方法；同时也要注意日志采集工具是否支持多行匹配



## 小结

从现状来看，`Go` 语言的 `Error Handling` 已趋于共识，。

后续差异点就在底层 `MyError` 这块的实现，我个人认为会有如下三个方向：

- 增加一些其余业务或系统的字段
- 对`Is`，`As` 等函数再进行一定的封装，使用起来更方便
- 区分不同的错误类型，来告诉调用方该如何处理，如 **普通错误**、**重试错误** 、**服务降级错误** 等



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

