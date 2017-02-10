package queue

import (
//constants "../constants"
//"container/list"
)

//var internalQueue [constants.NumberOfElevators]List
//var externalQueue [constants.NumberOfElevators]List

var nextFloorChannel *chan int

func InitQueue(nextFloorCh *chan int) {
	nextFloorChannel = nextFloorCh
}

func addInternalFloorOrder() {

}

func calculateCost() int {
	return 0
}
