package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type SubmitRatingInput struct {
	TripID  uuid.UUID
	RaterID uuid.UUID
	Rating  int
	Comment string
}

type SubmitRating struct {
	tripRepo     repositories.TripRepository
	ratingRepo   repositories.TripRatingRepository
	stateMachine *services.StateMachine
}

func NewSubmitRating(
	tripRepo repositories.TripRepository,
	ratingRepo repositories.TripRatingRepository,
	stateMachine *services.StateMachine,
) *SubmitRating {
	return &SubmitRating{
		tripRepo:     tripRepo,
		ratingRepo:   ratingRepo,
		stateMachine: stateMachine,
	}
}

func (uc *SubmitRating) Execute(ctx context.Context, input SubmitRatingInput) (*entities.TripRating, error) {
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	if trip.Status != entities.StatusRatingPending {
		return nil, domain.ErrInvalidStateTransition
	}

	// Determine who is being rated based on who is rating
	var ratedUserID uuid.UUID
	switch {
	case input.RaterID == trip.PassengerID:
		if trip.DriverID == nil {
			return nil, domain.ErrNotTripParticipant
		}
		ratedUserID = *trip.DriverID
	case trip.DriverID != nil && input.RaterID == *trip.DriverID:
		ratedUserID = trip.PassengerID
	default:
		return nil, domain.ErrNotTripParticipant
	}

	if input.Rating < 1 || input.Rating > 5 {
		return nil, domain.ErrInvalidRating
	}

	existing, err := uc.ratingRepo.FindByTripAndRater(ctx, input.TripID, input.RaterID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrAlreadyRated
	}

	now := time.Now()
	rating := &entities.TripRating{
		ID:          uuid.New(),
		TripID:      input.TripID,
		RaterID:     input.RaterID,
		RatedUserID: ratedUserID,
		Rating:      input.Rating,
		Comment:     input.Comment,
		CreatedAt:   now,
	}

	if err := uc.ratingRepo.Create(ctx, rating); err != nil {
		return nil, err
	}

	// Close the trip once both parties have rated
	count, err := uc.ratingRepo.CountByTripID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if count >= 2 {
		if err := uc.stateMachine.Transition(trip.Status, entities.StatusClosed); err != nil {
			return nil, domain.ErrInvalidStateTransition
		}
		if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusClosed); err != nil {
			return nil, err
		}
	}

	return rating, nil
}
