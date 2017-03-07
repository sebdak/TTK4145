package queue

import (
	network "../network"
	constants "../constants"
	elevator "../elevator"
	"fmt"
	"time"
)

var internalQueue []constants.Order
var externalQueue []constants.Order

var internalQueueMutex = make(chan bool, 1)
var newOrderCh chan constants.Order
var nextFloorCh chan constants.Order
var handledOrderCh chan constants.Order

var elevatorHeadingTx chan constants.ElevatorHeading
var elevatorHeadingRx chan constants.ElevatorHeading
var queuesTx chan []constants.Order
var queuesRx chan []constants.Order

var headings map[string]constants.ElevatorHeading = make(map[string]constants.ElevatorHeading)


func InitQueue(newOrderChannel chan constants.Order, nextFloorChannel chan constants.Order, handledOrderChannel chan constants.Order, elevatorHeadingTxChannel chan constants.ElevatorHeading, elevatorHeadingRxChannel chan constants.ElevatorHeading, queuesTxChannel chan []constants.Order, queuesRxChannel chan []constants.Order) {
	//Add channels for modulecommunication
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel

	//Channels for module-network communication
	elevatorHeadingTx = elevatorHeadingTxChannel
	elevatorHeadingRx = elevatorHeadingRxChannel
	queuesTx = queuesTxChannel
	queuesRx = queuesRxChannel

	//mutex := true
	internalQueueMutex <- true

	go lookForOrder()
	go lookForHandledOrder()
	go sendElevatorHeading()
	go updateElevatorHeadings()
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

//Skriv transmit og receive av queues!!!!!!!


func Debug() {
	//<-internalQueueMutex
	for i := 0; i < len(internalQueue); i++ {
		fmt.Println("Orders ", internalQueue[i].Floor, internalQueue[i].Direction)
	}
	//internalQueueMutex <- true
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

func deleteOrderFromQueues(order constants.Order) {
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
		if CheckIfNewOrder(order) {
			queueAddOrder(order)
			updateElevatorNextFloor()
		}

		time.Sleep(time.Millisecond)
	}
}


func CheckIfNewOrder(order constants.Order) bool {
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

func updateElevatorNextFloor() {
	<-internalQueueMutex

	if len(internalQueue) != 0 {
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

/* Bruk i externalqueue
func internalQueueAddOrder(order constants.Order) {
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
