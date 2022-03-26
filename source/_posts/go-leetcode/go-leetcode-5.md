---
title: Go算法实战 - 5.【最长回文子串LeetCode-5】
date: 2021-07-19 12:00:00
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

这道题中，我们也要注意**奇数和偶数**的边界情况。

<!-- more -->

## 基本解法

### 枚举所有情况

刚拿到这道题时，我们很容易想到一个解法：

- 枚举出s所有的子字符串
- 找到最长的回文串

这个解法代码我就不专门写了，复杂度很高，但如果你想不到解决方案时，可以先写出这个解法，再逐步优化。



### 利用递归拆分成子问题

```go
func longestPalindrome(s string) string {
    if len(s) <= 1 {
        return s
    }

    // 查找以s[0]为开头的最长回文子串
    right := len(s)
    for ; right >= 0; right-- {
        // 找到最长的回文子串就退出
        if isPalindrome(s[:right]) {
            break
        }
    }

    // 用递归的方式获取s[1:]的长度，与上面的结果对比
    subS := longestPalindrome(s[1:])
    if len(subS) > right {
        return subS
    }
    return s[:right]
}

// 头尾两个指针不断往中间缩进
func isPalindrome(s string) bool {
    if len(s) <= 1 {
        return true
    }
    left, right := 0, len(s) - 1
    
    // 如果s长度为奇数，退出时刚好left=right，偶数则为 left > right
    for left < right {
        if s[left] != s[right] {
            return false
        }
        left++
        right--
    }
    return true
}
```

这个解法的代码思路相对清晰，核心思路是将 **以s[0]开头的回文字符串** 与 **s[1:]中最长回文子串** 进行对比，其中后者是递归的子问题。

在有一定的算法基础后，我个人比较喜欢先用 **递归思路去简化问题**，将代码拆分成子问题，能大量地简化代码。



## 进阶解法

### 利用对称点

从回文串的特点来看，它存在一个对称点：

- 长度为奇数2n+1的字符串，对称点为第n+1个字符
- 长度为偶数2n的字符串，对称点为第n与n+1个字符之间，我们不妨命名为空白blank

所以，我们就对字符串s进行遍历，查找它的**每个对称点上最长的回文串**

```go
func longestPalindrome(s string) string {
	var longest string
	// 是否为空白
	// 我们把字符串s看作为： s[0]-blank-s[1]-...-s[n-1]
	var isBlank bool
	var mid int

	for mid < len(s) {
		var sub string
		if isBlank {
			// 空白的话，我们认为这个空白是在mid-1和mid之间的，从两边开始对比
			sub = findLongestPalindromeByMid(s, mid-1, mid)
			isBlank = false
			mid++
		} else {
			// 非空白的话，从索引mid的两边开始对比
			sub = findLongestPalindromeByMid(s, mid-1, mid+1)
			isBlank = true
		}
		if len(sub) > len(longest) {
			longest = sub
		}
	}
	return longest
}

func findLongestPalindromeByMid(s string, left, right int) string {
	// left和right就是两个指针，不断往两边移动，直到找到不相同的两个字符
	for left >= 0 && right < len(s) {
		if s[left] != s[right] {
			break
		}
		left--
		right++
	}

	// 注意这里的索引，很容易搞混，我们要取的是left+1到right-1，这里切片后面的索引为right
	return s[left+1 : right]
}
```



### 简化空白的处理

从上面的代码来看，空白的处理挺麻烦，我们可以直接从代码结构去优化。

大家可以仔细的阅读`isBlank`相关的代码，可以直接消除这个变量：

```go
func longestPalindrome(s string) string {
    var longest string
    var mid int 
    
    for mid < len(s) {
        // 对称点为某个字符，长度为奇数
        sub := findLongestPalindromeByMid(s, mid-1, mid)
        if len(sub) > len(longest) {
            longest = sub
        }
        
        // 对称点为空白，长度为偶数
        sub = findLongestPalindromeByMid(s, mid-1, mid+1)
        if len(sub) > len(longest) {
            longest = sub
        }
        mid++
    }
    return longest
}

func findLongestPalindromeByMid(s string, left, right int) string {
    for left >= 0 && right < len(s) {
        if s[left] != s[right] {
            break
        }
        left--
        right++
    }

    return s[left+1:right]
}
```



## 动态规划

动态规划是算法中大量出现的一个解法，我们以这道题为例，进行探索。

> 动态规划的细节请自行搜索，如 https://www.zhihu.com/question/39948290 

