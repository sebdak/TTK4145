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
var peerDisconnectsCh chan string
var hallLightCh chan []constants.Order

var elevatorHeadingTx chan constants.ElevatorHeading
var elevatorHeadingRx chan constants.ElevatorHeading
var queuesTx chan []constants.Order
var queuesRx chan []constants.Order
var externalOrderTx chan constants.Order
var externalOrderRx chan constants.Order
var handledExternalOrderTx chan constants.Order
var handledExternalOrderRx chan constants.Order

var headings map[string]constants.ElevatorHeading = make(map[string]constants.ElevatorHeading)


func InitQueue(newOrderChannel chan constants.Order, nextFloorChannel chan constants.Order, handledOrderChannel chan constants.Order, peerDisconnectsChannel chan string, hallLightChannel chan []constants.Order, elevatorHeadingTxChannel chan constants.ElevatorHeading, elevatorHeadingRxChannel chan constants.ElevatorHeading, queuesTxChannel chan []constants.Order, queuesRxChannel chan []constants.Order, externalOrderTxChannel chan constants.Order, externalOrderRxChannel chan constants.Order, handledExternalOrderTxChannel chan constants.Order, handledExternalOrderRxChannel chan constants.Order) {
	//Add channels for modulecommunication
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel
	peerDisconnectsCh = peerDisconnectsChannel
	hallLightCh = hallLightChannel

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

	go handleNewCabOrder()
	go handleExternalButtonOrder()
	go handleCompletedCabOrder()

	go sendElevatorHeading()
	go getElevatorHeadings()

	go sendExternalQueue()
	go getAndUpdateExternalQueues()
}


func initQueues() {
	internalQueue = make([]constants.Order)

	externalQueues = make([][]constants.Order, constants.QueueCopies)
	for i := 0; i < constants.QueueCopies; i++ {
		externalQueues[i] = make([]constants.Order, 1)
		externalQueues[i] = []constants.Order{}
	}
}

// -----------Sending and update of externalqueue----------------------------------
func sendExternalQueue() {
	for {
		if len(externalQueues[0]) > 0 && network.Master == true {
			<- externalQueuesMutex
			compareAndFixExternalQueues()
			queuesTx <- externalQueues[0]
			externalQueuesMutex <- true
		}
		time.Sleep(time.Millisecond * 100)			
	}
}

func getAndUpdateExternalQueues() {
	for {
		q = <- queuesRx
		
		//Send external queue for elevator to set hall lights
		qCopy := q
		hallLightCh <- qCopy
		
		<- externalQueuesMutex
		for i := 0; i < constants.QueueCopies; i++ {
			externalQueues[i] = q
		}
		addExternalOrdersForThisElevator()
		//stop spamming of order if in new queue
		updateOrdersThatNeedToBeAdded()
		//stop spamminf of handled order if not in new queue
		updateOrdersThatAreHandled()
		externalQueuesMutex <- true
	}
}

