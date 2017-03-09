package queue

import (
	network "../network"
	constants "../constants"
	elevator "../elevator"
	"fmt"
	"time"
)

var internalQueue []constants.Order
var externalQueues [][]constants.Order
var ordersThatNeedToBeAdded []constants.Order
var ordersThatAreHandled []constants.Order

var internalQueueMutex = make(chan bool, 1)
var externalQueuesMutex = make(chan bool, 1)
var ordersThatNeedToBeAddedMutex = make(chan bool, 1)
var ordersThatAreHandledMutex = make(chan bool, 1)

var newOrderCh chan constants.Order
var nextFloorCh chan constants.Order
var handledOrderCh chan constants.Order

var elevatorHeadingTx chan constants.ElevatorHeading
var elevatorHeadingRx chan constants.ElevatorHeading
var queuesTx chan []constants.Order
var queuesRx chan []constants.Order
var externalOrderTx chan constants.Order
var externalOrderRx chan constants.Order
var handledExternalOrderTx chan constants.Order
var handledExternalOrderRx chan constants.Order

var headings map[string]constants.ElevatorHeading = make(map[string]constants.ElevatorHeading)


func InitQueue(newOrderChannel chan constants.Order, nextFloorChannel chan constants.Order, handledOrderChannel chan constants.Order, elevatorHeadingTxChannel chan constants.ElevatorHeading, elevatorHeadingRxChannel chan constants.ElevatorHeading, queuesTxChannel chan []constants.Order, queuesRxChannel chan []constants.Order, externalOrderTxChannel chan constants.Order, externalOrderRxChannel chan constants.Order, handledExternalOrderTxChannel chan constants.Order, handledExternalOrderRxChannel chan constants.Order) {
	//Add channels for modulecommunication
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel

	//Channels for module-network communication
	elevatorHeadingTx = elevatorHeadingTxChannel
	elevatorHeadingRx = elevatorHeadingRxChannel
	queuesTx = queuesTxChannel
	queuesRx = queuesRxChannel
	externalOrderTx = externalOrderTxChannel
	externalOrderRx = externalOrderRxChannel
	handledExternalOrderTx = handledExternalOrderTxChannel
	handledExternalOrderRx = handledExternalOrderRxChannel

	//mutex := true
	internalQueueMutex <- true
	externalQueuesMutex <- true
	ordersThatNeedToBeAddedMutex <- true
	ordersThatAreHandledMutex <- true

	//init external queue
	initQueues()

	go lookForCabOrder()
	go lookForHandledCabOrder()
	go sendElevatorHeading()
	go updateElevatorHeadings()
	go sendExternalQueue()
	go handleExternalButtonOrder()
}


func initQueues() {
	internalQueue = make([]constants.Order)

	externalQueues = make([][]constants.Order, constants.NumberOfElevators)
	for i := 0; i < constants.NumberOfElevators; i++ {
		externalQueues[i] = make([]constants.Order, 1)
		externalQueues[i] = []constants.Order{}
	}
}


func sendExternalQueue() {
	for {
		if len(externalQueues[0]) > 0 {
			<- externalQueuesMutex
			compareAndFixExternalQueues()
			queuesTx <- externalQueues[0]
			externalQueuesMutex <- true
		}
		time.Sleep(time.Millisecond * 100)			
	}
}


func compareAndFixExternalQueues() {
	count := 0
	var correctQueueIndex int
	var majority int = constants.NumberOfElevators/2 + 1
	fmt.Println("majority: ", majority)

	for i := 0; i < constants.NumberOfElevators; i++ {
		for j := 0; j < constants.NumberOfElevators; j++ {
			if reflect.DeepEqual(externalQueues[i], externalQueues[j]) {
				fmt.Println("Reflect: ", reflect.DeepEqual(externalQueues[i], externalQueues[j]))
				count++				
			}
		}

		// if all queuea are alike
		if count >= majority {
			correctQueueIndex = i
		}
		
		count = 0
	}

	//if queues dont match
	if(count < constants.NumberOfElevators){
		for i := 0; i < constants.NumberOfElevators; i++ {
			if i != correctQueueIndex {
				externalQueues[i] = externalQueues[correctQueueIndex]				
			}
		}
	}
}


func updateExternalQueues() {
	for {
		q = <- queuesRx
		<- externalQueuesMutex
		for i := 0; i < constants.NumberOfElevators; i++ {
			externalQueues[i] = q
		}
		addNewOrdersForThisElevator()
		//stop spamming of order if in new queue
		updateOrdersThatNeedToBeAdded()
		//stop spamminf of handled order if not in new queue
		updateOrdersThatAreHandled()
		externalQueuesMutex <- true
	}
}

