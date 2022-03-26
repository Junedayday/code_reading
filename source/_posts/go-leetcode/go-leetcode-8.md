---
title: Go算法实战 - 8.【三数之和LeetCode-15】
date: 2021-08-08 12:00:00
categories: 
- 算法实战
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-15 三数之和

原题链接 https://leetcode-cn.com/problems/3sum/

```go
func threeSum(nums []int) [][]int {
}
```

<!-- more -->

## 基础解法

### 基本思路

在看到这道题的后，我们很自然地可以想到简单的解法，例如穷举出所有的值。这个代码我就不专门写了。

### 利用排序进行优化

由于这道题返回的`[][]int`要求是对应的值，而不是索引，所以我们可以利用排序做一定的加速，示例代码如下：

```go

func threeSum(nums []int) [][]int {
	// 先排序，为了方便处理
	sort.Ints(nums)
	// 用于去重
	var solutionMap = make(map[[3]int]struct{})

	for i := 0; i < len(nums)-2; i++ {
		for j := i + 1; j < len(nums)-1; j++ {
			for k := j + 1; k <= len(nums)-1; k++ {
				if nums[i]+nums[j]+nums[k] == 0 {
					// 取到一个解即可
					solutionMap[[3]int{nums[i], nums[j], nums[k]}] = struct{}{}
					break
				} else if nums[i]+nums[j]+nums[k] > 0 {
					// 如果已经大于0了，由于nums是递增的，无需继续循环下去了
					break
				}
			}
		}
	}
	// 取一下去重后的解
	var result [][]int
	for k := range solutionMap {
		result = append(result, append([]int{}, k[0], k[1], k[2]))
	}
	return result
}
```

但运行下来，在数据量大的情况下还是会超出时间限制

### 利用二分查找加速

我们把目光聚焦到`k`，在一个有序的数组中，可以利用二分查找加速

```go
func threeSum(nums []int) [][]int {
	sort.Ints(nums)
	var solutionMap = make(map[[3]int]struct{})

	for i := 0; i < len(nums)-2; i++ {
		for j := i + 1; j < len(nums)-1; j++ {
			// 用二分查找加速
			// 但需要注意的是，go里的sort并不是精确匹配，所以需要二次判断
			k := sort.SearchInts(nums[j+1:], -nums[i]-nums[j])
			if k < len(nums[j+1:]) && nums[i]+nums[j]+nums[j+k+1] == 0 {
				solutionMap[[3]int{nums[i], nums[j], nums[j+k+1]}] = struct{}{}
			}
		}
	}
	var result [][]int
	for k := range solutionMap {
		result = append(result, append([]int{}, k[0], k[1], k[2]))
	}
	return result
}
```

至此，代码已经可以通过验证，我们再看看有什么进一步的优化空间。



### 优化1：减少元素的存储

```go
func threeSum(nums []int) [][]int {
	sort.Ints(nums)
	// 缩小元素的存储
	var solutionMap = make(map[[2]int]struct{})

	for i := 0; i < len(nums)-2; i++ {
		for j := i + 1; j < len(nums)-1; j++ {
			k := sort.SearchInts(nums[j+1:], -nums[i]-nums[j])
			if k < len(nums[j+1:]) && nums[i]+nums[j]+nums[j+k+1] == 0 {
				solutionMap[[2]int{nums[i], nums[j]}] = struct{}{}
			}
		}
	}
	var result [][]int
	for k := range solutionMap {
		result = append(result, append([]int{}, k[0], k[1], -k[0]-k[1]))
	}
	return result
}
```

### 优化2：初始化切片大小，防止扩容效率

```go
func threeSum(nums []int) [][]int {
	sort.Ints(nums)
	var solutionMap = make(map[[2]int]struct{})

	for i := 0; i < len(nums)-2; i++ {
		for j := i + 1; j < len(nums)-1; j++ {
			k := sort.SearchInts(nums[j+1:], -nums[i]-nums[j])
			if k < len(nums[j+1:]) && nums[i]+nums[j]+nums[j+k+1] == 0 {
				solutionMap[[2]int{nums[i], nums[j]}] = struct{}{}
			}
		}
	}

	// 初始化切片空间
	var result = make([][]int, len(solutionMap))
	i := 0
	for k := range solutionMap {
		result[i] = []int{k[0], k[1], -k[0] - k[1]}
		i++
	}
	return result
}
```



### 优化3：利用map加速查询

```go
func threeSum(nums []int) [][]int {
	sort.Ints(nums)
	var dataCountMap = make(map[int]int)
	for _, v := range nums {
		if _, ok := dataCountMap[v]; !ok {
			dataCountMap[v] = 1
		} else {
			dataCountMap[v]++
		}
	}
	var solutionMap = make(map[[2]int]struct{})

	for i := 0; i < len(nums)-2; i++ {
		for j := i + 1; j < len(nums)-1; j++ {
			expected := -nums[i] - nums[j]
			if num, ok := dataCountMap[expected]; ok && expected >= nums[j] {
				if expected == nums[j] {
					num--
				}
				if expected == nums[i] {
					num--
				}
				if num > 0 {
					solutionMap[[2]int{nums[i], nums[j]}] = struct{}{}
				}
			}
		}
	}

	var result = make([][]int, len(solutionMap))
	i := 0
	for k := range solutionMap {
		result[i] = []int{k[0], k[1], -k[0] - k[1]}
		i++
	}
	return result
}
```



## 进阶思路 -  双指针

我们把眼光放回到这个问题。通过排序，其实我们已经将问题变得比较清晰了。

在这个题目中，有三个关键的变量，我们可以将其中一个固定，例如`i`，将问题简化为`nums[j]+nums[k]=-nums[i]`。

于是，问题就在于`j`和`k`这两个坐标的移动。整体的代码思路并不难，但性能的提升集中在**对剪枝情况的处理**，尤其是值相同的元素。

```go
func threeSum(nums []int) [][]int {
	sort.Ints(nums)
	var result [][]int

	for i := 0; i < len(nums)-2; i++ {
		// 剪枝：最小值大于0时无需再遍历
		if nums[i] > 0 {
			break
		}
		// 剪枝：最小值和前一个值一样时，上一个循环已经判断过，无需再判断
		if i > 0 && nums[i] == nums[i-1] {
			continue
		}
		// j,k 为两个指针，分别从最左边和最右边开始移动
		j, k := i+1, len(nums)-1
		for j < k {
			left, right := nums[j], nums[k]
			if nums[i]+nums[j]+nums[k] == 0 {
				result = append(result, []int{nums[i], nums[j], nums[k]})
				// 减枝：跳过 nums[j] == nums[j+1] 的情况
				for j < k && nums[j] == left {
					j++
				}
				// 减枝：跳过 nums[k] == nums[k-1] 的情况
				for j < k && nums[k] == right {
					k--
				}
			} else if nums[i]+nums[j]+nums[k] < 0 {
				// 和小于0，则增大最左边的j
				j++
			} else {
				// 和大于0，则减少最右边的k
				k--
			}
		}
	}
	return result
}
```



## 总结

这道题的难度并不高，我们可以快速地实现这块代码。

与此同时，我们将更多的注意力放在了**剪枝**的情况，也就成为了最终算法是否高效的关键因素。在实际的工程中，**剪枝**是一个很重要的思想，我们经常要**根据具体的数据特征进行策略调整**。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

