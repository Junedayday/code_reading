package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func atomicAdd() {
	var data, data2 int64
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			atomic.AddInt64(&data, 1)
			data2++
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("data", data, "data2", data2)
}

// From medium: https://medium.com/a-journey-with-go/go-how-to-reduce-lock-contention-with-the-atomic-package-ba3b2664b549
type Config struct {
	a []int
}

// 用 go run -race main.go atomic.go 观察data races
func dataRaces() {
	cfg := &Config{}
	go func() {
		i := 0
		for {
			i++
			cfg.a = []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}
		}
	}()

	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for n := 0; n < 100; n++ {
				fmt.Printf("%v\n", cfg)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func dataRacesWithLock() {
	var mu sync.RWMutex
	cfg := &Config{}
	go func() {
		i := 0
		for {
			i++
			mu.Lock()
			cfg.a = []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}
			mu.Unlock()
		}
	}()

	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for n := 0; n < 100; n++ {
				mu.RLock()
				fmt.Printf("%v\n", cfg)
				mu.RUnlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func dataRacesWithAtomic() {
	var v atomic.Value
	go func() {
		i := 0
		for {
			i++
			cfg := &Config{a: []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}}
			v.Store(cfg)
		}
	}()

	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for n := 0; n < 100; n++ {
				cfg := v.Load()
				fmt.Printf("%v\n", cfg)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// cas相关的代码，建议直接参考源码中的相关实现
