package proxy

import (
	"log"
	"net"
	"time"
)

func connCloseWithLog(conn net.Conn) {
	if err := conn.Close(); err != nil {
		log.Printf("proxy: handler: conn close err: %v", err)
	}
}

func DefaultDialer() *net.Dialer {
	d := &net.Dialer{
		Timeout: time.Second * 15,
	}
	return d
}
