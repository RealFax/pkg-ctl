package pkgCtl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
)

type Root struct {
	Context    context.Context
	CancelFunc context.CancelFunc

	logger       Logger
	destroyUnits []DestroyUnit
	closeSignal  chan struct{}
}

func (r *Root) Stop() {
	if r.CancelFunc == nil {
		return
	}
	r.logger.Info("context cancel signal")
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
	r.logger.Debug("starting destroy unit")
	sort.Slice(r.destroyUnits, func(i, j int) bool {
		return r.destroyUnits[i].Seq > r.destroyUnits[j].Seq
	})
	var err error
	for _, unit := range r.destroyUnits {
		if err = unit.Unit.Destroy(); err != nil {
			r.logger.Error(fmt.Sprintf("unit %s destroy fail, error: %s", unit.Name, err.Error()))
			return err
		}
	}
	r.Stop()
	r.logger.Info("all unit are unmount")
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
			r.logger.Error(fmt.Sprintf("unit %s create fail, error: %s", unit.Name, err.Error()))
			return err
		}

		r.destroyUnits = append(r.destroyUnits, DestroyUnit{
			Seq:  unit.Seq,
			Name: unit.Name,
			Unit: handle,
		})

		if handle.Async() {
			go func() {
				r.logger.Info(fmt.Sprintf("[Ctl(Startup<ASYNC>)] unit %s startup", unit.Name))
				if aErr := handle.Start(); aErr != nil {
					r.logger.Error(fmt.Sprintf(
						"[Ctl(Startup<ASYNC>)] unit %s start fail, error: %s",
						unit.Name,
						aErr.Error(),
					))
					return
				}
			}()
			continue
		}

		if err = handle.Start(); err != nil {
			r.logger.Error(fmt.Sprintf("unit %s start fail, error: %s", unit.Name, err.Error()))
			return err
		}
		r.logger.Info("unit %s startup", unit.Name)
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

func New(ctx context.Context, logger Logger) *Root {
	if ctx == nil {
		ctx = context.Background()
	}

	root := &Root{
		logger:       logger,
		destroyUnits: make([]DestroyUnit, 0),
		closeSignal:  make(chan struct{}, 1),
	}

	ctx = context.WithValue(ctx, "___PKG_CTL___", root)

	root.Context, root.CancelFunc = context.WithCancel(ctx)

	return root
}
