package tikdog_env

import (
	"os"
	"strings"

	"github.com/pubgo/xerror"
)

var Prefix string

func Set(key, value string) error {
	if Prefix != "" {
		key = Prefix + "_" + key
	}

	return xerror.Wrap(SetSys(key, value))
}

func SetSys(key, value string) error {
	return xerror.Wrap(os.Setenv(strings.ToUpper(key), value))
}

func Get(env *string, names ...string) {
	var nms []string
	if Prefix == "" {
		nms = names
	} else {
		for i := range names {
			nms = append(nms, strings.ToUpper(strings.Join([]string{Prefix, names[i]}, "_")))
		}
	}

	GetSys(env, nms...)
}

func GetSys(val *string, names ...string) {
	for _, name := range names {
		env, ok := os.LookupEnv(strings.ToUpper(name))
		env = strings.TrimSpace(env)
		if ok && env != "" {
			*val = env
		}
	}
}
