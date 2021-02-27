---
title: Go编程模式 - 5.函数式选项
date: 2021-02-20 18:31:59
categories: 
- 经典品读
tags:
- Go-Programming-Patterns
---

>  注：本文的灵感来源于GOPHER 2020年大会陈皓的分享，原PPT的[链接](https://www2.slideshare.net/haoel/go-programming-patterns?from_action=save)可能并不方便获取，所以我下载了一份[PDF](https://github.com/Junedayday/code_reading/tree/master/doc/Go_Programming_Patterns.pdf)到git仓，方便大家阅读。我将结合自己的实际项目经历，与大家一起细品这份文档。



## 目录

- [一个常见的HTTP服务器](#ServerConfig)
- [拆分可选配置](#SplitConfig)
- [函数式选项](#Functional-Option)
- [更进一步](#Further)



## ServerConfig

我们先来看看一个常见的HTTP服务器的配置，它区分了2个必填参数与4个非必填参数

```go
type ServerCfg struct {
	Addr     string        // 必填
	Port     int           // 必填
	Protocol string        // 非必填
	Timeout  time.Duration // 非必填
	MaxConns int           // 非必填
	TLS      *tls.Config   // 非必填
}

// 我们要实现非常多种方法，来支持各种非必填的情况，示例如下
func NewServer(addr string, port int) (*Server, error)                                   {}
func NewTLSServer(addr string, port int, tls *tls.Config) (*Server, error)               {}
func NewServerWithTimeout(addr string, port int, timeout time.Duration) (*Server, error) {}
func NewTLSServerWithMaxConnAndTimeout(addr string, port int, maxconns int, timeout time.Duration, tls *tls.Config) (*Server, error) {}
```



## SplitConfig

编程的一大重点，就是要 `分离变化点和不变点`。这里，我们可以将必填项认为是不变点，而非必填则是变化点。

我们将非必填的选项拆分出来。

```go
type Config struct {
	Protocol string
	Timeout  time.Duration
	MaxConns int
	TLS      *tls.Config
}

type Server struct {
	Addr string
	Port int
	Conf *Config
}

func NewServer(addr string, port int, conf *Config) (*Server, error) {
	return &Server{
		Addr: addr,
		Port: port,
		Conf: conf,
	}, nil
}

func main() {
	srv1, _ := NewServer("localhost", 9000, nil)

	conf := Config{Protocol: "tcp", Timeout: 60 * time.Second}
	srv2, _ := NewServer("localhost", 9000, &conf)

	fmt.Println(srv1, srv2)
}
```

到这里，其实已经满足大部分的开发需求了。那么，我们将进入今天的重点。



## Functional Option

```go
type Server struct {
	Addr     string
	Port     int
	Protocol string
	Timeout  time.Duration
	MaxConns int
	TLS      *tls.Config
}

// 定义一个Option类型的函数，它操作了Server这个对象
type Option func(*Server)

// 下面是对四个可选参数的配置函数
func Protocol(p string) Option {
	return func(s *Server) {
		s.Protocol = p
	}
}

func Timeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.Timeout = timeout
	}
}

func MaxConns(maxconns int) Option {
	return func(s *Server) {
		s.MaxConns = maxconns
	}
}

func TLS(tls *tls.Config) Option {
	return func(s *Server) {
		s.TLS = tls
	}
}

// 用到了不定参数的特性，将任意个option应用到Server上
func NewServer(addr string, port int, options ...Option) (*Server, error) {
	// 先填写默认值
	srv := Server{
		Addr:     addr,
		Port:     port,
		Protocol: "tcp",
		Timeout:  30 * time.Second,
		MaxConns: 1000,
		TLS:      nil,
	}
	// 应用任意个option
	for _, option := range options {
		option(&srv)
	}
	return &srv, nil
}

func main() {
	s1, _ := NewServer("localhost", 1024)
	s2, _ := NewServer("localhost", 2048, Protocol("udp"))
	s3, _ := NewServer("0.0.0.0", 8080, Timeout(300*time.Second), MaxConns(1000))

	fmt.Println(s1, s2, s3)
}
```

耗子哥给出了6个点，但我感受最深的是以下两点：

1. 可读性强，将配置都转化成了对应的函数项option
2. 扩展性好，新增参数只需要增加一个对应的方法

那么对应的代价呢？就是需要编写多个Option函数，代码量会有所增加。



如果大家对这个感兴趣，可以去看一下Rob Pike的[这篇blog](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html) 。



## Further

顺着耗子叔的例子，我们再思考一下，如果配置的过程中有参数限制，那么我们该怎么办呢？

首先，我们改造一下函数Option

```go
// 返回错误
type OptionWithError func(*Server) error
```

然后，我们改造一下其中两个函数作为示例

```go
func Protocol(p string) OptionWithError {
	return func(s *Server) error {
		if p == "" {
			return errors.New("empty protocol")
		}
		s.Protocol = p
		return nil
	}
}

func Timeout(timeout time.Duration) Option {
	return func(s *Server) error {
		if timeout.Seconds() < 1 {
			return errors.New("time out should not less than 1s")
		}
		s.Timeout = timeout
		return nil
	}
}
```

我们再做一次改造

```go
func NewServer(addr string, port int, options ...OptionWithError) (*Server, error) {
	srv := Server{
		Addr:     addr,
		Port:     port,
		Protocol: "tcp",
		Timeout:  30 * time.Second,
		MaxConns: 1000,
		TLS:      nil,
	}
	// 增加了一个参数验证的步骤
	for _, option := range options {
		if err := option(&srv); err != nil {
			return nil, err
		}
	}
	return &srv, nil
}
```

改造基本到此完成，希望能给大家带来一定的帮助。



> Github: https://github.com/Junedayday/code_reading
>
> Blog: http://junes.tech/
>
> Bilibili：https://space.bilibili.com/293775192
>
> 公众号：golangcoding

