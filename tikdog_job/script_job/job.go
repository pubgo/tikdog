package script_job

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/pubgo/tikdog/tikdog_cron"
	"github.com/pubgo/tikdog/tikdog_runtime/js_runtime"
	"github.com/pubgo/tikdog/tikdog_watcher"
	"github.com/pubgo/xerror"
	"io/ioutil"
)

const mainCode = "\nmain();\n"

type ss struct {
}

func (t *ss) Print() {
	fmt.Println("oksss")
}

func NewFromCode(path, code string) *job {
	j := &job{vm: js_runtime.New(), path: path, code: code}
	_, err := j.vm.RunString(code)
	xerror.Exit(err)

	j.name = path
	xerror.Exit(j.vm.ExportNameTo("name", &j.name))
	xerror.Exit(j.vm.ExportNameTo("version", &j.version))
	xerror.Exit(j.vm.ExportNameTo("kind", &j.kind))
	xerror.Exit(j.vm.ExportNameTo("cron", &j.cron))
	xerror.Exit(j.vm.ExportNameTo("main", &j.main))
	j.vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	j.vm.Set("print", fmt.Println)
	j.vm.Set("ss", &ss{})

	return j
}

type job struct {
	path    string
	name    string
	version string
	kind    string
	code    string
	cron    string
	vm      *js_runtime.Runtime
	main    func()
}

func (t *job) Cron() string {
	return t.cron
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

func (t *job) Name() string {
	return t.name
}

func (t *job) Type() string {
	return "script"
}

func (t *job) Close() error {
	return nil
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
		case tikdog_watcher.IsDeleteEvent(event):
			xerror.Panic(t.remove())
		case tikdog_watcher.IsRenameEvent(event):
		case tikdog_watcher.IsWriteEvent(event):
			dt, err := ioutil.ReadFile(event.Name)
			xerror.Panic(err)

			job := NewFromCode(event.Name, string(dt))
			xerror.Panic(t.remove())
			xerror.Panic(job.load())
		}

		return nil
	case tikdog_cron.Event:
		xerror.Panic(xerror.Try(t.main))
		return nil
	default:
		return nil
	}
}
