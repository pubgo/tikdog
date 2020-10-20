package tikdog_watcher

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/xerror"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func isHiddenDir(path string) bool {
	return len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".")
}

func cleanPath(path string) string {
	return strings.TrimSuffix(strings.TrimSpace(path), "/")
}

func (t *watcherManager) withLock(f func()) {
	t.mu.Lock()
	f()
	t.mu.Unlock()
}

func expandPath(path string) (string, error) {
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

func isDir(path string) bool {
	pf, err := os.Stat(path)
	return err == nil && pf.IsDir()
}

func validEvent(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Create == fsnotify.Create ||
		ev.Op&fsnotify.Write == fsnotify.Write ||
		ev.Op&fsnotify.Remove == fsnotify.Remove
}

func removeEvent(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Remove == fsnotify.Remove
}

func cmdPath(path string) string {
	return strings.Split(path, " ")[0]
}

func getShell() ([]string, error) {
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
