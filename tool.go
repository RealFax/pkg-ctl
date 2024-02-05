package pkgCtl

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
)

type Logger interface {
	Debug(msg string, v ...any)
	Info(msg string, v ...any)
	Warn(msg string, v ...any)
	Error(msg string, v ...any)
}

type Context struct {
	context.Context
	CancelFunc context.CancelFunc

	logger  Logger
	units   []DUnit
	csignal chan struct{}

	mu     sync.RWMutex
	values map[string]any
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func (c *Context) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.values, key)
}

func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, found := c.values[key]
	return val, found
}

func (c *Context) Stop() {
	if c.CancelFunc == nil {
		return
	}
	c.logger.Warn("context cancel signal")
	c.CancelFunc()
}

func (c *Context) Exit() error {
	c.Stop()
	if len(c.csignal) != 0 {
		return errors.New("exiting")
	}
	c.csignal <- struct{}{}
	return nil
}

func (c *Context) Destroy() error {
	c.logger.Debug("starting destroy unit", slog.Int("unit_count", len(c.units)))

	sort.Slice(c.units, func(i, j int) bool {
		return c.units[i].Seq > c.units[j].Seq
	})

	var err error
	for _, unit := range c.units {
		if err = unit.Unit.Destroy(); err != nil {
			c.logger.Error("unit destroy fail", slog.String("unit", unit.Name), slog.String("error", err.Error()))
			return err
		}
	}

	c.Stop()
	c.logger.Info("all unit are unmount")
	return nil
}

func (c *Context) Startup() error {
	if c.csignal == nil {
		return errors.New("unable to start pkg-ctl which has exited")
	}
	if len(units) == 0 {
		return errors.New("no unit require startup")
	}

	c.logger.Debug("starting startup unit", slog.Int("unit_count", len(units)))

	var err error
	for _, unit := range units {
		handle := unit.Handle(c)
		if err = handle.Create(); err != nil {
			c.logger.Error("unit create fail", slog.String("unit", unit.Name), slog.String("error", err.Error()))
			return err
		}

		c.units = append(c.units, DUnit{
			Seq:  unit.Seq,
			Name: unit.Name,
			Unit: handle,
		})

		if handle.Async() {
			ext := unit
			go func() {
				c.logger.Info("<Async> startup", slog.String("unit", ext.Name))
				if aErr := handle.Start(); aErr != nil {
					c.logger.Error("<Async> startup fail", slog.String("unit", ext.Name), slog.String("error", err.Error()))
					return
				}
			}()
			continue
		}

		if err = handle.Start(); err != nil {
			c.logger.Error("startup fail", slog.String("unit", unit.Name), slog.String("error", err.Error()))
			return err
		}
		c.logger.Info("startup", slog.String("unit", unit.Name))
	}

	return nil
}

func (c *Context) ListenAndDestroy() error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-c.csignal:
		return nil
	case <-sig:
		return c.Destroy()
	}
}

func New(opts ...Option) *Context {
	root := &Context{
		logger:  slog.Default(),
		units:   make([]DUnit, 0),
		csignal: make(chan struct{}, 1),
		values:  make(map[string]any),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(root)
		}
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
