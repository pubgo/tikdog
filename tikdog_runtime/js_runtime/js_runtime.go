package js_runtime

import (
	"errors"
	"fmt"
	"github.com/lithdew/quickjs"
)

var rt = quickjs.NewRuntime()

func Free() {
	rt.Free()
}

func AttrExtract(code string, names ...string) map[string]quickjs.Value {
	ctx := rt.NewContext()
	defer ctx.Free()

	g := ctx.Globals()

	val, err := ctx.Eval(code)
	check(err)
	defer val.Free()

	var ret = make(map[string]quickjs.Value)
	for _, name := range names {
		ret[name] = g.Get(name)
	}

	return ret
}

func Eval(code string) (quickjs.Value, quickjs.Value) {
	ctx := rt.NewContext()
	defer ctx.Free()

	g := ctx.Globals()

	val, err := ctx.Eval(code)
	check(err)
	defer val.Free()
	return g, val
}

func New(values ...quickjs.Value) *quickjs.Context {
	ctx := rt.NewContext()
	return ctx
}

func check(err error) {
	if err != nil {
		var evalErr *quickjs.Error
		if errors.As(err, &evalErr) {
			fmt.Println(evalErr.Cause)
			fmt.Println(evalErr.Stack)
		}
		panic(err)
	}
}
