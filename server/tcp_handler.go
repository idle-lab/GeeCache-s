package server

import (
	"log"
	"net"
)

type TCPHandler struct {
	Conn net.Conn
}

func NewTCPHandler(conn net.Conn) *TCPHandler {
	return &TCPHandler{
		Conn: conn,
	}
}

func (t *TCPHandler) handshake() bool {
	return true
}

func (t *TCPHandler) Process() {
	defer func() {
		_ = t.Conn.Close()
	}()
	if !t.handshake() {
		log.Printf("handshake failed with connection[%s]", t.Conn.RemoteAddr().String())
		return
	}

	for {

	}
}

func (t *TCPHandler) Close() {
	t.Conn.Close()
}
