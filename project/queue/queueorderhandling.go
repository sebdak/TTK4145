package queue

import (
	constants "../constants"
	elevator "../elevator"
	network "../network"
	"fmt"
	"reflect"
	"time"
)


// -----------New button orders----------------------------------
func handleNewCabOrder() {
	for {

		order := <-newOrderCh
		if checkIfNewInternalOrder(order) {
			<-internalQueueMutex
			internalQueue = append(internalQueue, order)
			updateElevatorNextOrder()
			internalQueueMutex <- true
		}

		time.Sleep(time.Millisecond)
	}
}

func handleExternalButtonOrder() {
	//ordersThatNeedToBeAdded = make([]constants.Order)
	go spamExternalOrdersThatNeedToBeAdded()

	for {

		order := <-newExternalOrderCh
		if checkIfNewExternalOrder(order) && !isOrderInNeedToBeAddedList(order) {
			<-ordersThatNeedToBeAddedMutex
			ordersThatNeedToBeAdded = append(ordersThatNeedToBeAdded, order)
			ordersThatNeedToBeAddedMutex <- true
		}

	}
}

func isOrderInNeedToBeAddedList(order constants.Order) bool {
	<-ordersThatNeedToBeAddedMutex
	for i := 0; i < len(ordersThatNeedToBeAdded); i++ {
		if ordersThatNeedToBeAdded[i] == order {
			ordersThatNeedToBeAddedMutex <- true
			return true
		}
	}

	ordersThatNeedToBeAddedMutex <- true
	return false
}


// -----------Completed order----------------------------------
func handleCompletedCabOrder() {
	//ordersThatAreHandled = make([]constants.Order)
	go spamExternalOrdersThatAreHandled()

	for {

		order := <-handledOrderCh
		//Check if order was external and that it is not already in handledOrderslist
		if order.Direction != constants.DirStop && !isOrderInHandledOrdersList(order) {
			<-ordersThatAreHandledMutex
			ordersThatAreHandled = append(ordersThatAreHandled, order)
			ordersThatAreHandledMutex <- true
		}

		<-internalQueueMutex
		deleteOrderFromInternalQueue(order)
		updateElevatorNextOrder()
		internalQueueMutex <- true
	}
}

func isOrderInHandledOrdersList(order constants.Order) bool {
	<-ordersThatAreHandledMutex
	for i := 0; i < len(ordersThatAreHandled); i++ {
		if ordersThatAreHandled[i] == order {
			ordersThatAreHandledMutex <- true
			return true
		}
	}

	ordersThatAreHandledMutex <- true
	return false
}


// -----------General functions on orders----------------------------------
func checkIfNewExternalOrder(order constants.Order) bool {
	<-externalQueuesMutex
	for j := 0; j < len(externalQueues[0]); j++ {
		if externalQueues[0][j].Floor == order.Floor && externalQueues[0][j].Direction == order.Direction {
			externalQueuesMutex <- true
			return false
		}
	}

	externalQueuesMutex <- true
	return true
}

