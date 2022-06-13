---
title: Go算法实战 - 6.【正则表达式匹配LeetCode-10】
date: 2022-04-09 08:00:00
categories: 
- 算法实战
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

## 题解

```go
func isMatch(s string, p string) bool {
	row, col := len(s), len(p)

	// dp 就是核心的状态转移方程，这里注意要+1，是为了空字符串这个边界条件
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
			// 如果p[j-1]为*，则可以截断*和它前面的一个字符，表示匹配0个对应字符
			dp[0][j] = dp[0][j-2]
		}
	}

	// 填充整个dp数组，注意i和j在dp中不变，但对应到字符串s/p中都要-1
	for i := 1; i < row+1; i++ {
		for j := 1; j < col+1; j++ {
            // 最后一个字符是*的话
			if p[j-1] == '*' {
				if s[i-1] == p[j-2] || p[j-2] == '.' {
                    // *匹配上一个字符，要么截断s一个字符，要么去掉*和前一个字符
					dp[i][j] = dp[i][j-2] || dp[i-1][j]
				} else {
					// 如果不匹配，则认为*没匹配上，只能去掉*和前一个字符
					dp[i][j] = dp[i][j-2]
				}
			} else if s[i-1] == p[j-1] || p[j-1] == '.' {
				// 如果精确匹配或者匹配上了.，就各自截断后往前找
				dp[i][j] = dp[i-1][j-1]
			}
		}
	}

	return dp[row][col]
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