func compareAndFixExternalQueues() {
	count := 0
	var correctQueueIndex int
	var majority int = constants.QueueCopies/2 + 1
	fmt.Println("majority: ", majority)

	for i := 0; i < constants.QueueCopies; i++ {
		for j := 0; j < constants.QueueCopies; j++ {
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
	if(count < constants.QueueCopies){
		for i := 0; i < constants.QueueCopies; i++ {
			if i != correctQueueIndex {
				externalQueues[i] = externalQueues[correctQueueIndex]				
			}
		}
	}
}



// -----------Sending of elevator headings----------------------------------


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


func getElevatorHeadings() {
	var heading constants.ElevatorHeading
	for {
		heading = <- elevatorHeadingRx
		fmt.Println("Heading ", heading.LastFloor, heading.Id )
		headings[heading.Id] = heading
	}
}

// -----------Order handling----------------------------------

func addExternalOrdersForThisElevator() {
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

func handleNewCabOrder() {
	for {

		order := <-newOrderCh
		if checkIfNewCabOrder(order) {
			<-internalQueueMutex
			internalQueue = append(internalQueue, order)
			internalQueueMutex <- true
			updateElevatorNextOrder()
		}

		time.Sleep(time.Millisecond)
	}
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


func handleCompletedCabOrder() {
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

func deleteOrderFromExternalQueue(order constants.Order) {
	<-externalQueuesMutex

	for i := 0; i < len(externalQueues[0]); i++ {
		if externalQueues[0][i].Floor == order.Floor && (externalQueues[0][i].Direction == order.Direction || externalQueues[0][i].Direction == constants.DirStop) {
			for j:= 0; j < constants.QueueCopies; j++{
				externalQueues[j] = append(externalQueues[j][:i], externalQueues[j][(i+1):]...)
			}
			break
		}
	}

	externalQueuesMutex <- true
}

func checkIfNewExternalOrder(order constants.Order) bool {
	<-externalQueuesMutex

	for j := 0; j < len(externalQueue[0]); j++ {
		if (externalQueues[0][j].Floor == order && externalQueues[0][j].Direction == order.Direction) {
			externalQueuesMutex <- true
			return false
		}
	}

	externalQueuesMutex <- true
	return true
}

func checkIfNewCabOrder(order constants.Order) bool {
	<-internalQueueMutex
	for i := 0; i < len(internalQueue); i++ {
		if internalQueue[i] == order {
			internalQueueMutex <- true
			return false
		}
	}

	internalQueueMutex <- true
	return true
}

func takeAllExternalOrders() {
	<- externalQueuesMutex		
	for i := 0; i < len(externalQueues[0]); i++ {
		if externalQueues[0][i].ElevatorID != network.Id && checkIfNewCabOrder(externalQueues[0][i]) {
			<- internalQueueMutex
			internalQueue = append(internalQueue, externalQueues[0][i])
			internalQueueMutex <- true
		}
	}
	// Clean external queue
	for i := 0; i < constants.QueueCopies; i++ {
		externalQueues[i] = make([]constants.Order)
	}
	externalQueuesMutex <- true	
}

func redistOrders(peerId string) {
	<- externalQueuesMutex
	<- ordersThatNeedToBeAddedMutex
	//go through external queue and find orders for peerId
	for i := 0; i < len(externalQueues[0]); i++ {
		order := externalQueues[0][i]
		if order.ElevatorID == peerId {
			
			order.ElevatorID = make(string)
			ordersThatNeedToBeAdded = append(ordersThatNeedToBeAdded, order)
			deleteOrderFromExternalQueue(order)
		}
	}

	ordersThatNeedToBeAddedMutex <- true
	externalQueuesMutex <- true
}

func handlePeerDisconnects() {
	for {
		peerId := <- peerDisconnectsCh

		if peerId == network.Id {
			// Push all other elevators orders to its own internal queue
			takeAllExternalOrders()
		} else {
			// If master == true, redist orders
			if network.Master == true {
				redistOrders(peerId)
			}
		}			
	}

}

// -----------Spamming of orders that need to  be added/removed----------------------------------

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

func spamExternalOrdersThatNeedToBeAdded() {
	for {
		<- ordersThatNeedToBeAddedMutex

		if len(ordersThatNeedToBeAdded) > 0 {
			for i := 0, i < len(ordersThatNeedToBeAdded); i++ {
				externalOrderTx <- ordersThatNeedToBeAdded[i]
				
			}
		}

		ordersThatNeedToBeAddedMutex <- true
		time.Sleep(time.Millisecond*10)
	}
}

func supdateOrdersThatNeedToBeAdded() {
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

// -----------Getting of orders that need to  be added/removed by MEISTER----------------------------------
func getExternalOrdersThatAreHandled(){
	for{
		if(network.Master == true){
			order := <- handledExternalOrderRx
			//Order may have been removed by master before slaves know it
			if(checkIfNewExternalOrder(order) == false){

			}
		}
	}
}



func getExternalOrdersThatNeedToBeAdded(){
	for{
		if(network.Master == true){
			order := <-externalOrderRx
			//Check because master can have already handled order while slave still is spamming
			if(checkIfNewExternalOrder(order) == true){

				order.ElevatorID = chooseElevatorThatTakesOrder(order)
				<- externalQueuesMutex
				externalQueues = append(externalQueues, order)
				externalQueuesMutex <- true

			}
		}
	}

}

// -----------Elevator next order stuff----------------------------------


func updateElevatorNextOrder() {
	<-internalQueueMutex

	if len(internalQueue) > 0 {
		var bestFloorSoFar constants.Order
		var bestDistSoFar int = 100
		var dist int

		for i := 0; i < len(internalQueue); i++ {

			dist = findDistToFloor(internalQueue[i],elevator.Direction, elevator.LastFloor)

			if dist < bestDistSoFar {

				bestDistSoFar = dist
				bestFloorSoFar = internalQueue[i]

			}

		}

		nextFloorCh <- bestFloorSoFar
	}

	internalQueueMutex <- true
}

func chooseElevatorThatTakesOrder(order constants.Order) string{
	var bestElevatorSoFar string
	var bestDistSoFar int = 100
	var dist int

	for i:=0; i < len(network.PeersInfo.Peers); i++{
		currentElevator := headings[network.PeersInfo.Peers[i]]
		currentElevatorDir := getElevatorDirection(headings[currentElevatorID].CurrentOrder, headings[currentElevatorID].LastFloor)
		dist = findDistToFloor(currentElevator.CurrentOrder, currentElevatorDir, currentElevator.LastFloor)
		
		if dist < bestDistSoFar {
			bestDistSoFar = dist
			bestElevatorSoFar = currentElevator.Id
		}
	}

	return bestElevatorSoFar
}

func getElevatorDirection(order constants.Order, elevatorLastFloor int) constants.ElevatorDirection {
	var direction constants.ElevatorDirection = constants.DirStop
	if order.Floor > elevatorLastFloor.LastFloor {
		direction = constants.DirUp
	} else {
		direction = constants.DirDown
	}
	return direction
}

func findDistToFloor(destinationOrder constants.Order, elevatorDir constants.ElevatorDirection, currentFloor int) int{
	dist := 100

	if elevatorDir == constants.DirUp {
		if destinationOrder.Floor > currentFloor && (destinationOrder.Direction == constants.DirUp || destinationOrder.Direction == constants.DirStop) {
			dist = destinationOrder.Floor - currentFloor
		} else if destinationOrder.Floor > currentFloor && (destinationOrder.Direction == constants.DirDown) {
			dist = ((constants.NumberOfFloors - 1) - currentFloor) + destinationOrder.Floor
		} else if destinationOrder.Floor <= currentFloor && (destinationOrder.Direction == constants.DirDown || destinationOrder.Direction == constants.DirStop) {
			dist = (constants.NumberOfFloors - 1) - currentFloor + (constants.NumberOfFloors - 1) - destinationOrder.Floor
		} else if destinationOrder.Floor <= currentFloor && (destinationOrder.Direction == constants.DirUp) {
			dist = (constants.NumberOfFloors - 1) - currentFloor + (constants.NumberOfFloors - 1) + destinationOrder.Floor
		}
	} else if elevatorDir == constants.DirDown {
		if destinationOrder.Floor < currentFloor && (destinationOrder.Direction == constants.DirDown || destinationOrder.Direction == constants.DirStop) {
			dist = currentFloor - destinationOrder.Floor
		} else if destinationOrder.Floor < currentFloor && (destinationOrder.Direction == constants.DirUp) {
			dist = currentFloor + destinationOrder.Floor
		} else if destinationOrder.Floor >= currentFloor && (destinationOrder.Direction == constants.DirUp || destinationOrder.Direction == constants.DirStop) {
			dist = currentFloor + destinationOrder.Floor
		} else if destinationOrder.Floor >= currentFloor && (destinationOrder.Direction == constants.DirDown) {
			dist = currentFloor + (constants.NumberOfFloors - 1) + (constants.NumberOfFloors - destinationOrder.Floor)
		}
	}

	return dist
}