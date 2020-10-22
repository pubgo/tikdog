package tikdog_cron

import (
	"fmt"
	"github.com/pubgo/xerror"
	"github.com/robfig/cron/v3"
	"sync"
)

const notFoundEntryID = cron.EntryID(-1)

type Event struct {
}
type Handler func(event interface{}) error

var EmptyEntry = cron.Entry{}

type cronManager struct {
	sync.RWMutex
	cron *cron.Cron
	data map[string]cron.EntryID
}

func (t *cronManager) loadID(name string) cron.EntryID {
	val, ok := t.data[name]
	if ok {
		return val
	}
	return notFoundEntryID
}

func (t *cronManager) Add(name string, spec string, cmd Handler) (grr error) {
	defer xerror.RespErr(&grr)

	t.Lock()
	defer t.Unlock()

	oldID := t.loadID(name)

	id, err := t.cron.AddFunc(spec, func() {
		go func() {
			if err := xerror.Parse(cmd(Event{})); err != nil {
				fmt.Println(err.Println())
			}
		}()
	})
	xerror.Panic(err)
	t.data[name] = id

	if oldID == notFoundEntryID {
		return nil
	}
	t.cron.Remove(id)

	return nil
}

func (t *cronManager) Get(name string) cron.Entry {
	t.RLock()
	defer t.RUnlock()

	id := t.loadID(name)
	if id == notFoundEntryID {
		return EmptyEntry
	}
	return t.cron.Entry(id)
}

func (t *cronManager) List() map[string]cron.Entry {
	t.RLock()
	defer t.RUnlock()

	var data = make(map[string]cron.Entry, len(t.data))
	for k, v := range t.data {
		data[k] = t.cron.Entry(v)
	}
	return data
}

func (t *cronManager) Remove(name string) error {
	t.Lock()
	defer t.Unlock()

	id := t.loadID(name)
	if id == notFoundEntryID {
		return nil
	}

	t.cron.Remove(id)
	delete(t.data, name)
	return nil
}

func (t *cronManager) Start() {
	t.cron.Start()
}

func (t *cronManager) Stop() {
	t.cron.Stop()
}

func New(opts ...cron.Option) *cronManager {
	return &cronManager{
		cron: cron.New(opts...),
		data: make(map[string]cron.EntryID),
	}
}
