// Go 1.2
// go run hello-world.go

package main

import (
	. "fmt"
	"runtime"
	"time"
)

var count int = 0

func someGoroutine(ch chan int) {
	Println("Hello from a goroutine!")

	for i := 0; i < 1000000; i++ {
		val := <-ch
		val++
		ch <- val
	}
}

func someGoroutine2(ch chan int) {
	Println("Hello from a goroutine!")

	for i := 0; i < 1000000; i++ {
		val := <-ch
		val--
		ch <- val
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // I guess this is a hint to what GOMAXPROCS does...
	// Try doing the exercise both with and without it!
	ch := make(chan int, 1)
	ch <- count

	//Println(ch)

	go someGoroutine(ch) // This spawns someGoroutine() as a goroutine
	go someGoroutine2(ch)

	// We have no way to wait for the completion of a goroutine (without additional syncronization of some sort)
	// We'll come back to using channels in Exercise 2. For now: Sleep.
	time.Sleep(500 * time.Millisecond)

	count = <-ch

	Println("Count equals: ", count)
	Println("Hello from main!")
}
