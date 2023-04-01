package pkgCtl

import (
	"context"
	"log"
)

var (
	// Deprecated
	closeListener = make(chan struct{}, 1)
	// Deprecated
	cancelFunc   context.CancelFunc
	units        = make([]Unit, 0)
	destroyUnits = make([]DestroyUnit, 0)
)

var Log *log.Logger

func init() {
	Log = log.Default()
}
