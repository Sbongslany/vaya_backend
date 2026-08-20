package notifications

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/google/uuid"
)

type FirebaseNotificationService struct {
	client    *messaging.Client
	tokenRepo repositories.DeviceTokenRepository
}

func NewFirebaseNotificationService(tokenRepo repositories.DeviceTokenRepository) (*FirebaseNotificationService, error) {
	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS_PATH environment variable not set")
	}

	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase messaging: %v", err)
	}

	return &FirebaseNotificationService{
		client:    client,
		tokenRepo: tokenRepo,
	}, nil
}

func (s *FirebaseNotificationService) SendPushToUser(ctx context.Context, userID string, title string, body string, data map[string]string) error {
	// We need to parse the userID string back to UUID to query the DB
	// For simplicity, we'll just query the string representation if your repo supports it, 
	// but since our repo expects uuid.UUID, let's parse it.
	// import "github.com/google/uuid"
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	tokens, err := s.tokenRepo.FindByUserID(ctx, uid)
	if err != nil || len(tokens) == 0 {
		return nil // Silently fail if user has no devices registered
	}

	for _, t := range tokens {
		message := &messaging.Message{
			Token: t.Token,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
		}

		response, err := s.client.Send(ctx, message)
		if err != nil {
			log.Printf("Failed to send FCM to token %s: %v", t.Token, err)
			// Optional: Delete invalid tokens here
			continue
		}
		log.Printf("Successfully sent FCM message: %s", response)
	}

	return nil
}