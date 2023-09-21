package pkgCtl

import (
	"context"
	"sort"
)

type Handler interface {
	Create() error
	Start() error
	Destroy() error
	Async() bool
}

type HandlerFunc func(*context.Context) Handler

type Unit struct {
	Seq    int
	Name   string
	Handle HandlerFunc
}

type DestroyUnit struct {
	Seq  int
	Name string
	Unit Handler
}

func RegisterHandler(seq int, name string, handler HandlerFunc) {
	units = append(units, Unit{
		Seq:    seq,
		Name:   name,
		Handle: handler,
	})
}

func Bootstrap(activeFunc func()) {
	if activeFunc == nil {
		panic("===== unset active func =====")
	}
	activeFunc()
	sort.Slice(units, func(i, j int) bool {
		return units[i].Seq < units[j].Seq
	})
}

func Self(ctx context.Context) (*Root, bool) {
	val, ok := ctx.Value("___PKG_CTL___").(*Root)
	return val, ok
}
