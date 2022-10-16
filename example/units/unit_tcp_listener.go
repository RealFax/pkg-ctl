package units

import (
	"github.com/RealFax/pkg-ctl"
	"github.com/RealFax/pkg-ctl/example/service"
)

func init() {
	pkgCtl.RegisterHandler(1, "tcp-listener", service.NewTCPListener)
}
