package main

import (
	"fmt"
)


func main() {

	str := "yo"
 	messages := make(chan string, 1)
    


    messages <- str

    fmt.Println(str)


}

