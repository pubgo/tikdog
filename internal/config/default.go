package config

import (
	"github.com/mitchellh/go-homedir"
	"github.com/pubgo/xerror"
	"path/filepath"

	"github.com/spf13/viper"
)

// 默认的全局配置
var (
	CfgType = "toml"
	CfgName = "config"
	Project = "tikdog"
	Debug   = true
	Trace   = false
	Mode    = "dev"
	Home    = xerror.PanicStr(filepath.Abs(filepath.Dir("")))
	// RunMode 项目运行环境
	RunMode = struct {
		Dev     string
		Test    string
		Stag    string
		Prod    string
		Release string
	}{
		Dev:     "dev",
		Test:    "test",
		Stag:    "stag",
		Prod:    "prod",
		Release: "release",
	}
)

func init() {
	getEnv(&Home, "home")
	getEnv(&Mode, "mode")

	{
		// 判断run mode格式
		switch Mode {
		case RunMode.Dev, RunMode.Stag, RunMode.Prod, RunMode.Test, RunMode.Release:
		default:
			xerror.Panic(xerror.Fmt("running mode does not match, mode: %s", Mode))
		}

		// 判断debug模式
		switch Mode {
		case RunMode.Dev, RunMode.Test, "":
			Debug = true
		}
	}

	{
		// 配置viper
		initViperEnv(Project)

		// 配置文件名字和类型
		viper.SetConfigType(CfgType)
		viper.SetConfigName(CfgName)

		// 监控当前工作目录
		_pwd := xerror.PanicStr(filepath.Abs(filepath.Dir("")))
		viper.AddConfigPath(_pwd)
		viper.AddConfigPath(filepath.Join(_pwd, CfgName))

		// 监控Home工作目录
		_home := xerror.PanicErr(homedir.Dir()).(string)
		viper.AddConfigPath(filepath.Join(_home, "."+Project))
		viper.AddConfigPath(filepath.Join(_home, "."+Project, CfgName))
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			xerror.ExitF(err, "read config failed")
		}
	}

	// 获取配置文件所在目录
	Home = filepath.Dir(xerror.PanicErr(filepath.Abs(viper.ConfigFileUsed())).(string))
}
