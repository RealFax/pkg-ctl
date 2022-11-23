package service

import (
	"context"
	"github.com/RealFax/pkg-ctl"
	"net"
)

type service struct {
	listener net.Listener
}

func NewTCPListener(ctx *context.Context) pkgCtl.Handler {
	return &service{}
}

func (s *service) Create() (err error) {
	s.listener, err = net.Listen("tcp", "127.0.0.1:12345")
	return err
}

func (s *service) Start() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			c.Write([]byte("hello, pkg-ctl"))
			c.Close()
		}(conn)
	}
}

func (s *service) Destroy() error {
	return s.listener.Close()
}

func (s *service) IsAsync() bool {
	return true
}
