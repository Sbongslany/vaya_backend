package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
)

const redemptionColumns = `id, promotion_id, user_id, trip_id, discount_applied, redeemed_at`

func scanRedemption(rs interface{ Scan(dest ...any) error }) (*entities.PromotionRedemption, error) {
	r := &entities.PromotionRedemption{}
	if err := rs.Scan(
		&r.ID, &r.PromotionID, &r.UserID, &r.TripID, &r.DiscountApplied, &r.RedeemedAt,
	); err != nil {
		return nil, err
	}
	return r, nil
}

type RedemptionRepository struct {
	pool *pgxpool.Pool
}

func NewRedemptionRepository(pool *pgxpool.Pool) *RedemptionRepository {
	return &RedemptionRepository{pool: pool}
}

func (r *RedemptionRepository) Create(ctx context.Context, redemption *entities.PromotionRedemption) error {
	query := `INSERT INTO promotion_redemptions (` + redemptionColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6)`

	_, err := r.pool.Exec(ctx, query,
		redemption.ID, redemption.PromotionID, redemption.UserID, redemption.TripID,
		redemption.DiscountApplied, redemption.RedeemedAt,
	)
	return err
}

func (r *RedemptionRepository) CountByUserAndPromotion(ctx context.Context, userID, promotionID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM promotion_redemptions WHERE user_id = $1 AND promotion_id = $2`
	var count int
	err := r.pool.QueryRow(ctx, query, userID, promotionID).Scan(&count)
	return count, err
}

func (r *RedemptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.PromotionRedemption, error) {
	query := `SELECT ` + redemptionColumns + ` FROM promotion_redemptions
		WHERE user_id = $1 ORDER BY redeemed_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var redemptions []*entities.PromotionRedemption
	for rows.Next() {
		redemption, err := scanRedemption(rows)
		if err != nil {
			return nil, err
		}
		redemptions = append(redemptions, redemption)
	}
	return redemptions, rows.Err()
}

func (r *RedemptionRepository) FindByTripID(ctx context.Context, tripID uuid.UUID) (*entities.PromotionRedemption, error) {
	query := `SELECT ` + redemptionColumns + ` FROM promotion_redemptions WHERE trip_id = $1`

	redemption, err := scanRedemption(r.pool.QueryRow(ctx, query, tripID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return redemption, nil
}
