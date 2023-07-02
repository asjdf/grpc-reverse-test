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

	onConnect func(conn net.Conn) error
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Listen(l net.Listener) *Server {
	s.listener = l
	return s
}

func (s *Server) Init() (err error) {
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				fmt.Printf("accept error: %v\n", err)
				continue
			}
			muxConf := smux.DefaultConfig()
			muxConf.Version = 2
			s.mux, err = smux.Server(conn, muxConf)
			if err != nil {
				fmt.Printf("init smux error: %v\n", err)
				continue
			}
			stream, err := s.mux.OpenStream()
			if err != nil {
				fmt.Printf("open stream error: %v\n", err)
				continue
			}
			go s.onConnect(stream)
		}
	}()
	return nil
}

func (s *Server) OnConnect(f func(conn net.Conn) error) *Server {
	s.onConnect = f
	return s
}
