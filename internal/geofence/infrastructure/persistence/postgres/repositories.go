package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/geofence/domain/entities"
)

// --- Geofence Repository ---

type GeofenceRepository struct {
	pool *pgxpool.Pool
}

func NewGeofenceRepository(pool *pgxpool.Pool) *GeofenceRepository {
	return &GeofenceRepository{pool: pool}
}

func (r *GeofenceRepository) Create(ctx context.Context, fence *entities.Geofence) error {
	query := `INSERT INTO geofences (id, name, type, coordinates, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		fence.ID, fence.Name, fence.Type, fence.Coordinates, fence.IsActive, fence.CreatedAt, fence.UpdatedAt,
	)
	return err
}

func (r *GeofenceRepository) List(ctx context.Context, activeOnly bool) ([]*entities.Geofence, error) {
	query := `SELECT id, name, type, coordinates, is_active, created_at, updated_at FROM geofences`
	if activeOnly {
		query += ` WHERE is_active = TRUE`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fences []*entities.Geofence
	for rows.Next() {
		f := &entities.Geofence{}
		if err := rows.Scan(&f.ID, &f.Name, &f.Type, &f.Coordinates, &f.IsActive, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fences = append(fences, f)
	}
	return fences, rows.Err()
}

func (r *GeofenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Geofence, error) {
	query := `SELECT id, name, type, coordinates, is_active, created_at, updated_at FROM geofences WHERE id = $1`

	f := &entities.Geofence{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&f.ID, &f.Name, &f.Type, &f.Coordinates, &f.IsActive, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func (r *GeofenceRepository) FindZonesContainingPoint(ctx context.Context, lat, lng float64) ([]*entities.Geofence, error) {
	// Uses Postgres geometric operator @> to check if POINT is inside POLYGON
	query := `
		SELECT id, name, type, coordinates, is_active, created_at, updated_at
		FROM geofences
		WHERE is_active = TRUE
		AND POLYGON(coordinates) @> POINT($1, $2)
	`

	rows, err := r.pool.Query(ctx, query, lng, lat) // Postgres POINT takes (x, y) which is (lng, lat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fences []*entities.Geofence
	for rows.Next() {
		f := &entities.Geofence{}
		if err := rows.Scan(&f.ID, &f.Name, &f.Type, &f.Coordinates, &f.IsActive, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fences = append(fences, f)
	}
	return fences, rows.Err()
}

// --- Zone Assignment Repository ---

type ZoneAssignmentRepository struct {
	pool *pgxpool.Pool
}

func NewZoneAssignmentRepository(pool *pgxpool.Pool) *ZoneAssignmentRepository {
	return &ZoneAssignmentRepository{pool: pool}
}

func (r *ZoneAssignmentRepository) AssignDriver(ctx context.Context, assignment *entities.ZoneAssignment) error {
	query := `INSERT INTO driver_zone_assignments (id, driver_id, zone_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (driver_id, zone_id) DO UPDATE SET status = $4`
	_, err := r.pool.Exec(ctx, query,
		assignment.ID, assignment.DriverID, assignment.ZoneID, assignment.Status, assignment.CreatedAt,
	)
	return err
}

func (r *ZoneAssignmentRepository) RemoveDriver(ctx context.Context, driverID, zoneID uuid.UUID) error {
	query := `DELETE FROM driver_zone_assignments WHERE driver_id = $1 AND zone_id = $2`
	_, err := r.pool.Exec(ctx, query, driverID, zoneID)
	return err
}

func (r *ZoneAssignmentRepository) GetDriverAssignments(ctx context.Context, driverID uuid.UUID) ([]*entities.ZoneAssignment, error) {
	query := `SELECT id, driver_id, zone_id, status, created_at FROM driver_zone_assignments WHERE driver_id = $1`

	rows, err := r.pool.Query(ctx, query, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*entities.ZoneAssignment
	for rows.Next() {
		a := &entities.ZoneAssignment{}
		if err := rows.Scan(&a.ID, &a.DriverID, &a.ZoneID, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}
