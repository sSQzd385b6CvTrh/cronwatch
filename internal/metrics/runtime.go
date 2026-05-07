package metrics

import (
	"runtime"
	"sync"
	"time"
)

// RuntimeSample holds a point-in-time snapshot of Go runtime statistics.
type RuntimeSample struct {
	Timestamp   time.Time
	Goroutines  int
	HeapAllocMB float64
	HeapSysMB   float64
	GCCycles    uint32
	NextGCMB    float64
}

// RuntimeCollector periodically samples Go runtime metrics.
type RuntimeCollector struct {
	interval time.Duration
	mu       sync.RWMutex
	latest   RuntimeSample
	stop     chan struct{}
	wg       sync.WaitGroup
}

const defaultRuntimeInterval = 15 * time.Second

// NewRuntimeCollector creates a RuntimeCollector that samples every interval.
// If interval is zero, defaultRuntimeInterval is used.
func NewRuntimeCollector(interval time.Duration) *RuntimeCollector {
	if interval <= 0 {
		interval = defaultRuntimeInterval
	}
	return &RuntimeCollector{
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start begins background sampling. It is safe to call once.
func (rc *RuntimeCollector) Start() {
	rc.sample()
	rc.wg.Add(1)
	go func() {
		defer rc.wg.Done()
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

// Stop halts background sampling and waits for the goroutine to exit.
func (rc *RuntimeCollector) Stop() {
	close(rc.stop)
	rc.wg.Wait()
}

// Latest returns the most recently collected RuntimeSample.
func (rc *RuntimeCollector) Latest() RuntimeSample {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.latest
}

func (rc *RuntimeCollector) sample() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	s := RuntimeSample{
		Timestamp:   time.Now().UTC(),
		Goroutines:  runtime.NumGoroutine(),
		HeapAllocMB: float64(ms.HeapAlloc) / (1024 * 1024),
		HeapSysMB:   float64(ms.HeapSys) / (1024 * 1024),
		GCCycles:    ms.NumGC,
		NextGCMB:    float64(ms.NextGC) / (1024 * 1024),
	}
	rc.mu.Lock()
	rc.latest = s
	rc.mu.Unlock()
}
