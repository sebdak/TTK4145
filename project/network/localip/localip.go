package localip

import (
	"net"
	"strings"
	"time"
)

func LocalIP() (string, error) {

	conn, err := net.DialTimeout("tcp4", "8.8.8.8:53", time.Second)

	if err != nil {
		return "", err
	}

	defer conn.Close()
	localIP := strings.Split(conn.LocalAddr().String(), ":")[0]

	return localIP, nil
}
