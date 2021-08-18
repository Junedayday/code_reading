---
title: Go算法实战 - 9.【电话号码的字母组合LeetCode-17】
date: 2021-08-18 12:00:00
categories: 
- 成长分享
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

## 基础解法

### 基本思路

这道题的思路并不复杂，我们逐个处理`digits`里的字符，追加到结果上

```go
var numToLetter = map[string][]string{
	"2": {"a", "b", "c"},
	"3": {"d", "e", "f"},
	"4": {"g", "h", "i"},
	"5": {"j", "k", "l"},
	"6": {"m", "n", "o"},
	"7": {"p", "q", "r", "s"},
	"8": {"t", "u", "v"},
	"9": {"w", "x", "y", "z"},
}

func letterCombinations(digits string) []string {
	if digits == "" {
		return []string{}
	}
	
	return matchNext([]string{}, digits)
}

// 核心递归，逐个处理剩余的字符串left
func matchNext(current []string, left string) []string {
	if left == "" {
		return current
	}
	// 从题目来看肯定是matched的
	matched, ok := numToLetter[string(left[0])]
	if !ok {
		return current
	}
	if len(current) == 0 {
		return matchNext(matched, left[1:])
	}
	next := make([]string, len(current)*len(matched))
	for i := range next {
	  // 利用位操作加速
		next[i] = current[i/len(matched)] + matched[i%len(matched)]
	}
	return matchNext(next, left[1:])
}
```



### 减少内存空间

从运行结果来看：

- 执行用时：0 ms, 在所有 Go 提交中击败了100.00%的用户

运行速度已经很难优化了，我们就想办法减少一下内存空间。

```go
// 减少空间，把切片转变成字符串
var numToLetter = map[string]string{
	"2": "abc",
	"3": "def",
	"4": "ghi",
	"5": "jkl",
	"6": "mno",
	"7": "pqrs",
	"8": "tuv",
	"9": "wxyz",
}

func letterCombinations(digits string) []string {
	if digits == "" {
		return []string{}
	}

	return matchNext([]string{}, digits)
}

func matchNext(current []string, left string) []string {
	if left == "" {
		return current
	}
	matched, ok := numToLetter[string(left[0])]
	if !ok {
		return current
	}
	if len(current) == 0 {
		next := make([]string, len(matched))
		for i := range matched {
			next[i] = string(matched[i])
		}
		return matchNext(next, left[1:])
	}
	next := make([]string, len(current)*len(matched))
	for i := range next {
		// 利用位操作加速
		next[i] = current[i/len(matched)] + string(matched[i%len(matched)])
	}
	return matchNext(next, left[1:])
}
```

## 进阶：引入函数式编程

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
	var matchNext func(current string, left string)
	matchNext = func(current string, left string) {
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
	matchNext("", digits)
	return result
}
```

整个思路比较简单，最主要的优点在于将`numToLetter`和`matchNext`收敛到了函数中，对外暴露的细节就减少了。

## 总结

本题的解法比较通俗易懂，最主要的切入点是引入**递归的思想**，来不断地缩减传入的字符串`digits`的长度，直到为0。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

