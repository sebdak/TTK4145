package main

import (
	constants "./constants"
	elevator "./elevator"
	queue "./queue"
)

func main() {

	newOrderChannel := make(chan constants.NewOrder)
	nextFloorChannel := make(chan int)

	elevator.InitElev(newOrderChannel, nextFloorChannel)
	queue.InitQueue(newOrderChannel, nextFloorChannel)

	go elevator.Run()

	runForever := make(chan bool)
	<-runForever
}
