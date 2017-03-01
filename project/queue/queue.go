package queue

import (
	constants "../constants"
	elevator "../elevator"
	//"container/list"
	"fmt"
	"time"
)

var internalQueue []constants.NewOrder
var externalQueue []constants.ExternalOrder
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
			fmt.Println("Added new order ", order.Floor, order.Direction)
			internalQueue = append(internalQueue, order)
			if order.Direction == constants.DirStop {
				internalQueue = append(internalQueue, order)
			}
			else {
				externalQueue = append(externalQueue,order)
			}
		}

		time.Sleep(time.Millisecond)
	}
}

func checkIfNewOrder(order constants.NewOrder) bool {
	for i := 0; i < len(internalQueue); i++ {
		if internalQueue[i] == order {
			return false
		}
	}
	for j := 0; j < len(externalQueue); j++ {
		if externalQueue[i] == order {
			return false
		}
	}
	return true
}

func calculateCost() int {
	return 0
}


func internalQueueAddOrder(order constants.NewOrder) {
	//Vi har  lastfloor, orderedfloor, direction
	orderDir := getNeededElevatorDirection(order)

	//DERSOM ORDERDIR OG ELEVATORDIR IKKE STEMMER OVERENS, VENT TIL DE STEMMER
	//INTERNALQUEUE BURDE HENTES UT FRA CHANNEL HVIS IKKE KAN CONQUERRENCY PROBLEMER OPPSTÅ

	for i := 0; i < len(internalQueue); i++ {
		queueOrderDir := getNeededElevatorDirection(internalQueue[i])

		switch{
		case orderDir == queueOrderDir:
			if(orderDir == constants.DirUp){

				if(order.Floor < internalQueue[i].Floor){
					//Legg inn ordre på index i
				} else if(i==len(internalQueue)-1){
					//Legg inn ordre bakerst
				}

			} else{

				if(order.Floor > internalQueue[i].Floor){
					//Legg inn ordre på index i
				} else if(i==len(internalQueue)-1){
					//Legg inn ordre bakerst
				}

			}

		case orderDir != queueOrderDir && i == len(internalQueue)-1:
			//Legg til ordre bakerst

		default:

		}



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

func getNeededElevatorDirection(order constants.NewOrder){
		direction := constants.DirStop
	if order.Floor > elevator.LastFloor {
		direction = constants.DirUp
	}
	else {
		direction = constants.DirDown
	}
	return direction
}
//NEXT TIME
//CREATE INTERNAL ORDER ARRAY AND MAKE ELEVATOR HANDLE INTERNAL QUEUE ORDERS
