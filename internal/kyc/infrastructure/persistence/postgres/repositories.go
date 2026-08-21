package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/kyc/domain/entities"
)

type KYCRepository struct {
	pool *pgxpool.Pool
}

func NewKYCRepository(pool *pgxpool.Pool) *KYCRepository {
	return &KYCRepository{pool: pool}
}

func (r *KYCRepository) ListPendingDrivers(ctx context.Context, limit, offset int) ([]*entities.DriverKYCSummary, error) {
	query := `
		SELECT u.id, u.email, u.onboarding_status,
			COUNT(d.id) as total_docs,
			COUNT(d.id) FILTER (WHERE d.status = 'PENDING') as pending_count,
			COUNT(d.id) FILTER (WHERE d.status = 'APPROVED') as approved_count,
			COUNT(d.id) FILTER (WHERE d.status = 'REJECTED') as rejected_count
		FROM auth.users u
		LEFT JOIN driver_documents d ON u.id = d.user_id
		WHERE u.role = 'DRIVER' AND u.onboarding_status = 'PENDING_REVIEW'
		GROUP BY u.id, u.email, u.onboarding_status
		ORDER BY u.created_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*entities.DriverKYCSummary
	for rows.Next() {
		s := &entities.DriverKYCSummary{}
		if err := rows.Scan(&s.UserID, &s.Email, &s.OnboardingStatus, &s.TotalDocuments, &s.PendingCount, &s.ApprovedCount, &s.RejectedCount); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}

func (r *KYCRepository) GetDocumentsByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.DriverDocument, error) {
	query := `SELECT id, user_id, document_type, file_url, status, rejection_reason, created_at, updated_at
		FROM driver_documents WHERE user_id = $1 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*entities.DriverDocument
	for rows.Next() {
		d := &entities.DriverDocument{}
		if err := rows.Scan(&d.ID, &d.UserID, &d.DocumentType, &d.FileURL, &d.Status, &d.RejectionReason, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *KYCRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*entities.DriverDocument, error) {
	query := `SELECT id, user_id, document_type, file_url, status, rejection_reason, created_at, updated_at
		FROM driver_documents WHERE id = $1`

	d := &entities.DriverDocument{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&d.ID, &d.UserID, &d.DocumentType, &d.FileURL, &d.Status, &d.RejectionReason, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}

func (r *KYCRepository) UpdateDocumentStatus(ctx context.Context, id uuid.UUID, status entities.DocumentStatus, reason *string) error {
	query := `UPDATE driver_documents SET status = $1, rejection_reason = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, reason, id)
	return err
}

func (r *KYCRepository) UpdateUserOnboardingStatus(ctx context.Context, userID uuid.UUID, status entities.OnboardingStatus) error {
	query := `UPDATE auth.users SET onboarding_status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, userID)
	return err
}

func (r *KYCRepository) CountDocumentStatuses(ctx context.Context, userID uuid.UUID) (total, pending, approved, rejected int, err error) {
	query := `
		SELECT
			COUNT(id),
			COUNT(id) FILTER (WHERE status = 'PENDING'),
			COUNT(id) FILTER (WHERE status = 'APPROVED'),
			COUNT(id) FILTER (WHERE status = 'REJECTED')
		FROM driver_documents WHERE user_id = $1
	`
	err = r.pool.QueryRow(ctx, query, userID).Scan(&total, &pending, &approved, &rejected)
	return
}
