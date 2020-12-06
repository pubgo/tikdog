package tikdog_sync

import (
	"github.com/pubgo/xlog"
	"go.uber.org/atomic"
)

func NewWaiter() *Waiter {
	return &Waiter{
		data: make(map[string]*atomic.Uint32),
		skip: make(map[string]*atomic.Uint32),
	}
}

type Waiter struct {
	data map[string]*atomic.Uint32
	skip map[string]*atomic.Uint32
}

func (t *Waiter) check(key string) {
	if _, ok := t.data[key]; !ok {
		t.skip[key] = atomic.NewUint32(0)
		t.data[key] = atomic.NewUint32(0)
	}
}

func (t *Waiter) Report(key string, c *atomic.Uint32) {
	t.check(key)

	if c.Load() != 0 {
		t.data[key].Store(0)
		return
	}

	t.data[key].Inc()
}

func (t *Waiter) Skip(key string) bool {
	t.check(key)

	if t.data[key].Load() == 0 {
		t.skip[key].Store(0)
		return false
	}

	t.skip[key].Inc()
	if t.skip[key].Load() > 5 {
		t.skip[key].Store(0)
		xlog.Debug("no skip")
		return false
	}

	xlog.Debug("skip")
	return true
}
