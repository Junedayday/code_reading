---
title: Go语言技巧 - 6.【深入Go Module】探索最小版本选择的机制
date: 2021-07-09 12:00:00
categories: 
- 成长分享
tags:
- Go-Tip
---

![Go-Study](https://i.loli.net/2021/05/05/2bmr98tG3xDneL5.jpg)

## 从一个示例讲起

用一个简单列表来表示我们的模块A依赖：

- B1`v1.0.0`
  - C1`v1.1.0`
  - C2 `v1.2.0`
- B2 `v1.2.0`
  - C1`v1.1.2`
  - C3 `v1.2.0`

表示 *A依赖B1与B2，而B1又依赖C1、C2，B2依赖C1、C3*。

这里，我们把关注点放到有争议的C1，它存在两个版本`v1.1.0`与`v1.1.2`。而最终A选择的是`v1.1.2`版本的C1。

- B1`v1.0.0`
  - ~~C1`v1.1.0`~~
  - C2 `v1.2.0`
- B2 `v1.2.0`
  - C1`v1.1.2`
  - C3 `v1.2.0`



### 问题1：为什么要选择较高版本的C1？

也许你会疑惑，为什么原则名字叫**最小版本选择**，但反而选择了较高那个版本呢？

我们要明确一点，**最小版本选择**这个概念不是应用在这个场景的！

从两个版本号的语义来看，`v1.1.2`和`v1.1.0`的主版本号都是`v1`，说明是向下兼容的。所以我们自然会选择较高的`v1.1.2`，毕竟如果用了`v1.1.0`，可能导致B2具体的代码不可用。



### 问题2：如果同时出现了v1和v2怎么办？

如果场景变化，C1的依赖版本为`v1.1.0`和`v2.0.0`，也就是大版本发生了变化。

从版本号的语义来看，两者是**不兼容**的！所以，这时不会出现**高版本覆盖低版本**的情况。

这时，就会出现依赖2个版本的C1。



### 问题3：那什么是最小版本选择中的“最小”呢？

在C1这个库中，我们能看到很多tag，例如`v1.1.0`，`v1.1.1`，`v1.1.2`，`v1.1.3`。而我们用到的是`v1.1.2`和`v1.1.0`。

从兼容性来看，`v1.1.3`肯定能兼容前面的版本。但这时，根据**最小版本选择**，我们引用到`v1.1.2`。

为什么要用这个最小版本原则，而不是每次都去拉取最新的tag？大家不妨思考思考，我这里列两个我能想到的点：

1. 保证项目依赖的稳定性：如果存在某个依赖库高频更新，会导致整个项目也频繁升级，造成风险；
2. 完全向下兼容并不可靠：毕竟软件存在不稳定性，最新的tag很有可能会导致代码变更；



## 结合源码巩固知识点

在阅读源码之前，我们先明确本次阅读源码的预期：**不要为了掌握所有代码细节而读代码，而是希望能通过了解这部分功能的一个大致实现，巩固理论知识**。

这里，我以`go语言1.15.11`版本为例，具体的代码路径在`src/cmd/go/internal/modcmd`下。

`go mod tidy`是整理Go Module最常用的指令之一，这里我们就来看看`tidy.go`文件。

### tidy的简介

```go
var cmdTidy = &base.Command{
	UsageLine: "go mod tidy [-v]",
	Short:     "add missing and remove unused modules",
	Long: `
Tidy makes sure go.mod matches the source code in the module.
It adds any missing modules necessary to build the current module's
packages and dependencies, and it removes unused modules that
don't provide any relevant packages. It also adds any missing entries
to go.sum and removes any unnecessary ones.

The -v flag causes tidy to print information about removed modules
to standard error.
	`,
}
```

`tidy`主要是把缺失的module加入到模块中，并删除弃用的modules。加上`-v`的标记位，就能把信息打印到标注错误。

### 核心代码

核心的数据结构为，储存Go Module的路径Path和版本Version:

```go
type Version struct {
	Path string
	Version string `json:",omitempty"`
}
```

而加载模块的代码，则是下面的`mvs.Req`函数：

```go
// cmd/go/internal/mvs
mvs.Req(Target, direct, &mvsReqs{buildList: keep})
```

这个函数的功能，我进行了一定的简化，大家关注重点标注出来的几行。

```go
func Req(target module.Version, base []string, reqs Reqs) ([]module.Version, error) {
  // 保存模块与其依赖module的map，用map是为了防止依赖库重复
	reqCache := map[module.Version][]module.Version{}
	reqCache[target] = nil
  
  // 第一次遍历：walk函数，用于遍历整个依赖
	var walk func(module.Version) error
	walk = func(m module.Version) error {
		// 获取m的依赖库required，保存到map中
		required, err := reqs.Required(m)
		if err != nil {
			return err
		}
		reqCache[m] = required
    // 继续遍历依赖的依赖，保证不缺失
		for _, m1 := range required {
			if err := walk(m1); err != nil {
				return err
			}
		}
		postorder = append(postorder, m)
		return nil
	}
  
  // 真正运行第一次walk的地方
	for _, m := range list {
		if err := walk(m); err != nil {
			return nil, err
		}
	}

  // 第二次遍历：再次定义一个walk函数，取最大的版本号
	have := map[module.Version]bool{}
	walk = func(m module.Version) error {
		if have[m] {
			return nil
		}
		have[m] = true
		for _, m1 := range reqCache[m] {
			walk(m1)
		}
		return nil
	}
	max := map[string]string{}
	for _, m := range list {
		if v, ok := max[m.Path]; ok {
      // 只保存较大的版本号
      // 而v1与v2的问题也是在这里解决的：两者的Path路径不同
			max[m.Path] = reqs.Max(m.Version, v)
		} else {
			max[m.Path] = m.Version
		}
	}
	
  // 真正运行第二次walk的地方
	var min []module.Version
	for _, path := range base {
		m := module.Version{Path: path, Version: max[path]}
		min = append(min, m)
		walk(m)
	}
	
  // 根据名称排序
	sort.Slice(min, func(i, j int) bool {
		return min[i].Path < min[j].Path
	})
	return min, nil
```



## 小结

**Minimal version selection (MVS)** 的整体实现看起来不复杂，但其实里面做了很多兼容性的工作，尤其是`indirect`和`incompatible`等特性。这其实在另一层面提醒了我们：**一项功能尽可能在前期做好设计，靠后期补救往往会增加大量兼容性的工作**。

整个Go Module的核心实现在于2点：

1. 2个`walk`函数，一个用于查找所有依赖，另一个选择最大依赖版本；
2. 选择最大依赖版本的核心依赖一个map，`max[m.Path] = reqs.Max(m.Version, v)`

至此，对Go Module的讲解告一段落了。而更多的细节问题，需要大家结合上一篇提到的排查问题工具，边实践、边加深理解。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili: https://space.bilibili.com/293775192
>
> 公众号: golangcoding
>
>  ![二维码](https://i.loli.net/2021/02/28/RPzy7Hjc9GZ8I3e.jpg)

