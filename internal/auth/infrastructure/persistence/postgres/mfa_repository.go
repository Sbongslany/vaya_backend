package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type MFARepository struct {
	pool *pgxpool.Pool
}

func NewMFARepository(pool *pgxpool.Pool) *MFARepository {
	return &MFARepository{pool: pool}
}

func (r *MFARepository) Upsert(ctx context.Context, secret *entities.MFASecret) error {
	query := `
		INSERT INTO auth.mfa_secrets (id, user_id, secret_encrypted, method, is_enabled, confirmed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			secret_encrypted = EXCLUDED.secret_encrypted,
			method = EXCLUDED.method,
			is_enabled = EXCLUDED.is_enabled,
			confirmed_at = EXCLUDED.confirmed_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, query,
		secret.ID, secret.UserID, secret.SecretEncrypted, secret.Method,
		secret.IsEnabled, secret.ConfirmedAt, secret.CreatedAt, secret.UpdatedAt,
	)
	return err
}

func (r *MFARepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.MFASecret, error) {
	secret := &entities.MFASecret{}
	query := `
		SELECT id, user_id, secret_encrypted, method, is_enabled, confirmed_at, created_at, updated_at
		FROM auth.mfa_secrets
		WHERE user_id = $1
	`
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&secret.ID, &secret.UserID, &secret.SecretEncrypted, &secret.Method,
		&secret.IsEnabled, &secret.ConfirmedAt, &secret.CreatedAt, &secret.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMFANotEnabled
	}
	return secret, err
}

func (r *MFARepository) Enable(ctx context.Context, userID uuid.UUID, confirmedAt time.Time) error {
	query := `
		UPDATE auth.mfa_secrets 
		SET is_enabled = TRUE, confirmed_at = $1, updated_at = $1 
		WHERE user_id = $2
	`
	_, err := r.pool.Exec(ctx, query, confirmedAt, userID)
	return err
}

func (r *MFARepository) Disable(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE auth.mfa_secrets 
		SET is_enabled = FALSE, confirmed_at = NULL, updated_at = NOW() 
		WHERE user_id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}