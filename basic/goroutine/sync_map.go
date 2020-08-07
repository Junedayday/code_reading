package main

import (
	"fmt"
	"sync"
)

func syncMap() {
	var m sync.Map
	m.Store("a", 1)
	fmt.Println(m.Load("a"))
	// Tip: LoadOrStore 有就加载值，没有就保存值
	fmt.Println(m.LoadOrStore("a", 1))
	m.Delete("a")
	fmt.Println(m.LoadOrStore("a", 1))
	fmt.Println(m.LoadOrStore("b", 2))
	// Tip: 遍历
	m.Range(func(key, value interface{}) bool {
		fmt.Println(key, value)
		return false
	})
}