### 动态规划解法

```go
func longestPalindrome(s string) string {
	// 边界条件判断
	length := len(s)
	if length <= 1 {
		return s
	}

	// 回文子串至少为1，即单个字符
	begin, maxLen := 0, 1
	// 动态规划的关键数组dp
	var dp = make([][]bool, length)
	for k1 := range dp {
		dp[k1] = make([]bool, length)
		for k2 := range dp[k1] {
			// 初始化均为false
			dp[k1][k2] = false
		}
		// 单个字符默认为回文串
		dp[k1][k1] = true
	}

	for size := 2; size <= length; size++ {
		for start := 0; start <= length-size; start++ {
			end := start + size - 1

			// 如果首位字符串相同
			if s[start] == s[end] {
				if size == 2 {
					// 边界情况，即只有2个字符且相等
					dp[start][end] = true
				} else {
					// 核心：动态规划的核心推导函数
					dp[start][end] = dp[start+1][end-1]
				}
			}

			// 更新最长回文子串
			if dp[start][end] && size > maxLen {
				maxLen = size
				begin = start
			}
		}
	}
	return s[begin : begin+maxLen]
}
```

最核心的公式为`dp[start][end] = dp[start+1][end-1]`，其余都是对边界条件的处理。

但运行结果的效率偏低

- 执行用时：104 ms, 在所有 Go 提交中击败了42.53%的用户
- 内存消耗：7 MB, 在所有 Go 提交中击败了24.68%的用户

### 优化1 优化条件分支

我们分析一下上面的代码：从空间复杂度来看，动态规划的数组`dp`是不可或缺的，所以减少内存消耗比较困难了。

因此，我们把目光放在执行时间上，那么重点就是`for`循环里的代码复杂度。

我们注意到一个关键的变量`dp[start][end]`，对它进行改造

```go
func longestPalindrome(s string) string {
	length := len(s)
	if length <= 1 {
		return s
	}

	begin, maxLen := 0, 1
	var dp = make([][]bool, length)
	for k1 := range dp {
		dp[k1] = make([]bool, length)
		for k2 := range dp[k1] {
			dp[k1][k2] = false
		}
		dp[k1][k1] = true
	}

	for size := 2; size <= length; size++ {
		for start := 0; start <= length-size; start++ {
			end := start + size - 1

			if s[start] == s[end] {
				if size == 2 || dp[start+1][end-1] {
					dp[start][end] = true
					// 减少if-else的分支判断次数
					if size > maxLen {
						maxLen = size
						begin = start
					}
				}
			}
		}
	}
	return s[begin : begin+maxLen]
}
```

效果很明显，代码执行时间减少

- 执行用时：72 ms, 在所有 Go 提交中击败了50.93%的用户
- 内存消耗：7 MB, 在所有 Go 提交中击败了28.37%的用户



### 优化2 分析循环内代码执行逻辑

我们继续看代码里的`size > maxLen`条件，发现会出现如下情况：

- 如果一个字符有多个相同`size`的回文子串，这个`if`内的语句会被执行多次
- 但我们只需要获得最长的回文子串之一，所以只需要记录第一次即可

于是我们尝试改造：

```go
func longestPalindrome(s string) string {
	length := len(s)
	if length <= 1 {
		return s
	}

	begin, maxLen := 0, 1
	var dp = make([][]bool, length)
	for k1 := range dp {
		dp[k1] = make([]bool, length)
		for k2 := range dp[k1] {
			dp[k1][k2] = false
		}
		dp[k1][k1] = true
	}

	for size := 2; size <= length; size++ {
		// 用来表示对应size的字符串已经找到
		founded := false
		for start := 0; start <= length-size; start++ {
			end := start + size - 1
			if s[start] == s[end] {
				if size == 2 || dp[start+1][end-1] {
					dp[start][end] = true
					// 只记录第一个即可
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

- 执行用时：80 ms, 在所有 Go 提交中击败了49.30%的用户
- 内存消耗：7 MB, 在所有 Go 提交中击败了31.54%的用户

然而，执行结果告诉我们这部分的优化无明显效果，甚至还降低了评分。但我们无需沮丧，**因为这道题针对的是小规模场景的算法题，这个思路可能在大规模的计算场景下带来很明显的提升**。



## 总结

Leetcode第五题的难度不高，也让我们初次接触了 **动态规划** 这个思路。

同时，我们也遇到了一个**“失败的优化案例”**。这也从侧面告诉了我们，**抛开具体场景的优化都是不可靠的**。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

