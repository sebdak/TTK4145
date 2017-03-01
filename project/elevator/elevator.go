package elevator

import (
	constants "../constants"
	driver "../driver"
	"fmt"
	"time"
	//timer
)

var lastFloor int
var orderedFloor int
var State constants.ElevatorState

var nextFloorCh chan int
var newOrderCh chan constants.NewOrder

func Run() {

	for {
		switch State {
		case constants.Initializing:
			//Reboot or something
			break

		case constants.AtFloor:
			if lastFloor != orderedFloor {
				goToOrderedFloor()
				State = constants.Moving
			}

			break

		case constants.Moving:
			lookForChangeInFloor()
			if lastFloor == orderedFloor && driver.GetFloorSensor() != -1 {
				orderedFloorReachedRoutine()
				State = constants.AtFloor
			}

			break
		}

		time.Sleep(time.Millisecond)
	}
}

func debug() {
	fmt.Println("Yo")
}

func UpdateNextOrder(floor int) {

	orderedFloor = floor

}

func orderedFloorReachedRoutine() {
	driver.SetMotorDir(constants.DirStop)
	driver.SetButtonLamp(constants.ButtonCallUp, lastFloor, 0)
	driver.SetButtonLamp(constants.ButtonCallDown, lastFloor, 0)
	driver.SetButtonLamp(constants.ButtonCommand, lastFloor, 0)
	//Tell queue ordered has been handled and ask for new order
	//Set door open light and start floortimer

}

func ReadState() {

}

func Reboot() {

}

func InitElev(newOrderChannel chan constants.NewOrder, nextFloorChannel chan int) {
	//Add channels
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel

	//Elevator stuff
	State = constants.Initializing
	driver.InitElev()
	if driver.GetFloorSensor() != 0 {
		driver.SetMotorDir(constants.DirDown)
	}
	for driver.GetFloorSensor() != 0 {
	}

	driver.SetMotorDir(constants.DirStop)
	//Go online
	State = constants.AtFloor
	go lookForButtonPress()
}

func goToOrderedFloor() {
	//start timer
	if lastFloor > orderedFloor {
		driver.SetMotorDir(constants.DirDown)
	} else {
		driver.SetMotorDir(constants.DirUp)
	}

}

func lookForChangeInFloor() {
	var currentFloorSignal = driver.GetFloorSensor()
	if currentFloorSignal != -1 && lastFloor != currentFloorSignal {

		lastFloor = currentFloorSignal
		driver.SetFloorIndicator(lastFloor)

	}
}

func lookForButtonPress() {
	var newInternalOrder constants.NewOrder

	for {
		for floor := 0; floor < constants.NumberOfFloors; floor++ {

			if driver.GetButtonSignal(constants.ButtonCommand, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCommand, floor, 1)
				newInternalOrder.Floor = floor
				newInternalOrder.Direction = constants.DirStop
				newOrderCh <- newInternalOrder
			}

			if driver.GetButtonSignal(constants.ButtonCallUp, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCallUp, floor, 1)
				newInternalOrder.Floor = floor
				newInternalOrder.Direction = constants.DirUp
				newOrderCh <- newInternalOrder

			}

			if driver.GetButtonSignal(constants.ButtonCallDown, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCallDown, floor, 1)
				newInternalOrder.Floor = floor
				newInternalOrder.Direction = constants.DirDown
				newOrderCh <- newInternalOrder

			}

		}

		time.Sleep(time.Millisecond * 10)
	}
}