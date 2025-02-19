package server

import (
	"net"
	"time"
)

type TCPServer struct {
	Addr     net.TCPAddr
	Listener net.Listener
	Conns    []Handler

	is_running bool
}

func NewTCPServer(ip string, port uint16) *TCPServer {
	addr := net.TCPAddr{
		IP:   net.IP(ip),
		Port: int(port),
	}

	return &TCPServer{
		Addr: addr,
	}
}

func (t *TCPServer) dispatch() {
	for t.is_running {
		conn, err := t.Listener.Accept()
		if err != nil {
			panic(err)
		}

		handler := NewTCPHandler(conn)
		go handler.Process()
	}
}

func (t *TCPServer) Start() {
	t.is_running = true
	if t.Listener == nil {
		listener, err := net.Listen("tcp", t.Addr.String())
		if err != nil {
			panic(err)
		}
		t.Listener = listener
	}

	go t.dispatch()

	for t.is_running {
		time.Sleep(100 * time.Microsecond)

		// do something
	}
}

func (t *TCPServer) Stop() {
	t.is_running = false
}
