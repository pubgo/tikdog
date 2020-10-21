package js_runtime

import (
	"errors"
	"fmt"
	"github.com/lithdew/quickjs"
	"github.com/pubgo/xerror"
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

func New() *quickjs.Context {
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

func PropertyNames(ctx *quickjs.Context) {
	names, err := ctx.Globals().PropertyNames()
	xerror.Panic(err)

	for _, name := range names {
		val := ctx.Globals().GetByAtom(name.Atom)
		defer val.Free()
		//fmt.Printf("'%s': %s\n", name, val)
	}
}
