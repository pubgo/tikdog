package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

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

func ScriptPath() string {
	return filepath.Join(Home, "scripts")
}
