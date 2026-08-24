package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/admin/domain/entities"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

// ==========================================
// PHASE 17: EXISTING DASHBOARD IMPLEMENTATIONS
// ==========================================

func (r *AdminRepository) GetPlatformStats(ctx context.Context) (*entities.PlatformStats, error) {
	stats := &entities.PlatformStats{}
	query := `
		SELECT 
			(SELECT COUNT(*) FROM auth.users) as total_users,
			(SELECT COUNT(*) FROM auth.users WHERE role = 'DRIVER') as total_drivers,
			(SELECT COUNT(*) FROM auth.users WHERE role = 'PASSENGER') as total_passengers,
			(SELECT COUNT(*) FROM trips) as total_trips,
			(SELECT COUNT(*) FROM trips WHERE status IN ('REQUESTED', 'DRIVER_ASSIGNED', 'DRIVER_EN_ROUTE', 'ARRIVED_AT_PICKUP', 'IN_PROGRESS')) as active_trips,
			COALESCE((SELECT SUM(final_fare) FROM trips WHERE status IN ('PAYMENT_COMPLETED', 'TRIP_COMPLETED')), 0) as total_revenue
	`
	err := r.pool.QueryRow(ctx, query).Scan(
		&stats.TotalUsers, &stats.TotalDrivers, &stats.TotalPassengers,
		&stats.TotalTrips, &stats.ActiveTrips, &stats.TotalRevenue,
	)
	if err != nil {
		return nil, err
	}
	stats.TotalCommission = stats.TotalRevenue * 0.20 
	return stats, nil
}

func (r *AdminRepository) GetFinancialSummary(ctx context.Context, startDate, endDate time.Time) (*entities.FinancialSummary, error) {
	summary := &entities.FinancialSummary{}
	query := `
		SELECT 
			COALESCE(SUM(final_fare), 0),
			COALESCE(SUM(final_fare * 0.20), 0),
			0, 
			0  
		FROM trips
		WHERE status IN ('PAYMENT_COMPLETED', 'TRIP_COMPLETED')
	`
	args := []interface{}{}
	if !startDate.IsZero() && !endDate.IsZero() {
		query += ` AND created_at BETWEEN $1 AND $2`
		args = append(args, startDate, endDate)
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&summary.TotalGrossFare, &summary.TotalCommission,
		&summary.TotalDriverPayouts, &summary.TotalRefunds,
	)
	return summary, err
}

func (r *AdminRepository) ListUsers(ctx context.Context, limit, offset int, role, status string) ([]*entities.UserSummary, error) {
	query := `SELECT id, email, role, status, created_at FROM auth.users WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if role != "" {
		query += fmt.Sprintf(" AND role = $%d", argID)
		args = append(args, role)
		argID++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, status)
		argID++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
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

func (r *AdminRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status entities.UserStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE auth.users SET status = $1, updated_at = NOW() WHERE id = $2`, status, userID)
	return err
}

func (r *AdminRepository) ListAllTrips(ctx context.Context, limit, offset int, status string) ([]*entities.TripSummary, error) {
	query := `SELECT id, passenger_id, driver_id, status, trip_type, estimated_fare, final_fare, created_at FROM trips WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, status)
		argID++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
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

// ==========================================
// PHASE B: NEW LIVE OPERATIONS IMPLEMENTATIONS
// ==========================================

func (r *AdminRepository) ListActiveTrips(ctx context.Context) ([]*entities.LiveTrip, error) {
	query := `
		SELECT id, status, passenger_id, driver_id, pickup_address, dropoff_address,
			pickup_latitude, pickup_longitude, dropoff_latitude, dropoff_longitude,
			estimated_fare, created_at
		FROM trips
		WHERE status IN ('REQUESTED', 'DRIVER_ASSIGNED', 'DRIVER_EN_ROUTE', 'ARRIVED_AT_PICKUP', 'IN_PROGRESS', 'TRIP_COMPLETED')
		ORDER BY created_at DESC LIMIT 200
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*entities.LiveTrip
	for rows.Next() {
		t := &entities.LiveTrip{}
		if err := rows.Scan(&t.ID, &t.Status, &t.PassengerID, &t.DriverID, &t.PickupAddress, &t.DropoffAddress,
			&t.PickupLat, &t.PickupLng, &t.DropoffLat, &t.DropoffLng, &t.EstimatedFare, &t.CreatedAt); err != nil {
			return nil, err
		}
		trips = append(trips, t)
	}
	return trips, rows.Err()
}

