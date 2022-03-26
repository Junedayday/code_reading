---
title: Go算法实战 - 4.【寻找两个正序数组的中位数LeetCode-4】
date: 2021-07-19 12:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-4 寻找两个正序数组的中位数

原题链接 https://leetcode-cn.com/problems/median-of-two-sorted-arrays/

```go
func findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
}
```

在解这个题之前，我们要注意**奇数和偶数**的边界情况。

- 奇数2n+1个，我们要取第n+1小的数
- 偶数2n个，我们要取第n和n+1小的数

在Go语言中，因为是强类型的，切片`nums1`与`nums2`是整数，返回值则是浮点数

> 这是我们遇到的第一道hard级别的题目，让我们一起尝试攻克它！

<!-- more -->

## 基本解法

### 常规思路1 - 逐个寻找

常规思路来看，我们就是找第X小的数，那我们就一个一个找：

```go
func findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
    length1, length2 := len(nums1), len(nums2)
    // 注意，这里是向下取整
    mid := (length1 + length2) / 2

    // 区分一下奇数与偶数，奇数为mid+1，偶数为mid/mid+1
    // 奇数为2n+1个，mid=n,这样下一个就是中位数
    // 偶数为2n个，mid=n，所以处理一下，让mid=n-1，这样接下来两个就是中位数
    var isOdd bool
    if (length1 + length2) % 2 == 1 {
        isOdd = true
    } else {
        mid =  mid - 1
    }
    
    var i1, i2 int
    // 移动两个索引i1与i2，找到最小的mid个
    // 关键是注意两个边界情况的判定
    for mid > 0 {
        if i1 >= len(nums1) {
            i2++
        } else if i2 >= len(nums2) {
            i1++
        } else if nums1[i1] > nums2[i2] {
            i2++
        } else {
            i1++
        }
        
        mid--
    }

    // 找到mid后下一个
    var n1 int
    if i1 >= len(nums1) {
        n1 = nums2[i2]
        i2++
    } else if i2 >= len(nums2) {
        n1 = nums1[i1]
        i1++
    } else if nums1[i1] > nums2[i2] {
        n1 = nums2[i2]
        i2++
    } else {
        n1 = nums1[i1]
        i1++
    }

    // 奇数直接返回结果
    if isOdd {
        return float64(n1)
    }

    // 偶数找到再下一个，再返回
    var n2 int
    if i1 >= len(nums1) {
        n2 = nums2[i2]
    } else if i2 >= len(nums2) {
        n2 = nums1[i1]
    } else if nums1[i1] > nums2[i2] {
        n2 = nums2[i2]
    } else {
        n2 = nums1[i1]
    }

    return float64(n1 + n2) / 2
}
```

这个代码很长，但性能不差，因为都是顺序的逻辑。

分析一下复杂度，空间复杂度是O(1)，时间复杂度为O(m+n)，其中m=len(nums1)，n=len(nums2)。



### 常规思路2 - 排序后直接根据索引查找

```go
func findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
    // Go里面不支持对 []int 直接进行排序，必须通过 sort.IntSlice 做一次转化
    var intSlice sort.IntSlice
    intSlice = append(intSlice, nums1...)
    intSlice = append(intSlice, nums2...)

    sort.Sort(intSlice)
    if len(intSlice) % 2 == 1 {
        // 注意这里的索引，切勿再+1
        return float64(intSlice[len(intSlice) / 2])
    }
    return float64(intSlice[len(intSlice) / 2 - 1] + intSlice[len(intSlice) / 2 ]) / 2
}
```

我个人觉得Go语言里的排序函数`sort.Sort`对使用者的体验不是很好，尤其是对一些基础类型的支持。

这个解法的问题是对nums1和nums2进行了重新排序，没有充分利用nums1与nums2为**有序数组**这个条件。



## 进阶解法

### 递归的基本思路

我们注意到，这道题中给出的两个数组都是 **已排序** 的，所以可以利用**数组的随机访问**特性，做一定的加速。

这道题的问题是取中位数，但由于**数组长度的奇偶性**问题，这个中位数很难递归。所以，我们需要将问题做一个转化，实现**找到第K小的数字**，也就是下面的`findKthSortedArrays`函数。

```go
func findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
    length1, length2 := len(nums1), len(nums2)
    // 奇数可以一步计算得出
    if (length1 + length2) % 2 == 1 {
        return findKthSortedArrays(nums1, nums2, (length1 + length2) / 2 + 1)
    }
    // 偶数的拆分为两个子问题
    return (findKthSortedArrays(nums1, nums2, (length1 + length2) / 2) +
        findKthSortedArrays(nums1, nums2, (length1 + length2) / 2 + 1)) / 2
}

func findKthSortedArrays(nums1 []int, nums2 []int, k int) float64 {
    length1, length2 := len(nums1), len(nums2)
    // 技巧1：保证nums1比nums2长，能减少下面很多条件的判断
    if length1 < length2 {
        return findKthSortedArrays(nums2, nums1, k)
    } else if length2 == 0 {
        return float64(nums1[k - 1])
    } else if k == 1 {
        if nums1[0] > nums2[0] {
            return float64(nums2[0])
        }
        return float64(nums1[0])
    }

    // 我们要保证i1和i2都要小于k，否则下面很难递归
    i2 := k / 2
    if i2 > length2 {
      i2 = length2
    }
    i1 := k - i2

    // 递归的核心思路
    if nums1[i1 - 1] < nums2[i2 - 1] {
        return findKthSortedArrays(nums1[i1:], nums2, k - i1)
    } else if (nums1[i1 - 1] > nums2[i2 - 1]) {
        return findKthSortedArrays(nums1, nums2[i2:], k - i2)
    }
    return float64(nums1[i1 - 1])
}
```

解决这道题的难点在于3点：

- 大量**边界条件**的编写，很容易发生遗漏导致运行失败
- 第n个数字转化为数组索引时，自带一个`-1`操作，在递归时容易混淆
- 递归的核心思路：**将第k个元素转化为2个数组的索引之和，并保证不小于各自数组的长度**



## 总结

解决本题的难点在于大量的条件判断，存在大量`if-else`的代码，很容易让我们在编写代码时产生混乱，常常需要大量的调试。

我个人比较推荐用**常规解法1**这种笨办法来思路梳理，完整地处理好各种边界条件。接下来，再通过**进阶解法**来提升自己的抽象水平。

恭喜我们，正式解决了第一道困难的Leetcode！



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

