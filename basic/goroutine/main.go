package main

func main() {
	/*
		Tip: sync.WaitGroup - 最简单的并发逻辑处理
	*/
	//wg()
	//errWg1()
	//errWg2()
	//errWg3()

	/*
		Tip: context.Context - 上下文管理
	*/
	//ctxCancel()
	//ctxTimeout()
	//ctxDeadline()
	//ctxValue()

	/*
		Tip: channel - 协程间的通信利器
	*/
	//ch()
	//chLimit()
	//chClose()
	//chCloseErr()
	//chTask()
	//chBuffer()
	//chBufferRange()

	//passBall()
	//passBallWithClose()

	/*
		Tip: sync.Map - 多协程下的map实现
	*/
	//syncMap()

	/*
		Tip: sync.Cond - 多协程下的首发通道
	*/
	//syncCondErr()
	//syncCondExplain()
	syncCond()
}
