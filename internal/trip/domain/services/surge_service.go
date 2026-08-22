package services

import "context"

type SurgeZone struct {
	ZoneKey    string  `json:"zone_key"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Demand     int64   `json:"demand"`
	Multiplier float64 `json:"multiplier"`
}

type SurgeService interface {
	RecordDemand(ctx context.Context, lat, lng float64) error
	GetMultiplier(ctx context.Context, lat, lng float64) (float64, error)
	GetHeatmap(ctx context.Context) ([]SurgeZone, error)
}
