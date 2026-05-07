package metrics

import (
	"runtime"
	"sync"
	"time"
)

// RuntimeSnapshot holds a point-in-time view of Go runtime statistics.
type RuntimeSnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	GoRoutines   int       `json:"goroutines"`
	HeapAllocMB  float64   `json:"heap_alloc_mb"`
	HeapSysMB    float64   `json:"heap_sys_mb"`
	GCCycles     uint32    `json:"gc_cycles"`
	NextGCMB     float64   `json:"next_gc_mb"`
}

// RuntimeCollector periodically samples Go runtime metrics.
type RuntimeCollector struct {
	mu       sync.RWMutex
	latest   RuntimeSnapshot
	stop     chan struct{}
	interval time.Duration
}

// NewRuntimeCollector creates a RuntimeCollector that samples every interval.
// Call Start to begin collection and Stop when done.
func NewRuntimeCollector(interval time.Duration) *RuntimeCollector {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &RuntimeCollector{
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start begins background sampling. It is safe to call once.
func (rc *RuntimeCollector) Start() {
	rc.sample() // immediate first sample
	go func() {
		ticker := time.NewTicker(rc.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rc.sample()
			case <-rc.stop:
				return
			}
		}
	}()
}

// Stop halts background sampling.
func (rc *RuntimeCollector) Stop() {
	close(rc.stop)
}

// Latest returns the most recently collected RuntimeSnapshot.
func (rc *RuntimeCollector) Latest() RuntimeSnapshot {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.latest
}

func (rc *RuntimeCollector) sample() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	snap := RuntimeSnapshot{
		Timestamp:   time.Now().UTC(),
		GoRoutines:  runtime.NumGoroutine(),
		HeapAllocMB: float64(ms.HeapAlloc) / (1024 * 1024),
		HeapSysMB:   float64(ms.HeapSys) / (1024 * 1024),
		GCCycles:    ms.NumGC,
		NextGCMB:    float64(ms.NextGC) / (1024 * 1024),
	}

	rc.mu.Lock()
	rc.latest = snap
	rc.mu.Unlock()
}
