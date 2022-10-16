package pkgCtl

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Destroy(cancel context.CancelFunc) (err error) {
	cancel()
	for i := 0; i < len(destroys); i++ {
		if err = destroys[i].Unit.Destroy(); err != nil {
			log.Printf(
				"unit %s destroy fail, error: %s",
				destroys[i].Name,
				err.Error(),
			)
		}
	}
	log.Println("all unit are unmount")
	return
}

func Startup(ctx *context.Context) error {
	sortCreates()
	size := len(creates)
	if size == 0 {
		return errors.New("no unit require register")
	}
	var (
		ec  = make(chan error, 1)
		err error
	)
	for i := 0; i < size; i++ {
		if len(ec) == 1 {
			return <-ec
		}
		unit := creates[i]
		uHandler := unit.Handle(ctx)
		if err = uHandler.Create(); err != nil {
			log.Printf("unit %s create fail, error: %s", unit.Name, err.Error())
			return err
		}
		registerDestroy(unit.Seq, unit.Name, uHandler)
		if !uHandler.IsAsync() {
			if err = uHandler.Start(); err != nil {
				log.Printf("unit %s start fail, error: %s", unit.Name, err.Error())
				return err
			}
			log.Printf("unit %s startup", unit.Name)
		} else {
			go func() {
				log.Printf("[Ctl(Startup<ASYNC>)] unit %s startup", unit.Name)
				if er := uHandler.Start(); err != nil {
					log.Printf("[Ctl(Startup<ASYNC>)] unit %s start fail, error: %s", unit.Name, er.Error())
				}
			}()
		}
	}
	return nil
}

func ListenAndDestroy(cancel context.CancelFunc) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
	return Destroy(cancel)
}
