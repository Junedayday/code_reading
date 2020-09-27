package main

import (
	"fmt"
	"unsafe"
)

func slice() {
	fmt.Println("Slice Init")
	var s []int
	// Tip: 对比一下map和slice的make函数，前者在类型后可跟0个或1个参数，而后者是1个和2个参数
	s = make([]int, 2)
	fmt.Println(len(s), cap(s))
	s = make([]int, 2, 4)
	fmt.Println(len(s), cap(s))

	// Tip: 元素个数小于cap时，append不会改变cap，只会增加len
	fmt.Println("Slice Assign")
	s[0] = 1
	s[1] = 2
	s = append(s, 4)
	fmt.Println(len(s), cap(s))

	// Tip: 元素个数超过cap时，会进行扩容
	s = append(s, 8, 16)
	fmt.Println(len(s), cap(s))

	// Tip: Slice没有显式的删除语句
	fmt.Println("Slice Delete")
	s = append(s[0:1], s[2:]...)
	fmt.Println(s)

	fmt.Println("Slice Range")
	for i, v := range s {
		fmt.Println(i, v)
		fmt.Printf("%p %p\n", &i, &v)
	}
}

/*
	slice 的源码部分

	slice基础结构slice:
	包括保存数据的array、长度len与容量cap

	初始化函数makeslice:
	math.MulUintptr：根据元素大小和容量cap，计算所需的内存空间
	mallocgc: 分配内存， 32K作为一个临界值，小的分配在P的cache中，大的分配在heap堆中

	扩容growslice:
	当长度小于1024时，cap翻倍；大于1024时，增加1/4。 但这个并不是绝对的，会根据元素的类型尽心过一定的优化

	拷贝slicecopy:
	核心函数为memmove，from=>to移动size大小的数据，size为 元素大小 * from和to中长度较小的个数

	拷贝slicestringcopy：
	基本与上面类似，字符串的拷贝
*/

func sliceAddr() {
	fmt.Println("Part 1")
	var s = make([]int, 2, 2)
	s[0] = 1
	fmt.Println(unsafe.Pointer(&s[0]))
	s[1] = 2
	fmt.Println(unsafe.Pointer(&s[0]))
	// Tip: 扩容后，slice的array的地址会重新分配
	s = append(s, 3)
	fmt.Println(unsafe.Pointer(&s[0]))

	fmt.Println("Part 2")
	// Tip: a虽然是一个新的地址，但指向的array是和a一致的
	a := s[:2]
	fmt.Printf("%p %p\n", &s, &a)
	fmt.Println(unsafe.Pointer(&a[0]))
	a[0] = 2
	fmt.Println(a, s)
	// Tip: 如果要进行slice拷贝，使用copy方法
	b := make([]int, 2)
	copy(b, s)
	fmt.Printf("%p %p\n", &s, &b)
	fmt.Println(unsafe.Pointer(&b[0]))

	fmt.Println("Part 3")
	// Tip: sNil的array指向nil，而sEmpty的array指向一个内存地址
	var sNil []int
	var sEmpty = make([]int, 0)
	fmt.Println(len(sNil), len(sEmpty), cap(sNil), cap(sEmpty))
}
