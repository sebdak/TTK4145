package main

import (
	driver "./driver"
	"fmt"
)

var State elevatorState

func main() {
	initE()
	fmt.Println("hello world")
	for {
		
	}
}

func SetNextOrder(floor int) {
	
}

func ReadState() {

}

func Reboot() {

}

func initE() {
	driver.InitElev()
	driver.SetFloorIndicator(2)
}


