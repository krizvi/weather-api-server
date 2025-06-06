package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/krizvi/weather-app-server/internal/service"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// WeatherHandler handles HTTP requests
// will delegate all processing to the service
type WeatherHandler struct {
	weatherService     service.WeatherService
	externalApiTimeout int
}

// New creates a new WeatherHandler instance
func New(weatherService service.WeatherService, externalApiTimeout int) *WeatherHandler {
	return &WeatherHandler{
		weatherService:     weatherService,
		externalApiTimeout: externalApiTimeout,
	}
}

// GetWeather handles GET requests to /weather endpoint
func (wh *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request
	slog.Info("GetWeather", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.String("remote-address", r.RemoteAddr))

	// Only allow GET requests
	if r.Method != http.MethodGet {
		wh.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse and validate query parameters
	lat, lon, err := wh.parseCoordinates(r)
	if err != nil {
		wh.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create context with timeout for the external API call
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(wh.externalApiTimeout)*time.Second)
	defer cancel()

	// Fetch weather data
	weatherData, err := wh.weatherService.GetWeather(ctx, lat, lon)
	if err != nil {
		log.Printf("Error fetching weather data: %v", err)
		wh.sendErrorResponse(w, http.StatusServiceUnavailable, "Unable to fetch weather data")
		return
	}

	// Send successful response
	wh.sendJSONResponse(w, http.StatusOK, weatherData)
	log.Printf("Successfully served weather data for coordinates (%.4f, %.4f)", lat, lon)
}

// parseCoordinates extracts and validates latitude and longitude from query parameters
func (wh *WeatherHandler) parseCoordinates(r *http.Request) (float64, float64, error) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	if latStr == "" || lonStr == "" {
		return 0, 0, fmt.Errorf("lat and lon query parameters are required")
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude value: %s", latStr)
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude value: %s", lonStr)
	}

	// Validate geographical bounds
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude must be between -90 and 90, got: %.4f", lat)
	}

	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("longitude must be between -180 and 180, got: %.4f", lon)
	}

	return lat, lon, nil
}

// sendJSONResponse sends a JSON response with the given status code and data
func (wh *WeatherHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendErrorResponse sends a JSON error response
func (wh *WeatherHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResp := ErrorResponse{Error: message}
	wh.sendJSONResponse(w, statusCode, errorResp)
}

// HealthCheck provides a simple health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
