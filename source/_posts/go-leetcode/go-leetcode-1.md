---
title: Go算法实战 - 1.【两数相加LeetCode-2】递归解法
date: 2021-07-10 12:00:00
categories: 
- 成长分享
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-2 两数相加

原题链接 https://leetcode-cn.com/problems/add-two-numbers/

```go
type ListNode struct {
    Val int
    Next *ListNode
}

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
}
```

为了能保证代码都能执行，我会贴出所有代码，**重点会用注释着重说明**。

<!-- more -->

## 递归实现的思路

### 简化问题

这道题的难点在于处理**进位**。那我们就先**简化问题、把框架搭起来**，看看先不考虑进位的大致代码怎么写的：

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    // 先判断边界情况
    if l1 == nil && l2 == nil {
        return nil
    }

    var node = new(ListNode)
    if l1 != nil && l2 != nil {
        node.Val = l1.Val + l2.Val
        node.Next = addTwoNumbers(l1.Next, l2.Next)
    } else if l1 != nil {
        node.Val = l1.Val 
        node.Next = addTwoNumbers(l1.Next, l2)
    } else if l2 != nil {
        node.Val = l2.Val 
        node.Next = addTwoNumbers(l1, l2.Next)
    }

    return node
}
```

但这块代码有个问题- 当`l1`和`l2`都为空时，还会进入一次`addTwoNumbers`，导致最高位必定是0。

所以，我们需要保证**最高位不要产生一个冗余，也就是l1和l2都为nil时，不要再进入addTwoNumbers函数**。



### 修复最高位的问题

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    var node = new(ListNode)
    // 这里l1和l2作为指针传递下去
    if l1 != nil && l2 != nil {
        node.Val = l1.Val + l2.Val
        l1, l2 = l1.Next, l2.Next
    } else if l1 != nil {
        node.Val = l1.Val 
        l1 = l1.Next
    } else if l2 != nil {
        node.Val = l2.Val
        l2 = l2.Next
    }

    // 如果都为空，无需继续处理
    if l1 == nil && l2 == nil {
        return node
    }

    // 继续处理下一个节点
    node.Next = addTwoNumbers(l1, l2)
    return node
}
```



### 递归函数增加进位参数carry

进位carry是一个在不同位中传递的参数，所以必须要加到函数签名中，所以我们得对递归函数进行改造。

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    return addTwoNumbersWithCarry(l1, l2, 0)
}

// 新的函数参数 carry
func addTwoNumbersWithCarry(l1 *ListNode, l2 *ListNode, carry int) *ListNode {
    var node = new(ListNode)
    if l1 != nil && l2 != nil {
        node.Val = l1.Val + l2.Val + carry
        l1, l2 = l1.Next, l2.Next
    } else if l1 != nil {
        node.Val = l1.Val + carry
        l1 = l1.Next
    } else if l2 != nil {
        node.Val = l2.Val + carry
        l2 = l2.Next
    } 

    var newCarry int
    if node.Val > 9 {
        node.Val = node.Val - 10
        newCarry = 1
    }

    if l1 == nil && l2 == nil && newCarry == 0 {
        return node
    }

    node.Next = addTwoNumbersWithCarry(l1, l2, newCarry)
    return node
}
```



### 边界条件修复

到了这里，我们看似完成了功能，但还有个边界条件没有修复：引入进位后，当`l1/l2`为nil，carry为1时，我们很容易就修复了

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    return addTwoNumbersWithCarry(l1, l2, 0)
}

func addTwoNumbersWithCarry(l1 *ListNode, l2 *ListNode, carry int) *ListNode {
    var node = new(ListNode)
    if l1 != nil && l2 != nil {
        node.Val = l1.Val + l2.Val + carry
        l1, l2 = l1.Next, l2.Next
    } else if l1 != nil {
        node.Val = l1.Val + carry
        l1 = l1.Next
    } else if l2 != nil {
        node.Val = l2.Val + carry
        l2 = l2.Next
    } else { 
        node.Val = carry // 修复进位的边界问题
    }

    var newCarry int
    if node.Val > 9 {
        node.Val = node.Val - 10
        newCarry = 1
    }

    if l1 == nil && l2 == nil && newCarry == 0 {
        return node
    }

    node.Next = addTwoNumbersWithCarry(l1, l2, newCarry)
    return node
}
```



