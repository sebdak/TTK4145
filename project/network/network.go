package network

import (
	bcast "./bcast"
	localip "./localip"
	peers "./peers"
	constants "../constants"
	"flag"
	"fmt"
	"os"
	"time"

)

var Master bool = false

var peerUpdateCh chan peers.PeerUpdate
var newOrderCh chan constants.Order 
var newExternalOrderCh chan constants.Order

var elevatorHeadingTx chan constants.ElevatorHeading
var elevatorHeadingRx chan constants.ElevatorHeading

var p peers.PeerUpdate
var Id string

/*
func main(){
	//StartUDPPeersBroadcast()
	InitNetwork(make(chan constants.Order))
}
*/

func InitNetwork(newOrderChannel chan constants.Order, newExternalOrderChannel chan constants.Order, elevatorHeadingTxChannel chan constants.ElevatorHeading , elevatorHeadingRxChannel chan constants.ElevatorHeading){
	//Store channels
	newOrderCh = newOrderChannel
	newExternalOrderCh = newExternalOrderChannel
	elevatorHeadingTx = elevatorHeadingTxChannel
	elevatorHeadingRx = elevatorHeadingRxChannel

	//Tries to go online 
	for(!testIfOnline()){
		fmt.Println("Not online, trying to reconnect")
		time.Sleep(time.Second*3)
	}

	//start peers broadcast
	StartUDPPeersBroadcast()

	go masterBroadcast()
	//Update peers on network 
	go lookForChangeInPeers()

	//wait one second for peers to come online
	time.Sleep(time.Second)

	checkIfMasterIsAlive()

	go transmitNewExternalOrders()

	go transceiveElevatorHeading()

}

func transceiveElevatorHeading() {
	go bcast.Transmitter(constants.HeadingPort, elevatorHeadingTx)
	go bcast.Receiver(constants.HeadingPort, elevatorHeadingRx)
}

func lookForChangeInPeers() {
	for {
		p = <- peerUpdateCh
		lookForLostElevator()	
	}
}

func checkIfMasterIsAlive() {
	if(Master != true){


		masterRx := make(chan string)
		go bcast.Receiver(constants.MasterPort, masterRx)
		timer := time.NewTimer(time.Millisecond * 500)
		
		select {
		case <- timer.C:
		chooseMasterSlave()
		case <- masterRx:
			Master = false
			fmt.Println("other master on network")
		}

	}
}

func chooseMasterSlave() {
	smallestID := p.Peers[0]

	for i := 1; i < len(p.Peers); i++ {
		if p.Peers[i] < smallestID {
			smallestID = p.Peers[i]
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
		if (Master) {
	 		masterTx <- Id
		}
		time.Sleep(time.Millisecond * 50)		
	}

}

func transmitNewExternalOrders(){
	orderTx := make(chan constants.Order)
	go bcast.Transmitter(constants.OrderPort, orderTx)

	for{
		order := <- newExternalOrderCh
		orderTx <- order
		time.Sleep(time.Millisecond)
	}
}

func lookForOrderFromNetwork(){
	orderRx := make(chan constants.Order)
	go bcast.Receiver(constants.OrderPort, orderRx)

	for{
		order := <- orderRx
		newOrderCh <- order
	}
}

/*
type HelloMsg struct {
	Message string
	Iter    int
}


func DebugSendMessage(){
	master := true
	//counter := 0
	//broadcast which state you are in
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	peerUpdateCh := make(chan peers.PeerUpdate)

	
	go StartUDPBroadcast(helloTx, helloRx, peerUpdateCh)
	iter := 0
	go func() {
		for {
			if master {
				iter++
				MsgSend := HelloMsg{Message: "Hello", Iter: iter}
				helloTx <- MsgSend
				time.Sleep(4 * time.Second)
			}
		}
	}()

	for {

		select {

		case rx := <-helloRx:

			message := rx.Message
			iter = rx.Iter
			fmt.Println("RECEIVED: ", iter, message)

		}

	}
}
*/


func lookForLostElevator(){
	if(len(p.Lost) > 0){
		if(testIfOnline()){
			//Elevator is still alive
			checkIfMasterIsAlive()
		} else{
			fmt.Println("Elevator is off network")
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

func StartUDPPeersBroadcast(){
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
