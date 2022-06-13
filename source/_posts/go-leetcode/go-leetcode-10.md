---
title: Go算法实战 - 10.【圆环回原点问题】
date: 2022-02-19 12:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## 经典面试题 圆环回原点问题

0-12共13个数构成一个环，从0出发，每次走1步，走n步回到0共有多少种走法？

```go
func CircleToOrigin(n int) int {
}
```

<!-- more -->

## 题解

```go
func CircleToOrigin(n int) int {
	var result = make([][]int, n+1)
	// 初始化
	for i := range result {
		result[i] = make([]int, 13)
	}
	// 初始处于原点，i表示走的步数，j表示走到的位置
	result[0][0] = 1
	for i := 1; i <= n; i++ {
		for j := 0; j < 13; j++ {
			// 动态规划
			result[i][j] = result[i-1][(j+1+13)%13] + result[i-1][(j-1+13)%13]
		}
	}
	return result[n][0]
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

