---
title: Go算法实战 - 6.【正则表达式匹配LeetCode-10】
date: 2021-07-28 12:00:00
categories: 
- 成长分享
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-10 正则表达式匹配

原题链接 https://leetcode-cn.com/problems/regular-expression-matching/

```go
func isMatch(s string, p string) bool {
}
```

<!-- more -->

## 基础解法

我们先理一下正则匹配的大致思路：**逐个对比s和p两个字符串，匹配则继续往后，发现不匹配直接退出**。

那么，我们先简化一下问题，看看代码的大致结构：



### 普通字符串匹配

```go
func isMatch(s string, p string) bool {
    if len(s) == 0 && len(p) == 0 {
        return true
    } else if len(s) == 0 || len(p) == 0 || s[0] != p[0] {
        return false
    }

    return isMatch(s[1:], p[1:])
}
```

这里的代码思路很清晰，我们重点要解决的是两个通配符`.`和`*`：

### 单个字符匹配

```go
func isMatch(s string, p string) bool {
    if len(s) == 0 && len(p) == 0 {
        return true
    } else if len(s) == 0 || len(p) == 0 {
        return false
    }

    // 如果p[0]为. ，则必定匹配
    if s[0] != p[0] && p[0] != '.' {
        return false
    }

    return isMatch(s[1:], p[1:])
}
```

## *匹配

接下来，我们就要解决最复杂的*匹配，也就是star符号。具体的解法我在下面给出，大家可以参考注释阅读：

```go

func isMatch(s string, p string) bool {
	if len(s) == 0 && len(p) == 0 {
		return true
	} else if len(s) == 0 {
		// 边界情况，即s为空，p前两个为 x*
		if len(p) >= 2 && p[1] == '*' {
			return isMatch(s, p[2:])
		}
		return false
	} else if len(p) == 0 {
		return false
	}

	// p是否为 x* 形式
	var hasStar bool
	if len(p) >= 2 && p[1] == '*' {
		hasStar = true
	}

	// isMatch表示s与p的第一个字符是否匹配
	var isMatched = true
	if s[0] != p[0] && p[0] != '.' {
		isMatched = false
	}

	if hasStar {
		if isMatched {
			// 情况1： 有星且第一个字符匹配，则递归包括2个情况：s去掉第一个字符，p去掉star这两个字符
			return isMatch(s[1:], p) || isMatch(s, p[2:])
		}
		// 情况2：有星且不匹配，则去掉p的前两个字符继续匹配
		return isMatch(s, p[2:])
	} else if !isMatched {
		// 情况3：没星且不匹配，则直接返回不匹配
		return false
	}
	// 情况4：没有星但是匹配，s和p删掉匹配的第一个字符，继续匹配
	return isMatch(s[1:], p[1:])
}
```

这个解法虽然看过去复杂，但是比较直观，核心在于两个变量`hasStar`和`isMatch`，以及它们组合起来的四个情况。



## 动态规划解

动态规划是一个面试高频的题，其核心是**状态转移方程**。这道题很符合动态规划的特征，我们通过了上面的递归解法，其实已经有了基本的思路：**递归中的四种情况，其实就是状态转移方程的大致思路**。

```go
func isMatch(s string, p string) bool {
	row, col := len(s), len(p)

	// dp 就是核心的状态转移方程，这里注意要+1，是为了空字符串这个边界条件
	// 所以后面的i/j默认都要-1
	dp := make([][]bool, row+1)
	for i := 0; i < len(dp); i++ {
		dp[i] = make([]bool, col+1)
	}

	// 填充dp[0]数组，也就是s为空字符串
	for j := 0; j < col+1; j++ {
		if j == 0 {
			// p为空字符串的情况
			dp[0][0] = true
		} else if p[j-1] == '*' {
			// 如果p[j-1]为*，则可以认为匹配p和p[0:j-2]一样，类似于情况2
			dp[0][j] = dp[0][j-2]
		}
	}

	// 填充整个dp数组，注意i和j在dp中不变，但对应到字符串s/p中都要-1
	for i := 1; i < row+1; i++ {
		for j := 1; j < col+1; j++ {
			if p[j-1] == '*' {
				if i != 0 && (s[i-1] == p[j-2] || p[j-2] == '.') {
					// 对应情况1，有星且第一个字符匹配
					dp[i][j] = dp[i][j-2] || dp[i-1][j]
				} else {
					// 对应情况2，有星且不匹配
					dp[i][j] = dp[i][j-2]
				}
			} else if i != 0 && (s[i-1] == p[j-1] || p[j-1] == '.') {
				// 对应情况4，没有星但是匹配
				dp[i][j] = dp[i-1][j-1]
			}
			// 其余的对应情况3，没星且不匹配，即默认false
		}
	}

	return dp[row][col]
}
```

只要有了递归解法的思路，动态规划的难度并不高。



## 总结

我们又完成了一道hard级别的题目！

这道题，让我们看到了递归与动态规划存在共性。其中，递归解法的核心思路是**将问题拆解为复杂度更低的子问题，直到边界情况**，而动态规划解法的核心思路是**从边界情况开始推导，从复杂度低的问题推导出复杂度更高的问题**。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

