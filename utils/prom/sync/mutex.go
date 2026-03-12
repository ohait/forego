package sync

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ohait/forego/ctx/log"
	"github.com/ohait/forego/utils/prom"
)

var mutrics = prom.Register("mutex", &prom.Histogram{
	Desc:   "mutex operation latency",
	Labels: []string{"op", "src"},
	Buckets: []float64{
		0.0001, 0.0002, 0.0005, // 100µs ..
		0.001, 0.002, 0.005,
		0.01, 0.02, 0.05,
		0.1, 0.2, 0.5,
		1, 2, 5,
		15, 30, 60, // .. 1min
	},
})

// Mutex is a wrapper on forego sync.Mutex that generates Prometheus metrics for lock and unlock operations, and logs if they take too long.
type Mutex struct {
	m      sync.Mutex
	lock   time.Time
	holder string
}

func (m *Mutex) Lock() {
	_, file, line, _ := runtime.Caller(1)
	holder := fmt.Sprintf("%s:%d", file, line)
	t0 := time.Now()
	m.m.Lock()
	m.lock = time.Now()
	m.holder = holder
	elapsed := m.lock.Sub(t0)
	mutrics.Observe(elapsed.Seconds(), "lock", holder)
	if elapsed > 100*time.Millisecond {
		log.Warnf(nil, "lock() took %v at %s", elapsed, holder)
	}
}

func (m *Mutex) Unlock() {
	dt := time.Since(m.lock)
	holder := m.holder
	m.m.Unlock()
	mutrics.Observe(dt.Seconds(), "unlock", holder)
	if dt > 500*time.Millisecond {
		log.Warnf(nil, "unlock() took %v at %s", dt, holder)
	}
}
