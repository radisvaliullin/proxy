package proxy

import (
	"log"
	"net"
)

func connCloseWithLog(conn net.Conn) {
	if err := conn.Close(); err != nil {
		log.Printf("proxy: handler: conn close err: %v", err)
	}
}
