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

	return xerror.Wrap(os.Setenv(strings.ToUpper(key), value))
}

func Get(val *string, names ...string) {
	for _, name := range names {
		nm := name
		if Prefix != "" {
			nm = Prefix + "_" + nm
		}

		env, ok := os.LookupEnv(strings.ToUpper(nm))
		env = strings.TrimSpace(env)
		if ok && env != "" {
			*val = env
		}
	}
}

// ExpandEnv replaces ${var} or $var in the string according to the values
// of the current environment variables. References to undefined
// variables are replaced by the empty string.
func ExpandEnv(s string) string {
	if Prefix != "" {
		s = Prefix + "_" + s
	}

	return os.ExpandEnv(s)
}

func Clear() {
	os.Clearenv()
}

func LookupEnv(key string) (string, bool) {
	if Prefix != "" {
		key = Prefix + "_" + key
	}

	return os.LookupEnv(key)
}
