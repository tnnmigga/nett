package idef

import (
	"reflect"
)

type IModule interface {
	Name() ModName
	Assign(any)
	MQ() chan any
	Run()
	Stop()
	RegisterHandler(mType reflect.Type, handler any)
	Before(state ServerState, hook func() error)
	After(state ServerState, hook func() error)
	Hook(state ServerState, stage int) []func() error
	Async(f func() (any, error), cb func(any, error))
}
