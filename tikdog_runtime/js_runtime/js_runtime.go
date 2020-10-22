package js_runtime

import "github.com/dop251/goja"

type FunctionCall = goja.FunctionCall
type Value = goja.Value
type ConstructorCall = goja.ConstructorCall
type Object = goja.Object

type Function func(call goja.FunctionCall) goja.Value
type Class func(ConstructorCall) *Object

func New() *Runtime {
	return &Runtime{Runtime: goja.New()}
}

type Runtime struct {
	*goja.Runtime
}

func (t *Runtime) ExportNameTo(name string, target interface{}) error {
	switch val := t.Get(name); val {
	case goja.Undefined(), goja.NaN(), goja.Null():
		return nil
	default:
		return t.ExportTo(val, target)
	}
}
