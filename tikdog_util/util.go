package tikdog_util

import (
	"github.com/pubgo/xerror"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func IsHiddenDir(path string) bool {
	return len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".")
}

func CleanPath(path string) string {
	return strings.TrimSuffix(strings.TrimSpace(path), "/")
}

func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		return home + path[1:], nil
	}
	var err error
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if path == "." {
		return wd, nil
	}
	if strings.HasPrefix(path, "./") {
		return wd + path[1:], nil
	}

	return path, nil
}

func IsNotExist(name string) bool {
	_, err := os.Stat(name)
	return os.IsNotExist(err)
}

func IsDir(path string) bool {
	pf, err := os.Stat(path)
	return err == nil && pf.IsDir()
}

func GetShell() ([]string, error) {
	if path, err := exec.LookPath("bash"); err == nil {
		return []string{path, "-c"}, nil
	}

	if path, err := exec.LookPath("sh"); err == nil {
		return []string{path, "-c"}, nil
	}

	// even windows, there still has git-bash or mingw
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c"}, nil
	}

	return nil, xerror.New("Could not find bash or sh on path.")
}
