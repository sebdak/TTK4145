package queue

import (
	constants "../constants"
	network "../network"
	driver "../driver"
	"reflect"
	"strconv"
	"os"
	"fmt"
	"io/ioutil"
	"io"
	"strings"
)

var headings map[string]constants.ElevatorHeading = make(map[string]constants.ElevatorHeading)

var internalQueue []constants.Order
var externalQueues [][]constants.Order
var ordersThatNeedToBeAdded []constants.Order
var ordersThatAreHandled []constants.Order

var internalQueueMutex chan bool = make(chan bool, 1)
var externalQueuesMutex chan bool = make(chan bool, 1)
var ordersThatNeedToBeAddedMutex chan bool = make(chan bool, 1)
var ordersThatAreHandledMutex chan bool = make(chan bool, 1)

var hallLightCh chan []constants.Order
var newOrderCh chan constants.Order
var nextFloorCh chan constants.Order
var handledOrderCh chan constants.Order
var newExternalOrderCh chan constants.Order
var peerDisconnectsCh chan string

var elevatorHeadingTx chan constants.ElevatorHeading
var elevatorHeadingRx chan constants.ElevatorHeading
var queuesTx chan []constants.Order
var queuesRx chan []constants.Order
var externalOrderTx chan constants.Order
var externalOrderRx chan constants.Order
var handledExternalOrderTx chan constants.Order
var handledExternalOrderRx chan constants.Order


func InitQueue(newOrderChannel chan constants.Order, newExternalOrderChannel chan constants.Order, nextFloorChannel chan constants.Order, handledOrderChannel chan constants.Order, peerDisconnectsChannel chan string, hallLightChannel chan []constants.Order, elevatorHeadingTxChannel chan constants.ElevatorHeading, elevatorHeadingRxChannel chan constants.ElevatorHeading, queuesTxChannel chan []constants.Order, queuesRxChannel chan []constants.Order, externalOrderTxChannel chan constants.Order, externalOrderRxChannel chan constants.Order, handledExternalOrderTxChannel chan constants.Order, handledExternalOrderRxChannel chan constants.Order) {
	//Add channels for modulecommunication
	newOrderCh = newOrderChannel
	nextFloorCh = nextFloorChannel
	handledOrderCh = handledOrderChannel
	newExternalOrderCh = newExternalOrderChannel
	peerDisconnectsCh = peerDisconnectsChannel
	hallLightCh = hallLightChannel

	//Channels for module-network communication
	elevatorHeadingTx = elevatorHeadingTxChannel
	elevatorHeadingRx = elevatorHeadingRxChannel
	queuesTx = queuesTxChannel
	queuesRx = queuesRxChannel
	externalOrderTx = externalOrderTxChannel
	externalOrderRx = externalOrderRxChannel
	handledExternalOrderTx = handledExternalOrderTxChannel
	handledExternalOrderRx = handledExternalOrderRxChannel

	//mutex := true
	internalQueueMutex <- true
	externalQueuesMutex <- true
	ordersThatNeedToBeAddedMutex <- true
	ordersThatAreHandledMutex <- true

	//init external queue
	initQueues()

	go handleNewCabOrder()
	go handleExternalButtonOrder()
	go handleCompletedCabOrder()

	//readInternalQueueFromFile
	readInternalQueueFromFile()

	go sendElevatorHeading()
	go getElevatorHeadings()

	go masterSendExternalQueue()
	go getExternalQueuesAndUpdate()

	go handlePeerDisconnects()

	go masterGetExternalOrdersThatAreHandled()
	go masterGetExternalOrdersThatNeedToBeAdded()
}

func initQueues() {
	//internalQueue = make([]constants.Order)

	externalQueues = make([][]constants.Order, constants.QueueCopies)
	for i := 0; i < constants.QueueCopies; i++ {
		externalQueues[i] = make([]constants.Order, 1)
		externalQueues[i] = []constants.Order{}
	}
}

func compareAndFixExternalQueues() {
	fmt.Println("28")
	count := 0
	totalCount := 0
	var correctQueueIndex int
	var majority int = constants.QueueCopies/2 + 1

	for i := 0; i < constants.QueueCopies; i++ {
		for j := 0; j < constants.QueueCopies; j++ {
			if reflect.DeepEqual(externalQueues[i], externalQueues[j]) {
				count++
			}
		}

		// If current queue is equal more than 50% of the queue copies
		if count >= majority {
			correctQueueIndex = i
		}
		totalCount += count
		count = 0
	}

	//if there was a mismatch
	if totalCount < constants.QueueCopies*constants.QueueCopies {
		for i := 0; i < constants.QueueCopies; i++ {
			if i != correctQueueIndex {
				externalQueues[i] = externalQueues[correctQueueIndex]
			}
		}
	}
}

func writeInternalQueueToFile() {
	fmt.Println("29")
	fo, err := os.Create("internalQueue.txt")
	if err != nil {
	}
	defer fo.Close()

	_, err = io.Copy(fo, strings.NewReader(internalQueueToString()))
	if err != nil {
	}
	
}

func readInternalQueueFromFile() {
	fmt.Println("30")
	b, err := ioutil.ReadFile("internalQueue.txt")
    if err != nil {
        fmt.Print(err)
    } else {
    	str := string(b)
    	
    	s := strings.Split(str, "\n")
    	for i := 0; i < len(s)-1; i++{
    		floor, _ := strconv.Atoi(s[i])
    		order := constants.Order{Floor: floor, Direction: constants.DirStop, ElevatorID: network.Id}

    		if checkIfNewInternalOrder(order) {
    			//set lights
    			driver.SetButtonLamp(constants.ButtonCommand, order.Floor, 1)
    			newOrderCh <- order
    		}
    	}
    }
}

func internalQueueToString() string {
	fmt.Println("31")
	var internalQueueString string;

	for i := 0; i < len(internalQueue); i++ {
		internalQueueString = internalQueueString + strconv.Itoa(internalQueue[i].Floor) + "\n"
	}

	return internalQueueString
}



