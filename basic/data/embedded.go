package main

import (
	"fmt"
)

// Tip: 代码复用

// 一个通用的订单，包含价格price字段
type Order struct {
	id    int
	price float32
}

func (o *Order) GetOrderPrice() float32 {
	return o.price
}

type FoodOrder struct {
	Order
	foodNames []string
}

func (fo *FoodOrder) GetFoodNames() []string {
	return fo.foodNames
}

func embedded() {
	fo := &FoodOrder{}
	// 可以直接调用embedded的Order的方法
	fmt.Println(fo.GetOrderPrice())
	fmt.Println(fo.GetFoodNames())
}

// Tip： 覆盖方法
type OverWriteOrder struct {
	Order
}

func (o *OverWriteOrder) GetOrderPrice() float32 {
	return o.price + 1
}

func overwrite() {
	owo := &OverWriteOrder{}
	// 会覆盖embedded的Order的方法
	fmt.Println(owo.GetOrderPrice())
}

// Tip: 方法采用指针和结构体
type AnimalPointer struct{}

func (a *AnimalPointer) Age() int {
	return 1
}

type AnimalStruct struct{}

func (a AnimalStruct) Age() int {
	return 0
}

func pointerAndStruct() {
	ap := &AnimalPointer{}
	as := &AnimalStruct{}
	ap.Age()
	as.Age()

	ap2 := AnimalPointer{}
	as2 := AnimalStruct{}
	ap2.Age()
	as2.Age()
}

// Tip: interface定义的小技巧，让组合更加高效
type Pusher interface {
	Pusher()
	Name()
}

type Puller interface {
	Puller()
	Name()
}

// 新版本已经支持两个interface中有相同的方法定义了
type PushPuller interface {
	//Pusher
	//Puller
}
