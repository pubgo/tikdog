package tikdog_cron

import (
	"context"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"github.com/robfig/cron/v3"
	"sync"
)

type Event struct{ context.Context }
type CallBack func(event interface{}) error

const notFoundEntryID = cron.EntryID(-1)

var EmptyEntry = cron.Entry{}

type cronManager struct {
	sync.RWMutex
	cron *cron.Cron
	data sync.Map
}

func (t *cronManager) loadID(name string) cron.EntryID {
	val, ok := t.data.Load(name)
	if ok {
		return val.(cron.EntryID)
	}
	return notFoundEntryID
}

func (t *cronManager) Add(name string, spec string, cmd CallBack) (grr error) {
	defer xerror.RespErr(&grr)

	oldID := t.loadID(name)

	id, err := t.cron.AddFunc(spec, func() {
		xprocess.Go(func(ctx context.Context) error {
			return xerror.Wrap(cmd(Event{Context: ctx}))
		})
	})
	xerror.Panic(err)
	t.data.Store(name, id)

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

	var data = make(map[string]cron.Entry)
	t.data.Range(func(key, value interface{}) bool {
		data[key.(string)] = t.cron.Entry(value.(cron.EntryID))
		return true
	})
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
	t.data.Delete(name)
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
	}
}
