package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
)

const callColumns = `id, trip_id, caller_id, receiver_id, status, started_at, ended_at, duration_seconds, created_at`

type CallSessionRepository struct {
	pool *pgxpool.Pool
}

func NewCallSessionRepository(pool *pgxpool.Pool) *CallSessionRepository {
	return &CallSessionRepository{pool: pool}
}

func (r *CallSessionRepository) Create(ctx context.Context, session *entities.CallSession) error {
	query := `INSERT INTO call_sessions (` + callColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		session.ID, session.TripID, session.CallerID, session.ReceiverID,
		session.Status, session.StartedAt, session.EndedAt, session.DurationSeconds, session.CreatedAt,
	)
	return err
}

func (r *CallSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.CallSession, error) {
	query := `SELECT ` + callColumns + ` FROM call_sessions WHERE id = $1`

	session := &entities.CallSession{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&session.ID, &session.TripID, &session.CallerID, &session.ReceiverID,
		&session.Status, &session.StartedAt, &session.EndedAt, &session.DurationSeconds, &session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

func (r *CallSessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.CallStatus) error {
	query := `UPDATE call_sessions SET status = $1 WHERE id = $2`

	// If status becomes CONNECTED, set started_at
	if status == entities.CallStatusConnected {
		query = `UPDATE call_sessions SET status = $1, started_at = NOW() WHERE id = $2`
	}

	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

func (r *CallSessionRepository) EndCall(ctx context.Context, id uuid.UUID, status entities.CallStatus, durationSeconds int) error {
	query := `UPDATE call_sessions SET status = $1, ended_at = NOW(), duration_seconds = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, durationSeconds, id)
	return err
}
