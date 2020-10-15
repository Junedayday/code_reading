package main

const (
	mutexLocked = 1 << iota // mutex is locked
	mutexWoken
	mutexStarving
	mutexWaiterShift = iota
)

func main() {
	/*
		map part
	*/
	//mapper()

	//mapAddr()
	//mapModify()
	//mapReplace()
	//mapSet()

	/*
		slice part
	*/
	//slice()
	//sliceAddr()

	/*
		interface
	*/
	//dataInterface()

	/*
		embedded
	*/
	//embedded()
	//overwrite()
	//pointerAndStruct()
}
