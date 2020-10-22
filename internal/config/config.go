package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
	"path/filepath"
	"reflect"
	"strings"
)

// Decode
// decode config data
func Decode(name string, fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	if viper.Get(name) == nil {
		return nil
	}

	if fn == nil {
		return xerror.New("fn should not be nil")
	}

	vfn := reflect.ValueOf(fn)
	if vfn.Type().Kind() != reflect.Func {
		return xerror.New("fn should be the a function")
	}

	if vfn.Type().NumIn() != 1 {
		return xerror.New("fn input num should be one")
	}

	mthIn := reflect.New(vfn.Type().In(0).Elem())
	ret := reflect.ValueOf(viper.UnmarshalKey).Call(
		[]reflect.Value{
			reflect.ValueOf(name),
			mthIn,
			reflect.ValueOf(func(cfg *mapstructure.DecoderConfig) {
				cfg.TagName = CfgType
			}),
		},
	)
	if !ret[0].IsNil() {
		return xerror.WrapF(ret[0].Interface().(error), "config decode error")
	}

	vfn.Call([]reflect.Value{mthIn})

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
