package main

import (
	"context"
	"github.com/RealFax/pkg-ctl"
	"github.com/RealFax/pkg-ctl/example/units"
	"log"
)

func main() {

	// register services
	pkgCtl.SetupActive(units.Active)

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
