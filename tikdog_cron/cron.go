package tikdog_cron

import (
	"github.com/pubgo/xerror"
	"github.com/robfig/cron/v3"
	"sync"
)

const notFoundEntryID = cron.EntryID(-1)

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

func (t *cronManager) Add(name string, spec string, cmd func()) (err error) {
	defer xerror.RespErr(&err)

	t.Lock()
	defer t.Unlock()

	id := t.loadID(name)
	if id == notFoundEntryID {
		id, err := t.cron.AddFunc(spec, cmd)
		xerror.Panic(err)
		t.data[name] = id
		return nil
	}

	id, err = t.cron.AddFunc(spec, cmd)
	xerror.Panic(err)
	t.data[name] = id

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

func (t *cronManager) Remove(name string) {
	t.Lock()
	defer t.Unlock()

	id := t.loadID(name)
	if id == notFoundEntryID {
		return
	}
	t.cron.Remove(id)
	delete(t.data, name)
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
