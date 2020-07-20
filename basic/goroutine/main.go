package main

func main() {
	/*
		Tip: sync.WaitGroup - 最简单的并发逻辑处理
	*/
	wg()
	//errWg1()
	//errWg2()
	//errWg3()

	/*
		Tip: context.Context
	*/
	ctxCancel()
	ctxTimeout()
	ctxDeadline()
	ctxValue()
}
