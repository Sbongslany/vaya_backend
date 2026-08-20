package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
)

const promotionColumns = `id, code, name, description, discount_type, discount_value,
	max_discount_amount, min_trip_fare, usage_limit, used_count, per_user_limit,
	valid_from, valid_until, status, created_by, created_at, updated_at`

func scanPromotion(rs interface{ Scan(dest ...any) error }) (*entities.Promotion, error) {
	p := &entities.Promotion{}
	if err := rs.Scan(
		&p.ID, &p.Code, &p.Name, &p.Description, &p.DiscountType, &p.DiscountValue,
		&p.MaxDiscountAmount, &p.MinTripFare, &p.UsageLimit, &p.UsedCount, &p.PerUserLimit,
		&p.ValidFrom, &p.ValidUntil, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return p, nil
}

type PromotionRepository struct {
	pool *pgxpool.Pool
}

func NewPromotionRepository(pool *pgxpool.Pool) *PromotionRepository {
	return &PromotionRepository{pool: pool}
}

func (r *PromotionRepository) Create(ctx context.Context, promo *entities.Promotion) error {
	query := `INSERT INTO promotions (` + promotionColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`

	_, err := r.pool.Exec(ctx, query,
		promo.ID, promo.Code, promo.Name, promo.Description, promo.DiscountType, promo.DiscountValue,
		promo.MaxDiscountAmount, promo.MinTripFare, promo.UsageLimit, promo.UsedCount, promo.PerUserLimit,
		promo.ValidFrom, promo.ValidUntil, promo.Status, promo.CreatedBy, promo.CreatedAt, promo.UpdatedAt,
	)
	return err
}

func (r *PromotionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Promotion, error) {
	query := `SELECT ` + promotionColumns + ` FROM promotions WHERE id = $1`

	promo, err := scanPromotion(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return promo, nil
}

func (r *PromotionRepository) GetByCode(ctx context.Context, code string) (*entities.Promotion, error) {
	query := `SELECT ` + promotionColumns + ` FROM promotions WHERE code = $1`

	promo, err := scanPromotion(r.pool.QueryRow(ctx, query, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return promo, nil
}

func (r *PromotionRepository) Update(ctx context.Context, promo *entities.Promotion) error {
	query := `UPDATE promotions SET
		code = $1, name = $2, description = $3, discount_type = $4, discount_value = $5,
		max_discount_amount = $6, min_trip_fare = $7, usage_limit = $8, per_user_limit = $9,
		valid_from = $10, valid_until = $11, status = $12, updated_at = NOW()
		WHERE id = $13`

	_, err := r.pool.Exec(ctx, query,
		promo.Code, promo.Name, promo.Description, promo.DiscountType, promo.DiscountValue,
		promo.MaxDiscountAmount, promo.MinTripFare, promo.UsageLimit, promo.PerUserLimit,
		promo.ValidFrom, promo.ValidUntil, promo.Status, promo.ID,
	)
	return err
}

func (r *PromotionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.PromotionStatus) error {
	query := `UPDATE promotions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

func (r *PromotionRepository) IncrementUsedCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE promotions SET used_count = used_count + 1, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PromotionRepository) List(ctx context.Context, status *entities.PromotionStatus, limit, offset int) ([]*entities.Promotion, error) {
	query := `SELECT ` + promotionColumns + ` FROM promotions`
	args := []interface{}{}

	if status != nil {
		query += ` WHERE status = $1`
		args = append(args, *status)
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promos []*entities.Promotion
	for rows.Next() {
		promo, err := scanPromotion(rows)
		if err != nil {
			return nil, err
		}
		promos = append(promos, promo)
	}
	return promos, rows.Err()
}

func (r *PromotionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM promotions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PromotionRepository) TryIncrementUsedCount(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `
		UPDATE promotions
		SET used_count = used_count + 1, updated_at = NOW()
		WHERE id = $1 AND (usage_limit IS NULL OR used_count < usage_limit)
	`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}
