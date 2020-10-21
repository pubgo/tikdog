package tikdog_cron

import (
	"errors"
	"github.com/pubgo/xerror"
	"github.com/robfig/cron/v3"
)

// cron.WithChain(cron.Recover(cron.DefaultLogger))
var defaultCron = New(cron.WithSeconds())

func SetDefault(c *cronManager) {
	defaultCron = c
}

func getDefault() *cronManager {
	if defaultCron != nil {
		return defaultCron
	}

	xerror.Exit(errors.New("please init cronManager"))
	return nil
}

func Start() {
	getDefault().Start()
}
func Stop() {
	getDefault().Stop()
}

func List() map[string]cron.Entry {
	return getDefault().List()
}
func Get(name string) cron.Entry {
	return getDefault().Get(name)
}
func Add(name string, spec string, cmd Handler) error {
	return getDefault().Add(name, spec, cmd)
}
func Remove(name string) {
	getDefault().Remove(name)
}
