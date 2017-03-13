package main

import (
	constants "./constants"
	elevator "./elevator"
	network "./network"
	queue "./queue"
	"fmt"
)

func main() {
	newOrderChannel := make(chan constants.Order,1)
	newExternalOrderChannel := make(chan constants.Order,1)
	nextFloorChannel := make(chan constants.Order,1)
	handledOrderChannel := make(chan constants.Order,1)
	peerDisconnectsChannel := make(chan string,1)
	hallLightChannel := make(chan []constants.Order,1)

	elevatorHeadingTxChannel := make(chan constants.ElevatorHeading,1)
	elevatorHeadingRxChannel := make(chan constants.ElevatorHeading,1)
	queuesTxChannel := make(chan []constants.Order,1)
	queuesRxChannel := make(chan []constants.Order,1)
	externalOrderTx := make(chan constants.Order,1)
	externalOrderRx := make(chan constants.Order,1)
	handledExternalOrderTx := make(chan constants.Order,1)
	handledExternalOrderRx := make(chan constants.Order,1)

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
