package queue

import (
	"container/list"
	constants "../constants"
)

type 

var internalQueue [constants.NumberOfElevators]List
var externalQueue [constants.NumberOfElevators]List

chan nextFloorChannel int

func InitQueue(nextFloorCh *chan int){
	nextFloorChannel = nextFloorCh
}

func addInternalFloorOrder() {

}



func calculateCost int()