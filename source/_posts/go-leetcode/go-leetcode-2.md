---
title: Go算法实战 - 2.【两数相加LeetCode-2】非递归解法
date: 2021-07-10 12:00:00
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

为了能保证代码都能执行，我会贴出所有代码，**重点会用注释着重说明**。

> 我个人认为，非递归比递归写法更加麻烦，所以放到了第二讲。一开始直接上手用非递归的解法，很容易迷失在 边界条件 和 循环条件 中，排查问题也比较麻烦。

<!-- more -->

## 非递归实现的思路

### 简化问题

我们不考虑进位问题，看看大致的代码架构：

### 不考虑进位的解法

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    // 哨兵节点，也就是作为初始化的节点
    // 在单向链表时引入这个哨兵，有利于我们找到起始的点
    var sentinel = new(ListNode)
    // walker节点，也就是用于遍历的节点
    var walker = sentinel

    for l1 != nil || l2 != nil {
        var node = new(ListNode)
        if l1 != nil {
            node.Val += l1.Val
            l1 = l1.Next
        }
        if l2 != nil {
            node.Val += l2.Val
            l2l1 = l2.Next
        }
        // 把node追加到后面，walker继续往后走
        walker.Next = node
        walker = walker.Next
    }

    return sentinel.Next
}
```



### 增加进位参数carry

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    var sentinel = new(ListNode)
    var carry int
    var walker = sentinel

    for l1 != nil || l2 != nil || carry > 0{
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
        carry, node.Val = node.Val/10, node.Val%10 // 利用位操作

        walker.Next = node
        walker = walker.Next
    }

    return sentinel.Next
}
```

一般情况下，**非递归的实现会比递归的性能更高，但可读性较差**。

1. 非递归减少了函数的堆栈，所以性能更高；
2. 递归通过递归函数简化了复杂度，而非递归则需要循环；



## 持续优化

### A - 简单优化代码结构（推荐）

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    var sentinel = new(ListNode)
    var carry int
    var walker = sentinel

    for l1 != nil || l2 != nil || carry > 0{
        var node = new(ListNode)
        node.Val = carry // 将carry直接放到初始化位置
        if l1 != nil {
            node.Val += l1.Val
            l1 = l1.Next
        }
        if l2 != nil {
            node.Val += l2.Val
            l2 = l2.Next
        }
        carry, node.Val = node.Val/10, node.Val%10 

        // 将两个指针赋值放在一起，含义比较清晰
        // 建议此类表达式尽量用于 多个强相关的变量 赋值，而不要贪图方便
        walker.Next, walker = node, node 
    }

    return sentinel.Next
}
```

- 内存消耗：4.7 MB, 在所有 Go 提交中击败了29.47%的用户

### B - 进一步节省空间

至此，其实我们的代码已经相当简洁了，但有同学追求更好的数据。

这里，我们可以看一下，因为`l1`和`l2`这两个链表相加后，新的链表长度肯定是大于这两者的。所以，我们可以尝试复用一下其中一个链表，节省一下内存空间。这里，我们尝试复用一下链表`l1`。

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    var carry int
    var walker = l1

    // walker用于遍历l1，而l1指针自身不动，用于返回
    for walker != nil || l2 != nil || carry > 0{
        if l2 != nil {
            walker.Val += l2.Val
            l2 = l2.Next
        }
        walker.Val += carry 

        carry, walker.Val = walker.Val/10, walker.Val%10 
        walker = walker.Next
    }

    return l1
}
```

这段代码的整体逻辑是正确的，但存在边界问题：如果`l1`比`l2`短时，后续的元素怎么生成？

于是，我们就有了改进：

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    var carry int
    var walker = l1

    for walker != nil || l2 != nil || carry > 0{
        if l2 != nil {
            walker.Val += l2.Val
            l2 = l2.Next
        }
        walker.Val += carry 

        carry, walker.Val = walker.Val/10, walker.Val%10 
        // 当walker下个节点为nil时，但后续节点还需要继续遍历，就新建一个Node
        if walker.Next == nil && (l2 != nil || carry > 0) {
            walker.Next = new(ListNode)
        }
        walker = walker.Next
    }

    return l1
}
```

- 内存消耗：4.4 MB, 在所有 Go 提交中击败了96.97%的用户

### C - 再次节省空间

从上一个例子不难想到，我们还有继续优化的空间：如果`l2`比`l1`长时，我们想办法把walker节点指向`l2`，于是就有了下面的代码

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    var carry int
    // 由于两个链表均为非空，所以初始化会简单一点
    var sentinel = l1
    var walker = l1

    for l1 != nil || l2 != nil || carry > 0{
        if l1 != nil {
            // walker与l1为统一节点时，Val已经有值了
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
        
        // 这里就是去找下一个walker节点，先看l1，再看l2，最后看carry位有没有
        if l1 != nil {
            walker.Next = l1
            walker = walker.Next
        } else if l2 != nil {
            walker.Next = l2
            walker = walker.Next
        } else if carry > 0 {
            walker.Next = new(ListNode)
            walker = walker.Next
        }
    }

    return sentinel
}
```

我们再次做一个简单的优化

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    var carry int
    var sentinel, walker = l1, l1

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
        
        if l1 != nil {
            walker.Next = l1
        } else if l2 != nil {
            walker.Next = l2
        } else if carry > 0 {
            walker.Next = new(ListNode)
        } else {
            // 从循环中跳出，也就是l1/l2为nil,carry=0
            break
        }
        // 将上面三个判断分支中的共性提取出
        walker = walker.Next
    }

    return sentinel
}
```



## 总结

在解单向链表的问题时，`Sentinel哨兵` + `Walker遍历`是一个很好的组合。

- `Sentinel`放在单向链表的起始，指向我们的链表，能解决很多初始情况问题，例如链表本身为`nil`
- `Walker`是一个遍历指针，聚焦于`walker = walker.Next`这个关键的移动操作

总体来说，非递归的代码可读性会比递归的差一点，比较考验程序员的解题思路。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

