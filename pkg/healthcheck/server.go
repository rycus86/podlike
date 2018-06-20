package healthcheck

import (
	"fmt"
	"net"
)

var (
	networkType    = "unix"
	networkAddress = "/.podlike.health.sock"
)

type Server struct {
	srv    net.Listener
	closed bool
}

func Serve() (*Server, error) {
	srv, err := net.Listen(networkType, networkAddress)
	if err != nil {
		return nil, err
	}

	server := &Server{
		srv: srv,
	}

	go server.handleRequests()

	return server, nil
}

func (s *Server) handleRequests() {
	for {
		if s.closed {
			return
		}

		conn, err := s.srv.Accept()
		if err != nil {
			if s.closed {
				return
			}

			fmt.Println("Failed to accept an incoming healthcheck request:", err)
			continue
		}

		conn.Write([]byte(State()))
		conn.Close()
	}
}

func (s *Server) Close() error {
	s.closed = true

	return s.srv.Close()
}
