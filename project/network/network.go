package main

import (
	bcast "./bcast"
	localip "./localip"
	peers "./peers"
	constants "../constants"
	queue "../queue"
	"flag"
	"fmt"
	"os"
	"time"

)

var peerUpdateCh chan peers.PeerUpdate
var newExternalOrderCh chan constants.Order 


var p peers.PeerUpdate
var id string


func main(){
	//StartUDPPeersBroadcast()
	InitNetwork(make(chan constants.Order))
}

func InitNetwork(newExternalOrderChannel chan constants.Order){
	//Store channels
	newExternalOrderCh = newExternalOrderChannel

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


	go lookForNewExternalOrder()

	var foo chan int = make(chan int)
	<- foo
}

func lookForChangeInPeers() {
	for {
		p = <- peerUpdateCh
		lookForLostElevator()	
	}
}

func checkIfMasterIsAlive() {
	if(queue.Master != true){


		masterRx := make(chan string)
		go bcast.Receiver(constants.MasterPort, masterRx)
		timer := time.NewTimer(time.Millisecond * 500)
		
		select {
		case <- timer.C:
		chooseMasterSlave()
		case <- masterRx:
			queue.Master = false
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

	if id == smallestID {
		//you are master
		//start masterBroadcast
		queue.Master = true
		fmt.Println("master because of smallestID")
		
	} else {
		queue.Master = false
		fmt.Println("not master because of ID")
	}
}

func masterBroadcast() {
	masterTx := make(chan string)
	go bcast.Transmitter(constants.MasterPort, masterTx)
	for {
		if (queue.Master) {
	 		masterTx <- id
		}
		time.Sleep(time.Millisecond * 50)		
	}

}

func lookForNewExternalOrder(){
	for{
		order := <- newExternalOrderCh
		if(queue.CheckIfNewOrder(order)){
			//broadcast order
		}
		time.Sleep(time.Millisecond)
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
	
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	peerTxEnable := make(chan bool)
	go peers.Transmitter(constants.PeersPort, id, peerTxEnable)
	go peers.Receiver(constants.PeersPort, peerUpdateCh)

}

func StartUDPOrderBroadcast(helloTx chan constants.Order, helloRx chan constants.Order, peerUpdateCh chan peers.PeerUpdate) {

		// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	//peerUpdateCh := make(chan peers.PeerUpdate) COMMENTED OUT
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(constants.PeersPort, id, peerTxEnable)
	go peers.Receiver(constants.PeersPort, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	//helloTx := make(chan HelloMsg)
	//helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(constants.MessagePort, helloTx)
	go bcast.Receiver(constants.MessagePort, helloRx)



}