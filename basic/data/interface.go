package main

import (
	"fmt"
)

type Dog interface {
	Hi()
}

type Dog1 struct{}

func (d Dog1) Hi() {}

type Dog2 struct{}

func (d Dog2) Hi() {}

/*
	查看汇编，了解调用:
1. go build -gcflags '-l' -o if main.go interface.go
2. go tool objdump -s "main\.dataInterface" if

（汇编会因平台不同，调用结果会不一样）调用了 CALL runtime.XXX
对应代码在runtime/iface.go下，在这里，我们看到了interface的两个核心结构 eface 和 iface，分别对应convT2E和convT2I

type iface struct {
	tab  *itab
	data unsafe.Pointer
}

type eface struct {
	_type *_type
	data  unsafe.Pointer
}

其中data指向了具体保存数据的内存地址，为一个通用结构。而itab中嵌套了_type，比较复杂，我们先从简单的eface看起
*/

func dataInterface() {
	var i interface{} = 1
	fmt.Println(i)
	// Tip: 类型定义，除非是100%确定成功，否则尽量用两个参数，否则会导致panic
	v, ok := i.(string)
	fmt.Printf(v, ok)

	var d1 Dog1
	var d2 Dog2
	var dList = []Dog{d1, d2}
	for _, v := range dList {
		v.Hi()
	}
}

/*
Tip: eface part，静态的数据结构
type _type struct {
	size       uintptr
	ptrdata    uintptr // size of memory prefix holding all pointers
	hash       uint32
	tflag      tflag
	align      uint8
	fieldAlign uint8
	kind       uint8
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	equal func(unsafe.Pointer, unsafe.Pointer) bool
	// gcdata stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, gcdata is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	gcdata    *byte
	str       nameOff
	ptrToThis typeOff
}

Tip: runtime/type.go 类型定义包含了go语言常见的类型，我们可以从 t.kind & kindMask 看到Go里支持的所有类型
*/

/*
Tip: iface part，面向对象中的interface
type itab struct {
	inter *interfacetype
	_type *_type
	hash  uint32 // copy of _type.hash. Used for type switches.
	_     [4]byte
	fun   [1]uintptr // variable sized. fun[0]==0 means _type does not implement inter.
}

Tip: interfacetype 为接口的定义方法集； _type保存了具体的类型；而具体的方法的地址都被保存在fun中，是一个数组
type interfacetype struct {
	typ     _type
	pkgpath name
	mhdr    []imethod
}
type imethod struct {
	name nameOff
	ityp typeOff
}

methods := (*[1 << 16]unsafe.Pointer)(unsafe.Pointer(&m.fun[0]))[:ni:ni]

Tip: interface的类型推断，依赖于一个 itabTable，即一个map+lock，将匹配成功的保存进来，下次可直接查询
*/
