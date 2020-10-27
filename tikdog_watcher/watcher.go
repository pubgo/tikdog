package tikdog_watcher

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/tikdog/tikdog_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrPathNotFound = errors.New("error: path not found")
)

type CallBack func(event interface{}) error
type Event struct {
	fsnotify.Event
	Watcher *fsnotify.Watcher
}

// watcherManager ...
type watcherManager struct {
	data sync.Map

	watcher *fsnotify.Watcher
	exitCh  chan struct{}
}

func New() (*watcherManager, error) {
	watcher, err := fsnotify.NewWatcher()
	xerror.Exit(err)

	return &watcherManager{
		watcher: watcher,
		exitCh:  make(chan struct{}),
	}, nil
}

func (t *watcherManager) add(name string) (err error) {
	defer xerror.RespErr(&err)

	// check file existed
	if tikdog_util.IsNotExist(name) {
		return xerror.Wrap(ErrPathNotFound)
	}

	// filter file
	//for i := range t.config.ExcludePattern {
	//	matched, err := regexp.MatchString(t.config.ExcludePattern[i], name)
	//	xerror.Panic(err)
	//	if matched {
	//		return nil
	//	}
	//}

	return xerror.Wrap(t.watcher.Add(name))
}

func (t *watcherManager) List() []string {
	var data []string
	t.data.Range(func(key, _ interface{}) bool {
		data = append(data, key.(string))
		return true
	})
	return data
}

func (t *watcherManager) RemoveRecursive(name string) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(handlePath(&name))
	xerror.Panic(t.Remove(name))

	if !tikdog_util.IsDir(name) {
		return nil
	}

	return xerror.Wrap(filepath.Walk(name, func(path string, info os.FileInfo, err error) (gerr error) {
		defer xerror.RespErr(&gerr)

		xerror.Panic(err)

		if info == nil {
			return nil
		}

		xerror.Panic(handlePath(&name))
		return xerror.Wrap(t.Remove(path))
	}))
}

func (t *watcherManager) Remove(name string) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(handlePath(&name))

	if tikdog_util.IsNotExist(name) {
		return nil
	}

	if _, ok := t.data.Load(name); !ok {
		return nil
	}

	xerror.Panic(t.watcher.Remove(name))
	t.data.Delete(name)
	return nil
}

func (t *watcherManager) Add(name string, h CallBack) (err error) {
	defer xerror.RespErr(&err)

	if h == nil {
		return xerror.New("CallBack is nil")
	}

	xerror.Panic(handlePath(&name))
	xerror.Panic(t.add(name))

	t.data.Store(name, h)
	return nil
}

func (t *watcherManager) AddRecursive(name string, h CallBack) (err error) {
	defer xerror.RespErr(&err)

	if h == nil {
		return xerror.New("CallBack is nil")
	}

	xerror.Panic(handlePath(&name))

	if !tikdog_util.IsDir(name) {
		xerror.Panic(t.add(name))
		t.data.Store(name, h)
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
		t.data.Store(name, h)
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

				fn, ok := t.data.Load(event.Name)
				if ok {
					if err := fn.(CallBack)(Event{Watcher: t.watcher, Event: event}); err != nil {
						fmt.Println(xerror.Parse(xerror.WrapF(err, event.String())).Println())
					}
				}
			case err, ok := <-t.watcher.Errors:
				if !ok {
					return
				}

				if err == nil {
					continue
				}

				xlog.Error(err.Error())
			}
		}
	}()
}

// Stop
func (t *watcherManager) Stop() {
	t.exitCh <- struct{}{}
}

func IsWriteEvent(ev Event) bool {
	return ev.Op&fsnotify.Write == fsnotify.Write
}

func IsDeleteEvent(ev Event) bool {
	return ev.Op&fsnotify.Remove == fsnotify.Remove
}

func IsCreateEvent(ev Event) bool {
	return ev.Op&fsnotify.Create == fsnotify.Create
}

func IsUpdateEvent(ev Event) bool {
	op := ev.Op & fsnotify.Create
	switch op {
	case fsnotify.Write, fsnotify.Rename:
		return true
	default:
		return false
	}
}

func IsRenameEvent(ev Event) bool {
	return ev.Op&fsnotify.Rename == fsnotify.Rename
}

func handlePath(name *string) (err error) {
	defer xerror.RespErr(&err)

	nm := *name
	nm = filepath.Clean(nm)
	nme := xerror.PanicStr(filepath.EvalSymlinks(*name))
	*name = xerror.PanicStr(filepath.Abs(nme))

	return nil
}
