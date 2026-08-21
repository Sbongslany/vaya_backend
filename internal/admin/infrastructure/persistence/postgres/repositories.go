package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/admin/domain/entities"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) GetPlatformStats(ctx context.Context) (*entities.PlatformStats, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM auth.users) as total_users,
			(SELECT COUNT(*) FROM auth.users WHERE role = 'DRIVER') as total_drivers,
			(SELECT COUNT(*) FROM auth.users WHERE role = 'PASSENGER') as total_passengers,
			(SELECT COUNT(*) FROM trips) as total_trips,
			(SELECT COUNT(*) FROM trips WHERE status IN ('REQUESTED', 'DRIVER_ASSIGNED', 'IN_PROGRESS', 'DRIVER_EN_ROUTE')) as active_trips,
			(SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'COMPLETED') as total_revenue,
			(SELECT COALESCE(SUM(amount), 0) FROM ledger_entries WHERE reference_type = 'PLATFORM_COMMISSION') as total_commission
	`

	stats := &entities.PlatformStats{}
	err := r.pool.QueryRow(ctx, query).Scan(
		&stats.TotalUsers, &stats.TotalDrivers, &stats.TotalPassengers,
		&stats.TotalTrips, &stats.ActiveTrips, &stats.TotalRevenue, &stats.TotalCommission,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *AdminRepository) GetFinancialSummary(ctx context.Context) (*entities.FinancialSummary, error) {
	query := `
		SELECT
			(SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'COMPLETED') as total_gross_fare,
			(SELECT COALESCE(SUM(amount), 0) FROM ledger_entries WHERE reference_type = 'PLATFORM_COMMISSION') as total_commission,
			(SELECT COALESCE(SUM(amount), 0) FROM payouts WHERE status = 'COMPLETED') as total_driver_payouts,
			(SELECT COALESCE(SUM(amount), 0) FROM ledger_entries WHERE reference_type = 'REFUND') as total_refunds
	`

	summary := &entities.FinancialSummary{}
	err := r.pool.QueryRow(ctx, query).Scan(
		&summary.TotalGrossFare, &summary.TotalCommission, &summary.TotalDriverPayouts, &summary.TotalRefunds,
	)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func (r *AdminRepository) ListUsers(ctx context.Context, role *entities.UserRole, status *entities.UserStatus, limit, offset int) ([]*entities.UserSummary, error) {
	query := `
		SELECT id, email, role, status, created_at
		FROM auth.users
		WHERE ($1::VARCHAR IS NULL OR role = $1)
		AND ($2::VARCHAR IS NULL OR status = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, role, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.UserSummary
	for rows.Next() {
		u := &entities.UserSummary{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *AdminRepository) ListTrips(ctx context.Context, status *string, limit, offset int) ([]*entities.TripSummary, error) {
	query := `
		SELECT id, passenger_id, driver_id, status, trip_type, estimated_fare, final_fare, created_at
		FROM trips
		WHERE ($1::VARCHAR IS NULL OR status = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*entities.TripSummary
	for rows.Next() {
		t := &entities.TripSummary{}
		if err := rows.Scan(&t.ID, &t.PassengerID, &t.DriverID, &t.Status, &t.TripType, &t.EstimatedFare, &t.FinalFare, &t.CreatedAt); err != nil {
			return nil, err
		}
		trips = append(trips, t)
	}
	return trips, rows.Err()
}

func (r *AdminRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status entities.UserStatus) error {
	query := `UPDATE auth.users SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, userID)
	return err
}
