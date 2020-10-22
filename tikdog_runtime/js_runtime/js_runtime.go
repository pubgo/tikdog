package js_runtime

import (
	"github.com/dop251/goja"
	"time"
)

var TagName string

func init() {
	TagName = "json"
}

type FunctionCall = goja.FunctionCall
type Value = goja.Value
type ConstructorCall = goja.ConstructorCall
type Object = goja.Object

type Function func(call goja.FunctionCall) goja.Value
type Class func(ConstructorCall) *Object

func New() *Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper(TagName, true))
	vm.SetRandSource(func() float64 { return float64(time.Now().UnixNano()) })
	vm.SetTimeSource(func() time.Time { return time.Now() })
	return &Runtime{Runtime: vm}
}

type Runtime struct {
	*goja.Runtime
}

func (t *Runtime) JsExportTo(name string, target interface{}) error {
	switch val := t.Get(name); val {
	case goja.Undefined(), goja.NaN(), goja.Null():
		return nil
	default:
		return t.ExportTo(val, target)
	}
}
