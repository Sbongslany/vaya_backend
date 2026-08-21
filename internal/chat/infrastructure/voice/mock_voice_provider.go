package voice

import (
	"context"
	"fmt"
)

type MockVoiceProvider struct{}

func NewMockVoiceProvider() *MockVoiceProvider {
	return &MockVoiceProvider{}
}

func (m *MockVoiceProvider) GenerateToken(ctx context.Context, userID string, tripID string, isCaller bool) (string, error) {
	role := "publisher"
	if !isCaller {
		role = "subscriber"
	}

	// In production, this will call Twilio/Agora API to generate a JWT
	return fmt.Sprintf("mock-webrtc-token-%s-%s-%s", userID, tripID, role), nil
}
