package pkgCtl

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ForceExit the ListenAndDestroy func
func ForceExit() error {
	if len(closeListener) != 0 {
		return errors.New("exiting")
	}
	closeListener <- struct{}{}
	return nil
}

// Exit unregister all services immediately after calling
//
// this func can only be called when ListenAndDestroy is used, otherwise an error will be return
func Exit() error {
	if cancelFunc == nil {
		return errors.New("unset cancelFunc")
	}
	return Destroy(cancelFunc)
}

func ExitWithTimeout(d time.Duration) error {
	err := make(chan error, 1)
	ticker := time.NewTicker(d)
	go func() {
		select {
		case <-ticker.C:
			ticker.Stop()
			log.Println("exit timeout! calling ForceExit")
			err <- ForceExit()
		}
	}()
	go func() {
		err <- Exit()
		ticker.Stop()
	}()
	return <-err
}

// Destroy unregister all services immediately after calling
func Destroy(cancel context.CancelFunc) (err error) {
	cancel()
	for i := 0; i < len(destroyUnits); i++ {
		if err = destroyUnits[i].Unit.Destroy(); err != nil {
			Log.Printf(
				"unit %s destroy fail, error: %s",
				destroyUnits[i].Name,
				err.Error(),
			)
		}
	}
	Log.Println("all unit are unmount")
	closeListener <- struct{}{}
	return
}

// Startup all registered services in order
func Startup(rootCtx *context.Context) error {
	if len(units) == 0 {
		return errors.New("no unit require register")
	}
	var (
		errorChan = make(chan error, 1)
		err       error
	)
	for _, unit := range units {
		if len(errorChan) == 1 {
			return <-errorChan
		}
		handler := unit.Handle(rootCtx)
		if err = handler.Create(); err != nil {
			Log.Printf("unit %s create fail, error: %s", unit.Name, err.Error())
			return err
		}
		registerDestroy(unit.Seq, unit.Name, handler)
		if handler.IsAsync() {
			go func() {
				Log.Printf("[Ctl(Startup<ASYNC>)] unit %s startup", unit.Name)
				if hErr := handler.Start(); hErr != nil {
					Log.Printf("[Ctl(Startup<ASYNC>)] unit %s start fail, error: %s", unit.Name, hErr.Error())
					return
				}
			}()
			continue
		}
		if err = handler.Start(); err != nil {
			Log.Printf("unit %s start fail, error: %s", unit.Name, err.Error())
			return err
		}
		Log.Printf("unit %s startup", unit.Name)
	}
	return nil
}

func ListenAndDestroy(cancel context.CancelFunc) error {
	cancelFunc = cancel
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-closeListener:
		return nil
	case <-c:
		return Destroy(cancel)
	}
}