func checkIfNewInternalOrder(order constants.Order) bool {
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

func deleteOrderFromInternalQueue(order constants.Order) {
	//Assume function that calls this has internalqueuemutex

	for i := 0; i < len(internalQueue); i++ {
		if(order.Floor == internalQueue[i].Floor){
			if order.Direction == constants.DirUp{
				if(internalQueue[i].Direction == constants.DirUp || internalQueue[i].Direction == constants.DirStop){
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				} 
			} else if order.Direction == constants.DirDown{
				if(internalQueue[i].Direction == constants.DirDown || internalQueue[i].Direction == constants.DirStop){
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				} 
			} else if order.Direction == constants.DirStop{
				if elevator.Direction == constants.DirUp && internalQueue[i].Direction == constants.DirUp{
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				} else if elevator.Direction == constants.DirDown && internalQueue[i].Direction == constants.DirDown{
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				} else if internalQueue[i].Direction == constants.DirStop{
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				}
			}
		}
	}
}


func deleteOrderFromExternalQueue(order constants.Order) {
	//Assume that externalqueuesmutex is taken in function that calls this
	for i := 0; i < len(externalQueues[0]); i++ {
		if externalQueues[0][i].Floor == order.Floor && externalQueues[0][i].Direction == order.Direction {
			externalQueues[0] = append(externalQueues[j][:i], externalQueues[j][(i+1):]...)
			break
		}
	}

	//Update all other copies
	for i := 1; i < constants.QueueCopies; i++ {
		externalQueues[i] = externalQueues[0]
	}

}

func addExternalOrdersForThisElevator() {
	for i := 0; i < len(externalQueues[0]); i++ {
		//Check if new external order is not in internalqueue
		newOrder := true
		if externalQueues[0][i].ElevatorID == network.Id {
			if !checkIfNewInternalOrder(externalQueues[0][i]) {
				newOrder = false
			}else  {
				//Check if new external order has not just been handled
				for j := 0; j < len(ordersThatAreHandled); j++ {
					if externalQueues[0][i] == ordersThatAreHandled[j] {
						newOrder = false
						break
					}
				}
				
			}

			if newOrder == true {
				<-internalQueueMutex
				internalQueue = append(internalQueue, externalQueues[0][i])
				updateElevatorNextOrder()
				internalQueueMutex <- true
			}
		}
	}
}

// -----------Disconnected peer functions----------------------------------

func handlePeerDisconnects() {
	for {
		elevatorId := <-peerDisconnectsCh

		if elevatorId == network.Id {
			// Push all other elevators orders to its own internal queue
			disconnectedTakeAllExternalOrders()
		} else {
			masterRedistOrders(elevatorId)
		}
	}
}

func disconnectedTakeAllExternalOrders() {
	<-externalQueuesMutex
	for i := 0; i < len(externalQueues[0]); i++ {
		if checkIfNewInternalOrder(externalQueues[0][i]) {
			<-internalQueueMutex
			internalQueue = append(internalQueue, externalQueues[0][i])
			internalQueueMutex <- true
		}
	}
	// Clean external queue
	for i := 0; i < constants.QueueCopies; i++ {
		externalQueues[i] = externalQueues[i][:0]
	}
	externalQueuesMutex <- true
}

func masterRedistOrders(elevatorId string) {
	if network.Master == true {
		<-externalQueuesMutex
		<-ordersThatNeedToBeAddedMutex
		//go through external queue and find orders for elevatorId
		for i := 0; i < len(externalQueues[0]); i++ {
			order := externalQueues[0][i]
			if order.ElevatorID == elevatorId {
	
				order.ElevatorID = ""
				deleteOrderFromExternalQueue(order)
				ordersThatNeedToBeAdded = append(ordersThatNeedToBeAdded, order)
			}
		}
	
		ordersThatNeedToBeAddedMutex <- true
		externalQueuesMutex <- true
	}
}


// -----------Elevator next order stuff----------------------------------

func updateElevatorNextOrder() {

	if len(internalQueue) > 0 {

		var bestFloorSoFar constants.Order
		var bestDistSoFar int = 100

		for i := 0; i < len(internalQueue); i++ {

			dist := findDistToFloor(internalQueue[i], elevator.Direction, elevator.LastFloor)
			if dist < bestDistSoFar {

				bestDistSoFar = dist
				bestFloorSoFar = internalQueue[i]

			}

		}
		fmt.Println("best order: ", bestFloorSoFar)
		fmt.Println("extqueue: ", externalQueues[0])
		nextFloorCh <- bestFloorSoFar
	}

}

func chooseElevatorThatTakesOrder(order constants.Order) string {
	var bestElevatorSoFar string
	var bestDistSoFar int = 100
	for i := 0; i < len(network.PeersInfo.Peers); i++ {
		currentElevator := headings[network.PeersInfo.Peers[i]]
		currentElevatorDir := getElevatorDirection(currentElevator.CurrentOrder, currentElevator.LastFloor)
		dist := findDistToFloor(currentElevator.CurrentOrder, currentElevatorDir, currentElevator.LastFloor)

		if dist < bestDistSoFar {
			bestDistSoFar = dist
			bestElevatorSoFar = currentElevator.Id
		}
	}

	return bestElevatorSoFar
}

func getElevatorDirection(order constants.Order, elevatorLastFloor int) constants.ElevatorDirection {
	var direction constants.ElevatorDirection = constants.DirStop
	if order.Floor > elevatorLastFloor {
		direction = constants.DirUp
	} else {
		direction = constants.DirDown
	}
	return direction
}

func findDistToFloor(destinationOrder constants.Order, elevatorDir constants.ElevatorDirection, currentFloor int) int {
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