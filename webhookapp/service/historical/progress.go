package historical

import (
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

type progressSnapshot struct {
	success int64
	failure int64
	elapsed time.Duration
	rate    float64
}

func (p *ProgressTracker) RecordSuccess() {
	snapshot := p.incrementAndCheckLog(true)
	if snapshot != nil {
		p.logProgress(snapshot)
	}
}

func (p *ProgressTracker) RecordFailure() {
	snapshot := p.incrementAndCheckLog(false)
	if snapshot != nil {
		p.logProgress(snapshot)
	}
}

func (p *ProgressTracker) incrementAndCheckLog(isSuccess bool) *progressSnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()

	if isSuccess {
		p.successCount++
	} else {
		p.failureCount++
	}

	now := time.Now()
	if now.Sub(p.lastUpdate) <= 10*time.Second {
		return nil
	}

	p.lastUpdate = now
	elapsed := now.Sub(p.startTime)
	total := p.successCount + p.failureCount

	var rate float64
	if elapsed.Seconds() == 0 {
		rate = 0.0
	} else {
		rate = float64(total) / elapsed.Seconds()
	}

	return &progressSnapshot{
		success: p.successCount,
		failure: p.failureCount,
		elapsed: elapsed,
		rate:    rate,
	}
}

func (p *ProgressTracker) logProgress(snapshot *progressSnapshot) {
	logger.L().Info("Progress update",
		"success", snapshot.success,
		"failed", snapshot.failure,
		"rate", snapshot.rate,
		"elapsed", snapshot.elapsed.Round(time.Second))
}

func (p *ProgressTracker) PrintFinalReport() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	duration := time.Since(p.startTime)
	total := p.successCount + p.failureCount

	var avgRate float64
	if duration.Seconds() == 0 {
		avgRate = 0.0
	} else {
		avgRate = float64(total) / duration.Seconds()
	}

	logger.L().Info("Historical fetch completed",
		"success", p.successCount,
		"failed", p.failureCount,
		"total", total,
		"duration", duration.Round(time.Second),
		"avg_rate", avgRate)
}
