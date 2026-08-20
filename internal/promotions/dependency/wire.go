package dependency

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/promotions/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/services"
	"github.com/yourorg/ehailing/backend/internal/promotions/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/promotions/interfaces/http/handlers"
)

type PromotionsContainer struct {
	AdminHandler     *handlers.AdminPromotionHandler
	PassengerHandler *handlers.PassengerPromotionHandler
	RedeemUC         *usecases.RedeemPromotion
}

func WirePromotions(pgPool *pgxpool.Pool) *PromotionsContainer {
	// Repositories
	promoRepo := postgres.NewPromotionRepository(pgPool)
	redemptionRepo := postgres.NewRedemptionRepository(pgPool)

	// Domain services
	discountCalc := services.NewDiscountCalculator()

	// Use cases — admin
	createPromoUC := usecases.NewCreatePromotion(promoRepo)
	updatePromoUC := usecases.NewUpdatePromotion(promoRepo)
	activatePromoUC := usecases.NewActivatePromotion(promoRepo)
	pausePromoUC := usecases.NewPausePromotion(promoRepo)
	listPromosUC := usecases.NewListPromotions(promoRepo)
	getPromoUC := usecases.NewGetPromotion(promoRepo)

	// Use cases — passenger
	validatePromoUC := usecases.NewValidatePromoCode(promoRepo, redemptionRepo, discountCalc)
	redeemPromoUC := usecases.NewRedeemPromotion(promoRepo, redemptionRepo, discountCalc)
	getRedemptionsUC := usecases.NewGetUserRedemptions(redemptionRepo)

	// Handlers
	adminHandler := handlers.NewAdminPromotionHandler(
		createPromoUC, updatePromoUC, activatePromoUC,
		pausePromoUC, listPromosUC, getPromoUC,
	)
	passengerHandler := handlers.NewPassengerPromotionHandler(
		validatePromoUC, getRedemptionsUC,
	)

	return &PromotionsContainer{
		AdminHandler:     adminHandler,
		PassengerHandler: passengerHandler,
		RedeemUC:         redeemPromoUC,
	}
}

// PromotionRedeemerAdapter adapts the RedeemPromotion use case to the
// trip domain's PromotionRedeemer interface.
type PromotionRedeemerAdapter struct {
	redeemUC *usecases.RedeemPromotion
}

func NewPromotionRedeemerAdapter(redeemUC *usecases.RedeemPromotion) *PromotionRedeemerAdapter {
	return &PromotionRedeemerAdapter{redeemUC: redeemUC}
}

func (a *PromotionRedeemerAdapter) RedeemForTrip(
	ctx context.Context,
	code string,
	userID uuid.UUID,
	tripID uuid.UUID,
	tripFare float64,
) (float64, uuid.UUID, error) {
	result, err := a.redeemUC.Execute(ctx, usecases.RedeemPromotionInput{
		Code:     code,
		UserID:   userID,
		TripID:   tripID,
		TripFare: tripFare,
	})
	if err != nil {
		return 0, uuid.Nil, err
	}
	return result.Discount, result.PromotionID, nil
}
