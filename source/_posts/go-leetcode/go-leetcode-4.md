---
title: Go算法实战 - 4.【寻找两个正序数组的中位数LeetCode-4】
date: 2022-04-07 08:00:00
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

<!-- more -->

## 题解

```go

func findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
	l1, l2 := len(nums1), len(nums2)
	if (l1+l2)%2 == 1 {
		return float64(findKthInSortedArrays(nums1, nums2, (l1+l2)/2+1))
	}
	return float64(findKthInSortedArrays(nums1, nums2, (l1+l2)/2)+findKthInSortedArrays(nums1, nums2, (l1+l2)/2+1)) / 2
}

func findKthInSortedArrays(nums1 []int, nums2 []int, k int) int {
	fmt.Println(nums1, nums2, k)
	length1, length2 := len(nums1), len(nums2)
	if length1 < length2 {
		return findKthInSortedArrays(nums2, nums1, k)
	} else if length2 == 0 {
		return nums1[k-1]
	} else if k == 1 {
		if nums1[0] > nums2[0] {
			return nums2[0]
		}
		return nums1[0]
	}

	// 取i1/i2为k/2，并处理越界
	i1, i2 := k/2, k-k/2
	if i2 > length2 {
		i2 = length2
		i1 = k - length2
	}

	// 截断小的数组后，继续递归查找
	if nums1[i1-1] < nums2[i2-1] {
		return findKthInSortedArrays(nums1[i1:], nums2, k-i1)
	} else if nums1[i1-1] > nums2[i2-1] {
		return findKthInSortedArrays(nums1, nums2[i2:], k-i2)
	}
	return nums1[i1-1]
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

