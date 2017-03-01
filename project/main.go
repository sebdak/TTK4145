package main

import (
	constants "./constants"
	elevator "./elevator"
	queue "./queue"
)

func main() {

	newOrderChannel := make(chan constants.NewOrder)
	nextFloorChannel := make(chan int)
	handledOrderChannel := make(chan constants.NewOrder)

	elevator.InitElev(newOrderChannel, nextFloorChannel, handledOrderChannel)
	queue.InitQueue(newOrderChannel, nextFloorChannel, handledOrderChannel)

	go elevator.Run()

	runForever := make(chan bool)
	<-runForever
}
