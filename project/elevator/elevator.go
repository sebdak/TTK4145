package elevator

import (
	constants "../constants"
	driver "../driver"
	"fmt"
	"time"
	//timer
)

var LastFloor int
var currentOrder constants.NewOrder
var Direction constants.ElevatorDirection = constants.DirStop
var state constants.ElevatorState
var ID int = 1

var nextFloorCh chan constants.NewOrder
var newOrderCh chan constants.NewOrder
var handledOrderCh chan constants.NewOrder

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

func InitElev(newOrderChannel chan constants.NewOrder, nextFloorChannel chan constants.NewOrder, handledOrderChannel chan constants.NewOrder) {
	//Add channels
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel

	//Elevator stuff
	currentOrder = constants.NewOrder{Floor: 0, Direction: constants.DirStop, ElevatorID: -1}
	state = constants.Initializing
	driver.InitElev()

	goToFirstFloor()

	//Go online
	state = constants.AtFloor
	go lookForButtonPress()
	go lookForNewQueueOrder()
}

func orderedFloorReachedRoutine() {
	driver.SetMotorDir(constants.DirStop)
	driver.SetButtonLamp(constants.ButtonCallUp, LastFloor, 0)
	driver.SetButtonLamp(constants.ButtonCallDown, LastFloor, 0)
	driver.SetButtonLamp(constants.ButtonCommand, LastFloor, 0)
	setDirection()

	//Start floortimer
	waitAtFloorTimer := time.NewTimer(time.Second * 2)
	<-waitAtFloorTimer.C

	//Tell queue order has been handled
	handledOrderCh <- currentOrder
}

func goToFirstFloor() {
	if driver.GetFloorSensor() != 0 {
		driver.SetMotorDir(constants.DirDown)
	}

	for driver.GetFloorSensor() != 0 {
	}

	driver.SetMotorDir(constants.DirStop)
	setDirection()
}

func lookForNewQueueOrder() {
	for {
		currentOrder = <-nextFloorCh
		fmt.Println("New elevator order: ", currentOrder.Floor)
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
	var newOrder constants.NewOrder
	newOrder.ElevatorID = -1

	for {
		for floor := 0; floor < constants.NumberOfFloors; floor++ {

			if driver.GetButtonSignal(constants.ButtonCommand, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCommand, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirStop
				newOrder.ElevatorID = ID
				newOrderCh <- newOrder
			}

			if driver.GetButtonSignal(constants.ButtonCallUp, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCallUp, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirUp
				newOrderCh <- newOrder

			}

			if driver.GetButtonSignal(constants.ButtonCallDown, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCallDown, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirDown
				newOrderCh <- newOrder

			}

		}

		time.Sleep(time.Millisecond * 10)
	}
}
