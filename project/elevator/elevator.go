package elevator

import (
	constants "../constants"
	driver "../driver"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"time"
)

var LastFloor int
var lastOrder constants.Order
var CurrentOrder constants.Order
var Direction constants.ElevatorDirection
var state constants.ElevatorState

var newOrderCh chan constants.Order
var newExternalOrderCh chan constants.Order
var nextFloorCh chan constants.Order
var handledOrderCh chan constants.Order
var hallLightCh chan []constants.Order


func InitElev(newOrderChannel chan constants.Order, newExternalOrderChannel chan constants.Order, nextFloorChannel chan constants.Order, handledOrderChannel chan constants.Order, hallLightChannel chan []constants.Order) {
	//Add channels
	newOrderCh = newOrderChannel
	newExternalOrderCh = newExternalOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel
	hallLightCh = hallLightChannel

	//Elevator stuff
	state = constants.Initializing
	driver.InitElev()
	run()

}

func run() {

	for {
		switch state {
		case constants.Initializing:
			if(testElevator() == true){
				state = constants.AtFloor
				go lookForOrderButtonPress()
				go lookForNewQueueOrder()
				go setHallLights()
			} else {
				state = constants.Broken
			}
			break

		case constants.AtFloor:
			if CurrentOrder != lastOrder {
				moveTowardsOrderedFloor()
				state = constants.Moving
			}

			break

		case constants.Moving:

			secureElevatorIsMoving()

			if LastFloor == CurrentOrder.Floor && (CurrentOrder.Direction == elevator.Direction ||CurrentOrder.Direction == DirStop) { 

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

func testElevator() bool {
	failedToReachFloorTimer := time.NewTimer(time.Second * 12)

	if driver.GetFloorSensor() != 0 {
		//Send elevator to 1 floor and check that it's reached
		CurrentOrder = constants.Order{Floor: 0, Direction: constants.DirStop, ElevatorID: "-1"}
		moveTowardsOrderedFloor()
		
		select {
		case <-failedToReachFloorTimer.C:
			return false
		case driver.GetFloorSensor() == 0:
			break
		}

	} else {
		//Elevator is in 1 floor so send it up and down one floor to confirm that it's working
		CurrentOrder = constants.Order{Floor: 1, Direction: constants.DirStop, ElevatorID: "-1"}
		moveTowardsOrderedFloor()
		
		select {
		case <-failedToReachFloorTimer.C:
			return false
		case driver.GetFloorSensor() == 1:
			break
		}

		CurrentOrder = constants.Order{Floor: 0, Direction: constants.DirStop, ElevatorID: "-1"}
		moveTowardsOrderedFloor()

		select {
		case <-failedToReachFloorTimer.C:
			return false
		case driver.GetFloorSensor() == 0:
			break
		}
	}

	driver.SetMotorDir(constants.DirStop)
	lastOrder = CurrentOrder
	LastFloor = 0
	return true
}

func lookForNewQueueOrder() {
	for {
		CurrentOrder = <-nextFloorCh
		fmt.Println("New elevator order: ", CurrentOrder.Floor, CurrentOrder.Direction, Direction)
		time.Sleep(time.Millisecond * 5)
	}
}

func moveTowardsOrderedFloor() {
	setDirection()
	driver.SetMotorDir(Direction)
}

func setDirection() {
	if LastFloor == 0 {
		Direction = constants.DirUp
	} else if LastFloor == constants.NumberOfFloors-1 {
		Direction = constants.DirDown
	} else if CurrentOrder.Floor > LastFloor{
		Direction = constants.DirUp
	} else if CurrentOrder.Floor < LastFloor{
		Direction = constants.DirDown
	} else if CurrentOrder.Floor == LastFloor{
		Direction = constants.DirStop
	}
}

func secureElevatorIsMoving() {
	//start timer first
	failedToReachFloorTimer := time.NewTimer(time.Second * 12)

	//lookForChangeInFloor will return false if the timer times out
	if !lookForChangeInFloor(failedToReachFloorTimer) {
		//change state to "Broken". maybe make it more precice later
		state = constants.Broken
	}
}

func lookForChangeInFloor(failedToReachFloorTimer *time.Timer) bool {

	for {
		currentFloorSignal := driver.GetFloorSensor()
		if currentFloorSignal != -1 && (LastFloor != currentFloorSignal || CurrentOrder.Floor == currentFloorSignal){
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
		time.Sleep(time.Millisecond * 10)
	}
}

func orderedFloorReachedRoutine() {
	driver.SetMotorDir(constants.DirStop)
	driver.SetButtonLamp(constants.ButtonCommand, LastFloor, 0) //Cab order lights can be directly shut off by elevator

	lastOrder = CurrentOrder

	//Tell queue order has been handled
	handledOrderCh <- CurrentOrder

	//Start floortimer and open doors
	driver.SetDoorOpenLamp(1)
	waitAtFloorTimer := time.NewTimer(time.Second * 2)
	<-waitAtFloorTimer.C
	driver.SetDoorOpenLamp(0)
}

func lookForOrderButtonPress() {
	var newOrder constants.Order

	for {
		for floor := 0; floor < constants.NumberOfFloors; floor++ {

			if driver.GetButtonSignal(constants.ButtonCommand, floor) == 1 {
				driver.SetButtonLamp(constants.ButtonCommand, floor, 1)
				newOrder.Floor = floor
				newOrder.Direction = constants.DirStop
				newOrderCh <- newOrder
			} else if driver.GetButtonSignal(constants.ButtonCallUp, floor) == 1 {
				newOrder.Floor = floor
				newOrder.Direction = constants.DirUp
				newExternalOrderCh <- newOrder

			} else if driver.GetButtonSignal(constants.ButtonCallDown, floor) == 1 {
				newOrder.Floor = floor
				newOrder.Direction = constants.DirDown
				newExternalOrderCh <- newOrder

			}

		}

		time.Sleep(time.Millisecond * 10)
	}
}

func updateHallLights() {
	var qCopy []constants.Order
	for {
		q := <-hallLightCh
		if reflect.DeepEqual(q, qCopy) == false && len(q) > 0 { //To avoid setting lights all the time without receiving new external queue
			qCopy = q 
			for i := 0; i < constants.NumberOfFloors; i++ {

				//Go through all up-hall-buttons
				for j := 0; j < len(q); j++ {
					if i == q[j].Floor && constants.DirUp == q[j].Direction {
						driver.SetButtonLamp(constants.ButtonCallUp, i, 1)
						break //Only one button-up-order for each floor
					} else {
						driver.SetButtonLamp(constants.ButtonCallUp, i, 0)
					}
				}

				//Go through all down-hall-buttons
				for j := 0; j < len(q); j++ {
					if i == q[j].Floor && constants.DirDown == q[j].Direction {
						driver.SetButtonLamp(constants.ButtonCallDown, i, 1)
						break //Only one button-down-order for each floor
					} else {
						driver.SetButtonLamp(constants.ButtonCallDown, i, 0)
					}
				}
			}
		} else if len(q) == 0 { //No ext. orders -> no lights should be lit
			for i := 0; i < constants.NumberOfFloors; i++ {
				driver.SetButtonLamp(constants.ButtonCallUp, i, 0)
				driver.SetButtonLamp(constants.ButtonCallDown, i, 0)
			}
		}
	}
}

func shutdownRoutine() {
	//spawn new process
	fmt.Println("Elevator broken, attempting reboot")

	exec.Command("gnome-terminal", "-x", "go", "run", "main.go").Output()

	//kill this process
	proc, _ := os.FindProcess(os.Getpid())
	proc.Kill()
}
