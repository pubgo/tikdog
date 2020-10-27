package tikdog

import (
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pubgo/tikdog/internal/config"
	"github.com/pubgo/tikdog/tikdog_cron"
	"github.com/pubgo/tikdog/tikdog_job/script_job"
	"github.com/pubgo/tikdog/tikdog_watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"
)

func New() *tikdog {
	return &tikdog{}
}

type tikdog struct{}

func initDevLog() {
	zl, err := xlog_config.NewZapLoggerFromConfig(xlog_config.NewDevConfig())
	xerror.Exit(err)

	zl = zl.WithOptions(zap.AddCaller(), zap.AddCallerSkip(2)).Named(config.Project)
	xerror.Exit(xlog.SetLog(xlog.New(zl)))
}

func (t *tikdog) loadLog() {
	initDevLog()

	xerror.Exit(config.Decode("log", func(cfg *xlog_config.Config) {
		zapL := xerror.PanicErr(xlog_config.NewZapLoggerFromConfig(*cfg)).(*zap.Logger)
		zapL = zapL.WithOptions(xlog.AddCaller(), xlog.AddCallerSkip(2)).Named(config.Project)
		xerror.Exit(xlog.SetLog(xlog.New(zapL)))
	}))
}

func (t *tikdog) Run() (grr error) {
	defer xerror.RespErr(&grr)

	t.loadLog()
	t.loadScripts()

	xerror.Panic(t.Start())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch

	xerror.Panic(t.Stop())
	return nil
}

func (t *tikdog) loadScripts() {
	var scriptPath = config.ScriptPath()
	xerror.Exit(tikdog_watcher.Add(scriptPath, script_job.SimpleEvent()))
	xerror.Exit(filepath.Walk(scriptPath, func(path string, info os.FileInfo, err error) (grr error) {
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
		xerror.Panic(job.Load())
		return nil
	}))
}

func (t *tikdog) Start() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(tikdog_cron.Start())
	tikdog_watcher.Start()

	return nil
}

func (t *tikdog) Stop() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(tikdog_cron.Stop())
	tikdog_watcher.Stop()

	return nil
}
