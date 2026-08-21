package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
)

const payoutColumns = `id, user_id, amount, currency, status, bank_name, bank_account_number,
	bank_account_name, paystack_transfer_reference, paystack_transfer_id, failure_reason,
	requested_at, processed_at, created_at, updated_at`

func scanPayout(rs interface{ Scan(dest ...any) error }) (*entities.Payout, error) {
	p := &entities.Payout{}
	if err := rs.Scan(
		&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.BankName,
		&p.BankAccountNumber, &p.BankAccountName, &p.PaystackTransferRef,
		&p.PaystackTransferID, &p.FailureReason, &p.RequestedAt, &p.ProcessedAt,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return p, nil
}

type PayoutRepository struct {
	pool *pgxpool.Pool
}

func NewPayoutRepository(pool *pgxpool.Pool) *PayoutRepository {
	return &PayoutRepository{pool: pool}
}

func (r *PayoutRepository) Create(ctx context.Context, payout *entities.Payout) error {
	query := `INSERT INTO payouts (` + payoutColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`

	_, err := r.pool.Exec(ctx, query,
		payout.ID, payout.UserID, payout.Amount, payout.Currency, payout.Status,
		payout.BankName, payout.BankAccountNumber, payout.BankAccountName,
		payout.PaystackTransferRef, payout.PaystackTransferID, payout.FailureReason,
		payout.RequestedAt, payout.ProcessedAt, payout.CreatedAt, payout.UpdatedAt,
	)
	return err
}

func (r *PayoutRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Payout, error) {
	query := `SELECT ` + payoutColumns + ` FROM payouts WHERE id = $1`

	payout, err := scanPayout(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return payout, nil
}

func (r *PayoutRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Payout, error) {
	query := `SELECT ` + payoutColumns + ` FROM payouts
		WHERE user_id = $1 ORDER BY requested_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []*entities.Payout
	for rows.Next() {
		payout, err := scanPayout(rows)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, payout)
	}
	return payouts, rows.Err()
}

func (r *PayoutRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.PayoutStatus, failureReason *string) error {
	query := `UPDATE payouts SET status = $1, failure_reason = $2, processed_at = NOW(), updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, failureReason, id)
	return err
}

func (r *PayoutRepository) UpdatePaystackFields(ctx context.Context, id uuid.UUID, transferRef string, transferID string) error {
	query := `UPDATE payouts SET paystack_transfer_reference = $1, paystack_transfer_id = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, transferRef, transferID, id)
	return err
}
