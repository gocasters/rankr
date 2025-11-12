package historical

import (
	"fmt"
	"sync"
	"time"

	"github.com/gocasters/rankr/pkg/logger"
)

type ProgressTracker struct {
	successCount int64
	failureCount int64
	startTime    time.Time
	lastUpdate   time.Time
	mu           sync.RWMutex
}

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{}
}

func (p *ProgressTracker) Start() {
	p.startTime = time.Now()
	p.lastUpdate = time.Now()
}

func (p *ProgressTracker) RecordSuccess() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.successCount++
	p.maybeLogProgress()
}

func (p *ProgressTracker) RecordFailure() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failureCount++
	p.maybeLogProgress()
}

func (p *ProgressTracker) maybeLogProgress() {
	now := time.Now()
	if now.Sub(p.lastUpdate) > 10*time.Second {
		p.lastUpdate = now

		elapsed := now.Sub(p.startTime)
		total := p.successCount + p.failureCount
		rate := float64(total) / elapsed.Seconds()

		logger.L().Info("Progress update",
			"success", p.successCount,
			"failed", p.failureCount,
			"rate", fmt.Sprintf("%.2f events/sec", rate),
			"elapsed", elapsed.Round(time.Second))
	}
}

func (p *ProgressTracker) PrintFinalReport() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	duration := time.Since(p.startTime)
	total := p.successCount + p.failureCount

	logger.L().Info("Historical fetch completed",
		"success", p.successCount,
		"failed", p.failureCount,
		"total", total,
		"duration", duration.Round(time.Second),
		"avg_rate", fmt.Sprintf("%.2f events/sec", float64(total)/duration.Seconds()))
}
