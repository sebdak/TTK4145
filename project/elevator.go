package main

import (
	constants "./constants"
	driver "./driver"
)


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
	driver.InitElev()
	if (driver.GetFloorSensor() != 0) {
		setDirection(constants.DIR_DOWN)
	}
	for (driver.GetFloorSensor() != 0) {}

	setDirection(constants.DIR_STOP)
}

func setDirection(dir constants.ElevatorDirection){
	driver.SetMotorDir(dir)
}
