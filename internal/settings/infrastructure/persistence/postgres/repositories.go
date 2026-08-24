package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/settings/domain/entities"
)

// --- Vehicle Type Repository ---

type VehicleTypeRepository struct {
	pool *pgxpool.Pool
}

func NewVehicleTypeRepository(pool *pgxpool.Pool) *VehicleTypeRepository {
	return &VehicleTypeRepository{pool: pool}
}

func (r *VehicleTypeRepository) Create(ctx context.Context, v *entities.VehicleType) error {
	query := `INSERT INTO vehicle_types (id, name, slug, base_fare, per_km_rate, per_min_rate, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(ctx, query, v.ID, v.Name, v.Slug, v.BaseFare, v.PerKmRate, v.PerMinRate, v.IsActive, v.CreatedAt)
	return err
}

func (r *VehicleTypeRepository) Update(ctx context.Context, v *entities.VehicleType) error {
	query := `UPDATE vehicle_types SET name = $1, slug = $2, base_fare = $3, per_km_rate = $4, per_min_rate = $5, is_active = $6 WHERE id = $7`
	_, err := r.pool.Exec(ctx, query, v.Name, v.Slug, v.BaseFare, v.PerKmRate, v.PerMinRate, v.IsActive, v.ID)
	return err
}

func (r *VehicleTypeRepository) GetBySlug(ctx context.Context, slug string) (*entities.VehicleType, error) {
	query := `SELECT id, name, slug, base_fare, per_km_rate, per_min_rate, is_active, created_at FROM vehicle_types WHERE slug = $1`

	v := &entities.VehicleType{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(&v.ID, &v.Name, &v.Slug, &v.BaseFare, &v.PerKmRate, &v.PerMinRate, &v.IsActive, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return v, nil
}

func (r *VehicleTypeRepository) List(ctx context.Context) ([]*entities.VehicleType, error) {
	query := `SELECT id, name, slug, base_fare, per_km_rate, per_min_rate, is_active, created_at FROM vehicle_types ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []*entities.VehicleType
	for rows.Next() {
		v := &entities.VehicleType{}
		if err := rows.Scan(&v.ID, &v.Name, &v.Slug, &v.BaseFare, &v.PerKmRate, &v.PerMinRate, &v.IsActive, &v.CreatedAt); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, rows.Err()
}

func (r *VehicleTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM vehicle_types WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// --- Settings Repository ---

type SettingsRepository struct {
	pool *pgxpool.Pool
}

func NewSettingsRepository(pool *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{pool: pool}
}

func (r *SettingsRepository) Get(ctx context.Context) (*entities.PlatformSettings, error) {
	query := `SELECT id, commission_percentage, cancellation_fee, updated_at FROM platform_settings LIMIT 1`

	s := &entities.PlatformSettings{}
	err := r.pool.QueryRow(ctx, query).Scan(&s.ID, &s.CommissionPercentage, &s.CancellationFee, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SettingsRepository) Update(ctx context.Context, s *entities.PlatformSettings) error {
	query := `UPDATE platform_settings SET commission_percentage = $1, cancellation_fee = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, s.CommissionPercentage, s.CancellationFee, s.ID)
	return err
}
