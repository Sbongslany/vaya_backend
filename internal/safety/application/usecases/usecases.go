package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	tripEntities "github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	tripRepos "github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"

	"github.com/yourorg/ehailing/backend/internal/safety/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/safety/domain/repositories"
)

var (
	ErrTripNotActive         = errors.New("trip_is_not_active")
	ErrSOSAlreadyActive      = errors.New("sos_already_active")
	ErrInvalidOrExpiredToken = errors.New("invalid_or_expired_token")
)

// --- Trigger SOS ---

type TriggerSOS struct {
	sosRepo  repositories.SOSRepository
	tripRepo tripRepos.TripRepository
}

func NewTriggerSOS(sosRepo repositories.SOSRepository, tripRepo tripRepos.TripRepository) *TriggerSOS {
	return &TriggerSOS{sosRepo: sosRepo, tripRepo: tripRepo}
}

func (uc *TriggerSOS) Execute(ctx context.Context, tripID, userID uuid.UUID) (*entities.SOSAlert, error) {
	trip, err := uc.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if trip == nil || !isTripActive(trip.Status) {
		return nil, ErrTripNotActive
	}

	// Check if SOS is already active
	existing, err := uc.sosRepo.GetActiveByTripID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSOSAlreadyActive
	}

	alert := &entities.SOSAlert{
		ID:          uuid.New(),
		TripID:      tripID,
		TriggeredBy: userID,
		Status:      entities.SOSStatusActive,
		TriggeredAt: time.Now(),
	}

	if err := uc.sosRepo.Create(ctx, alert); err != nil {
		return nil, err
	}

	// TODO: Integrate with FCM/Email service to alert admins and trusted contacts

	return alert, nil
}

// --- Generate Share Link ---

type GenerateShareLink struct {
	shareRepo repositories.ShareTokenRepository
	tripRepo  tripRepos.TripRepository
	baseURL   string
}

func NewGenerateShareLink(shareRepo repositories.ShareTokenRepository, tripRepo tripRepos.TripRepository, baseURL string) *GenerateShareLink {
	return &GenerateShareLink{shareRepo: shareRepo, tripRepo: tripRepo, baseURL: baseURL}
}

func (uc *GenerateShareLink) Execute(ctx context.Context, tripID uuid.UUID, durationHours int) (string, error) {
	trip, err := uc.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return "", err
	}
	if trip == nil {
		return "", ErrTripNotActive
	}

	if durationHours <= 0 || durationHours > 48 {
		durationHours = 24 // Default 24 hours
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	now := time.Now()
	shareToken := &entities.TripShareToken{
		ID:        uuid.New(),
		TripID:    tripID,
		Token:     tokenStr,
		ExpiresAt: now.Add(time.Duration(durationHours) * time.Hour),
		CreatedAt: now,
	}

	if err := uc.shareRepo.Create(ctx, shareToken); err != nil {
		return "", err
	}

	return uc.baseURL + "/track/" + tokenStr, nil
}

// --- View Shared Trip (Public/Unauthenticated) ---

type ViewSharedTrip struct {
	shareRepo repositories.ShareTokenRepository
	tripRepo  tripRepos.TripRepository
}

func NewViewSharedTrip(shareRepo repositories.ShareTokenRepository, tripRepo tripRepos.TripRepository) *ViewSharedTrip {
	return &ViewSharedTrip{shareRepo: shareRepo, tripRepo: tripRepo}
}

func (uc *ViewSharedTrip) Execute(ctx context.Context, token string) (*entities.SharedTripView, error) {
	shareToken, err := uc.shareRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if shareToken == nil || time.Now().After(shareToken.ExpiresAt) {
		return nil, ErrInvalidOrExpiredToken
	}

	trip, err := uc.tripRepo.GetByID(ctx, shareToken.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, ErrInvalidOrExpiredToken
	}

	view := &entities.SharedTripView{
		TripID:         trip.ID,
		Status:         string(trip.Status),
		PickupAddress:  trip.PickupAddress,
		DropoffAddress: trip.DropoffAddress,
		RoutePolyline:  trip.RoutePolyline,
		PickupLatLng:   map[string]float64{"lat": trip.PickupLatitude, "lng": trip.PickupLongitude},
		DropoffLatLng:  map[string]float64{"lat": trip.DropoffLatitude, "lng": trip.DropoffLongitude},
	}

	return view, nil
}

// Helper to check if trip is in an active state
func isTripActive(status tripEntities.TripStatus) bool {
	activeStatuses := []tripEntities.TripStatus{
		tripEntities.StatusDriverAssigned,
		tripEntities.StatusDriverEnRoute,
		tripEntities.StatusDriverArrived,
		tripEntities.StatusTripInProgress,
	}
	for _, s := range activeStatuses {
		if status == s {
			return true
		}
	}
	return false
}