func updateOrdersThatNeedToBeAdded() {
	indexesToDelete []int = make([]int)
	<- ordersThatNeedToBeAddedMutex
	
	for i := 0; i < len(ordersThatNeedToBeAdded); i++ {
		for j := 0; j < len(externalQueues[0]); j++ {
			if ordersThatNeedToBeAdded[i] == externalQueues[0][j] {
				indexesToDelete = append(indexesToDelete, i)
			}
			
		}
	}

	for i := 0; i < len(indexesToDelete); i++ {
		index := indexesToDelete[i]
		ordersThatNeedToBeAdded = append(ordersThatNeedToBeAdded[:index], ordersThatNeedToBeAdded[(index+1):]...)
	}
	
	ordersThatNeedToBeAddedMutex <- true
}


func updateOrdersThatAreHandled() {
	indexesToDelete []int = make([]int)

	<- ordersThatAreHandledMutex
	for i := 0; i < len(ordersThatAreHandled); i++ {
		safeToDelete := true
		for j := 0; j < len(externalQueues[0]); j++ {
			if ordersThatAreHandled[i] == externalQueues[0][j] {
				safeToDelete = false
				break
			}
			
		}

		if(safeToDelete==true){
			indexesToDelete = append(indexesToDelete, i)
		} 

		safeToDelete = true
	}

	for i := 0; i < len(indexesToDelete); i++ {
		index := indexesToDelete[i]
		ordersThatAreHandled = append(ordersThatAreHandled[:index], ordersThatAreHandled[(index+1):]...)
	}	

	ordersThatAreHandledMutex <- true
}


func addNewOrdersForThisElevator() {
	newOrder := true
	<- internalQueueMutex

	for i := 0; i < len(externalQueues[0]); i++ {
		//Check if new external order is not in internalqueue
		if externalQueues[0][i].ElevatorID == network.Id {
			for j := 0; j < len(internalQueue); j++ {
				if (externalQueues[0][i] == internalQueue[j]) {
					newOrder = false
					break
				}
			}
		}

		//Check if new external order has not just been handled
		if(newOrder == true){
			for j:= 0; j< len(ordersThatAreHandled);j++{
				if(externalQueues[0][i] == ordersThatAreHandled[j]){
					newOrder = false
					break
				}
			}
		}


		if(newOrder == true){
			internalQueue = append(internalQueue, externalQueues[0][i])

		} 

		newOrder = true
	}
	internalQueueMutex <- true

}


func sendElevatorHeading() {
	for {
		heading := constants.ElevatorHeading{
			LastFloor: elevator.LastFloor,
			CurrentOrder: elevator.CurrentOrder,
			Id: network.Id,
		}

		if(headings[heading.Id] != heading){

			elevatorHeadingTx <- heading

		} 

		time.Sleep(time.Millisecond * 20)	
	}
}


func updateElevatorHeadings() {
	var heading constants.ElevatorHeading
	for {
		heading = <- elevatorHeadingRx
		fmt.Println("Heading ", heading.LastFloor, heading.Id )
		headings[heading.Id] = heading
	}
}


func lookForHandledCabOrder() {
	go spamExternalOrdersThatAreHandled()

	for {

		order := <- handledOrderCh

		//Check if order was external
		if order.Direction != constants.DirStop {
			<- ordersThatAreHandledMutex
			ordersThatAreHandled = append(ordersThatAreHandled, order)
			ordersThatNeedToBeAddedMutex <- true
		}
		deleteOrderFromInternalQueue(order)
		updateElevatorNextOrder()

	}
}


func spamExternalOrdersThatAreHandled() {
	for {
		if len(ordersThatAreHandled) > 0 {
			for i := 0, i < len(ordersThatAreHandled); i++ {
				handledExternalOrderTx <- ordersThatAreHandled[i]
				time.Sleep(time.Millisecond)
			}
		}
	}	
}


func deleteOrderFromInternalQueue(order constants.Order) {
	<-internalQueueMutex

	for i := 0; i < len(internalQueue); i++ {
		if internalQueue[i].Floor == order.Floor && (internalQueue[i].Direction == order.Direction || internalQueue[i].Direction == constants.DirStop) {
			internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
			i--
		}
	}

	internalQueueMutex <- true
}


func handleExternalButtonOrder() {
	ordersThatNeedToBeAdded = make([]constants.Order)
	go spamExternalOrdersThatNeedToBeAdded()

	for {

		order := <-newExternalOrderCh

		if checkIfNewExternalOrder(order) {
			<- ordersThatNeedToBeAddedMutex
			ordersThatNeedToBeAdded = append(ordersThatNeedToBeAdded, order)
			ordersThatNeedToBeAddedMutex <- true
			externalOrderTx <- order
		}

	}
}


