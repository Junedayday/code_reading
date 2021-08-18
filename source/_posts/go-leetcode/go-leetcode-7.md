---
title: Go算法实战 - 7.【盛最多水的容器LeetCode-11】
date: 2021-08-02 12:00:00
categories: 
- 成长分享
tags:
- Go-Leetcode
---

![Go-Leetcode](https://i.loli.net/2021/07/10/SbG3k5XFRlsJdOV.jpg)

## Leetcode-11 盛最多水的容器

原题链接 https://leetcode-cn.com/problems/container-with-most-water/

```go
func maxArea(height []int) int {
}
```

<!-- more -->

## 基础解法

### 基本的递归

我们先通过递归来解一下这个问题：

```go
func maxArea(height []int) int {
    if len(height) <= 1 {
        return 0
    } else if len(height) == 2 {
        if height[0] > height[1] {
            return height[1]
        }
        return height[0]
    }

    // 右边固定为height[len(height) - 1]，左边不断移动，寻找最大的区域
    var max int
    for i := 0; i < len(height) - 1; i++ {
        right := height[len(height) - 1]
        if height[i] < right {
            right = height[i]
        }
        area := (len(height) - 1 - i) * right
        if area > max {
            max = area
        }
    }

    // height去掉最右边的一个点，拆解为子问题
    subArea := maxArea(height[:len(height) - 1])
    if subArea > max{
        return subArea
    }
    return max
}
```

整个代码的逻辑没有问题，但在线上执行的结果是**超出了时间限制**，也就是递归太深。

我们能否想个办法，做到**减枝**？我们尝试下将已经算出来的区域传递下去。

### 利用减枝

```go
func maxArea(height []int) int {
	// 处理初始边界条件
	if len(height) <= 1 {
		return 0
	} else if len(height) == 2 {
		if height[0] > height[1] {
			return height[1]
		}
		return height[0]
	}

	return maxAreaWithAera(height, 0)
}

func maxAreaWithAera(height []int, area int) int {
	if len(height) <= 1 {
		return area
	}

	right := height[len(height)-1]
	if right != 0 {
		leftIndex := len(height) - 2
		// 关键在于理解 len(height)-1-area/right，也就是左边至少从右边的边偏移area/right，才有可能大于area
		if area != 0 && leftIndex > len(height)-1-area/right {
			leftIndex = len(height) - 1 - area/right
		}
		// leftIndex往左偏移
		for ; leftIndex >= 0; leftIndex-- {
			right := height[len(height)-1]
			if height[leftIndex] < right {
				right = height[leftIndex]
			}
			a := (len(height) - 1 - leftIndex) * right
			if a > area {
				area = a
			}
		}
	}

	return maxAreaWithAera(height[:len(height)-1], area)
}
```

重点就是`len(height)-1-area/right`这个值，这里利用了传递的`area`进行**减枝**。



## 进阶解法

### 梳理思路

我们跳出代码，来思考一下整个问题解决的宏观思路

1. 当左边与右边确定时，假设索引为`l1`和`r1`，区域大小是固定的
   1. `area = min(height[l1], height[r1]) * (r1 - l1)`
2. 接下来，我们要简化问题，也就是要将**[]height的左边界往右移或者右边界往左移**
   1. 无论如何移动，**x轴是不断缩小的**，所以问题在于`左边界height[l2]`和右边界的高度`height[r2]`
   2. `area2 = min(height[l2], height[r2]) * (r2 - l2)`
3. `l2`和`r2`同时改变的话，整个计算方式就会很复杂，那我们就尝试固定其中一个不变，例如
   1. `l1 = l2` 并且 `r1 > r2`，即**右边界往左移动**，此时
   2. `area = min(height[l1], height[r1]) * (r1 - l1)`
   3. `area2 = min(height[l1], height[r2]) * (r2 - l1)`
4. 有什么办法可以对比`area`与`area2` 呢？
   1. `r1 - l1 > r2 - l1`可根据条件快速判断
   2. 核心在于对比 `min(height[l1], height[r1])` 和 `min(height[l1], height[r2])`
      1. 如果 `height[r1])` >= `min(height[l1]`，也就是**[]height高度最右边最高于左边**，那么 `min(height[l1], height[r1])` >=  `min(height[l1], height[r2])`成立，此时 area >= area2 也必定成立
      2. 如果 `height[r1])` < `min(height[l1]`，那么 area 与 area2 的关系没法判断
   3. 归纳一下上面这个情况：就是**当[]height高度最右边高于最左边时，移动右边面积肯定变小，移动左边面积变化未知**。
5. 用更通用的说法就是，如果要找到[]height子集中更大的面积，**固定较高边，移动较低边**。在编程中，这种接法往往称为**双指针**。



### 双指针解法

```go
func maxArea(height []int) int {
    left, right := 0, len(height)-1
    var max int
    for left < right {
        var area int
        if height[left] > height[right] {
            area = height[right] * (right - left)
            right--
        } else {
            area = height[left] * (right - left)
            left++
        }
        
        if area > max {
            max = area
        }
    }
    return max
}
```

**双指针解法很简洁**，但最重要的是推导过程。



## 总结

在编写代码的过程中，我们很难一步到位就写出最佳实现的解法，而前面的**递归+减枝**方法，虽然代码比较复杂，但是更符合我们直观逻辑的。**双指针解法**并不直观，这也就是体现出了刷题的价值。

值得一提的是，如果你上手就写出双指针解法，面试官会认为你是靠刷题记忆的，所以在面试算法的过程中，我们更应该关注**解决问题的递进式思路**，答案只是评价算法能力的其中一个重要项。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

