package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/trip/interfaces/http/handlers"
)

type TripContainer struct {
	Handler *handlers.TripHandler
}

func WireTrip(pgPool *pgxpool.Pool) *TripContainer {
	// Repositories
	tripRepo := postgres.NewTripRepository(pgPool)
	tripOfferRepo := postgres.NewTripOfferRepository(pgPool)

	// Domain services
	fareCalc := services.NewFareCalculator()
	stateMachine := services.NewStateMachine()

	// Use cases
	createTripUC := usecases.NewCreateTrip(tripRepo, fareCalc)
	getTripUC := usecases.NewGetTrip(tripRepo)
	getNearbyTripsUC := usecases.NewGetNearbyTrips(tripRepo)
	submitOfferUC := usecases.NewSubmitTripOffer(tripRepo, tripOfferRepo, stateMachine)
	getOffersUC := usecases.NewGetTripOffers(tripRepo, tripOfferRepo)
	acceptOfferUC := usecases.NewAcceptTripOffer(tripRepo, tripOfferRepo, stateMachine)
	confirmAssignmentUC := usecases.NewConfirmTripAssignment(tripRepo, stateMachine)

	// Handler
	handler := handlers.NewTripHandler(
		createTripUC,
		getTripUC,
		getNearbyTripsUC,
		submitOfferUC,
		getOffersUC,
		acceptOfferUC,
		confirmAssignmentUC,
	)

	return &TripContainer{
		Handler: handler,
	}
}
