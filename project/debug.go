package main

import (
	constants "./constants"
	//"time"
	"fmt"
	"reflect"
)


var externalQueues [][]constants.Order


func main() {
	externalQueues = make([][]constants.Order, constants.NumberOfElevators)
	for i := 0; i < constants.NumberOfElevators; i++ {
		externalQueues[i] = make([]constants.Order, 1)
		externalQueues[i] = []constants.Order{}
	}


	fmt.Println("len externalQueues ", len(externalQueues[0]))
	//if len(externalQueues[0]) > 0 {
	compareAndFixExternalQueues()	
	//}
	fmt.Println("new externalQueue 0 ", externalQueues[0])
}


func compareAndFixExternalQueues() {
	count := 0
	var correctQueueIndex int
	var majority int = constants.NumberOfElevators/2 + 1
	fmt.Println("majority: ", majority)

	for i := 0; i < constants.NumberOfElevators; i++ {
		for j := 0; j < constants.NumberOfElevators; j++ {
			if reflect.DeepEqual(externalQueues[i], externalQueues[j]) {
				fmt.Println("Reflect: ", reflect.DeepEqual(externalQueues[i], externalQueues[j]))
				count++				
			}
		}

		// if all queuea are alike
		if count >= majority {
			correctQueueIndex = i
		}

		count = 0
	}

	//if queues dont match
	if(count < constants.NumberOfElevators){
		for i := 0; i < constants.NumberOfElevators; i++ {
			if i != correctQueueIndex {
				externalQueues[i] = externalQueues[correctQueueIndex]				
			}
		}
	}
}