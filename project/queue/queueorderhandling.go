package queue

import (
	constants "../constants"
	elevator "../elevator"
	network "../network"
)

// -----------New button orders----------------------------------
func handleNewCabOrder() {

	for {
		
		order := <-newOrderCh

		if checkIfNewInternalOrder(order) {

			internalQueueMutex.Lock()
			internalQueue = append(internalQueue, order)
			writeInternalQueueToFile()
			updateElevatorNextOrder()
			internalQueueMutex.Unlock()

		}

	}

}

func handleExternalButtonOrder() {
	//ordersThatNeedToBeAdded = make([]constants.Order)
	go spamExternalOrdersThatNeedToBeAdded()

	for {
		
		order := <-newExternalOrderCh

		if checkIfNewExternalOrder(order) && !isOrderInNeedToBeAddedList(order) && network.Online == true {

			ordersThatNeedToBeAddedMutex.Lock()
			ordersThatNeedToBeAdded = append(ordersThatNeedToBeAdded, order)
			ordersThatNeedToBeAddedMutex.Unlock()

		}

	}

}

func isOrderInNeedToBeAddedList(order constants.Order) bool {
	
	ordersThatNeedToBeAddedMutex.Lock()
	defer ordersThatNeedToBeAddedMutex.Unlock()

	for i := 0; i < len(ordersThatNeedToBeAdded); i++ {

		if ordersThatNeedToBeAdded[i] == order {
			
			return true

		}
	}
	return false
}


// -----------Completed order----------------------------------

func handleCompletedCabOrder() {

	//ordersThatAreHandled = make([]constants.Order)
	go spamExternalOrdersThatAreHandled()

	for {
		
		order := <-handledOrderCh
		//Check if order was external and that it is not already in handledOrderslist. If elev is off network there's no need telling others order has been handled
		if order.Direction != constants.DirStop && !isOrderInHandledOrdersList(order) && network.Online == true {

			ordersThatAreHandledMutex.Lock()
			ordersThatAreHandled = append(ordersThatAreHandled, order)
			ordersThatAreHandledMutex.Unlock()

		}

		internalQueueMutex.Lock()
		deleteOrderFromInternalQueue(order)
		writeInternalQueueToFile()
		updateElevatorNextOrder()
		internalQueueMutex.Unlock()

	}

}

func isOrderInHandledOrdersList(order constants.Order) bool {
	
	ordersThatAreHandledMutex.Lock()
	defer ordersThatAreHandledMutex.Unlock()

	for i := 0; i < len(ordersThatAreHandled); i++ {

		if ordersThatAreHandled[i] == order {
			
			return true

		}

	}

	return false

}


// -----------General functions on orders----------------------------------
func checkIfNewExternalOrder(order constants.Order) bool {
	
	externalQueuesMutex.Lock()
	defer externalQueuesMutex.Unlock()

	for j := 0; j < len(externalQueues[0]); j++ {

		if externalQueues[0][j].Floor == order.Floor && externalQueues[0][j].Direction == order.Direction {

			return false

		}

	}

	return true

}

func checkIfNewInternalOrder(order constants.Order) bool {
	
	internalQueueMutex.Lock()
	defer internalQueueMutex.Unlock()

	for i := 0; i < len(internalQueue); i++ {

		if internalQueue[i] == order {

			return false

		}

	}

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
	
	for i:= 0; i < constants.QueueCopies; i++{

		for j := 0; j < len(externalQueues[i]); j++ {

			if externalQueues[i][j].Floor == order.Floor && externalQueues[i][j].Direction == order.Direction {

				externalQueues[i] = append(externalQueues[i][:j], externalQueues[i][(j+1):]...)
				j--

			}

		}

	}

}

func addExternalOrdersForThisElevator() {

	for i := 0; i < len(externalQueues[0]); i++ {

		//Assuming order is new
		newOrder := true

		if externalQueues[0][i].ElevatorID == network.Id { 

			if !checkIfNewInternalOrder(externalQueues[0][i]) {

				newOrder = false

			}else if isOrderInHandledOrdersList(externalQueues[0][i]) { 

				newOrder = false

			}

			if newOrder == true {

				internalQueueMutex.Lock()
				internalQueue = append(internalQueue, externalQueues[0][i])
				writeInternalQueueToFile()
				updateElevatorNextOrder()
				internalQueueMutex.Unlock()

			}

		}

	}

}

// -----------Disconnected peer functions----------------------------------


