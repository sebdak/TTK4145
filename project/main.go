package main

import (
	constants "./constants"
	elevator "./elevator"
	queue "./queue"
)

func main() {

	nextFloorCh := make(chan constants.InternalFloorOrder)
	elevator.InitElev(nextFloorCh)
	queue.InitQueue(nextFloorCh)

	go elevator.Run()

	runForever := make(chan bool)
	<-runForever
}