## 持续优化

首先，先明确一下优化的原则：

**我并不是单纯地为了提升性能而去优化，而是更应该从全局入手，考虑代码的可读性和扩展性！**

所以，下面的优化并不一定是性能最优的，但或多或少可能让你感受到代码的迭代升级。



### A - 复用变量

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    return addTwoNumbersWithCarry(l1, l2, 0)
}

func addTwoNumbersWithCarry(l1 *ListNode, l2 *ListNode, carry int) *ListNode {
    var node = new(ListNode)
    if l1 != nil && l2 != nil {
        node.Val = l1.Val + l2.Val + carry
        l1, l2 = l1.Next, l2.Next
    } else if l1 != nil {
        node.Val = l1.Val + carry
        l1 = l1.Next
    } else if l2 != nil {
        node.Val = l2.Val + carry
        l2 = l2.Next
    } else {
        node.Val = carry
    }

    if node.Val > 9 {
        node.Val = node.Val - 10
        carry = 1 // 复用carry变量
    } else {
        carry = 0
    }

    if l1 == nil && l2 == nil && carry == 0 {
        return node
    }

    node.Next = addTwoNumbersWithCarry(l1, l2, carry)
    return node
}
```

删除`newCarry`变量，节省了内存。

虽然这点改进很小，但我想表达的重点是：**大家不要小看变量的复用，尤其是在一些递归调用的场景下，能节省大量的空间。**上面的`l1`与`l2`这两个指针也进行了变量的复用。



### B - 增加位操作，去除if-else分支

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    return addTwoNumbersWithCarry(l1, l2, 0)
}

func addTwoNumbersWithCarry(l1 *ListNode, l2 *ListNode, carry int) *ListNode {
    var node = new(ListNode)
    if l1 != nil && l2 != nil {
        node.Val = l1.Val + l2.Val + carry
        l1, l2 = l1.Next, l2.Next
    } else if l1 != nil {
        node.Val = l1.Val + carry
        l1 = l1.Next
    } else if l2 != nil {
        node.Val = l2.Val + carry
        l2 = l2.Next
    } else {
        node.Val = carry
    }

    carry, node.Val = node.Val/10, node.Val%10 // 引入位操作

    if l1 == nil && l2 == nil && carry == 0 {
        return node
    }

    node.Next = addTwoNumbersWithCarry(l1, l2, carry)
    return node
}
```



### C - 增加代码的扩展性（推荐）

在这个代码里，我们只支持2个`ListNode`的相加，就引入了4个`if-else`的分支，这就很难支持大量`ListNode`的扩展。

**总体来说，我个人推荐这个解法，它的思路很清晰，也不会出现边界问题。**

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
    carry, node.Val = node.Val/10, node.Val%10 // 引入位操作

    if l1 == nil && l2 == nil && carry == 0 {
        return node
    }

    node.Next = addTwoNumbersWithCarry(l1, l2, carry)
    return node
}
```



## 实战化特性

在实际的项目中，我们会希望这个函数的扩展性能更好，例如支持多个输入参数。

### 引入不定参数的特性

我们进一步改造成**不定参数**形式的函数签名：

```go
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    return addTwoNumbersWithCarry(0, l1, l2)
}

// 不定参数必须是最后一个函数签名
func addTwoNumbersWithCarry(carry int, nodes ...*ListNode) *ListNode {
    var node = new(ListNode)
    for k := range nodes {
        if nodes[k] != nil {
            node.Val += nodes[k].Val
            nodes[k] = nodes[k].Next
        }
    }

    node.Val += carry
    carry, node.Val = node.Val/10, node.Val%10 // 引入位操作

    // 判断所有node是否为空
    var isEnd = true
    for k := range nodes {
        if nodes[k] != nil {
            isEnd = false
            break
        }
    }
    if isEnd && carry == 0 {
        return node
    }

    node.Next = addTwoNumbersWithCarry(carry, nodes...)
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

