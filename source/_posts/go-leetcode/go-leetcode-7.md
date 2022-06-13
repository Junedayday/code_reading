---
title: Go算法实战 - 7.【盛最多水的容器LeetCode-11】
date: 2022-04-09 11:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-11 盛最多水的容器

原题链接 https://leetcode-cn.com/problems/container-with-most-water/

```go
func maxArea(height []int) int {
}
```

<!-- more -->

### 题解

```go
func maxArea(height []int) int {
	// 左右两个指针往中间逼近
	left, right := 0, len(height)-1
	var max int
	for left < right {
		var area int
		// 哪边高度低，就挪哪边
		if height[left] < height[right] {
			area = (right - left) * height[left]
			left++
		} else {
			area = (right - left) * height[right]
			right--
		}
		if area > max {
			max = area
		}
	}

	return max
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

