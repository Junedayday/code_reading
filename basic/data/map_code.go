package main

import (
	"fmt"
)

func mapAddr() {
	var mp = map[int]int{1: 2, 2: 3, 3: 4}
	fmt.Printf("%p\n", &mp)

	fmt.Println("map range 1 start")
	// Tip 注意range中的k是不能保证顺序的
	for k, v := range mp {
		// Tip: k,v 是为了遍历、另外开辟内存保存的一对变量，不是map中key/value保存的地址
		fmt.Println(&k, &v)

		// Tip: 修改k/v后，不会生效
		k++
		v++

		// Tip 如果在这里加上一个delete操作，在k=1时会删除k=2的元素，然后就直接跳过元素2，遍历到元素3
		// map的delete是安全的，不会存在race等竞争问题，这里用到的k,v是值复制，删除是针对hmap的操作
		// delete(mp, k)

		// 编译错误，value不可寻址，因为这个在内部是频繁变化的
		// fmt.Printf("%p", &mp[k])
	}
	fmt.Println(mp)
	fmt.Println("map range 1 end")
}

func mapModify() {
	// Tip: 如果map最终的size比较大，就放到初始化的make中，会减少hmap扩容带来的内容重新分配
	//var mp2 = make(map[int]int,1000)
	var mp2 = make(map[int]int)
	mp2[1] = 2
	mp2[2] = 3
	mp2[3] = 4
	fmt.Println("map range 2 start")
	for k, v := range mp2 {
		// Tip: 在range的过程中，如果不断扩容，何时退出是不确定的，和是否需要sleep无关
		mp2[k+1] = v + 1
		// time.Sleep(10 * time.Millisecond)
	}
	fmt.Println(len(mp2))
	fmt.Println("map range 2 end")
}

// https://stackoverflow.com/questions/45132563/idiomatic-way-of-renaming-keys-in-map-while-ranging-over-the-original-map
func mapReplace() {
	o := make(map[string]string) // original map
	r := make(map[string]string) // replacement map original -> destination keys

	o["a"] = "x"
	o["b"] = "y"

	r["a"] = "1"
	r["b"] = "2"

	fmt.Println(o) // -> map[a:x b:y]

	// Tip: 因为o的k-v是在不断增加的，所以遍历何时结束是不确定的
	// 此时k可能为"1"或者"2"，对应的r[k]不存在，返回默认值空字符串 ""，所以结果会多一个key为""，value为"x"或"y"的异常值
	for k, v := range o {
		o[r[k]] = v
	}
	// Tip: 到这里，也许你会好奇，为什么每次运行的结果会不一致呢？
	// 1. 首先，我们要了解一点，遍历这个工作在hmap中是通过buckets进行的
	// 2. 因为多次运行的结果不一致，说明每一次运行时，分配的bucket是有随机的
	// 3. 仔细查看hmap这个结构体，我们不难发现hash0这个随机值，确认其对分配bucket的hash计算带来的影响

	delete(o, "a")
	delete(o, "b")

	fmt.Println(o)
}

func mapSet() {
	// 空struct是不占用空间的
	var mp = make(map[string]struct{})
	fmt.Println(mp)
}
