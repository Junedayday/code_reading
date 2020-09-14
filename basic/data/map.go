package main

import "fmt"

func mapper() {
	// var mp map[string]int 不会初始化
	mp := make(map[string]int)

	mp["Tom"] = 10

	age, ok := mp["Tom"]
	fmt.Println(age, ok)
	age1 := mp["Tom"]
	age2 := mp["Tom2"]
	fmt.Println(age1, age2)

	for key, value := range mp {
		fmt.Println(key, value)
	}

	for key := range mp {
		fmt.Println(key, mp[key])
	}
	var data = []int{1, 3, 5}
	for key := range data {
		fmt.Println(key, data[key])
	}
}

// go tool compile -S map.go
func mapCompile() {
	m := make(map[string]int, 9)
	key := "test"
	m[key] = 1
	_, ok := m[key]
	if ok {
		delete(m, key)
	}
}

/*
	源码文件：runtime/map.go
	初始化：makemap
		1. map中bucket的初始化 makeBucketArray
		2. overflow 的定义为哈希冲突的值，用链表法解决
	赋值：mapassign
		1. 不支持并发操作 h.flags&hashWriting
		2. key的alg算法包括两种，一个是equal，用于对比；另一个是hash，也就是哈希函数
		3. 位操作 h.flags ^= hashWriting 与 h.flags &^= hashWriting
		4. 根据hash找到bucket，遍历其链表下的8个bucket，对比hashtop值；如果key不在map中，判断是否需要扩容
	扩容：hashGrow
		1. 扩容时，会将原来的 buckets 搬运到 oldbuckets
	读取：mapaccess
		1. mapaccess1_fat 与 mapaccess2_fat 分别对应1个与2个返回值
		2. hash 分为低位与高位两部分，先通过低位快速找到bucket，再通过高位进一步查找，对后对比具体的key
		3. 访问到oldbuckets中的数据时，会迁移到buckets
	删除：mapdelete
		1. 引入了emptyOne与emptyRest，后者是为了加速查找
*/
