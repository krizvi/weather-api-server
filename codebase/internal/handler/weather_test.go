package handler

import (
	"context"
	"fmt"
	"github.com/krizvi/weather-app-server/internal/service"
	"net/http/httptest"
	"testing"
)

// Mock implementation for testing
type MockWeatherService struct {
	shouldError bool
	returnData  *service.WeatherData
}

func (m *MockWeatherService) GetWeather(ctx context.Context, lat, lon float64) (*service.WeatherData, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.returnData, nil
}

func TestWeatherHandler_Success(t *testing.T) {
	// Create mock service - NO real API calls
	mockService := &MockWeatherService{
		shouldError: false,
		returnData: &service.WeatherData{
			Condition:           "Clear",
			TemperatureCategory: "hot",
		},
	}

	// Test handler with mock - this is where interface matters!
	handler := New(mockService, 10) // Accepts WeatherService interface

	req := httptest.NewRequest("GET", "/weather?lat=40.7&lon=-74.0", nil)
	w := httptest.NewRecorder()

	handler.GetWeather(w, req)

	// Assert response without hitting OpenWeatherMap API
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestWeatherHandler_ServiceError(t *testing.T) {
	// Test error handling
	mockService := &MockWeatherService{shouldError: true}
	handler := New(mockService, 10)

	req := httptest.NewRequest("GET", "/weather?lat=40.7&lon=-74.0", nil)
	w := httptest.NewRecorder()

	handler.GetWeather(w, req)

	// Should return 503 when service fails
	if w.Code != 503 {
		t.Errorf("Expected 503, got %d", w.Code)
	}
}
