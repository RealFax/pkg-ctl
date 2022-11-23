package pkgCtl

import (
	"context"
	"log"
)

var (
	activeFunc    = func() {}
	closeListener = make(chan struct{}, 1)
	cancelFunc    context.CancelFunc
	creates       = make([]Unit, 0)
	destroys      = make([]DestroyUnit, 0)
)

var Log *log.Logger

func init() {
	Log = log.Default()
}
