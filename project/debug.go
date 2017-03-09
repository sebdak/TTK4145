package main

import (
	"fmt"
)


func main() {
	k := 6
	for i := 0; i < k; i++ {
		fmt.Println("i: ", i)
		k--
	}
}
