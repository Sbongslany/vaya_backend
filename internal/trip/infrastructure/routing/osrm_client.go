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

type OSRMClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOSRMClient(baseURL string) *OSRMClient {
	return &OSRMClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type osrmResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"` // meters
		Duration float64 `json:"duration"` // seconds
		Geometry string  `json:"geometry"` // encoded polyline
		Legs     []struct {
			Steps []struct {
				Maneuver struct {
					Type     string `json:"type"`
					Modifier string `json:"modifier"`
				} `json:"maneuver"`
				Name     string  `json:"name"`
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
			} `json:"steps"`
		} `json:"legs"`
	} `json:"routes"`
}

func (c *OSRMClient) CalculateRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (*services.RouteResult, error) {
	url := fmt.Sprintf(
		"%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=polyline&steps=true",
		c.baseURL, fromLng, fromLat, toLng, toLat,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSRM request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OSRM request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OSRM response: %w", err)
	}

	var osrmResp osrmResponse
	if err := json.Unmarshal(body, &osrmResp); err != nil {
		return nil, fmt.Errorf("failed to parse OSRM response: %w", err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return nil, fmt.Errorf("OSRM returned no route: %s", osrmResp.Code)
	}

	route := osrmResp.Routes[0]

	result := &services.RouteResult{
		DistanceKM:      route.Distance / 1000.0,
		DurationMinutes: int(route.Duration / 60.0),
		Polyline:        route.Geometry,
	}

	// Extract steps
	for _, leg := range route.Legs {
		for _, step := range leg.Steps {
			instruction := step.Maneuver.Type
			if step.Maneuver.Modifier != "" {
				instruction += " " + step.Maneuver.Modifier
			}
			if step.Name != "" {
				instruction += " onto " + step.Name
			}

			result.Steps = append(result.Steps, services.RouteStep{
				Instruction:     instruction,
				DistanceKM:      step.Distance / 1000.0,
				DurationMinutes: int(step.Duration / 60.0),
			})
		}
	}

	return result, nil
}
