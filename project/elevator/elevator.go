package elevator

import (
	constants "../constants"
	driver "../driver"
	"fmt"
	"time"
)

var LastFloor int
var currentOrder constants.Order
var Direction constants.ElevatorDirection = constants.DirStop
var state constants.ElevatorState
var ID int = 1

var newInternalOrderCh chan constants.Order
var newExternalOrderCh chan constants.Order
var nextFloorCh chan constants.Order
var handledOrderCh chan constants.Order

func Run() {

	for {
		switch state {
		case constants.Initializing:
			//Reboot or something
			break

		case constants.AtFloor:
			if LastFloor != currentOrder.Floor {
				goToOrderedFloor()
				state = constants.Moving
			}

			break

		case constants.Moving:
			lookForChangeInFloor()
			if LastFloor == currentOrder.Floor && driver.GetFloorSensor() != -1 {
				orderedFloorReachedRoutine()
				state = constants.AtFloor
			}

			break
		}

		time.Sleep(time.Millisecond)
	}
}

func debug() {
	fmt.Println("Yo")
}

func InitElev(newInternalOrderChannel chan constants.Order, newExternalOrderChannel chan constants.Order, nextFloorChannel chan constants.Order, handledOrderChannel chan constants.Order) {
	//Add channels
	newInternalOrderCh = newInternalOrderChannel
	newExternalOrderCh = newExternalOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel

	//Elevator stuff
	currentOrder = constants.Order{Floor: 0, Direction: constants.DirStop, ElevatorID: -1}
	state = constants.Initializing
	driver.InitElev()

	goToFirstFloor()
	setDirection()

	//Go online
	state = constants.AtFloor
	go lookForButtonPress()
	go lookForNewQueueOrder()
}

func orderedFloorReachedRoutine() {
	driver.SetMotorDir(constants.DirStop)
	setLights()
	setDirection()

	//Tell queue order has been handled
	handledOrderCh <- currentOrder

	//Start floortimer
	waitAtFloorTimer := time.NewTimer(time.Second * 2)
	<-waitAtFloorTimer.C
}

func setLights() {
	if currentOrder.Direction == constants.DirUp {
		driver.SetButtonLamp(constants.ButtonCallUp, LastFloor, 0)
	} else if currentOrder.Direction == constants.DirDown {
		driver.SetButtonLamp(constants.ButtonCallDown, LastFloor, 0)
	}
	driver.SetButtonLamp(constants.ButtonCommand, LastFloor, 0)
}

func goToFirstFloor() {
	if driver.GetFloorSensor() != 0 {
		driver.SetMotorDir(constants.DirDown)
	}

	for driver.GetFloorSensor() != 0 {
	}

	driver.SetMotorDir(constants.DirStop)
}

func lookForNewQueueOrder() {
	for {
		currentOrder = <-nextFloorCh
		fmt.Println("New elevator order: ", currentOrder.Floor, currentOrder.Direction, Direction)
		time.Sleep(time.Millisecond * 5)
	}
}

func goToOrderedFloor() {
	//start timer
	setDirection()
	driver.SetMotorDir(Direction)
}

func setDirection() {
	if LastFloor == 0 {
		Direction = constants.DirUp
	} else if LastFloor == constants.NumberOfFloors-1 {
		Direction = constants.DirDown
	} else if LastFloor < currentOrder.Floor && currentOrder.Floor != -1 {
		Direction = constants.DirUp
	} else if LastFloor > currentOrder.Floor && currentOrder.Floor != -1 {
		Direction = constants.DirDown
	}
}

func lookForChangeInFloor() {
	var currentFloorSignal = driver.GetFloorSensor()
	if currentFloorSignal != -1 && LastFloor != currentFloorSignal {

		LastFloor = currentFloorSignal
		driver.SetFloorIndicator(LastFloor)

	}
}

func lookForButtonPress() {
	var newOrder constants.Order
	newOrder.ElevatorID = 1

	for {
		for floor := 0; floor < constants.NumberOfFloors; floor++ {

			if driver.GetButtonSignal(constants.ButtonCommand, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCommand, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirStop
				newOrder.ElevatorID = ID
				newInternalOrderCh <- newOrder
			}

			if driver.GetButtonSignal(constants.ButtonCallUp, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCallUp, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirUp
				newExternalOrderCh <- newOrder

			}

			if driver.GetButtonSignal(constants.ButtonCallDown, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCallDown, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirDown
				newExternalOrderCh <- newOrder

			}

		}

		time.Sleep(time.Millisecond * 10)
	}
}
