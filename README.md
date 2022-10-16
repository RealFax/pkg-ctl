# Package controller
 - a simple golang task controller

## how to use  

 _using pkg-ctl is very simple_
 
 - First add pkg-ctl to your `go.mod`
 - Then refer to the example code to register your package in `pkg-ctl`


## example

__main.go__

```go
package main

import (
	"context"
	"github.com/RealFax/pkg-ctl"
	"log"
)

func main() {

	var (
		ctx, cancel = context.WithCancel(context.Background())
		err         error
	)

	if err = pkgCtl.Startup(&ctx); err != nil {
		log.Fatal(err)
	}
	
    // do something

	if err = pkgCtl.ListenAndDestroy(cancel); err != nil {
        log.Fatal(err)
	}
	
}
```

__service.go__

```go
package service

import (
	"context"
	"github.com/RealFax/pkg-ctl"
	"net"
)

type service struct {
	listener net.Listener
}

func NewService(ctx *context.Context) pkgCtl.Handler {
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
		conn.Write([]byte("hello, pkg-ctl"))
		conn.Close()
	}
}

func (s *service) Destroy() error {
	return s.listener.Close()
}

func (s *service) IsAsync() bool {
	return true
}

func init()  {
    pkgCtl.RegisterHandler(1, "tcp-listener", NewService)
}
```