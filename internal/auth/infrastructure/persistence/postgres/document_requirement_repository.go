package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type DocumentRequirementRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRequirementRepository(pool *pgxpool.Pool) *DocumentRequirementRepository {
	return &DocumentRequirementRepository{pool: pool}
}

func (r *DocumentRequirementRepository) GetAll(ctx context.Context) ([]*entities.DocumentRequirement, error) {
	query := `
		SELECT id, doc_type, is_mandatory, applies_to_vehicle, description, created_at, updated_at
		FROM auth.document_requirements ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requirements []*entities.DocumentRequirement
	for rows.Next() {
		req := &entities.DocumentRequirement{}
		if err := rows.Scan(
			&req.ID, &req.DocType, &req.IsMandatory, &req.AppliesToVehicle,
			&req.Description, &req.CreatedAt, &req.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requirements = append(requirements, req)
	}
	return requirements, nil
}

func (r *DocumentRequirementRepository) UpdateMandatoryStatus(ctx context.Context, docType string, isMandatory bool) error {
	query := `UPDATE auth.document_requirements SET is_mandatory = $1, updated_at = NOW() WHERE doc_type = $2`
	_, err := r.pool.Exec(ctx, query, isMandatory, docType)
	return err
}