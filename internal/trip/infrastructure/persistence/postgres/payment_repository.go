package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

const paymentColumns = `id, trip_id, passenger_id, amount, currency, method, status,
	paystack_reference, paystack_authorization_url, created_at, updated_at`

func scanPayment(rs interface{ Scan(dest ...any) error }) (*entities.Payment, error) {
	p := &entities.Payment{}
	if err := rs.Scan(
		&p.ID, &p.TripID, &p.PassengerID, &p.Amount, &p.Currency, &p.Method, &p.Status,
		&p.PaystackReference, &p.PaystackAuthorizationURL, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return p, nil
}

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *entities.Payment) error {
	query := `INSERT INTO payments (` + paymentColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		payment.ID, payment.TripID, payment.PassengerID, payment.Amount, payment.Currency,
		payment.Method, payment.Status, payment.PaystackReference, payment.PaystackAuthorizationURL,
		payment.CreatedAt, payment.UpdatedAt,
	)
	return err
}

func (r *PaymentRepository) GetByTripID(ctx context.Context, tripID uuid.UUID) (*entities.Payment, error) {
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE trip_id = $1`

	payment, err := scanPayment(r.pool.QueryRow(ctx, query, tripID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return payment, nil
}

func (r *PaymentRepository) GetByReference(ctx context.Context, reference string) (*entities.Payment, error) {
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE paystack_reference = $1`

	payment, err := scanPayment(r.pool.QueryRow(ctx, query, reference))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return payment, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

func (r *PaymentRepository) UpdatePaystackFields(ctx context.Context, id uuid.UUID, reference string, authURL string) error {
	query := `UPDATE payments SET paystack_reference = $1, paystack_authorization_url = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, reference, authURL, id)
	return err
}
