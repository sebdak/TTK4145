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
	queuesTxChannel := make(chan []constants.Order)
	queuesRxChannel := make(chan []constants.Order)
	externalOrderTx := make(chan constants.Order)
	externalOrderRx := make(chan constants.Order)
	handledExternalOrderTx := make(chan constants.Order)
	handledExternalOrderRx := make(chan constants.Order)

	elevator.InitElev(newOrderChannel,
		newExternalOrderChannel, 
		nextFloorChannel, 
		handledOrderChannel)
	network.InitNetwork(newOrderChannel, 
		newExternalOrderChannel, 
		elevatorHeadingTxChannel, 
		elevatorHeadingRxChannel, 
		queuesTxChannel, 
		queuesRxChannel,
		externalOrderTx,
		externalOrderRx,
		handledExternalOrderTx,
		handledExternalOrderRx)
	queue.InitQueue(newOrderChannel, 
		nextFloorChannel, 
		handledOrderChannel, 
		elevatorHeadingTxChannel, 
		elevatorHeadingRxChannel, 
		queuesTxChannel, 
		queuesRxChannel,
		externalOrderTx,
		externalOrderRx,
		handledExternalOrderTx,
		handledExternalOrderRx)

	go elevator.Run()

	runForever := make(chan bool)
	<-runForever
}
