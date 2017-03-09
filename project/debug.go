package main

import (
	"fmt"
)


func main() {

	var channel chan string = make(chan string)
	str := "yo"

	channel <- str

	fmt.Println(str)
}
