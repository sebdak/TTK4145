package main

import (
	constants "./constants"
	elevator "./elevator"
	network "./network"
	queue "./queue"
	"fmt"
)

func main() {
	newOrderChannel := make(chan constants.Order)
	newExternalOrderChannel := make(chan constants.Order)
	nextFloorChannel := make(chan constants.Order)
	handledOrderChannel := make(chan constants.Order)
	peerDisconnectsChannel := make(chan string)
	hallLightChannel := make(chan []constants.Order)

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
		handledOrderChannel,
		hallLightChannel)

	fmt.Println("Init elev ok")

	network.InitNetwork(newOrderChannel,
		peerDisconnectsChannel,
		elevatorHeadingTxChannel,
		elevatorHeadingRxChannel,
		queuesTxChannel,
		queuesRxChannel,
		externalOrderTx,
		externalOrderRx,
		handledExternalOrderTx,
		handledExternalOrderRx)

	fmt.Println("Init network ok")

	queue.InitQueue(newOrderChannel,
		newExternalOrderChannel,
		nextFloorChannel,
		handledOrderChannel,
		peerDisconnectsChannel,
		hallLightChannel,
		elevatorHeadingTxChannel,
		elevatorHeadingRxChannel,
		queuesTxChannel,
		queuesRxChannel,
		externalOrderTx,
		externalOrderRx,
		handledExternalOrderTx,
		handledExternalOrderRx)

	fmt.Println("Init queue ok")

	hopefullyRunSucessfullyForever := make(chan bool)
	<-hopefullyRunSucessfullyForever
}
