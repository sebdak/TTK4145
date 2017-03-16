package queue

import (
	constants "../constants"
	elevator "../elevator"
	network "../network"
	"time"
	"fmt"
)


func masterSendExternalQueue() {
	for {
		
		if network.Master == true { 

			externalQueuesMutex.Lock()
			compareAndFixExternalQueues()
			queuesTx <- externalQueues[0]
			externalQueuesMutex.Unlock()

		}
		time.Sleep(time.Millisecond * 50)

	}
}

func getExternalQueuesAndUpdate() {

	for {
		
		extQueue := <-queuesRx
		externalQueuesMutex.Lock()

		if (network.Master == false){

			for i := 0; i < constants.QueueCopies; i++ {
				externalQueues[i] = extQueue
			}

		}

		addExternalOrdersForThisElevator()

		//Stop spamming of order if in new queue
		updateOrdersThatNeedToBeAdded()

		//Stop spamming of handled order if not in new queue
		updateOrdersThatAreHandled()

		//Update lights in elevator
		hallLightCh <- externalQueues[0]

		fmt.Println("Ext. queue: ", externalQueues[0])

		externalQueuesMutex.Unlock()

	}

}

func updateOrdersThatNeedToBeAdded() {
	
	ordersThatNeedToBeAddedMutex.Lock()

	var indexesToDelete []int
	for i := 0; i < len(ordersThatNeedToBeAdded); i++ {

		for j := 0; j < len(externalQueues[0]); j++ {

			if ordersThatNeedToBeAdded[i].Floor == externalQueues[0][j].Floor && ordersThatNeedToBeAdded[i].Direction == externalQueues[0][j].Direction {
				indexesToDelete = append(indexesToDelete, i)
			}

		}

	}

	for i := 0; i < len(indexesToDelete); i++ {

		index := indexesToDelete[i]
		ordersThatNeedToBeAdded = append(ordersThatNeedToBeAdded[:index], ordersThatNeedToBeAdded[(index+1):]...)
		break

	}

	ordersThatNeedToBeAddedMutex.Unlock()
}

func updateOrdersThatAreHandled() {
	
	var indexesToDelete []int

	ordersThatAreHandledMutex.Lock()
	for i := 0; i < len(ordersThatAreHandled); i++ {

		safeToDelete := true
		for j := 0; j < len(externalQueues[0]); j++ {

			if ordersThatAreHandled[i].Floor == externalQueues[0][j].Floor && ordersThatAreHandled[i].Direction == externalQueues[0][j].Direction {
				safeToDelete = false
				break
			}

		}

		if safeToDelete == true {

			indexesToDelete = append(indexesToDelete, i)

		}

	}

	for i := 0; i < len(indexesToDelete); i++ {

		index := indexesToDelete[i]
		ordersThatAreHandled = append(ordersThatAreHandled[:index], ordersThatAreHandled[(index+1):]...)
		break

	}
	ordersThatAreHandledMutex.Unlock()
	
}

func spamExternalOrdersThatAreHandled() {

	for {
		
		if len(ordersThatAreHandled) > 0 {

			ordersThatAreHandledMutex.Lock()
			for i := 0; i < len(ordersThatAreHandled); i++ {

				handledExternalOrderTx <- ordersThatAreHandled[i]

			}
			ordersThatAreHandledMutex.Unlock()

		}

		time.Sleep(time.Millisecond * 50)
	}

}

func spamExternalOrdersThatNeedToBeAdded() {
	for {
		
		if len(ordersThatNeedToBeAdded) > 0 {

			ordersThatNeedToBeAddedMutex.Lock()
			for i := 0; i < len(ordersThatNeedToBeAdded); i++ {

				externalOrderTx <- ordersThatNeedToBeAdded[i]

			}
			ordersThatNeedToBeAddedMutex.Unlock()

		}

		time.Sleep(time.Millisecond * 50)
	}
}

func sendElevatorHeading() {

	for {
		
		heading := constants.ElevatorHeading{
			LastFloor:    elevator.LastFloor,
			CurrentOrder: elevator.CurrentOrder,
			Direction:    elevator.Direction,
			Id:           network.Id,
		}

		if headings[heading.Id] != heading {

			elevatorHeadingTx <- heading

		}

		time.Sleep(time.Millisecond * 50)
	}

}

func getElevatorHeadings() {

	for {
		
		heading := <-elevatorHeadingRx
		headings[heading.Id] = heading

	}

}

func masterGetExternalOrdersThatAreHandled() {

	for {
		
		if network.Master == true {

			order := <-handledExternalOrderRx

			//Order may have been removed by master before slaves know it
			if checkIfNewExternalOrder(order) == false && !isOrderInNeedToBeAddedList(order) {

				externalQueuesMutex.Lock()
				deleteOrderFromExternalQueue(order)
				externalQueuesMutex.Unlock()

			}

		}

		time.Sleep(time.Millisecond*40)

	}

}

func masterGetExternalOrdersThatNeedToBeAdded() {
	for {
		
		if network.Master == true {

			order := <-externalOrderRx

			//Check because master can have already handled order while slave still is spamming
			if checkIfNewExternalOrder(order) == true && !isOrderInHandledOrdersList(order) {

				order.ElevatorID = masterChooseElevatorThatTakesOrder(order)
				externalQueuesMutex.Lock()
				for i := 0; i < constants.QueueCopies; i++ {

					externalQueues[i] = append(externalQueues[i], order)

				}
				externalQueuesMutex.Unlock()

			}

		}

		time.Sleep(time.Millisecond*40)
	}

}

func handlePeerDisconnects() {
	for {
		
		elevatorId := <-peerDisconnectsCh

		if elevatorId == network.Id {

			disconnectedTakeAllExternalOrders()

		} else {

			masterRedistOrders(elevatorId)

		}
	}
}