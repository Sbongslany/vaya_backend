package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type GetUserRating struct {
	ratingRepo repositories.UserRatingRepository
}

func NewGetUserRating(ratingRepo repositories.UserRatingRepository) *GetUserRating {
	return &GetUserRating{ratingRepo: ratingRepo}
}

func (uc *GetUserRating) Execute(ctx context.Context, userID uuid.UUID) (*entities.RatingSummary, error) {
	summary, err := uc.ratingRepo.GetRatingSummary(ctx, userID)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return &entities.RatingSummary{
			UserID:      userID,
			RatingAvg:   0,
			RatingCount: 0,
		}, nil
	}
	return summary, nil
}
