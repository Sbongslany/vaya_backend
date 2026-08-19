package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Log(ctx context.Context, log *entities.AuditLog) error {
	query := `
		INSERT INTO auth.audit_logs (id, user_id, action, ip_address, device_id, user_agent, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		log.ID, log.UserID, log.Action, log.IPAddress, log.DeviceID, log.UserAgent, log.Metadata, log.CreatedAt,
	)
	// Audit logging should ideally never break the main flow, 
	// but we return the error so the caller can decide how to handle it.
	return err
}