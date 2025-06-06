package service

import (
	"context"
	"os"
	"testing"
)

// openweathermap_test.go - Tests actual service
func TestOpenWeatherMapService_GetWeather(t *testing.T) {
	// Skip if no API key (for CI/CD)
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test - no API key")
	}

	service := New(apiKey, "https://api.openweathermap.org/data/2.5", 10)

	ctx := context.Background()
	data, err := service.GetWeather(ctx, 40.7128, -74.0060)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.Condition == "" {
		t.Error("Expected condition to be set")
	}

	if data.TemperatureCategory == "" {
		t.Error("Expected temperature category to be set")
	}
}
