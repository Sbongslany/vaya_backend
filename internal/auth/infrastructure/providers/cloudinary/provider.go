package cloudinary

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/yourorg/ehailing/backend/internal/config"
)

type CloudinaryService struct {
	cfg *config.Config
}

func NewCloudinaryService(cfg *config.Config) *CloudinaryService {
	return &CloudinaryService{cfg: cfg}
}

// GenerateSignature creates a secure SHA1 signature for direct-to-cloud uploads
func (s *CloudinaryService) GenerateSignature(folder string, timestamp int64) string {
	// Cloudinary requires parameters to be concatenated with the API secret
	// Format: folder={folder}&timestamp={timestamp}{api_secret}
	stringToSign := fmt.Sprintf("folder=%s&timestamp=%d%s", folder, timestamp, s.cfg.CloudinaryAPISecret)

	h := sha1.New()
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}
