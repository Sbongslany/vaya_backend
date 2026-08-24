package jobs

import (
	"context"
	"log/slog"
	"time"

	tripRepo "github.com/yourorg/ehailing/backend/internal/trip/infrastructure/persistence/postgres"
)

type TripJobs struct {
	repo *tripRepo.TripRepository
	log  *slog.Logger
}

func NewTripJobs(repo *tripRepo.TripRepository, log *slog.Logger) *TripJobs {
	return &TripJobs{repo: repo, log: log}
}

// CancelStalledTrips runs every minute and cancels trips stuck in REQUESTED for > 15 mins
func (j *TripJobs) CancelStalledTrips(ctx context.Context) {
	cutoff := time.Now().Add(-15 * time.Minute)
	count, err := j.repo.CancelStalledRequestedTrips(ctx, cutoff)
	if err != nil {
		j.log.Error("failed to cancel stalled trips", "error", err)
		return
	}
	if count > 0 {
		j.log.Info("auto-cancelled stalled trips", "count", count)
	}
}

// ActivateScheduledTrips runs every minute and activates trips whose scheduled time has arrived
func (j *TripJobs) ActivateScheduledTrips(ctx context.Context) {
	now := time.Now()
	count, err := j.repo.ActivateScheduledTrips(ctx, now)
	if err != nil {
		j.log.Error("failed to activate scheduled trips", "error", err)
		return
	}
	if count > 0 {
		j.log.Info("activated scheduled trips", "count", count)
	}
}
