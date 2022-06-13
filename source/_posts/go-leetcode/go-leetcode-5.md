---
title: Go算法实战 - 5.【最长回文子串LeetCode-5】
date: 2022-04-08 08:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-5 最长回文子串

原题链接 https://leetcode-cn.com/problems/longest-palindromic-substring/

```go
func longestPalindrome(s string) string {
}
```

<!-- more -->

## 题解

递归版本

```go
func longestPalindrome(s string) string {
	if len(s) <= 1 {
		return s
	}
	s1 := longestPalindrome(s[1:])

	// 从右往左移动指针，查询最大子字符串
	right := len(s)
	for ; right > 0; right-- {
		if isOk(s[:right]) {
			break
		}
	}
	s2 := s[:right]

	if len(s1) > len(s2) {
		return s1
	}
	return s2
}

func isOk(s string) bool {
	if len(s) <= 1 {
		return true
	}
	// 移动双指针进行判断
	start, end := 0, len(s)-1
	for start < end {
		if s[start] != s[end] {
			return false
		}
		start++
		end--
	}
	return true
}
```

非递归版本

```go
func longestPalindrome(s string) string {
	length := len(s)
	if length <= 1 {
		return s
	}

	// 初始化
	begin, maxLen := 0, 1
	var dp = make([][]bool, length)
	for k1 := range dp {
		dp[k1] = make([]bool, length)
		for k2 := range dp[k1] {
			dp[k1][k2] = false
		}
		// 长度为1的字符串，为回文字符串
		dp[k1][k1] = true
	}

	for size := 2; size <= length; size++ {
		// founded用来表示对应size的回文字符串已经找到
		founded := false
		for start := 0; start <= length-size; start++ {
			end := start + size - 1
			if s[start] == s[end] {
				// 长度为2的不用继续查了
				if size == 2 || dp[start+1][end-1] {
					dp[start][end] = true
					// 全局最长的字符串，只记录第一个即可
					if !founded && size > maxLen {
						maxLen, begin = size, start
						founded = true
					}
				}
			}
		}
	}
	return s[begin : begin+maxLen]
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

