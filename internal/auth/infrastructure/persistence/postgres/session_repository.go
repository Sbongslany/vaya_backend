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

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, session *entities.Session) error {
	query := `
		INSERT INTO auth.sessions (id, user_id, refresh_token_hash, device_id, device_type, device_name, ip_address, user_agent, mfa_verified, created_at, last_used_at, expires_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserID, session.RefreshTokenHash,
		session.DeviceID, session.DeviceType, session.DeviceName,
		session.IPAddress, session.UserAgent, session.MFAVerified,
		session.CreatedAt, session.LastUsedAt, session.ExpiresAt, session.RevokedAt,
	)
	return err
}

func (r *SessionRepository) FindByRefreshTokenHash(ctx context.Context, hash string) (*entities.Session, error) {
	session := &entities.Session{}
	query := `
		SELECT id, user_id, refresh_token_hash, device_id, device_type, device_name, ip_address, user_agent, mfa_verified, created_at, last_used_at, expires_at, revoked_at
		FROM auth.sessions WHERE refresh_token_hash = $1
	`
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&session.ID, &session.UserID, &session.RefreshTokenHash,
		&session.DeviceID, &session.DeviceType, &session.DeviceName,
		&session.IPAddress, &session.UserAgent, &session.MFAVerified,
		&session.CreatedAt, &session.LastUsedAt, &session.ExpiresAt, &session.RevokedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	return session, err
}

func (r *SessionRepository) Update(ctx context.Context, session *entities.Session) error {
	query := `
		UPDATE auth.sessions 
		SET refresh_token_hash = $1, last_used_at = $2, expires_at = $3 
		WHERE id = $4
	`
	_, err := r.pool.Exec(ctx, query,
		session.RefreshTokenHash,
		session.LastUsedAt,
		session.ExpiresAt,
		session.ID,
	)
	return err
}

func (r *SessionRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID, lastUsedAt time.Time) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.sessions SET last_used_at = $1 WHERE id = $2", lastUsedAt, id)
	return err
}

func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.sessions SET revoked_at = $1 WHERE id = $2", revokedAt, id)
	return err
}

func (r *SessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL", revokedAt, userID)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
	session := &entities.Session{}
	query := `SELECT id, user_id, refresh_token_hash, expires_at, revoked_at FROM auth.sessions WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &session.ExpiresAt, &session.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	return session, err
}

func (r *SessionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, device_id, device_type, device_name, ip_address, user_agent, mfa_verified, created_at, last_used_at, expires_at, revoked_at
		FROM auth.sessions
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY last_used_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*entities.Session
	for rows.Next() {
		s := &entities.Session{}
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.RefreshTokenHash, &s.DeviceID, &s.DeviceType,
			&s.DeviceName, &s.IPAddress, &s.UserAgent, &s.MFAVerified,
			&s.CreatedAt, &s.LastUsedAt, &s.ExpiresAt, &s.RevokedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM auth.sessions WHERE expires_at < NOW() AND revoked_at IS NOT NULL")
	return err
}
