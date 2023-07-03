package grpcTun

import (
	"fmt"
	"github.com/xtaci/smux"
	"net"
)

// Server此时是调用服务的client端
type Server struct {
	listener net.Listener
	mux      *smux.Session

	onClientConnect func(conn net.Conn)
}

type ClientSession struct {
	mux *smux.Session
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Listen(l net.Listener) *Server {
	s.listener = l
	return s
}

func (s *Server) Init() {
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				fmt.Printf("accept error: %v\n", err)
				continue
			}

			session := &ClientSession{}
			muxConf := smux.DefaultConfig()
			muxConf.Version = 2
			session.mux, err = smux.Server(conn, muxConf)
			if err != nil {
				fmt.Printf("init smux error: %v\n", err)
				continue
			}
			connToClient, err := session.mux.OpenStream()
			if err != nil {
				fmt.Printf("open conn to client error: %v\n", err)
				continue
			}
			go s.onClientConnect(connToClient)
		}
	}()
	return
}

func (s *Server) OnClientConnect(f func(conn net.Conn)) *Server {
	s.onClientConnect = f
	return s
}
