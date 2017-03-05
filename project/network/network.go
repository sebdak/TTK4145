package main

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

var peerUpdateCh chan peers.PeerUpdate

type HelloMsg struct {
	Message string
	Iter    int
}

func main(){
	DebugAlive()

}
func DebugAlive(){
	for{
		testIfOnline()
		time.Sleep(time.Second * 2)
	}
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


func lookForLostElevator(){
	for {
		select {

		case p := <-peerUpdateCh:
			if(len(p.Lost) > 0){
				if(testIfOnline()){
					//Elevator is still alive
				} else{
					//Elevator is off network
				}
			}

		}
		
		time.Sleep(time.Millisecond)
	}
}

func testIfOnline() bool {
	_, err := localip.LocalIP()
		if err != nil {
			return false
		}
	return true
}

func StartUDPBroadcast(helloTx chan HelloMsg, helloRx chan HelloMsg, peerUpdateCh chan peers.PeerUpdate) {
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

	// The example message. We just send one of these every second.

	/*
		go func() {
			helloMsg := "I'm alive " + id
			for {
				//helloMsg.Iter++
				helloTx <- helloMsg
				time.Sleep(5 * time.Second)
			}
		}()
	*/

	fmt.Println("Started")
	/*
	for {
		
			select {
			case p := <-peerUpdateCh:
				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", p.Peers)
				fmt.Printf("  New:      %q\n", p.New)
				fmt.Printf("  Lost:     %q\n", p.Lost)

					case a := <-helloRx:
						fmt.Printf("Received: %#v\n", a)

			}
		
	}
	*/
}