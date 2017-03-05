package main

import (
	constants "./constants"
	elevator "./elevator"
	queue "./queue"
	network "./network"

)

func main() {

	newOrderChannel := make(chan constants.NewOrder)
	nextFloorChannel := make(chan constants.NewOrder)
	handledOrderChannel := make(chan constants.NewOrder)

	
	elevator.InitElev(newOrderChannel, nextFloorChannel, handledOrderChannel)
	queue.InitQueue(newOrderChannel, nextFloorChannel, handledOrderChannel)


	go elevator.Run()

	runForever := make(chan bool)
	<-runForever
}
