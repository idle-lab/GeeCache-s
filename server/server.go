package server

type Server interface {
	Start()
	Stop()
	Close()
}
