package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type TripEventService struct {
	eventRepo   repositories.TripEventRepository
	tripRepo    repositories.TripRepository
	broadcaster EventBroadcaster
	notifier    NotificationService
}

func NewTripEventService(
	eventRepo repositories.TripEventRepository,
	tripRepo repositories.TripRepository,
	broadcaster EventBroadcaster,
	notifier NotificationService,
) *TripEventService {
	return &TripEventService{
		eventRepo:   eventRepo,
		tripRepo:    tripRepo,
		broadcaster: broadcaster,
		notifier:    notifier,
	}
}

func (s *TripEventService) Record(
	ctx context.Context,
	tripID uuid.UUID,
	eventType string,
	actorID *uuid.UUID,
	fromStatus, toStatus string,
	metadata map[string]interface{},
) error {
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON = data
	}

	event := &entities.TripEvent{
		ID:         uuid.New(),
		TripID:     tripID,
		EventType:  eventType,
		ActorID:    actorID,
		FromStatus: &fromStatus,
		ToStatus:   &toStatus,
		Metadata:   metadataJSON,
		CreatedAt:  time.Now(),
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return err
	}

	// 1. Broadcast to WebSocket clients (In-App)
	if s.broadcaster != nil {
		s.broadcaster.Broadcast(tripID, event)
	}

	// 2. Send FCM Push Notification (Background)
	if s.notifier != nil {
		go s.handlePushNotification(tripID, toStatus)
	}

	return nil
}

// handlePushNotification determines who to notify based on the new trip status
func (s *TripEventService) handlePushNotification(tripID uuid.UUID, toStatus string) {
	ctx := context.Background()
	trip, err := s.tripRepo.GetByID(ctx, tripID)
	if err != nil || trip == nil {
		return
	}

	var targetUserID string
	var title, body string

	switch entities.TripStatus(toStatus) {
	case entities.StatusDriverEnRoute:
		targetUserID = trip.PassengerID.String()
		title = "Driver is on the way"
		body = "Your driver is heading to your pickup location."
	case entities.StatusDriverArrived:
		targetUserID = trip.PassengerID.String()
		title = "Driver has arrived"
		body = "Your driver is at the pickup location."
	case entities.StatusOffersReceived:
		targetUserID = trip.PassengerID.String()
		title = "New Trip Offer"
		body = "A driver has submitted an offer for your trip."
	case entities.StatusTripCompleted:
		if trip.DriverID != nil {
			targetUserID = trip.DriverID.String()
			title = "Trip Completed"
			body = "Please wait for the passenger to process payment."
		}
	case entities.StatusPaymentCompleted:
		if trip.DriverID != nil {
			targetUserID = trip.DriverID.String()
			title = "Payment Received"
			body = "The passenger has completed the payment. Please rate your trip."
		}
		targetUserIDPassenger := trip.PassengerID.String()
		_ = s.notifier.SendPushToUser(ctx, targetUserIDPassenger, "Payment Successful", "Your trip has been paid. Please rate your driver.", nil)
	}

	if targetUserID != "" {
		data := map[string]string{"trip_id": tripID.String(), "status": toStatus}
		if err := s.notifier.SendPushToUser(ctx, targetUserID, title, body, data); err != nil {
			log.Printf("FCM push failed for user %s: %v", targetUserID, err)
		}
	}
}
