package script_job

import (
	"errors"
	"fmt"
	"github.com/lithdew/quickjs"
	"github.com/pubgo/tikdog/tikdog_cron"
	"github.com/pubgo/tikdog/tikdog_runtime/js_runtime"
	"github.com/pubgo/tikdog/tikdog_watcher"
	"github.com/pubgo/xerror"
	"time"
)

const mainCode = "\nmain();\n"

func New() *job {
	return &job{
		rt: js_runtime.New(),
	}
}

func NewFromCode(path, code string) *job {
	j := &job{rt: js_runtime.New(), path: path, code: code}

	g := j.rt.Globals()
	_, err := j.rt.Eval(code)
	xerror.Exit(err)

	j.name = path
	if val := g.Get("name"); !val.IsUndefined() {
		j.name = val.String()
	}
	if val := g.Get("version"); !val.IsUndefined() {
		j.version = val.String()
	}
	if val := g.Get("kind"); !val.IsUndefined() {
		j.kind = val.String()
	}
	if val := g.Get("cron"); !val.IsUndefined() {
		j.cron = val.String()
	}

	return j
}

type job struct {
	path    string
	name    string
	version string
	kind    string
	code    string
	cron    string
	rt      *quickjs.Context
}

func (t *job) Cron() string {
	return t.cron
}

func (t *job) Name() string {
	return t.name
}

func (t *job) Type() string {
	return "script"
}

func (t *job) OnEvent(event interface{}) error {
	if event == nil {
		return nil
	}

	switch event := event.(type) {
	case tikdog_watcher.Event:
		switch {
		case tikdog_watcher.IsCreateEvent(event):
		case tikdog_watcher.IsDeleteEvent(event):
		case tikdog_watcher.IsRenameEvent(event):
		case tikdog_watcher.IsWriteEvent(event):
			fmt.Println(event.Name, event.String())
			fmt.Println(*t)
		}

		return nil
	case tikdog_cron.Event:
		js_runtime.PropertyNames(t.rt)

		//t.rt.Globals().SetFunction("print", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		//	fmt.Println(args)
		//	return ctx.Null()
		//})

		ctx := js_runtime.New()
		defer ctx.Free()

		//_, err := t.rt.Eval(t.code)
		_, err := ctx.Eval(t.code)
		if err != nil {
			time.Sleep(time.Millisecond * 10)
			var evalErr *quickjs.Error
			if errors.As(err, &evalErr) {
				fmt.Printf("%#v", err)
				fmt.Println(evalErr.Cause)
				fmt.Println(evalErr.Stack)
				return nil
			}

			return xerror.Wrap(err)
		}
		//defer val.Free()

		return nil
	default:
		return nil
	}
}
