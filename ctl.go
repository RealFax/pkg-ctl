package pkgCtl

import (
	"sort"
)

var (
	units = make([]Unit, 0, 16)
)

type (
	Handler interface {
		mustEmbedUnimplemented()
		Create() error
		Start() error
		Destroy() error
		Async() bool
	}
	HandlerFunc func(*Context) Handler
	Unit        struct {
		Seq    int
		Name   string
		Handle HandlerFunc
	}
	DUnit struct {
		Seq  int
		Name string
		Unit Handler
	}
	UnimplementedHandler struct{}
)

func (h UnimplementedHandler) mustEmbedUnimplemented() {}
func (h UnimplementedHandler) Create() error           { panic("Unimplemented Create") }
func (h UnimplementedHandler) Start() error            { panic("Unimplemented Create") }
func (h UnimplementedHandler) Destroy() error          { panic("Unimplemented Create") }
func (h UnimplementedHandler) Async() bool             { panic("Unimplemented Create") }

func Register(seq int, name string, handler HandlerFunc) {
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
