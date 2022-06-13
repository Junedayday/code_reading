---
title: Go算法实战 - 2.【两数相加LeetCode-2】非递归解法
date: 2022-04-06 08:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-2 两数相加

原题链接 https://leetcode-cn.com/problems/add-two-numbers/

我们继续看上一个题目，这次我们尝试写一个非递归的解法。

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
	// walker是为了在l1/l2里遍历，修改Next指针
	var carry, sentinel, walker = 0, l1, l1

	for {
		if l1 != nil {
			if walker != l1 {
				walker.Val += l1.Val
			}
			l1 = l1.Next
		}
		if l2 != nil {
			if walker != l2 {
				walker.Val += l2.Val
			}
			l2 = l2.Next
		}
		walker.Val += carry
		carry, walker.Val = walker.Val/10, walker.Val%10

		// 这里很重要，指定的是walker.Next的指向，能解决l1/l2跨链表的访问
		if l1 != nil {
			walker.Next = l1
		} else if l2 != nil {
			walker.Next = l2
		} else if carry > 0 {
			walker.Next = new(ListNode)
		} else {
			break
		}
		walker = walker.Next
	}

	return sentinel
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

