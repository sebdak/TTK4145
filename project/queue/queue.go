package queue

import (
	constants "../constants"
	//"container/list"
	"fmt"
	"time"
)

var internalQueue []constants.NewOrder

//var externalQueue [constants.NumberOfElevators]List

var newOrderCh chan constants.NewOrder
var nextFloorCh chan int

func InitQueue(newOrderChannel chan constants.NewOrder, nextFloorChannel chan int) {
	//Add channels
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel

	go lookForNewInternalOrder()
}

func lookForNewInternalOrder() {
	for {

		order := <-newOrderCh
		if checkIfNewOrder(order) {
			fmt.Println("Added new order ", order.Floor, order.Direction)
			internalQueue = append(internalQueue, order)
		}

		time.Sleep(time.Millisecond)
	}
}

func checkIfNewOrder(order constants.NewOrder) bool {
	for i := 0; i < len(internalQueue); i++ {
		if internalQueue[i] == order {
			return false
		}
	}
	return true
}

func calculateCost() int {
	return 0
}

//NEXT TIME
//CREATE INTERNAL ORDER ARRAY AND MAKE ELEVATOR HANDLE INTERNAL QUEUE ORDERS
