---
title: Go编程模式 - 3.继承与嵌入
date: 2021-02-20 18:31:54
categories: 
- 经典品读
tags:
- Go-Programming-Patterns
---

>  注：本文的灵感来源于GOPHER 2020年大会陈皓的分享，原PPT的[链接](https://www2.slideshare.net/haoel/go-programming-patterns?from_action=save)可能并不方便获取，所以我下载了一份[PDF](https://github.com/Junedayday/code_reading/tree/master/doc/Go_Programming_Patterns.pdf)到git仓，方便大家阅读。我将结合自己的实际项目经历，与大家一起细品这份文档。



## 目录

- [嵌入和委托](#Embedded)
- [反转控制](#IoC)



## Embedded

接口定义

```go
// 定义了两种interface
type Painter interface {
	Paint()
}

type Clicker interface {
	Click()
}
```

Label 实现了 Painter

```go
// 标准组件，用于嵌入
type Widget struct {
	X, Y int
}

// Label 实现了 Painter
type Label struct {
	Widget        // Embedding (delegation)
	Text   string // Aggregation
}

func (label Label) Paint() {
	fmt.Printf("%p:Label.Paint(%q)\n", &label, label.Text)
}
```

ListBox实现了Painter和Clicker

```go
// ListBox声明了Paint和Click，所以实现了Painter和Clicker
type ListBox struct {
	Widget          // Embedding (delegation)
	Texts  []string // Aggregation
	Index  int      // Aggregation
}

func (listBox ListBox) Paint() {
	fmt.Printf("ListBox.Paint(%q)\n", listBox.Texts)
}

func (listBox ListBox) Click() {
	fmt.Printf("ListBox.Click(%q)\n", listBox.Texts)
}
```

Button也实现了Painter和Clicker

```go
// Button 继承了Label，所以直接实现了Painter
// 接下来，Button又声明了Paint和Click，所以实现了Painter和Clicker，其中Paint方法被覆
type Button struct {
	Label // Embedding (delegation)
}

func (button Button) Paint() { // Override
	fmt.Printf("Button.Paint(%s)\n", button.Text)
}

func (button Button) Click() {
	fmt.Printf("Button.Click(%s)\n", button.Text)
}
```

方法调用

```go
func main() {
	label := Label{Widget{10, 70}, "Label"}
	button1 := Button{Label{Widget{10, 70}, "OK"}}
	button2 := Button{Label{Widget{50, 70}, "Cancel"}}
	listBox := ListBox{Widget{10, 40},
		[]string{"AL", "AK", "AZ", "AR"}, 0}

	for _, painter := range []Painter{label, listBox, button1, button2} {
		painter.Paint()
	}

	for _, widget := range []interface{}{label, listBox, button1, button2} {
		// 默认都实现了Painter接口，可以直接调用
		widget.(Painter).Paint()
		if clicker, ok := widget.(Clicker); ok {
			clicker.Click()
		}
	}
}
```

这个例子代码很多，我个人认为重点可以归纳为一句话：

**用嵌入实现方法的继承，减少代码的冗余度**



耗子叔的例子很精彩，不过我个人不太喜欢`interface`这个数据类型（main函数中），有没有什么优化的空间呢？

```go
// 定义两种方法的组合
type PaintClicker interface {
	Painter
	Clicker
}

func main() {
	// 在上面的例子中，interface传参其实不太优雅，有没有更优雅的实现呢？那就用组合的interface
	for _, widget := range []PaintClicker{listBox, button1, button2} {
		widget.Paint()
		widget.Click()
	}
}
```



## IoC

先看一个Int集合的最基本实现

```go
// Int集合，用于最基础的增删查
type IntSet struct {
	data map[int]bool
}

func NewIntSet() IntSet {
	return IntSet{make(map[int]bool)}
}

func (set *IntSet) Add(x int) { set.data[x] = true }

func (set *IntSet) Delete(x int) { delete(set.data, x) }

func (set *IntSet) Contains(x int) bool { return set.data[x] }
```

现在，需求来了，我们希望对这个Int集合的操作是可撤销的

```go
// 可撤销的Int集合，依赖于IntSet，我们看看基本实现
type UndoableIntSet struct { // Poor style
	IntSet    // Embedding (delegation)
	functions []func()
}

func NewUndoableIntSet() UndoableIntSet {
	return UndoableIntSet{NewIntSet(), nil}
}

// 新增
// 不存在元素时：添加元素，并新增撤销函数：删除
// 存在元素时：不做任何操作，并新增撤销函数：空
func (set *UndoableIntSet) Add(x int) { // Override
	if !set.Contains(x) {
		set.data[x] = true
		set.functions = append(set.functions, func() { set.Delete(x) })
	} else {
		set.functions = append(set.functions, nil)
	}
}

// 删除，与新增相反
// 存在元素时：删除元素，并新增撤销函数：新增
// 不存在元素时：不做任何操作，并新增撤销函数：空
func (set *UndoableIntSet) Delete(x int) { // Override
	if set.Contains(x) {
		delete(set.data, x)
		set.functions = append(set.functions, func() { set.Add(x) })
	} else {
		set.functions = append(set.functions, nil)
	}
}

// 撤销：执行最后一个撤销函数function
func (set *UndoableIntSet) Undo() error {
	if len(set.functions) == 0 {
		return errors.New("No functions to undo")
	}
	index := len(set.functions) - 1
	if function := set.functions[index]; function != nil {
		function()
		set.functions[index] = nil // For garbage collection
	}
	set.functions = set.functions[:index]
	return nil
}
```

上面的实现是一种顺序逻辑的思路，整体还是挺麻烦的。有没有优化思路呢？

定义一下Undo这个结构。

```go
type Undo []func()

func (undo *Undo) Add(function func()) {
	*undo = append(*undo, function)
}

func (undo *Undo) Undo() error {
	functions := *undo
	if len(functions) == 0 {
		return errors.New("No functions to undo")
	}
	index := len(functions) - 1
	if function := functions[index]; function != nil {
		function()
		functions[index] = nil // For garbage collection
	}
	*undo = functions[:index]
	return nil
}
```

细品一下这里的实现：

```go
type IntSet2 struct {
	data map[int]bool
	undo Undo
}

func NewIntSet2() IntSet2 {
	return IntSet2{data: make(map[int]bool)}
}

func (set *IntSet2) Undo() error {
	return set.undo.Undo()
}

func (set *IntSet2) Contains(x int) bool {
	return set.data[x]
}

func (set *IntSet2) Add(x int) {
	if !set.Contains(x) {
		set.data[x] = true
		set.undo.Add(func() { set.Delete(x) })
	} else {
		set.undo.Add(nil)
	}
}

func (set *IntSet2) Delete(x int) {
	if set.Contains(x) {
		delete(set.data, x)
		set.undo.Add(func() { set.Add(x) })
	} else {
		set.undo.Add(nil)
	}
}
```

我们看一下，这块代码的前后逻辑有了啥变化：

1. 之前，撤销函数是在Add/Delete时添加的，函数中包含了IntSet的操作，也就是 **Undo依赖IntSet**
2. 而修改之后，撤销函数被抽象为Undo，撤销相关的工作直接调用Undo相关的工作即可，也就是 **IntSet依赖Undo**

我们再来分析一下

- Undo是控制逻辑 - 撤销动作
- IntSet是业务逻辑 - 保存数据的功能。

**业务逻辑依赖控制逻辑，才能保证在复杂业务逻辑变化场景下，代码更健壮！**



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili：https://space.bilibili.com/293775192
>
> 公众号：golangcoding

