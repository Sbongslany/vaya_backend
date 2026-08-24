package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/yourorg/ehailing/backend/internal/workers/jobs"
)

type Manager struct {
	tripJobs    *jobs.TripJobs
	cleanupJobs *jobs.CleanupJobs
	log         *slog.Logger
	cancel      context.CancelFunc
}

func NewManager(tripJobs *jobs.TripJobs, cleanupJobs *jobs.CleanupJobs, log *slog.Logger) *Manager {
	return &Manager{
		tripJobs:    tripJobs,
		cleanupJobs: cleanupJobs,
		log:         log,
	}
}

func (m *Manager) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	m.log.Info("background workers started")

	// Ticker for minute-based jobs (Trips)
	go m.runMinuteJobs(ctx)

	// Ticker for hour-based jobs (Cleanup)
	go m.runHourJobs(ctx)
}

func (m *Manager) Stop() {
	if m.cancel != nil {
		m.cancel()
		m.log.Info("background workers stopped")
	}
}

func (m *Manager) runMinuteJobs(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run immediately on startup
	m.tripJobs.CancelStalledTrips(ctx)
	m.tripJobs.ActivateScheduledTrips(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.tripJobs.CancelStalledTrips(ctx)
			m.tripJobs.ActivateScheduledTrips(ctx)
		}
	}
}

func (m *Manager) runHourJobs(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.cleanupJobs.CleanExpiredShareTokens(ctx)
		}
	}
}
