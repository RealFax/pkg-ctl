package main

import (
	"context"
	"github.com/RealFax/pkg-ctl"
	"github.com/RealFax/pkg-ctl/example/units"
	"log"
	"time"
)

func main() {

	// register services
	pkgCtl.Bootstrap(example_units.Active)

	var (
		ctx, cancel = context.WithCancel(context.Background())
		err         error
	)

	if err = pkgCtl.Startup(&ctx); err != nil {
		log.Fatal(err)
	}

	// quit halfway
	go func() {
		time.Sleep(time.Second * 3)
		if er := pkgCtl.Exit(); er != nil {
			log.Println("call pkgCtl.Exit error:", er)
		}
		log.Println("call pkgCtl.Exit success")
	}()

	// do something

	if err = pkgCtl.ListenAndDestroy(cancel); err != nil {
		log.Fatal(err)
	}

}