func spamExternalOrdersThatNeedToBeAdded() {
	for {
		<- ordersThatNeedToBeAddedMutex

		if len(ordersThatNeedToBeAdded) > 0 {
			for i := 0, i < len(ordersThatNeedToBeAdded); i++ {
				orderTx <- ordersThatNeedToBeAdded[i]
				
			}
		}

		ordersThatNeedToBeAddedMutex <- true
		time.Sleep(time.Millisecond*10)
	}
}


func checkIfNewExternalOrder(order constants.Order) bool {
	<-externalQueueMutex

	for j := 0; j < len(externalQueue[0]); j++ {
		if (externalQueues[0][j].Floor == order && externalQueues[0][j].Direction == order.Direction) {
			externalQueuesMutex <- true
			return false
		}
	}

	externalQueueMutex <- true
	return true
}


func lookForCabOrder() {
	for {

		order := <-newOrderCh
		if checkIfNewCabOrder(order) {
			queueAddOrder(order)
			updateElevatorNextOrder()
		}

		time.Sleep(time.Millisecond)
	}
}


func checkIfNewCabOrder(order constants.Order) bool {
	<-internalQueueMutex
	for i := 0; i < len(internalQueue); i++ {
		if internalQueue[i] == order {
			internalQueueMutex <- true
			return false
		}
	}
	for j := 0; j < len(externalQueue); j++ {
		if externalQueue[j] == order {
			internalQueueMutex <- true
			return false
		}
	}

	internalQueueMutex <- true
	return true
}


func queueAddOrder(order constants.Order) {
	

	if order.Direction == constants.DirStop {
		<-internalQueueMutex
		internalQueue = append(internalQueue, order)
		internalQueueMutex <- true
	} else {
		//HANDLE EXTERNAL ORDER
	}
	Debug()
}


func updateElevatorNextOrder() {
	<-internalQueueMutex

	if len(internalQueue) > 0 {
		var bestFloorSoFar constants.Order
		var bestFloorSoFarDist int = 100
		var dist int

		for i := 0; i < len(internalQueue); i++ {

			if elevator.Direction == constants.DirUp {

				if internalQueue[i].Floor > elevator.LastFloor && (internalQueue[i].Direction == constants.DirUp || internalQueue[i].Direction == constants.DirStop) {

					dist = internalQueue[i].Floor - elevator.LastFloor

				} else if internalQueue[i].Floor > elevator.LastFloor && (internalQueue[i].Direction == constants.DirDown) {

					dist = ((constants.NumberOfFloors - 1) - elevator.LastFloor) + internalQueue[i].Floor

				} else if internalQueue[i].Floor <= elevator.LastFloor && (internalQueue[i].Direction == constants.DirDown || internalQueue[i].Direction == constants.DirStop) {

					dist = (constants.NumberOfFloors - 1) - elevator.LastFloor + (constants.NumberOfFloors - 1) - internalQueue[i].Floor

				} else if internalQueue[i].Floor <= elevator.LastFloor && (internalQueue[i].Direction == constants.DirUp) {

					dist = (constants.NumberOfFloors - 1) - elevator.LastFloor + (constants.NumberOfFloors - 1) + internalQueue[i].Floor

				}

			} else if elevator.Direction == constants.DirDown {
				if internalQueue[i].Floor < elevator.LastFloor && (internalQueue[i].Direction == constants.DirDown || internalQueue[i].Direction == constants.DirStop) {

					dist = elevator.LastFloor - internalQueue[i].Floor

				} else if internalQueue[i].Floor < elevator.LastFloor && (internalQueue[i].Direction == constants.DirUp) {

					dist = elevator.LastFloor + internalQueue[i].Floor

				} else if internalQueue[i].Floor >= elevator.LastFloor && (internalQueue[i].Direction == constants.DirUp || internalQueue[i].Direction == constants.DirStop) {

					dist = elevator.LastFloor + internalQueue[i].Floor

				} else if internalQueue[i].Floor >= elevator.LastFloor && (internalQueue[i].Direction == constants.DirDown) {

					dist = elevator.LastFloor + (constants.NumberOfFloors - 1) + (constants.NumberOfFloors - internalQueue[i].Floor)

				}
			}

			if dist <= bestFloorSoFarDist {

				bestFloorSoFarDist = dist
				bestFloorSoFar = internalQueue[i]

			}

		}

		nextFloorCh <- bestFloorSoFar
	}

	internalQueueMutex <- true
}


func getNeededElevatorDirection(order constants.Order) constants.ElevatorDirection {
	var direction constants.ElevatorDirection = constants.DirStop
	if order.Floor > elevator.LastFloor {
		direction = constants.DirUp
	} else {
		direction = constants.DirDown
	}
	return direction
}

//NEXT TIME
//CREATE INTERNAL ORDER ARRAY AND MAKE ELEVATOR HANDLE INTERNAL QUEUE ORDERS
