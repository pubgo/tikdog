package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func Env(env *string, names ...string) {
	getEnv(env, names...)
}

func SysEnv(env *string, names ...string) {
	getSysEnv(env, names...)
}

func getSysEnv(val *string, names ...string) {
	for _, name := range names {
		env, ok := os.LookupEnv(strings.ToUpper(name))
		env = strings.TrimSpace(env)
		if ok && env != "" {
			*val = env
		}
	}
}

func getEnv(val *string, names ...string) {
	for _, name := range names {
		env, ok := os.LookupEnv(strings.ToUpper(strings.Join([]string{Project, name}, "_")))
		env = strings.TrimSpace(env)
		if ok && env != "" {
			*val = env
		}
	}
}

// Decode
// decode config data
func Decode(name string, data interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.PanicF(viper.UnmarshalKey(name, data, func(cfg *mapstructure.DecoderConfig) {
		cfg.TagName = CfgType
	}), "config decode error")

	return
}

// initViperEnv
// sets to use Env variables if set.
func initViperEnv(prefix string) {
	viper.SetEnvPrefix(prefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "/"))
	viper.AutomaticEnv()
}
