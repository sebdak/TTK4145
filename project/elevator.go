package main

import (
	constants "./constants"
	driver "./driver"
	//timer
)



var lastFloor int
var orderedFloor int
var State constants.ElevatorState 

func main() {
	initElev()
	
	for {
		//switch	
	}
}

func SetNextOrder(floor int) {
	
}

func ReadState() {

}

func Reboot() {

}

func initElev() {
	State = constants.INITIALIZING
	driver.InitElev()
	if (driver.GetFloorSensor() != 0) {
		setDirection(constants.DIR_DOWN)
	}
	for (driver.GetFloorSensor() != 0) {}

	setDirection(constants.DIR_STOP)
	//Go online	
	State = constants.AT_FLOOR
}

func setDirection(dir constants.ElevatorDirection){
	driver.SetMotorDir(dir)
}

func goToFloor(floor int) {
	//start timer
	
}
