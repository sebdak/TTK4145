package main

import (
	constants "./constants"
	elevator "./elevator"
	queue "./queue"
	network "./network"

)

func main() {

	newInternalOrderChannel := make(chan constants.Order)
	newExternalOrderChannel := make(chan constants.Order)
	nextFloorChannel := make(chan constants.Order)
	handledOrderChannel := make(chan constants.Order)

	network.InitNetwork(newExternalOrderChannel)
	elevator.InitElev(newInternalOrderChannel,newExternalOrderChannel, nextFloorChannel, handledOrderChannel)
	queue.InitQueue(newInternalOrderChannel, nextFloorChannel, handledOrderChannel)


	go elevator.Run()

	runForever := make(chan bool)
	<-runForever
}
