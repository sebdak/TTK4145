// Go 1.2
// go run helloworld_go.go

package main

import (
	. "fmt"
	"runtime"
	"time"
)

var count int = 0

func someGoroutine() {
	Println("Hello from a goroutine!")
	for i := 0; i < 1000000; i++ {
		count++
	}
}

func someGoroutine2() {
	Println("Hello from a goroutine!")
	for i := 0; i < 1000000; i++ {
		count--
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // I guess this is a hint to what GOMAXPROCS does...
	// Try doing the exercise both with and without it!
	go someGoroutine() // This spawns someGoroutine() as a goroutine
	go someGoroutine2()

	// We have no way to wait for the completion of a goroutine (without additional syncronization of some sort)
	// We'll come back to using channels in Exercise 2. For now: Sleep.
	time.Sleep(100 * time.Millisecond)
	Println("Count equals: ", count)
	Println("Hello from main!")
}
