---
title: Go算法实战 - 11.【反转链表LeetCode-206】
date: 2022-04-01 12:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-206 反转链表

原题链接 https://leetcode-cn.com/problems/reverse-linked-list/

```go
type ListNode struct {
	Val  int
	Next *ListNode
}

func reverseList(head *ListNode) *ListNode {

}
```

<!-- more -->

## 题解

递归

```go
func reverseList(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}
	var p = reverseList(head.Next)
	// 重点：调整两个前置指针
	head.Next.Next = head
	head.Next = nil
	return p
}
```

非递归

```go
func reverseList(head *ListNode) *ListNode {
	// 这里初始为nil，解决了第一个head指向为nil
	var pre *ListNode
	for head != nil {
		// 存一下next指针，防止丢失
		next := head.Next
		// head指向前驱节点
		head.Next = pre
		// 两个指针往后挪，注意先后顺序
		pre = head
		head = next
	}
	return pre
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

