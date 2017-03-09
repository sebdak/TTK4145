package elevator

import (
	constants "../constants"
	driver "../driver"
	"fmt"
	"time"
)

var LastFloor int
var CurrentOrder constants.Order
var Direction constants.ElevatorDirection = constants.DirStop
var state constants.ElevatorState
var ID int = 1

var newOrderCh chan constants.Order
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
			if LastFloor != CurrentOrder.Floor {
				goToOrderedFloor()
				state = constants.Moving
			}

			break

		case constants.Moving:

			//start timer first
			failedToReachFloorTimer := time.NewTimer(time.Second*6)
			

			//lookForChangeInFloor will return false if the timer times out
			if !lookForChangeInFloor(failedToReachFloorTimer){
				//change state to "Broken". maybe make it more precice later
				state = constants.Broken	
			}


			if LastFloor == CurrentOrder.Floor && driver.GetFloorSensor() != -1 {
				
				orderedFloorReachedRoutine()
				state = constants.AtFloor
			}

			break
		
		case constants.Broken:
			// go off network?
			break
		}

		time.Sleep(time.Millisecond)
	}
}

func debug() {
	fmt.Println("Yo")
}

func InitElev(newOrderChannel chan constants.Order, newExternalOrderChannel chan constants.Order, nextFloorChannel chan constants.Order, handledOrderChannel chan constants.Order) {
	//Add channels
	newOrderCh = newOrderChannel
	newExternalOrderCh = newExternalOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel

	//Elevator stuff
	CurrentOrder = constants.Order{Floor: 0, Direction: constants.DirStop, ElevatorID: -1}
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
	handledOrderCh <- CurrentOrder

	//Start floortimer
	waitAtFloorTimer := time.NewTimer(time.Second * 2)
	<-waitAtFloorTimer.C
}

func setLights() {
	if CurrentOrder.Direction == constants.DirUp {
		driver.SetButtonLamp(constants.ButtonCallUp, LastFloor, 0)
	} else if CurrentOrder.Direction == constants.DirDown {
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
		CurrentOrder = <-nextFloorCh
		fmt.Println("New elevator order: ", CurrentOrder.Floor, CurrentOrder.Direction, Direction)
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
	} else if LastFloor < CurrentOrder.Floor && CurrentOrder.Floor != -1 {
		Direction = constants.DirUp
	} else if LastFloor > CurrentOrder.Floor && CurrentOrder.Floor != -1 {
		Direction = constants.DirDown
	}
}

func lookForChangeInFloor(failedToReachFloorTimer time.Timer) bool {
	var currentFloorSignal = driver.GetFloorSensor()

	for {
		
		if currentFloorSignal != -1 && LastFloor != currentFloorSignal {
			LastFloor = currentFloorSignal
			driver.SetFloorIndicator(LastFloor)

			return true
		}

		select {
		case <- failedToReachFloorTimer.C:
			return false
		default:
			//prevent timer from blocking
		}
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
				newOrderCh <- newOrder
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
