package main

import (
	"fmt"
)

var internalQueue []Order
type ElevatorDirection int


type Order struct {
	Floor      int
	Direction  ElevatorDirection
	ElevatorID string
}

const (
	DirUp   ElevatorDirection = 1
	DirStop                   = 0
	DirDown                   = -1
)
var direction ElevatorDirection

func main() {

	//order0 := Order{Floor: 3, Direction: DirStop, ElevatorID: "-1"}
	direction = DirDown
	order1 := Order{Floor: 2, Direction: DirStop, ElevatorID: "-1"}
	order2 := Order{Floor: 2, Direction: DirDown, ElevatorID: "-1"}
	order3 := Order{Floor: 2, Direction: DirUp, ElevatorID: "-1"}
	order4 := Order{Floor: 3, Direction: DirUp, ElevatorID: "-1"}

	//internalQueue = append(internalQueue, order0)
 	internalQueue = append(internalQueue, order1)
 	internalQueue = append(internalQueue, order2)
 	internalQueue = append(internalQueue, order3)
 	internalQueue = append(internalQueue, order4)

    fmt.Println(internalQueue)

    deleteOrderFromInternalQueue(order1)
    deleteOrderFromInternalQueue(order3)
    deleteOrderFromInternalQueue(order4)

    fmt.Println(internalQueue)
    fmt.Println(len(internalQueue))

}


func deleteOrderFromInternalQueue(order Order) {
	//Assume function that calls this has internalqueuemutex

	for i := 0; i < len(internalQueue); i++ {
		if(order.Floor == internalQueue[i].Floor){
			if order.Direction == DirUp{
				if(internalQueue[i].Direction == DirUp || internalQueue[i].Direction == DirStop){
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				} 
			} else if order.Direction == DirDown{
				if(internalQueue[i].Direction == DirDown || internalQueue[i].Direction == DirStop){
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				} 
			} else if order.Direction == DirStop{
				if direction == DirUp && internalQueue[i].Direction == DirUp{
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				} else if direction == DirDown && internalQueue[i].Direction == DirDown{
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				} else if internalQueue[i].Direction == DirStop{
					internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
					i--
				}
			}
		}

		fmt.Println(i)
	}

}

/*
			if (internalQueue[i].Direction == DirStop){
				internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
				i--
			} else if (internalQueue[i].Direction == order.Direction){
				internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
				i--
			} else if (internalQueue[i].Direction == DirUp && order.Direction == DirStop){
				internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
				i--
			} else if (internalQueue[i].Direction == DirDown && order.Direction == DirStop){
				internalQueue = append(internalQueue[:i], internalQueue[(i+1):]...)
				i--
			}

*/
