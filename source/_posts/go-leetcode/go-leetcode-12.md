---
title: Go算法实战 - 12.【二分查找LeetCode-704】
date: 2022-04-01 13:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-704 二分查找

原题链接 https://leetcode-cn.com/problems/binary-search/submissions/

```go
func search(nums []int, target int) int {

}
```

<!-- more -->

## 题解

```go
func search(nums []int, target int) int {
	start, end := 0, len(nums)-1
	// 注意，当nums长度为1时，start=end=0
	// 所以这个判断逻辑要注意
	for end >= start {
		mid := (start + end) / 2
		if nums[mid] == target {
			return mid
		} else if nums[mid] > target {
			end = mid - 1
		} else {
			start = mid + 1
		}
	}
	return -1
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

