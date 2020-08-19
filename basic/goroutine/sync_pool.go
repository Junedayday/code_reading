package main

import (
	"fmt"
	"reflect"
	"sync"
)

func syncPool() {
	var sp = sync.Pool{
		// Tip: 声明对象池的New函数，这里以一个简单的int为例
		New: func() interface{} {
			return 100
		},
	}
	// Tip: 从对象池里获取一个对象
	data := sp.Get().(int)
	fmt.Println(data)
	// Tip: 往对象池里放回一个对象
	sp.Put(data)

	fmt.Println(reflect.ValueOf(sp))
}
