package server

type Handler interface {
	Process()
	Close()
}
