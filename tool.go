package pkgCtl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
)

type Context struct {
	context.Context
	CancelFunc context.CancelFunc

	logger       Logger
	destroyUnits []DestroyUnit
	closeSignal  chan struct{}

	mu  sync.RWMutex
	env map[string]any
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.env[key] = value
}

func (c *Context) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.env, key)
}

func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, found := c.env[key]
	return val, found
}

func (c *Context) Stop() {
	if c.CancelFunc == nil {
		return
	}
	c.logger.Info("context cancel signal")
	c.CancelFunc()
}

func (c *Context) Exit() error {
	c.Stop()
	if len(c.closeSignal) != 0 {
		return errors.New("exiting")
	}
	c.closeSignal <- struct{}{}
	return nil
}

func (c *Context) Destroy() error {
	c.logger.Debug("starting destroy unit")
	sort.Slice(c.destroyUnits, func(i, j int) bool {
		return c.destroyUnits[i].Seq > c.destroyUnits[j].Seq
	})
	var err error
	for _, unit := range c.destroyUnits {
		if err = unit.Unit.Destroy(); err != nil {
			c.logger.Error(fmt.Sprintf("unit %s destroy fail, error: %s", unit.Name, err.Error()))
			return err
		}
	}
	c.Stop()
	c.logger.Info("all unit are unmount")
	return nil
}

func (c *Context) Startup() error {
	if c.closeSignal == nil {
		return errors.New("unable to start pkg-ctl which has exited")
	}
	if len(units) == 0 {
		return errors.New("no unit require startup")
	}

	var err error
	for _, unit := range units {
		handle := unit.Handle(c)
		if err = handle.Create(); err != nil {
			c.logger.Error(fmt.Sprintf("unit %s create fail, error: %s", unit.Name, err.Error()))
			return err
		}

		c.destroyUnits = append(c.destroyUnits, DestroyUnit{
			Seq:  unit.Seq,
			Name: unit.Name,
			Unit: handle,
		})

		if handle.Async() {
			ext := unit
			go func() {
				c.logger.Info(fmt.Sprintf("[Ctl(Startup<ASYNC>)] unit %s startup", ext.Name))
				if aErr := handle.Start(); aErr != nil {
					c.logger.Error(fmt.Sprintf(
						"[Ctl(Startup<ASYNC>)] unit %s start fail, error: %s",
						ext.Name,
						aErr.Error(),
					))
					return
				}
			}()
			continue
		}

		if err = handle.Start(); err != nil {
			c.logger.Error(fmt.Sprintf("unit %s start fail, error: %s", unit.Name, err.Error()))
			return err
		}
		c.logger.Info(fmt.Sprintf("unit %s startup", unit.Name))
	}

	return nil
}

func (c *Context) ListenAndDestroy() error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-c.closeSignal:
		return nil
	case <-sig:
		return c.Destroy()
	}
}

func New(logger Logger) *Context {
	if logger == nil {
		logger = DefaultLogger
	}

	root := &Context{
		logger:       logger,
		destroyUnits: make([]DestroyUnit, 0),
		closeSignal:  make(chan struct{}, 1),
		env:          make(map[string]any),
	}

	root.Context, root.CancelFunc = context.WithCancel(context.Background())

	return root
}

func Use(ctx context.Context) (*Context, bool) {
	c, ok := ctx.(*Context)
	return c, ok
}

func UseValue[T any](ctx *Context, key string) (T, bool) {
	val, found := ctx.Get(key)
	if !found {
		var zero T
		return zero, false
	}

	av, ok := val.(T)
	return av, ok
}
