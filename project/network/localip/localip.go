package localip

import (
	"net"
	"strings"
	"fmt"
)

func LocalIP() (string, error) {
	fmt.Println("1")

	select{
		case conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53}):
		break
	}
	
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
