package pkgCtl

import (
	"context"
	"sort"
)

type Handler interface {
	Create() error
	Start() error
	Destroy() error
	IsAsync() bool
}

type Unit struct {
	Seq    int
	Name   string
	Handle func(ctx *context.Context) Handler
}

type DestroyUnit struct {
	Seq  int
	Name string
	Unit Handler
}

func registerDestroy(seq int, name string, unit Handler) {
	destroyUnits = append(destroyUnits, DestroyUnit{
		Seq:  seq,
		Name: name,
		Unit: unit,
	})
}

func RegisterHandler(seq int, name string, handler func(ctx *context.Context) Handler) {
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