func (r *AdminRepository) ListOnlineDrivers(ctx context.Context) ([]*entities.LiveDriver, error) {
	query := `
		SELECT u.id, u.email, 'ONLINE' as status, 0.0 as latitude, 0.0 as longitude
		FROM auth.users u
		WHERE u.role = 'DRIVER' AND u.status = 'ACTIVE'
		LIMIT 200
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []*entities.LiveDriver
	for rows.Next() {
		d := &entities.LiveDriver{}
		if err := rows.Scan(&d.ID, &d.Email, &d.Status, &d.Latitude, &d.Longitude); err != nil {
			return nil, err
		}
		drivers = append(drivers, d)
	}
	return drivers, rows.Err()
}

func (r *AdminRepository) ForceCancelTrip(ctx context.Context, tripID uuid.UUID, reason string) error {
	_, err := r.pool.Exec(ctx, `UPDATE trips SET status = 'CANCELLED', cancellation_reason = $1, updated_at = NOW() WHERE id = $2`, reason, tripID)
	return err
}

func (r *AdminRepository) ForceCompleteTrip(ctx context.Context, tripID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE trips SET status = 'TRIP_COMPLETED', updated_at = NOW() WHERE id = $1`, tripID)
	return err
}

func (r *AdminRepository) ListActiveSOS(ctx context.Context) ([]*entities.LiveSOS, error) {
	query := `SELECT id, trip_id, triggered_by, status, triggered_at FROM sos_alerts WHERE status = 'ACTIVE' ORDER BY triggered_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*entities.LiveSOS
	for rows.Next() {
		a := &entities.LiveSOS{}
		if err := rows.Scan(&a.ID, &a.TripID, &a.TriggeredBy, &a.Status, &a.TriggeredAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

func (r *AdminRepository) CreateAuditLog(ctx context.Context, log *entities.AdminAuditLog) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO admin_audit_logs (id, admin_id, action, resource_type, resource_id, details, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		log.ID, log.AdminID, log.Action, log.ResourceType, log.ResourceID, log.Details, log.CreatedAt)
	return err
}

// ==========================================
// PHASE C: PAYOUT APPROVALS IMPLEMENTATIONS
// ==========================================

func (r *AdminRepository) ListPendingPayouts(ctx context.Context) ([]*entities.PayoutSummary, error) {
	query := `
		SELECT id, driver_id, amount, bank_name, account_number, status, created_at 
		FROM payouts 
		WHERE status = 'PENDING' 
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []*entities.PayoutSummary
	for rows.Next() {
		p := &entities.PayoutSummary{}
		if err := rows.Scan(&p.ID, &p.DriverID, &p.Amount, &p.BankName, &p.AccountNumber, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		payouts = append(payouts, p)
	}
	return payouts, rows.Err()
}

func (r *AdminRepository) GetPayoutByID(ctx context.Context, payoutID uuid.UUID) (*entities.PayoutDetails, error) {
	query := `SELECT id, driver_id, amount, bank_code, account_number, status FROM payouts WHERE id = $1`
	p := &entities.PayoutDetails{}
	err := r.pool.QueryRow(ctx, query, payoutID).Scan(&p.ID, &p.DriverID, &p.Amount, &p.BankCode, &p.AccountNumber, &p.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func (r *AdminRepository) ApprovePayout(ctx context.Context, payoutID uuid.UUID, adminID uuid.UUID) error {
	query := `UPDATE payouts SET status = 'APPROVED', updated_at = NOW() WHERE id = $1 AND status = 'PENDING'`
	_, err := r.pool.Exec(ctx, query, payoutID)
	if err != nil {
		return err
	}
	
	return r.CreateAuditLog(ctx, &entities.AdminAuditLog{
		ID: uuid.New(), AdminID: adminID, Action: "APPROVE_PAYOUT", ResourceType: "PAYOUT", ResourceID: &payoutID, Details: "Payout approved by admin", CreatedAt: time.Now(),
	})
}

func (r *AdminRepository) RejectPayout(ctx context.Context, payoutID uuid.UUID, adminID uuid.UUID, reason string) error {
	query := `UPDATE payouts SET status = 'REJECTED', updated_at = NOW() WHERE id = $1 AND status = 'PENDING'`
	_, err := r.pool.Exec(ctx, query, payoutID)
	if err != nil {
		return err
	}
	
	return r.CreateAuditLog(ctx, &entities.AdminAuditLog{
		ID: uuid.New(), AdminID: adminID, Action: "REJECT_PAYOUT", ResourceType: "PAYOUT", ResourceID: &payoutID, Details: "Rejected: " + reason, CreatedAt: time.Now(),
	})
}