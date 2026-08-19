package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type DocumentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{pool: pool}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *entities.DriverDocument) error {
	query := `
		INSERT INTO auth.driver_documents (id, driver_profile_id, vehicle_id, doc_type, file_key, file_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		doc.ID, doc.DriverProfileID, doc.VehicleID, doc.DocType,
		doc.FileKey, doc.FileURL, doc.Status, doc.CreatedAt, doc.UpdatedAt,
	)
	return err
}

func (r *DocumentRepository) GetByProfileID(ctx context.Context, profileID uuid.UUID) ([]*entities.DriverDocument, error) {
	query := `
		SELECT id, driver_profile_id, vehicle_id, doc_type, file_key, file_url, status, admin_notes, created_at, updated_at
		FROM auth.driver_documents WHERE driver_profile_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*entities.DriverDocument
	for rows.Next() {
		doc := &entities.DriverDocument{}
		if err := rows.Scan(
			&doc.ID, &doc.DriverProfileID, &doc.VehicleID, &doc.DocType,
			&doc.FileKey, &doc.FileURL, &doc.Status, &doc.AdminNotes, &doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (r *DocumentRepository) UpdateStatus(ctx context.Context, docID uuid.UUID, status string, adminNotes *string) error {
	query := `UPDATE auth.driver_documents SET status = $1, admin_notes = $2, updated_at = $3 WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, status, adminNotes, time.Now(), docID)
	return err
}
