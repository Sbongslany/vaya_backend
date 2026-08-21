package services

import "context"

// VoiceProvider abstracts the WebRTC/VoIP token generation.
// You can swap the MockProvider with Twilio or Agora later.
type VoiceProvider interface {
	GenerateToken(ctx context.Context, userID string, tripID string, isCaller bool) (string, error)
}
