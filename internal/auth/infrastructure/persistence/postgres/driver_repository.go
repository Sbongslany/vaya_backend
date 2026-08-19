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

type DriverRepository struct {
	pool *pgxpool.Pool
}

func NewDriverRepository(pool *pgxpool.Pool) *DriverRepository {
	return &DriverRepository{pool: pool}
}

func (r *DriverRepository) CreateProfile(ctx context.Context, profile *entities.DriverProfile) error {
	query := `
		INSERT INTO auth.driver_profiles (id, user_id, license_number, license_expiry, onboarding_step, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		profile.ID, profile.UserID, profile.LicenseNumber, profile.LicenseExpiry,
		profile.OnboardingStep, profile.Status, profile.CreatedAt, profile.UpdatedAt,
	)
	return err
}

func (r *DriverRepository) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*entities.DriverProfile, error) {
	profile := &entities.DriverProfile{}
	query := `
		SELECT id, user_id, license_number, license_expiry, onboarding_step, status, created_at, updated_at
		FROM auth.driver_profiles WHERE user_id = $1
	`
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.LicenseNumber, &profile.LicenseExpiry,
		&profile.OnboardingStep, &profile.Status, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return profile, err
}

func (r *DriverRepository) GetProfileByID(ctx context.Context, id uuid.UUID) (*entities.DriverProfile, error) {
	profile := &entities.DriverProfile{}
	query := `
		SELECT id, user_id, license_number, license_expiry, onboarding_step, status, created_at, updated_at
		FROM auth.driver_profiles WHERE id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&profile.ID, &profile.UserID, &profile.LicenseNumber, &profile.LicenseExpiry,
		&profile.OnboardingStep, &profile.Status, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return profile, err
}

func (r *DriverRepository) UpdateProfile(ctx context.Context, profile *entities.DriverProfile) error {
	query := `
		UPDATE auth.driver_profiles 
		SET license_number = $1, license_expiry = $2, onboarding_step = $3, status = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.pool.Exec(ctx, query,
		profile.LicenseNumber, profile.LicenseExpiry, profile.OnboardingStep, profile.Status, time.Now(), profile.ID,
	)
	return err
}

func (r *DriverRepository) UpdateOnboardingStep(ctx context.Context, profileID uuid.UUID, step string) error {
	query := `UPDATE auth.driver_profiles SET onboarding_step = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, step, profileID)
	return err
}

func (r *DriverRepository) UpdateStatus(ctx context.Context, profileID uuid.UUID, status string) error {
	query := `UPDATE auth.driver_profiles SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, profileID)
	return err
}

func (r *DriverRepository) CreateVehicle(ctx context.Context, vehicle *entities.Vehicle) error {
	query := `
		INSERT INTO auth.vehicles (id, driver_profile_id, make, model, year, color, plate_number, vehicle_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		vehicle.ID, vehicle.DriverProfileID, vehicle.Make, vehicle.Model, vehicle.Year,
		vehicle.Color, vehicle.PlateNumber, vehicle.VehicleType, vehicle.Status, vehicle.CreatedAt, vehicle.UpdatedAt,
	)
	return err
}

func (r *DriverRepository) GetVehiclesByProfileID(ctx context.Context, profileID uuid.UUID) ([]*entities.Vehicle, error) {
	query := `
		SELECT id, driver_profile_id, make, model, year, color, plate_number, vehicle_type, status, created_at, updated_at
		FROM auth.vehicles WHERE driver_profile_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []*entities.Vehicle
	for rows.Next() {
		v := &entities.Vehicle{}
		if err := rows.Scan(
			&v.ID, &v.DriverProfileID, &v.Make, &v.Model, &v.Year, &v.Color,
			&v.PlateNumber, &v.VehicleType, &v.Status, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}
