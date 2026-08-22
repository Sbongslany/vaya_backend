package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type GoogleMapsClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewGoogleMapsClient(apiKey string) *GoogleMapsClient {
	return &GoogleMapsClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type googleDirectionsResponse struct {
	Status string `json:"status"`
	Routes []struct {
		OverviewPolyline struct {
			Points string `json:"points"`
		} `json:"overview_polyline"`
		Legs []struct {
			Distance struct {
				Value int    `json:"value"` // meters
				Text  string `json:"text"`
			} `json:"distance"`
			Duration struct {
				Value int    `json:"value"` // seconds
				Text  string `json:"text"`
			} `json:"duration"`
			Steps []struct {
				HTMLInstructions string `json:"html_instructions"`
				Distance         struct {
					Value int `json:"value"`
				} `json:"distance"`
				Duration struct {
					Value int `json:"value"`
				} `json:"duration"`
			} `json:"steps"`
		} `json:"legs"`
	} `json:"routes"`
}

func (c *GoogleMapsClient) CalculateRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (*services.RouteResult, error) {
	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/directions/json?origin=%f,%f&destination=%f,%f&key=%s",
		fromLat, fromLng, toLat, toLng, c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Maps request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Google Maps request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Google Maps response: %w", err)
	}

	var gResp googleDirectionsResponse
	if err := json.Unmarshal(body, &gResp); err != nil {
		return nil, fmt.Errorf("failed to parse Google Maps response: %w", err)
	}

	if gResp.Status != "OK" || len(gResp.Routes) == 0 {
		return nil, fmt.Errorf("Google Maps returned no route: %s", gResp.Status)
	}

	route := gResp.Routes[0]
	leg := route.Legs[0]

	result := &services.RouteResult{
		DistanceKM:      float64(leg.Distance.Value) / 1000.0,
		DurationMinutes: leg.Duration.Value / 60,
		Polyline:        route.OverviewPolyline.Points,
	}

	for _, step := range leg.Steps {
		result.Steps = append(result.Steps, services.RouteStep{
			Instruction:     step.HTMLInstructions,
			DistanceKM:      float64(step.Distance.Value) / 1000.0,
			DurationMinutes: step.Duration.Value / 60,
		})
	}

	return result, nil
}
