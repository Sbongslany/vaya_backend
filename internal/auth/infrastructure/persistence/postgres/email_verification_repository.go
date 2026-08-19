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

type EmailVerificationRepository struct {
	pool *pgxpool.Pool
}

func NewEmailVerificationRepository(pool *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{pool: pool}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, token *entities.EmailVerificationToken) error {
	query := `
		INSERT INTO auth.email_verification_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *EmailVerificationRepository) FindByHash(ctx context.Context, hash string) (*entities.EmailVerificationToken, error) {
	token := &entities.EmailVerificationToken{}
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM auth.email_verification_tokens
		WHERE token_hash = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.UsedAt, &token.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTokenNotFound
	}
	return token, err
}

func (r *EmailVerificationRepository) MarkUsed(ctx context.Context, id uuid.UUID, usedAt time.Time) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.email_verification_tokens SET used_at = $1 WHERE id = $2", usedAt, id)
	return err
}

func (r *EmailVerificationRepository) InvalidatePrevious(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.email_verification_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL", userID)
	return err
}