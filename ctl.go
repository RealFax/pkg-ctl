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
	destroys = append(destroys, DestroyUnit{
		Seq:  seq,
		Name: name,
		Unit: unit,
	})
}

func SetupActive(fn func()) {
	activeFunc = fn
}

func RegisterHandler(seq int, name string, handler func(ctx *context.Context) Handler) {
	creates = append(creates, Unit{
		Seq:    seq,
		Name:   name,
		Handle: handler,
	})
}

func sortCreates() {
	sort.Slice(creates, func(i, j int) bool {
		return creates[i].Seq < creates[j].Seq
	})
}

func init() {
	activeFunc()
}
