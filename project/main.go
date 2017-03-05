package main

import (
	constants "./constants"
	elevator "./elevator"
	queue "./queue"
	network "./network"
)

func main() {

	newOrderChannel := make(chan constants.Order)
	newExternalOrderChannel := make(chan constants.Order)
	nextFloorChannel := make(chan constants.Order)
	handledOrderChannel := make(chan constants.Order)

	elevatorHeadingTxChannel := make(chan constants.ElevatorHeading)
	elevatorHeadingRxChannel := make(chan constants.ElevatorHeading)

	elevator.InitElev(newOrderChannel,newExternalOrderChannel, nextFloorChannel, handledOrderChannel)
	network.InitNetwork(newOrderChannel, newExternalOrderChannel, elevatorHeadingTxChannel, elevatorHeadingRxChannel)
	queue.InitQueue(newOrderChannel, nextFloorChannel, handledOrderChannel, elevatorHeadingTxChannel, elevatorHeadingRxChannel )

	go elevator.Run()

	runForever := make(chan bool)
	<-runForever
}