func disconnectedTakeAllExternalOrders() {
	
	externalQueuesMutex.Lock()

	for i := 0; i < len(externalQueues[0]); i++ {

		if checkIfNewInternalOrder(externalQueues[0][i]) {

			internalQueueMutex.Lock()
			internalQueue = append(internalQueue, externalQueues[0][i])
			internalQueueMutex.Unlock()

		}

	}

	// Clean external queue
	for i := 0; i < constants.QueueCopies; i++ {

		externalQueues[i] = externalQueues[i][:0]

	}

	externalQueuesMutex.Unlock()

}

func masterRedistOrders(elevatorId string) {
	
	if network.Master == true {

		externalQueuesMutex.Lock()
		ordersThatNeedToBeAddedMutex.Lock()

		for i := 0; i < len(externalQueues[0]); i++ {

			order := externalQueues[0][i]

			if order.ElevatorID == elevatorId {

				deleteOrderFromExternalQueue(order)
				order.ElevatorID = ""
				ordersThatNeedToBeAdded = append(ordersThatNeedToBeAdded, order)
				i--

			}

		}
	
		ordersThatNeedToBeAddedMutex.Unlock()
		externalQueuesMutex.Unlock()

	}

}


// -----------Elevator next order stuff----------------------------------

func updateElevatorNextOrder() {
	
	if len(internalQueue) > 0 {

		var bestFloorSoFar constants.Order
		var bestDistSoFar int = 100

		for i := 0; i < len(internalQueue); i++ {

			dist := findDistToFloor(internalQueue[i], elevator.Direction, elevator.LastFloor,internalQueue[i])

			if dist < bestDistSoFar {

				bestDistSoFar = dist
				bestFloorSoFar = internalQueue[i]

			}

		}

		nextFloorCh <- bestFloorSoFar

	} else {

		//Send order for floor "-1", telling elevator there are no new orders
		nextFloorCh <- constants.Order{Floor: -1, Direction: constants.DirStop, ElevatorID: "-1"} 
	
	}

}

func masterChooseElevatorThatTakesOrder(order constants.Order) string {

	var bestElevatorSoFar string
	var bestDistSoFar int = 100

	for i := 0; i < len(network.PeersInfo.Peers); i++ {

		currentElevator := headings[network.PeersInfo.Peers[i]]
		dist := findDistToFloor(order, currentElevator.Direction, currentElevator.LastFloor, currentElevator.CurrentOrder)

		if dist < bestDistSoFar {

			bestDistSoFar = dist
			bestElevatorSoFar = currentElevator.Id

		}

	}

	return bestElevatorSoFar

}

func findDistToFloor(destinationOrder constants.Order, elevatorDir constants.ElevatorDirection, currentFloor int, currentOrder constants.Order) int {
	dist := 100

	if elevatorDir == constants.DirUp {

		if destinationOrder.Floor > currentFloor && (destinationOrder.Direction == constants.DirUp || destinationOrder.Direction == constants.DirStop) {

			dist = destinationOrder.Floor - currentFloor

		} else if destinationOrder.Floor > currentFloor && (destinationOrder.Direction == constants.DirDown) {

			dist = ((constants.NumberOfFloors - 1) - currentFloor) + (constants.NumberOfFloors - 1) -  destinationOrder.Floor

		} else if destinationOrder.Floor <= currentFloor && (destinationOrder.Direction == constants.DirDown || destinationOrder.Direction == constants.DirStop) {

			dist = (constants.NumberOfFloors - 1) - currentFloor + (constants.NumberOfFloors - 1) - destinationOrder.Floor

		} else if destinationOrder.Floor <= currentFloor && (destinationOrder.Direction == constants.DirUp) {

			dist = (constants.NumberOfFloors - 1) - currentFloor + (constants.NumberOfFloors - 1) + destinationOrder.Floor

		}

		//Punish that elevator needs to stop before handling this order
		if(destinationOrder.Floor > currentOrder.Floor){
			dist += 1
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

		//Punish that elevator needs to stop before handling this order
		if(destinationOrder.Floor < currentOrder.Floor){
			dist += 1
		}

	} else if elevatorDir == constants.DirStop{

		if destinationOrder.Floor >= currentFloor{

			dist = destinationOrder.Floor - currentFloor

		} else if destinationOrder.Floor < currentFloor{

			dist = currentFloor - destinationOrder.Floor

		}
	}

	return dist

}