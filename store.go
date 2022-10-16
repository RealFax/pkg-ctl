package pkgCtl

import "log"

var (
	creates  = make([]Unit, 0)
	destroys = make([]DestroyUnit, 0)
)

var Log *log.Logger

func init() {
	Log = log.Default()
}
