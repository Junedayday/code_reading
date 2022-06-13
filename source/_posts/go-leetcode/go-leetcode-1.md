---
title: Go算法实战 - 1.【两数相加LeetCode-2】递归解法
date: 2022-04-05 08:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-2 两数相加

原题链接 - https://leetcode-cn.com/problems/add-two-numbers/

```go
type ListNode struct {
    Val int
    Next *ListNode
}

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
}
```

<!-- more -->

## 题解

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    return addTwoNumbersWithCarry(l1, l2, 0)
}

func addTwoNumbersWithCarry(l1 *ListNode, l2 *ListNode, carry int) *ListNode {
    var node = new(ListNode)
    if l1 != nil {
        node.Val += l1.Val
        l1 = l1.Next
    }

    if l2 != nil {
        node.Val += l2.Val
        l2 = l2.Next
    }

    node.Val += carry
    // 引入位操作
    carry, node.Val = node.Val/10, node.Val%10 

    // 没有后续节点
    if l1 == nil && l2 == nil && carry == 0 {
        return node
    }

    node.Next = addTwoNumbersWithCarry(l1, l2, carry)
    return node
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

