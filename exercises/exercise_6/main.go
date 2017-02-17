package main

import (
	network "./Netw"
	"fmt"
	"os/exec"
	"time"
)

func main() {

	//broadcast which state you are in
	helloTx := make(chan string)
	helloRx := make(chan string)

	go network.StartUDPBroadcast(helloTx, helloRx)

	go func() {
		iter := 1
		for {
			helloTx <- "Test"
			time.Sleep(10 * time.Second)
			iter++
		}
	}()

	for {
		select {
		case rx := <-helloRx:
			fmt.Println("THIS MESSAGE WAS RECEIVED: " + rx)
		}
	}

}

func spawnBackup() {
	fmt.Println("Spawning backup ")
	out, err := exec.Command("gnome-terminal", "-x", "main.go").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
}
