package script_job

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/lithdew/quickjs"
	"github.com/pubgo/tikdog/tikdog_runtime/js_runtime"
)

func New() *job {
	return &job{
		rt: js_runtime.New(),
	}
}

func NewFromCode(code string) *job {
	j := &job{rt: js_runtime.New()}
	attrs := js_runtime.AttrExtract(code, "name", "version", "kind")
	for k, v := range attrs {
		if v.IsUndefined() || v.IsNull() || v.IsException() {

		}

		switch k {
		case "name":
		case "version":
		case "kind":

		}
	}

	return j
}

type job struct {
	path    string
	name    string
	version string
	kind    string
	code    string
	rt      *quickjs.Context
}

func (t *job) Type() string {
	return "script"
}

func (t *job) OnEvent(ctx context.Context, event interface{}) error {
	if event == nil {
		return nil
	}

	switch event := event.(type) {
	case fsnotify.Event:
		fmt.Println(event.Name)
		t.rt.Eval()
		return nil
	default:
		return nil
	}
}
