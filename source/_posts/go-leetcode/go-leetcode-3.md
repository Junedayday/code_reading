---
title: Go算法实战 - 3.【无重复字符的最长子串LeetCode-3】
date: 2022-04-06 12:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-3 无重复字符的最长子串

原题链接 https://leetcode-cn.com/problems/longest-substring-without-repeating-characters/

```go
func lengthOfLongestSubstring(s string) int {
}
```

<!-- more -->

## 题解

```go
func lengthOfLongestSubstring(s string) int {
	// byte与其index，如果重复取最大的index覆盖
	var mp = make(map[byte]int)
	var left, max = 0, 0

	for i := 0; i < len(s); i++ {
		length := i - left + 1
		if _, ok := mp[s[i]]; ok {
			length2 := i - mp[s[i]]
			// 如果left+1在mp[s[i]]左边，则更新left指针到mp[s[i]]+1
			if left-1 < mp[s[i]] {
				length = length2
				left = mp[s[i]] + 1
			}
		}
		if length > max {
			max = length
		}
		mp[s[i]] = i
	}
	// 结尾的字符串
	if len(s)-left > max {
		max = len(s) - left
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

