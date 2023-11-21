package pkgCtl_test

import (
	pkgCtl "github.com/RealFax/pkg-ctl"
)

var ctx = pkgCtl.New(nil)

func ExampleUse() {
	c, ok := pkgCtl.Use(ctx)
	if !ok {
		// not pkgCtl.Context
		return
	}

	// call pkgCtl.Context api
	c.Get("Key")
}

func ExampleRegisterHandler() {
	pkgCtl.RegisterHandler(-1, "test-loader", func(c *pkgCtl.Context) pkgCtl.Handler {
		c.Set("ctx1", false)
		return nil
	})
}

func ExampleUseValue() {
	c, ok := pkgCtl.Use(ctx)
	if !ok {
		return
	}

	c.Set("K", "V")

	catch, ok := pkgCtl.UseValue[string](c, "K")
	if !ok {
		return
	}

	if catch != "V" {
		// fail
	}
	// success
}
