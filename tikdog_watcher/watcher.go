package tikdog_watcher

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

var (
	ErrPathNotFound = errors.New("error: path not found")
)

type Handler func(event fsnotify.Event) error

// FilterFileHookFunc is a function that is called to filter files during listings.
// If a file is ok to be listed, nil is returned otherwise ErrSkip is returned.
type FilterFileHookFunc func(info os.FileInfo, fullPath string) error

// watcherManager ...
type watcherManager struct {
	mu   sync.RWMutex
	data map[string]Handler

	config  *config
	watcher *fsnotify.Watcher

	eventCh chan string
	exitCh  chan bool
}

func New() (*watcherManager, error) {
	var err error
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &watcherManager{
		watcher: watcher,
		eventCh: make(chan string, 1000),
	}, nil
}

func isExist(name string) bool {
	_, err := os.Stat(name)
	return os.IsExist(err)
}

func (t *watcherManager) add(name string) (err error) {
	defer xerror.RespErr(&err)

	// check file existed
	if !isExist(name) {
		return xerror.Wrap(ErrPathNotFound)
	}

	// filter file
	for i := range t.config.ExcludePattern {
		matched, err := regexp.MatchString(t.config.ExcludePattern[i], name)
		xerror.Panic(err)
		if matched {
			return nil
		}
	}

	return xerror.Wrap(t.watcher.Add(name))
}

func handlePath(name *string) (err error) {
	defer xerror.RespErr(&err)

	nme := xerror.PanicStr(filepath.EvalSymlinks(*name))
	*name = xerror.PanicStr(filepath.Abs(nme))

	return nil
}

func (t *watcherManager) RemoveRecursive(name string) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Remove(name))

	if !isDir(name) {
		return nil
	}

	return xerror.Wrap(filepath.Walk(name, func(path string, info os.FileInfo, err error) (gerr error) {
		defer xerror.RespErr(&gerr)

		xerror.Panic(err)

		if info == nil {
			return nil
		}

		return t.Remove(path)
	}))
}

func (t *watcherManager) Remove(name string) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(handlePath(&name))

	if !isExist(name) {
		return nil
	}

	xerror.Panic(t.watcher.Remove(name))
	delete(t.data, name)
	return nil
}

func (t *watcherManager) Add(name string, h Handler) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(handlePath(&name))

	t.mu.Lock()
	defer t.mu.Unlock()

	xerror.Panic(t.add(name))

	t.data[name] = h
	return nil
}

func (t *watcherManager) AddRecursive(name string, h Handler) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(handlePath(&name))

	t.mu.Lock()
	defer t.mu.Unlock()

	if !isDir(name) {
		xerror.Panic(t.add(name))
		t.data[name] = h
		return nil
	}

	return xerror.Wrap(filepath.Walk(name, func(path string, info os.FileInfo, err error) (grr error) {
		defer xerror.RespErr(&grr)

		xerror.Panic(err)

		if info == nil {
			return nil
		}

		xerror.Panic(handlePath(&name))
		xerror.Panic(t.add(name))
		t.data[name] = h
		return nil
	}))
}

// Start
// Endless loop and never return
func (t *watcherManager) Start() {

	go func() {
		for {
			select {
			case <-t.exitCh:
				_ = t.watcher.Close()
				return
			case event, ok := <-t.watcher.Events:
				if !ok {
					return
				}

				t.mu.RLock()
				fn, ok := t.data[event.Name]
				t.mu.RUnlock()
				if ok {
					if err := fn(event); err != nil {
						xlog.Error(xerror.Parse(xerror.WrapF(err, event.String())).Stack(true))
					}
				}
			case err, ok := <-t.watcher.Errors:
				if !ok {
					return
				}

				xlog.Error(err.Error())
			}
		}
	}()
}

// Stop
func (t *watcherManager) Stop() {
	t.exitCh <- true
}
