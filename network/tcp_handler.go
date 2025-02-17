package network

import (
	"net"
)

type TCPHandler struct {
	Conn net.Conn
	Buf  []byte
}

func NewTCPHandler(conn net.Conn) *TCPHandler {
	return &TCPHandler{
		Conn: conn,
	}
}

func (h *TCPHandler) handshake() bool {
	return true
}

func (h *TCPHandler) Process() {

}
