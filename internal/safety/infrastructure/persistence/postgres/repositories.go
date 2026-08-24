package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/safety/domain/entities"
)

// --- SOS Repository ---

type SOSRepository struct {
	pool *pgxpool.Pool
}

func NewSOSRepository(pool *pgxpool.Pool) *SOSRepository {
	return &SOSRepository{pool: pool}
}

func (r *SOSRepository) Create(ctx context.Context, alert *entities.SOSAlert) error {
	query := `INSERT INTO sos_alerts (id, trip_id, triggered_by, status, triggered_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, alert.ID, alert.TripID, alert.TriggeredBy, alert.Status, alert.TriggeredAt)
	return err
}

func (r *SOSRepository) GetActiveByTripID(ctx context.Context, tripID uuid.UUID) (*entities.SOSAlert, error) {
	query := `SELECT id, trip_id, triggered_by, status, triggered_at, resolved_at, resolved_by
		FROM sos_alerts WHERE trip_id = $1 AND status = 'ACTIVE'`

	alert := &entities.SOSAlert{}
	err := r.pool.QueryRow(ctx, query, tripID).Scan(&alert.ID, &alert.TripID, &alert.TriggeredBy, &alert.Status, &alert.TriggeredAt, &alert.ResolvedAt, &alert.ResolvedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return alert, nil
}

func (r *SOSRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.SOSStatus, resolvedBy *uuid.UUID) error {
	query := `UPDATE sos_alerts SET status = $1, resolved_at = NOW(), resolved_by = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, resolvedBy, id)
	return err
}

func (r *SOSRepository) ListActive(ctx context.Context) ([]*entities.SOSAlert, error) {
	query := `SELECT id, trip_id, triggered_by, status, triggered_at, resolved_at, resolved_by
		FROM sos_alerts WHERE status = 'ACTIVE' ORDER BY triggered_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*entities.SOSAlert
	for rows.Next() {
		a := &entities.SOSAlert{}
		if err := rows.Scan(&a.ID, &a.TripID, &a.TriggeredBy, &a.Status, &a.TriggeredAt, &a.ResolvedAt, &a.ResolvedBy); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// --- Share Token Repository ---

type ShareTokenRepository struct {
	pool *pgxpool.Pool
}

func NewShareTokenRepository(pool *pgxpool.Pool) *ShareTokenRepository {
	return &ShareTokenRepository{pool: pool}
}

func (r *ShareTokenRepository) Create(ctx context.Context, token *entities.TripShareToken) error {
	query := `INSERT INTO trip_share_tokens (id, trip_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, token.ID, token.TripID, token.Token, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *ShareTokenRepository) GetByToken(ctx context.Context, token string) (*entities.TripShareToken, error) {
	query := `SELECT id, trip_id, token, expires_at, created_at
		FROM trip_share_tokens WHERE token = $1`

	t := &entities.TripShareToken{}
	err := r.pool.QueryRow(ctx, query, token).Scan(&t.ID, &t.TripID, &t.Token, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}
