package main

import (
	elevator "./elevator"
	queue "./queue"
)

func main() {

	nextFloorCh := make(chan int)
	elevator.InitElev(&nextFloorCh)
	queue.InitQueue(&nextFloorCh)

	go elevator.Run()

	runForever := make(chan bool)
	<-runForever
}
