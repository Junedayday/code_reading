---
title: Go算法实战 - 3.【无重复字符的最长子串LeetCode-3】
date: 2021-07-14 12:00:00
categories: 
- 成长分享
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



## 递归解题思路

从函数签名`func lengthOfLongestSubstring(s string) int `来看，我们可以将问题拆分成子问题

1. 首字符
   1. 包含s[0]的最长无重复子串
   2. lengthOfLongestSubstring(s[1:n])
2. 尾字符
   1. lengthOfLongestSubstring(s[0:n-1])
   2. 包含s[n]的最长无重复子串

### 利用首字符的递归问题

```go
func lengthOfLongestSubstring(s string) int {
    // guard clauses，也就是卫语句，在递归中非常重要
    if len(s) <= 1 {
        return len(s)
    }
    // 全局的字符串，用于保存
    var mp = make(map[byte]int)
    var i int
    for i = 0 ;i < len(s); i++ {
        if _, ok := mp[s[i]]; ok {
            break // 找到重复的，直接退出
        } else {
            mp[s[i]] = i // 没找到重复的，加上这个子串
        }
    }

    length := lengthOfLongestSubstring(s[1:])
    if i > length {
        return i
    }
    return length
}
```

递归的解法可读性较佳，但实际的时间复杂度会比较高，因为它的每一次 **递归操作都要维护一个新的数据栈**。



## 非递归实现思路

### 先用“笨办法”实现

我们先不追求**复杂度**，先写一个简单的解法，搭建出解题框架：

从左往右看字符串s，如果我们限制了起始字符，然后逐个往右查找、找到第一个重复字符串，就是最长子串。

所以，如果s为`b1b2b3...bn`，它就能被拆解为 **起始字符为`b1`，`b2`,`b3` .... `bn`这样n个子问题中的最大值**。

于是，我们尝试写一下代码：

```go
func lengthOfLongestSubstring(s string) int {
    var max int
    // 遍历s，以i作为起始点，找到最长的子串
    // 注意：在string中如果用range的方法遍历，类型不是byte，而是rune
    for i := 0 ;i < len(s); i++ {
        var mp = make(map[byte]int)
        mp[s[i]] = i
        // j 作为一个从i往后移动的游标，找到第一个重复的词或者达到len(s)，也就是末尾
        var j int
        for j = i + 1; j < len(s); j++ {
            if _, ok := mp[s[j]]; ok {
                break // 找到重复的，直接退出
            } else {
                mp[s[j]] = j // 没找到重复的，加上这个子串
            }
        }
        // 算出以i为起点的最长子串
        length := j - i
        if length > max {
            max = length
        }
    }
    return max
}
```

这里面存在一些边界条件的判定，需要大家认证思考。



### 从核心map切入

在上面的解法中，我们用到了一个`map[byte]int`，用来保存 **字符与位置的映射关系**。但在整个循环的过程中，我们反复地`var mp = make(map[byte]int)`创建了空间。

由于s是一个固定的字符串，我们可以换一个思路尝试，先写出一个纯过程式的代码

```go
func lengthOfLongestSubstring(s string) int {
    // 全局的字符串，用于保存
    var mp = make(map[byte]int)
    var left = 0
    var max int
    for i := 0 ;i < len(s); i++ {
        // Case1: 这是一个暂时未出现过的字符
        if _, ok := mp[s[i]]; !ok {
            length := i - left + 1 // 到最左边的距离
            if length > max {
                max = length
            }
            mp[s[i]] = i // 不存在的新元素，直接添加进来
            continue // 打断逻辑
        }
        // Case2: 这是一个出现过的重复字符
        length := i - left + 1 // 到最左边的距离
        length2 := i - mp[s[i]] // 到上一个重复字符串的距离
        if length > length2 {
            length = length2 // 取较短值
            left = mp[s[i]] + 1
        }
        if length > max {
            max = length
        }
        mp[s[i]] = i // 更新索引
    }
    // 处理一下最后一个字符串
    if len(s) - left > max {
        max = len(s) - left 
    }

    return max
}
```

这个代码的关键实现在于**两个索引index**：

1. `i`，用于遍历s
2. `left`，0 <= left <= i，s[left:i]是**不存在重复字符的字符串**，其中left尽量取最小。换一个说法，s[left:i]是**以s[i]为右节点的、无重复的、最长的子字符串**。

我们再回头看一下上面的代码，可读性有不少改进空间，我们尝试做一下优化，让可读性更好：

```go
func lengthOfLongestSubstring(s string) int {
    var mp = make(map[byte]int)
    var left = 0
    var max int
    for i := 0 ;i < len(s); i++ {
        length := i - left + 1 // 到最左边的距离
        if _, ok := mp[s[i]]; ok {
            length2 := i - mp[s[i]] // 到上一个重复字符串的距离
            if length > length2 {
                length = length2
                left = mp[s[i]] + 1
            }
        }
        if length > max {
            max = length
        }
        mp[s[i]] = i 
    }
    if len(s) - left > max {
        max = len(s) - left 
    }

    return max
}
```

改动点并不大，但抽离了不少共性的代码，整体的性能有所提升。



## 总结

面对**明显可用递归方案解决**的题目时，个人比较推荐的解题思路是：

- **用递归的解决方案理清思路，写出一个可用的方案，此时不要关注性能**
- **从复杂度的角度思考，哪部分的工作是重复性的，提取出一个非递归的方案**

如果一开始就去抠所谓的最佳方案，很容易陷入细节问题，而丢失了全局视野。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

