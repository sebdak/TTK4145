package main

import (
	constants "./constants"
	driver "./driver"
)

var State constants.ElevatorState 

func main() {
	initElev()
	//fmt.Println("hello world")
	for {
		
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
	State = constants.AT_FLOOR
}

func setDirection(dir constants.ElevatorDirection){
	driver.SetMotorDir(dir)
}


