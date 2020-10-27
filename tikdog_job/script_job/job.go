package script_job

import (
	"github.com/pubgo/tikdog/tikdog_cron"
	"github.com/pubgo/tikdog/tikdog_runtime/js_runtime"
	"github.com/pubgo/tikdog/tikdog_watcher"
	"github.com/pubgo/xerror"
	"io/ioutil"
)

func SimpleEvent() func(event interface{}) error {
	return (&job{}).OnEvent
}

func NewFromCode(path, code string) *job {
	j := &job{vm: js_runtime.New(), path: path, code: code, name: path}

	xerror.ExitErr(j.vm.RunString(code))
	xerror.Exit(j.vm.JsExportTo("name", &j.name))
	xerror.Exit(j.vm.JsExportTo("version", &j.version))
	xerror.Exit(j.vm.JsExportTo("kind", &j.kind))
	xerror.Exit(j.vm.JsExportTo("cron", &j.cron))
	xerror.Exit(j.vm.JsExportTo("main", &j.main))

	return j
}

type job struct {
	vm      *js_runtime.Runtime
	path    string
	name    string
	version string
	kind    string
	code    string
	cron    string
	main    func()
}

func (t *job) remove() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(tikdog_watcher.Remove(t.path))
	xerror.Panic(tikdog_cron.Remove(t.name))
	return nil
}

func (t *job) Load() (err error) {
	return t.load()
}

func (t *job) load() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(tikdog_watcher.Add(t.path, t.OnEvent))
	xerror.Panic(tikdog_cron.Add(t.name, t.cron, t.OnEvent))
	return nil
}

func (t *job) Type() string {
	return "script"
}

func (t *job) OnEvent(event interface{}) (err error) {
	defer xerror.RespErr(&err)

	if event == nil {
		return nil
	}

	switch event := event.(type) {
	case tikdog_watcher.Event:
		switch {
		case tikdog_watcher.IsCreateEvent(event):
			dt := xerror.PanicBytes(ioutil.ReadFile(event.Name))
			xerror.Panic(NewFromCode(event.Name, string(dt)).load())

		case tikdog_watcher.IsDeleteEvent(event):
			xerror.Panic(t.remove())

		case tikdog_watcher.IsUpdateEvent(event):
			dt := xerror.PanicBytes(ioutil.ReadFile(event.Name))
			xerror.Panic(NewFromCode(event.Name, string(dt)).load())
			xerror.Panic(t.remove())
			return nil
		}

		return nil
	case tikdog_cron.Event:
		xerror.Panic(xerror.Try(t.main))
		return nil
	default:
		return nil
	}
}
