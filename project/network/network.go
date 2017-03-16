package network

import (
	constants "../constants"
	bcast "./bcast"
	localip "./localip"
	peers "./peers"
	"fmt"
	"os"
	"time"
	//"reflect"
)

var Master bool = false

var peerUpdateCh chan peers.PeerUpdate
var newOrderCh chan constants.Order
var peerDisconnectsCh chan string

var elevatorHeadingTx chan constants.ElevatorHeading
var elevatorHeadingRx chan constants.ElevatorHeading
var queuesTx chan []constants.Order
var queuesRx chan []constants.Order
var externalOrderTx chan constants.Order
var externalOrderRx chan constants.Order
var handledExternalOrderTx chan constants.Order
var handledExternalOrderRx chan constants.Order
var masterRx chan string

var PeersInfo peers.PeerUpdate
var Id string = ""
var Online bool = false

func InitNetwork(newOrderChannel chan constants.Order, peerDisconnectsChannel chan string, elevatorHeadingTxChannel chan constants.ElevatorHeading, elevatorHeadingRxChannel chan constants.ElevatorHeading, queuesTxChannel chan []constants.Order, queuesRxChannel chan []constants.Order, externalOrderTxChannel chan constants.Order, externalOrderRxChannel chan constants.Order, handledExternalOrderTxChannel chan constants.Order, handledExternalOrderRxChannel chan constants.Order) {
	//Store channels for module->module communication
	newOrderCh = newOrderChannel
	peerDisconnectsCh = peerDisconnectsChannel

	//Store channels for module->network->module communication
	elevatorHeadingTx = elevatorHeadingTxChannel
	elevatorHeadingRx = elevatorHeadingRxChannel
	queuesTx = queuesTxChannel
	queuesRx = queuesRxChannel
	externalOrderTx = externalOrderTxChannel
	externalOrderRx = externalOrderRxChannel
	handledExternalOrderTx = handledExternalOrderTxChannel
	handledExternalOrderRx = handledExternalOrderRxChannel

	//Tries to go online
	for !testIfOnline() {
		fmt.Println("Not online, trying to reconnect")
		time.Sleep(time.Second * 3)
	}

	StartUDPPeersBroadcast()

	go masterBroadcast()
	go lookForChangeInPeers()

	listenForMaster()

	//Wait one second for other peers to come online
	time.Sleep(time.Second)

    checkIfMasterIsAlive()

	transceiveElevatorHeading()
	transceiveNewExternalOrder()
	transceiveHandledExternalOrder()
	transceiveQueues()

}

func StartUDPPeersBroadcast() {

	peerUpdateCh = make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(constants.PeersPort, Id, peerTxEnable)
	go peers.Receiver(constants.PeersPort, peerUpdateCh)
}

func masterBroadcast() {
	masterTx := make(chan string)
	go bcast.Transmitter(constants.MasterPort, masterTx)
	for {
		if Master == true {
			masterTx <- Id
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func lookForChangeInPeers() {
	for {
		PeersInfo = <-peerUpdateCh

		//Coming online again
		if len(PeersInfo.New) > 0 && PeersInfo.New == Id && Online == false{
			if(testIfOnline()){
				checkIfMasterIsAlive()
			}
		}

		if(len(PeersInfo.Lost) > 0){
			handleLostElevator()
		}
	}
}

func listenForMaster(){

	masterRx = make(chan string,1)
	go bcast.Receiver(constants.MasterPort, masterRx)
}

func checkIfMasterIsAlive() {
		
		L:
		//Emptying masterbuffer first
		for {
		    select {
		    	
		    case <-masterRx:
		    	time.Sleep(time.Millisecond*1)
		    default:
		       break L
		    }
		}
		noMasterTimer := time.NewTimer(time.Millisecond * 500)
		select {
		case <-noMasterTimer.C:
			chooseMasterSlave()
			break
		case <-masterRx:
			if(Master != true){
				Master = false
				fmt.Println("Slave")
			}
		}

	
}

func chooseMasterSlave() {
	smallestID := PeersInfo.Peers[0]

	for i := 1; i < len(PeersInfo.Peers); i++ {
		if PeersInfo.Peers[i] < smallestID {
			smallestID = PeersInfo.Peers[i]
		}
	}

	if Id == smallestID {

		Master = true
		fmt.Println("Master")

	} else {
		Master = false
		fmt.Println("Slave")
	}
}



func handleLostElevator() {

	if testIfOnline() {

		checkIfMasterIsAlive()

		//Tell queue which elevators disconnected
		for i:= 0; i < len(PeersInfo.Lost); i++{
			peerDisconnectsCh <- PeersInfo.Lost[i]
		}

	} else {
		fmt.Println("Slave")
		Master = false
		//Tell queue this elevator is offline
		peerDisconnectsCh <- Id
	}

}

func testIfOnline() bool {

	localIP, err := localip.LocalIP()
	if err != nil {

		Online = false
		fmt.Println("Offline", Id)
		return false

	} else if Id == "" {

		//Set network ID if online
		Id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())

	}
	
	fmt.Println("Online", Id)
	Online = true
	return true
}



func transceiveElevatorHeading() {
	go bcast.Transmitter(constants.HeadingPort, elevatorHeadingTx)
	go bcast.Receiver(constants.HeadingPort, elevatorHeadingRx)
}

func transceiveNewExternalOrder() {
	go bcast.Transmitter(constants.NewExternalOrderPort, externalOrderTx)
	go bcast.Receiver(constants.NewExternalOrderPort, externalOrderRx)
}

func transceiveHandledExternalOrder() {
	go bcast.Transmitter(constants.HandledExternalOrderPort, handledExternalOrderTx)
	go bcast.Receiver(constants.HandledExternalOrderPort, handledExternalOrderRx)
}

func transceiveQueues() {
	go bcast.Transmitter(constants.QueuePort, queuesTx)
	go bcast.Receiver(constants.QueuePort, queuesRx)
}
