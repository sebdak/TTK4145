package network

import (
	constants "../constants"
	bcast "./bcast"
	localip "./localip"
	peers "./peers"
	"flag"
	"fmt"
	"os"
	"time"
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

var PeersInfo peers.PeerUpdate
var Id string

func InitNetwork(newOrderChannel chan constants.Order, peerDisconnectsChannel chan string, elevatorHeadingTxChannel chan constants.ElevatorHeading, elevatorHeadingRxChannel chan constants.ElevatorHeading, queuesTxChannel chan []constants.Order, queuesRxChannel chan []constants.Order, externalOrderTxChannel chan constants.Order, externalOrderRxChannel chan constants.Order, handledExternalOrderTxChannel chan constants.Order, handledExternalOrderRxChannel chan constants.Order) {
	//Store channels for module communication
	newOrderCh = newOrderChannel
	peerDisconnectsCh = peerDisconnectsChannel

	//Store channels for module-network communication
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

	//start peers broadcast
	StartUDPPeersBroadcast()

	go masterBroadcast()
	//Update peers on network
	go lookForChangeInPeers()

	//wait one second for peers to come online
	time.Sleep(time.Second)

	checkIfMasterIsAlive()

	go transceiveElevatorHeading()
	go transceiveNewExternalOrder()
	go transceiveHandledExternalOrder()
	go transceiveQueues()

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

func lookForChangeInPeers() {
	for {
		PeersInfo = <-peerUpdateCh
		handleLostElevator()
	}
}

func transceiveQueues() {
	go bcast.Transmitter(constants.QueuePort, queuesTx)
	go bcast.Receiver(constants.QueuePort, queuesRx)
}

func chooseMasterSlave() {
	smallestID := PeersInfo.Peers[0]

	for i := 1; i < len(PeersInfo.Peers); i++ {
		if PeersInfo.Peers[i] < smallestID {
			smallestID = PeersInfo.Peers[i]
		}
	}

	if Id == smallestID {
		//you are master
		//start masterBroadcast
		Master = true
		fmt.Println("master because of smallestID")

	} else {
		Master = false
		fmt.Println("not master because of ID")
	}
}

func masterBroadcast() {
	masterTx := make(chan string)
	go bcast.Transmitter(constants.MasterPort, masterTx)
	for {
		if Master {
			masterTx <- Id
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func handleLostElevator() {
	if len(PeersInfo.Lost) > 0 {
		if testIfOnline() {
			//Elevator is still alive
			checkIfMasterIsAlive()
			peerDisconnectsCh <- PeersInfo.Lost[0]

		} else {
			fmt.Println("Elevator is off network")
			Master = false
			//dump external queue to internal
			peerDisconnectsCh <- PeersInfo.Lost[0]
		}

	}
	time.Sleep(time.Millisecond)
}

func testIfOnline() bool {
	_, err := localip.LocalIP()
	if err != nil {
		return false
	}
	return true
}

func checkIfMasterIsAlive() {
	if Master != true {

		masterRx := make(chan string)
		go bcast.Receiver(constants.MasterPort, masterRx)
		timer := time.NewTimer(time.Millisecond * 500)

		select {
		case <-timer.C:
			chooseMasterSlave()
		case <-masterRx:
			Master = false
			fmt.Println("other master on network")
		}

	}
}

func StartUDPPeersBroadcast() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	peerUpdateCh = make(chan peers.PeerUpdate)

	flag.StringVar(&Id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if Id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		Id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	peerTxEnable := make(chan bool)

	go peers.Transmitter(constants.PeersPort, Id, peerTxEnable)
	go peers.Receiver(constants.PeersPort, peerUpdateCh)
}
