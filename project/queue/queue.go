package queue

import (
	constants "../constants"
	elevator "../elevator"
	//"container/list"
	"fmt"
	"time"
)

var internalQueue []constants.NewOrder
var externalQueue []constants.NewOrder

//var externalQueue [constants.NumberOfElevators]List

var internalQueueMutex = make(chan bool, 1)
var newOrderCh chan constants.NewOrder
var nextFloorCh chan constants.NewOrder
var handledOrderCh chan constants.NewOrder

func InitQueue(newOrderChannel chan constants.NewOrder, nextFloorChannel chan constants.NewOrder, handledOrderChannel chan constants.NewOrder) {
	//Add channels
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel

	//mutex := true
	internalQueueMutex <- true

	go lookForOrder()
	go lookForHandledOrder()
}

func lookForHandledOrder() {
	for {

		order := <-handledOrderCh

		deleteOrderFromQueues(order)
		//Send ut ny kø til master
		Debug()

		updateElevatorNextFloor()

	}
}

func deleteOrderFromQueues(order constants.NewOrder) {
	<-internalQueueMutex

	for i := 0; i < len(internalQueue); i++ {
		if internalQueue[i].Floor == order.Floor && (internalQueue[i].Direction == order.Direction || internalQueue[i].Direction == constants.DirStop) {
			internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
		}
	}

	if order.Direction != constants.DirStop {
		for i := 0; i < len(externalQueue); i++ {

			if externalQueue[i].Floor == order.Floor && (externalQueue[i].Direction == order.Direction) {
				externalQueue = append(externalQueue[:i], externalQueue[(i+1):]...)
			}

		}
	}

	internalQueueMutex <- true

}

func lookForOrder() {
	for {

		order := <-newOrderCh
		if checkIfNewOrder(order) {
			queueAddOrder(order)
			updateElevatorNextFloor()
		}

		time.Sleep(time.Millisecond)
	}
}

func Debug() {
	//<-internalQueueMutex
	for i := 0; i < len(internalQueue); i++ {
		fmt.Println("Orders ", internalQueue[i].Floor, internalQueue[i].Direction)
	}
	//internalQueueMutex <- true
}

func checkIfNewOrder(order constants.NewOrder) bool {
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

func calculateCost() int {
	return 0
}

func queueAddOrder(order constants.NewOrder) {
	<-internalQueueMutex

	if order.Direction == constants.DirStop {
		internalQueue = append(internalQueue, order)
	} else {
		externalQueue = append(externalQueue, order)
		if order.ElevatorID == elevator.ID {
			internalQueue = append(internalQueue, order)
		}
	}
	Debug()
	internalQueueMutex <- true
}

func updateElevatorNextFloor() {
	<-internalQueueMutex

	if len(internalQueue) != 0 {
		var bestFloorSoFar constants.NewOrder
		var nextFloorDist int = 100
		var dist int

		for i := 0; i < len(internalQueue); i++ {

			if elevator.Direction == constants.DirUp {

				if internalQueue[i].Floor > elevator.LastFloor {

					dist = internalQueue[i].Floor - elevator.LastFloor
					if dist <= nextFloorDist {

						nextFloorDist = dist
						bestFloorSoFar = internalQueue[i]

					}

				} else if internalQueue[i].Floor <= elevator.LastFloor {
					dist = ((constants.NumberOfFloors - 1) - elevator.LastFloor) + ((constants.NumberOfFloors - 1) - internalQueue[i].Floor)

					if dist <= nextFloorDist {

						nextFloorDist = dist
						bestFloorSoFar = internalQueue[i]

					}

				}

			} else if elevator.Direction == constants.DirDown {
				if internalQueue[i].Floor < elevator.LastFloor {

					dist = elevator.LastFloor - internalQueue[i].Floor
					if dist <= nextFloorDist {

						nextFloorDist = dist
						bestFloorSoFar = internalQueue[i]

					}

				} else if internalQueue[i].Floor >= elevator.LastFloor {
					dist = elevator.LastFloor + internalQueue[i].Floor

					if dist <= nextFloorDist {

						nextFloorDist = dist
						bestFloorSoFar = internalQueue[i]
					}

				}
			}

		}

		nextFloorCh <- bestFloorSoFar
	}

	internalQueueMutex <- true
}

/* Bruk i externalqueue
func internalQueueAddOrder(order constants.NewOrder) {
	//Vi har  lastfloor, orderedfloor, direction
	orderDir := getNeededElevatorDirection(order)


	for i := 0; i < len(internalQueue); i++ {
		queueOrderDir := getNeededElevatorDirection(internalQueue[i])

		select{
		case orderDir == queueOrderDir:
			if(orderDir == constants.DirUp){

				if(order.Floor < internalQueue[i].Floor){
					//Legg inn ordre på index i
				}

			} else{

				if(order.Floor > internalQueue[i].Floor){
					//Legg inn ordre på index i
				}

			}

		case orderDir != queueOrderDir && orderDir == elevator.Direction:
			//Legg inn ordre på index i

		case orderDir != queueOrderDir && orderDir != elevator.Direction:
			//

		default:

		}



	}
}
*/

func getNeededElevatorDirection(order constants.NewOrder) constants.ElevatorDirection {
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
