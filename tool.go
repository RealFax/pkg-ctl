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
//
// Deprecated
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
//
// Deprecated
func Exit() error {
	if cancelFunc == nil {
		return errors.New("unset cancelFunc")
	}
	return Destroy(cancelFunc)
}

// ExitWithTimeout
//
// Deprecated
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
//
// Deprecated
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
//
// Deprecated
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
		if handler.Async() {
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

// ListenAndDestroy
//
// Deprecated
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

type Root struct {
	Context    context.Context
	CancelFunc context.CancelFunc

	destroyUnits []DestroyUnit
	closeSignal  chan struct{}
}

func (r *Root) Stop() {
	if r.CancelFunc == nil {
		return
	}
	r.CancelFunc()
}

func (r *Root) Exit() error {
	r.Stop()
	if len(r.closeSignal) != 0 {
		return errors.New("exiting")
	}
	r.closeSignal <- struct{}{}
	return nil
}

func (r *Root) Destroy() error {
	r.Stop()
	var err error
	for _, unit := range destroyUnits {
		if err = unit.Unit.Destroy(); err != nil {
			Log.Printf("unit %s destroy fail, error: %s", unit.Name, err.Error())
			return err
		}
	}
	Log.Println("all unit are unmount")
	return nil
}

func (r *Root) Startup() error {
	if r.closeSignal == nil {
		return errors.New("unable to start pkg-ctl which has exited")
	}
	if len(units) == 0 {
		return errors.New("no unit require startup")
	}

	var err error
	for _, unit := range units {
		handle := unit.Handle(&r.Context)
		if err = handle.Create(); err != nil {
			Log.Printf("unit %s create fail, error: %s", unit.Name, err.Error())
			return err
		}

		r.destroyUnits = append(r.destroyUnits, DestroyUnit{
			Seq:  unit.Seq,
			Name: unit.Name,
			Unit: handle,
		})

		if handle.Async() {
			go func() {
				Log.Printf("[Ctl(Startup<ASYNC>)] unit %s startup", unit.Name)
				if aErr := handle.Start(); aErr != nil {
					Log.Printf("[Ctl(Startup<ASYNC>)] unit %s start fail, error: %s", unit.Name, aErr.Error())
					return
				}
			}()
			continue
		}

		if err = handle.Start(); err != nil {
			Log.Printf("unit %s start fail, error: %s", unit.Name, err.Error())
			return err
		}
		Log.Printf("unit %s startup", unit.Name)
	}

	return nil
}

func (r *Root) ListenAndDestroy() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-r.closeSignal:
		return nil
	case <-c:
		return r.Destroy()
	}
}

func New(ctx context.Context) *Root {
	if ctx == nil {
		ctx = context.Background()
	}

	root := &Root{
		destroyUnits: make([]DestroyUnit, 0),
		closeSignal:  make(chan struct{}, 1),
	}

	ctx = context.WithValue(ctx, "___PKG_CTL___", root)

	root.Context, root.CancelFunc = context.WithCancel(ctx)

	return root
}
