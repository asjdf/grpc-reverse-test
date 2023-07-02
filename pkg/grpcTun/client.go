package grpcTun

import (
	"github.com/xtaci/smux"
	"net"
)

var _ net.Listener = (*Client)(nil)

type Client struct {
	sConn net.Conn
	mux   *smux.Session
}

func (c *Client) Accept() (net.Conn, error) {
	return c.mux.AcceptStream()
}

func (c *Client) Close() error {
	c.mux.Close()
	c.sConn.Close()
	return nil
}

func (c *Client) Addr() net.Addr {
	return c.mux.LocalAddr()
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Conn(conn net.Conn) *Client {
	c.sConn = conn
	return c
}

func (c *Client) Init() (err error) {
	muxConf := smux.DefaultConfig()
	muxConf.Version = 2
	c.mux, err = smux.Client(c.sConn, muxConf)
	return
}
