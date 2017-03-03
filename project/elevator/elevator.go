package elevator

import (
	constants "../constants"
	driver "../driver"
	"fmt"
	"time"
	//timer
)

var LastFloor int
var OrderedFloor int = -1
var Direction constants.ElevatorDirection = constants.DirStop
var state constants.ElevatorState

var nextFloorCh chan int
var newOrderCh chan constants.NewOrder
var handledOrderCh chan constants.NewOrder

func Run() {

	for {
		switch state {
		case constants.Initializing:
			//Reboot or something
			break

		case constants.AtFloor:
			if LastFloor != OrderedFloor {
				goToOrderedFloor()
				state = constants.Moving
			}

			break

		case constants.Moving:
			lookForChangeInFloor()
			if LastFloor == OrderedFloor && driver.GetFloorSensor() != -1 {
				orderedFloorReachedRoutine()
				state = constants.AtFloor
				Direction = constants.DirStop
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

	OrderedFloor = floor

}

func orderedFloorReachedRoutine() {
	driver.SetMotorDir(constants.DirStop)
	driver.SetButtonLamp(constants.ButtonCallUp, LastFloor, 0)
	driver.SetButtonLamp(constants.ButtonCallDown, LastFloor, 0)
	driver.SetButtonLamp(constants.ButtonCommand, LastFloor, 0)
	//Tell queue ordered has been handled and ask for new order
	//Set door open light and start floortimer

}

func ReadState() {

}

func Reboot() {

}

func InitElev(newOrderChannel chan constants.NewOrder, nextFloorChannel chan int, handledOrderChannel chan constants.NewOrder) {
	//Add channels
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel

	//Elevator stuff
	state = constants.Initializing
	driver.InitElev()
	if driver.GetFloorSensor() != 0 {
		driver.SetMotorDir(constants.DirDown)
	}
	i := 1
	for driver.GetFloorSensor() != 0 {
		fmt.Println(driver.GetFloorSensor(), i)
		i++
		time.Sleep(time.Millisecond * 100)
	}

	driver.SetMotorDir(constants.DirStop)
	//Go online
	state = constants.AtFloor
	go lookForButtonPress()
}

func goToOrderedFloor() {
	//start timer
	setDirection()
	driver.SetMotorDir(Direction)
}

func setDirection() {
	if LastFloor < OrderedFloor && OrderedFloor != -1 {
		Direction = constants.DirUp
	} else if LastFloor < OrderedFloor && OrderedFloor != -1 {
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
	var newInternalOrder constants.NewOrder
	newInternalOrder.Elevator = -1

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
