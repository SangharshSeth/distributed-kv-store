package server

import "net"

const (
	address = "localhost:9000"
)

type TCPServer struct {
	address string
}

func NewTCPServer(address string) *TCPServer {
	return &TCPServer{
		address: address,
	}
}

func (s *TCPServer) Listen() (net.Listener, error) {

	tcpListner, err := net.Listen("tcp", s.address)
	return tcpListner, err
}
