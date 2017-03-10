package elevator

import (
	constants "../constants"
	driver "../driver"
	"fmt"
	"reflect"
	"time"
)

var LastFloor int
var CurrentOrder constants.Order
var Direction constants.ElevatorDirection = constants.DirStop
var state constants.ElevatorState

var newOrderCh chan constants.Order
var newExternalOrderCh chan constants.Order
var nextFloorCh chan constants.Order
var handledOrderCh chan constants.Order
var hallLightCh chan []constants.Order

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

			secureFloorIsReached()

			if LastFloor == CurrentOrder.Floor && driver.GetFloorSensor() != -1 {

				orderedFloorReachedRoutine()
				state = constants.AtFloor
			}

			break

		case constants.Broken:
			shutdownRoutine()
			break
		}

		time.Sleep(time.Millisecond)
	}
}

func InitElev(newOrderChannel chan constants.Order, newExternalOrderChannel chan constants.Order, nextFloorChannel chan constants.Order, handledOrderChannel chan constants.Order, hallLightChannel chan []constants.Order) {
	//Add channels
	newOrderCh = newOrderChannel
	newExternalOrderCh = newExternalOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel
	hallLightCh = hallLightChannel

	//Elevator stuff
	CurrentOrder = constants.Order{Floor: 0, Direction: constants.DirStop, ElevatorID: "-1"}
	state = constants.Initializing
	driver.InitElev()

	goToFirstFloor()
	setDirection()

	//Go online
	state = constants.AtFloor
	go lookForButtonPress()
	go lookForNewQueueOrder()
	go setHallLights()
}

func setHallLights() {
	qCopy := make([]constants.Order)
	for {
		q := <-hallLightCh

		if reflect.DeepEqual(q, qCopy) == false {
			for i := 0; i < constants.NumberOfFloors; i++ {
				for j := 0; j < len(q); j++ {
					if i == q[j].Floor && q[j].Direction == constants.DirUp {
						//turn on for dir up
						driver.SetButtonLamp(constants.ButtonCallUp, i, 1)
						break
					} else {
						//turn off light for dir up
						driver.SetButtonLamp(constants.ButtonCallUp, i, 0)
					}
				}

				for j := 0; j < len(q); j++ {
					if i == q[j].Floor && q[j].Direction == constants.DirDown {
						//turn on for dir down
						driver.SetButtonLamp(constants.ButtonCallDown, i, 1)
						break
					} else {
						//turn off light dir down
						driver.SetButtonLamp(constants.ButtonCallDown, i, 1)
					}
				}
			}
			qCopy = q
		}
	}
}

func secureFloorIsReached() {
	//start timer first
	failedToReachFloorTimer := time.NewTimer(time.Second * 6)

	//lookForChangeInFloor will return false if the timer times out
	if !lookForChangeInFloor(failedToReachFloorTimer) {
		//change state to "Broken". maybe make it more precice later
		state = constants.Broken
	}
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

			failedToReachFloorTimer.Stop()
			return true
		}

		select {
		case <-failedToReachFloorTimer.C:
			return false
		default:
			//prevent timer from blocking
		}
	}
}

func lookForButtonPress() {
	var newOrder constants.Order

	for {
		for floor := 0; floor < constants.NumberOfFloors; floor++ {

			if driver.GetButtonSignal(constants.ButtonCommand, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCommand, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirStop
				newOrderCh <- newOrder
			}

			if driver.GetButtonSignal(constants.ButtonCallUp, floor) == 1 {
				//driver.SetButtonLamp(constants.ButtonCallUp, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirUp
				newExternalOrderCh <- newOrder

			}

			if driver.GetButtonSignal(constants.ButtonCallDown, floor) == 1 {
				//driver.SetButtonLamp(constants.ButtonCallDown, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirDown
				newExternalOrderCh <- newOrder

			}

		}

		time.Sleep(time.Millisecond * 10)
	}
}

func shutdownRoutine() {
	//go off network
	peerTxEnableCh <- false

	//spawn new process
	_, _ := exec.Command("gnome-terminal", "-x", "go", "run", "main.go").Output()

	//kill this process
	proc, _ := os.FindProcess(os.Getpid())
	proc.Kill()
}
