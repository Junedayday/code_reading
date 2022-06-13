---
title: Go算法实战 - 8.【三数之和LeetCode-15】
date: 2022-04-09 11:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-15 三数之和

原题链接 https://leetcode-cn.com/problems/3sum/

```go
func threeSum(nums []int) [][]int {
}
```

<!-- more -->

## 题解

```go
func threeSum(nums []int) [][]int {
	sort.Ints(nums)
	var result [][]int

	// -2是保证至少留下2个数
	for i := 0; i < len(nums)-2; i++ {
		// 剪枝：最小值大于0时无需再遍历
		if nums[i] > 0 {
			break
		}
		// 剪枝：最小值和前一个值一样时，上一个循环已经判断过，无需再判断
		if i > 0 && nums[i] == nums[i-1] {
			continue
		}
		// j,k 为两个指针，分别从最左边和最右边开始移动
		j, k := i+1, len(nums)-1
		for j < k {
			left, right := nums[j], nums[k]
			if nums[i]+nums[j]+nums[k] == 0 {
				result = append(result, []int{nums[i], nums[j], nums[k]})
				// 减枝：同值的话左边往右移，跳过 nums[j] == nums[j+1] 的情况
				for j < k && nums[j] == left {
					j++
				}
				// 减枝：同值的话右边往左移，跳过 nums[k] == nums[k-1] 的情况
				for j < k && nums[k] == right {
					k--
				}
			} else if nums[i]+nums[j]+nums[k] < 0 {
				// 和小于0，则增大最左边的j
				j++
			} else {
				// 和大于0，则减少最右边的k
				k--
			}
		}
	}
	return result
}
```

> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

