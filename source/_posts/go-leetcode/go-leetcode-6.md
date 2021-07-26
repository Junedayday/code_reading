---
title: Go算法实战 - 6.【正则表达式匹配LeetCode-10】
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

