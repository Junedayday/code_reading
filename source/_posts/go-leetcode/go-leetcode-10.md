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

## 基础思路

这道题需要一定的**动态规划**的基础，我们要对这道题做一定的简化，也就是将n变量变成n-1，看看前一步走到了哪。

`第n步走到0 = 第n-1步走到12 + 第n-1步走到1`

所以，我们要引入一个变量，就是当前走到了哪个位置。用代码表示，就是

`result[n][pos] = result[n-1][pos+1] + result[n-1][pos-1] `

我们再考虑到pos-1<0的情况，保证处于0到12的范围，代码就更新为:

`result[n][pos] = result[n-1][(pos+1+13)%13] + result[n-1][(pos-1+13)%13]`

## 题解

```go
func CircleToOrigin(n int) int {
	var result = make([][]int, n+1)
	// 初始化
	for i := range result {
		result[i] = make([]int, 13)
	}
	// 初始处于原点
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

整个代码实现很清晰。我们可以将这个`13`提取出来作为一个变量，也就是环的大小可以自行变化。

## 总结

动态规划最重要的是写出递推的公式，是我们重点需要记忆的部分，剩余的就是考虑边界条件即可。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

