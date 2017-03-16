package localip

import (
	"net"
	"strings"
	"fmt"
	"time"
)

func LocalIP() (string, error) {
	fmt.Println("1")

	//conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53}):

	conn, err := net.DialTimeout("tcp4", []byte{8, 8, 8, 8}, time.Second)
	
	fmt.Println("11")
	if err != nil {
		return "", err
	}
	fmt.Println("2")
	defer conn.Close()
	localIP := strings.Split(conn.LocalAddr().String(), ":")[0]
	fmt.Println("3")
	return localIP, nil
}
