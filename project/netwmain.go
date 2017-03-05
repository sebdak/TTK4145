package main

import (
	network "./Netw"
	peers "./Netw/network/peers"
	"fmt"
	"os/exec"
	"time"
)

func main() {
	master := true
	//counter := 0
	//broadcast which state you are in
	helloTx := make(chan network.HelloMsg)
	helloRx := make(chan network.HelloMsg)
	peerUpdateChan := make(chan peers.PeerUpdate)

	//exec.Command("gnome-terminal", "-x", "echo", "helloworld").Output()
	go network.StartUDPBroadcast(helloTx, helloRx, peerUpdateChan)

	timer := time.NewTimer(time.Second * 5)
	select {
	case <-helloRx:
		master = false
		timer.Stop()

	case <-timer.C:
		master = true
		spawnBackup()
	}
	iter := 0
	go func() {
		for {
			if master {
				iter++
				MsgSend := network.HelloMsg{Message: "Hello", Iter: iter}
				helloTx <- MsgSend
				time.Sleep(4 * time.Second)
			}
		}
	}()

	//var p peers.PeerUpdate

	for {

		timer := time.NewTimer(time.Second * 5)

		select {

		//case p = <-peerUpdateChan:

		case rx := <-helloRx:
			/*
				if len(p.Peers) < 2 {
					fmt.Println("Spawning backup")
					spawnBackup()
				}
			*/

			iter = rx.Iter
			fmt.Println("RECEIVED: ", iter)
			timer.Stop()

		case <-timer.C:
			fmt.Println("I am master")
			master = true
			spawnBackup()

		}
	}
}

func spawnBackup() {
	fmt.Println("Spawning backup ")
	out, err := exec.Command("gnome-terminal", "-x", "go", "run", "netwmain.go").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
}
