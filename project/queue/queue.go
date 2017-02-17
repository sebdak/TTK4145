package queue

import (
	constants "../constants"
	//"container/list"
	"fmt"
	"time"
)

//var internalQueue [constants.NumberOfElevators]List
//var externalQueue [constants.NumberOfElevators]List

var nextFloorChannel chan constants.InternalFloorOrder

func InitQueue(nextFloorCh chan constants.InternalFloorOrder) {
	nextFloorChannel = nextFloorCh
	go lookForNewInternalOrder()
}

func lookForNewInternalOrder() {
	for {

		newOrder := <-nextFloorChannel
		fmt.Println("Received order ", newOrder.Floor, newOrder.Direction)

		time.Sleep(time.Millisecond)
	}
}

func calculateCost() int {
	return 0
}

//NEXT TIME
//CREATE INTERNAL ORDER ARRAY AND MAKE ELEVATOR HANDLE INTERNAL QUEUE ORDERS
