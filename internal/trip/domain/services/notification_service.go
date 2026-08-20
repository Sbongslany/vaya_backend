package services

import "context"

// NotificationService abstracts push notifications (FCM, APNs, etc.)
type NotificationService interface {
	SendPushToUser(ctx context.Context, userID string, title string, body string, data map[string]string) error
}
