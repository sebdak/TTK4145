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

var newOrderCh chan constants.NewOrder
var nextFloorCh chan int
var handledOrderCh chan constants.NewOrder

func InitQueue(newOrderChannel chan constants.NewOrder, nextFloorChannel chan int, handledOrderChannel chan constants.NewOrder) {
	//Add channels
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel

	go lookForNewInternalOrder()
}

func lookForNewInternalOrder() {
	for {

		order := <-newOrderCh
		if checkIfNewOrder(order) {
			queueAddOrder(order)
			updateNextFloor()
		}

		time.Sleep(time.Millisecond)
	}
}

func Debug() {
	for i := 0; i < len(internalQueue); i++ {
		fmt.Println("Orders ", internalQueue[i].Floor, internalQueue[i].Direction)
	}
}

func checkIfNewOrder(order constants.NewOrder) bool {
	for i := 0; i < len(internalQueue); i++ {
		if internalQueue[i] == order {
			return false
		}
	}
	for j := 0; j < len(externalQueue); j++ {
		if externalQueue[j] == order {
			return false
		}
	}
	return true
}

func calculateCost() int {
	return 0
}

func queueAddOrder(order constants.NewOrder) {
	if order.Direction == constants.DirStop {
		internalQueue = append(internalQueue, order)
	} else {
		externalQueue = append(externalQueue, order)
	}
}

func updateNextFloor() {
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

		nextFloorCh <- bestFloorSoFar.Floor

	}
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
