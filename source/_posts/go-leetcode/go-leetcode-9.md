---
title: Go算法实战 - 9.【电话号码的字母组合LeetCode-17】
date: 2021-08-18 12:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-17 电话号码的字母组合

原题链接 https://leetcode-cn.com/problems/letter-combinations-of-a-phone-number/

```go
func letterCombinations(digits string) []string {
}
```

<!-- more -->

## 题解

```go
func letterCombinations(digits string) []string {
	if len(digits) == 0 {
		return []string{}
	}
	var result []string
	numToLetter := map[string]string{
		"2": "abc",
		"3": "def",
		"4": "ghi",
		"5": "jkl",
		"6": "mno",
		"7": "pqrs",
		"8": "tuv",
		"9": "wxyz",
	}

	// current表示当前匹配的字符串，left表示剩余要匹配的
	var matchNext func(current string, left string)
	matchNext = func(current string, left string) {
		// 完成匹配
		if len(left) == 0 {
			result = append(result, current)
			return
		}
		for _, v := range numToLetter[string(left[0])] {
			current = current + string(v)
			matchNext(current, left[1:])
			current = current[:len(current)-1] //回溯
		}
	}

	// 开始匹配，初始current为空
	matchNext("", digits)
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

