package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

const tripEventColumns = `id, trip_id, event_type, actor_id, from_status, to_status, metadata, created_at`

type TripEventRepository struct {
	pool *pgxpool.Pool
}

func NewTripEventRepository(pool *pgxpool.Pool) *TripEventRepository {
	return &TripEventRepository{pool: pool}
}

func (r *TripEventRepository) Create(ctx context.Context, event *entities.TripEvent) error {
	query := `INSERT INTO trip_events (trip_id, event_type, actor_id, from_status, to_status, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		event.TripID, event.EventType, event.ActorID, event.FromStatus, event.ToStatus,
		event.Metadata, event.CreatedAt,
	)
	return err
}

func (r *TripEventRepository) FindByTripID(ctx context.Context, tripID uuid.UUID) ([]*entities.TripEvent, error) {
	query := `SELECT ` + tripEventColumns + ` FROM trip_events WHERE trip_id = $1 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*entities.TripEvent
	for rows.Next() {
		e := &entities.TripEvent{}
		if err := rows.Scan(
			&e.ID, &e.TripID, &e.EventType, &e.ActorID, &e.FromStatus, &e.ToStatus,
			&e.Metadata, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}