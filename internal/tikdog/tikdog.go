package tikdog

import (
	"fmt"
	"github.com/pubgo/tikdog/internal/config"
	"github.com/pubgo/tikdog/tikdog_cron"
	"github.com/pubgo/tikdog/tikdog_job/script_job"
	"github.com/pubgo/tikdog/tikdog_watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func New() *tikdog {
	return &tikdog{}
}

type tikdog struct {
	cfg *option
}

func (t *tikdog) Run() (grr error) {
	defer xerror.RespErr(&grr)

	xerror.Panic(t.loadScripts())

	zl, err := xlog_config.NewZapLoggerFromConfig(xlog_config.NewDevConfig())
	xerror.Panic(err)

	zl = zl.WithOptions(zap.AddCaller(), zap.AddCallerSkip(2)).Named(config.Project)
	xerror.Panic(xlog.SetLog(xlog.New(zl)))

	xerror.Panic(t.Start())

	fmt.Println(tikdog_watcher.List())
	fmt.Println(tikdog_cron.List())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch

	xerror.Panic(t.Stop())
	return nil
}

func (t *tikdog) loadScripts() (err error) {
	fmt.Println(config.Home)
	return filepath.Walk(config.ScriptPath(), func(path string, info os.FileInfo, err error) (grr error) {
		defer xerror.RespErr(&err)
		xerror.Panic(err)

		if info == nil || info.IsDir() {
			return nil
		}

		code := string(xerror.PanicBytes(ioutil.ReadFile(path)))
		if code == "" {
			return xerror.New("code is empty", path)
		}

		job := script_job.NewFromCode(path, code)
		xerror.Panic(tikdog_watcher.Add(path, job.OnEvent))
		xerror.Panic(tikdog_cron.Add(job.Name(), job.Cron(), job.OnEvent))
		return nil
	})
}

func (t *tikdog) Start() (err error) {
	defer xerror.RespErr(&err)

	tikdog_cron.Start()
	tikdog_watcher.Start()

	return nil
}

func (t *tikdog) Stop() (err error) {
	defer xerror.RespErr(&err)

	tikdog_cron.Stop()
	tikdog_watcher.Stop()

	return nil
}
