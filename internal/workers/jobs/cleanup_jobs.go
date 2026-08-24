package jobs

import (
	"context"
	"log/slog"
	"time"

	safetyRepo "github.com/yourorg/ehailing/backend/internal/safety/infrastructure/persistence/postgres"
)

type CleanupJobs struct {
	shareRepo *safetyRepo.ShareTokenRepository
	log       *slog.Logger
}

func NewCleanupJobs(shareRepo *safetyRepo.ShareTokenRepository, log *slog.Logger) *CleanupJobs {
	return &CleanupJobs{shareRepo: shareRepo, log: log}
}

// CleanExpiredShareTokens runs once an hour to delete expired tracking links
func (j *CleanupJobs) CleanExpiredShareTokens(ctx context.Context) {
	now := time.Now()
	count, err := j.shareRepo.DeleteExpired(ctx, now)
	if err != nil {
		j.log.Error("failed to clean expired share tokens", "error", err)
		return
	}
	if count > 0 {
		j.log.Info("cleaned expired share tokens", "count", count)
	}
}
