package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type WaypointRepository struct {
	pool *pgxpool.Pool
}

func NewWaypointRepository(pool *pgxpool.Pool) *WaypointRepository {
	return &WaypointRepository{pool: pool}
}

func (r *WaypointRepository) CreateMany(ctx context.Context, waypoints []*entities.Waypoint) error {
	if len(waypoints) == 0 {
		return nil
	}

	query := `INSERT INTO trip_waypoints (id, trip_id, sequence, latitude, longitude, address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	batch := &pgx.Batch{}
	for _, wp := range waypoints {
		batch.Queue(query, wp.ID, wp.TripID, wp.Sequence, wp.Latitude, wp.Longitude, wp.Address, wp.CreatedAt)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(waypoints); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *WaypointRepository) FindByTripID(ctx context.Context, tripID uuid.UUID) ([]*entities.Waypoint, error) {
	query := `SELECT id, trip_id, sequence, latitude, longitude, address, created_at
		FROM trip_waypoints WHERE trip_id = $1 ORDER BY sequence ASC`

	rows, err := r.pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var waypoints []*entities.Waypoint
	for rows.Next() {
		wp := &entities.Waypoint{}
		if err := rows.Scan(&wp.ID, &wp.TripID, &wp.Sequence, &wp.Latitude, &wp.Longitude, &wp.Address, &wp.CreatedAt); err != nil {
			return nil, err
		}
		waypoints = append(waypoints, wp)
	}
	return waypoints, rows.Err()
}
