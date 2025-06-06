package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// WeatherData represents the weather information we return to clients
type WeatherData struct {
	ObservationTime     string
	Country             string
	City                string
	Condition           string
	TemperatureCategory string
}

// OpenWeatherMapResponse represents the response structure from OpenWeatherMap API
type OpenWeatherMapResponse struct {
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	UnixSeconds int64 `json:"dt"` // this is definitely seconds from Epoch (01011970)
	Location    struct {
		Country string `json:"country"`
	} `json:"sys"`
	Name string `json:"name"`

	// COD is the HTTP status code piggy-backed in the response payload
	// Same as the actual HTTP response status but included in JSON for convenience
	// 200 = success, 429 = rate limited, 401 = bad api key
	// Reference: https://openweathermap.org/appid
	HttpCode int    `json:"cod"`
	Message  string `json:"message,omitempty"` // error details when something goes wrong
}

func (response *OpenWeatherMapResponse) weatherCheckTime() string {
	weatherCheckedTime := time.Unix(response.UnixSeconds, 0)
	return weatherCheckedTime.Format("2006-01-02 15:04:05 MST")
}

// WeatherService defines the interface for weather data retrieval
type WeatherService interface {
	GetWeather(ctx context.Context, lat, lon float64) (*WeatherData, error)
}

// OpenWeatherMapService implements WeatherService using OpenWeatherMap API
type OpenWeatherMapService struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// New creates a new instance of OpenWeatherMapService
func New(apiKey string, baseURL string, timeoutSec int) *OpenWeatherMapService {
	return &OpenWeatherMapService{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			// Global timeout for the entire HTTP request lifecycle
			// (connection + sending + receiving + processing)
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
	}
}

// GetWeather fetches weather data for the given coordinates
func (srv *OpenWeatherMapService) GetWeather(ctx context.Context, lat, lon float64) (*WeatherData, error) {
	// Build the API URL with query parameters
	apiURL, err := srv.buildAPIURL(lat, lon)
	if err != nil {
		return nil, fmt.Errorf("failed to build API URL: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Weather-API-Go/1.0")

	// Make the HTTP request
	resp, err := srv.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var mapResponse OpenWeatherMapResponse
	if err := json.Unmarshal(body, &mapResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check API status (gets detailed error message)
	if mapResponse.HttpCode != 200 {
		return nil, fmt.Errorf("OpenWeatherMap API error (code %d): %s", mapResponse.HttpCode, mapResponse.Message)
	}
	// Convert temperature from Kelvin to Fahrenheit
	tempFahrenheit := (mapResponse.Main.Temp-273.15)*9/5 + 32

	return &WeatherData{
		ObservationTime:     mapResponse.weatherCheckTime(),
		Country:             mapResponse.Location.Country,
		City:                mapResponse.Name,
		Condition:           mapResponse.Weather[0].Main,
		TemperatureCategory: categorizeTemperature(tempFahrenheit),
	}, nil
}

// buildAPIURL constructs the OpenWeatherMap API URL with the given coordinates
func (srv *OpenWeatherMapService) buildAPIURL(lat, lon float64) (string, error) {
	baseURL, err := url.Parse(srv.baseURL + "/weather")
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Add("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	params.Add("appid", srv.apiKey)

	baseURL.RawQuery = params.Encode()
	return baseURL.String(), nil
}

// categorizeTemperature implements the assignment requirement to classify temperature as
// "hot, cold, or moderate" using my discretion for temperature ranges.
// Using Fahrenheit thresholds: 50DegF and 68DegF as reasonable comfort boundaries.
func categorizeTemperature(tempFahrenheit float64) string {
	switch {
	case tempFahrenheit < 50:
		return "cold"
	case tempFahrenheit >= 50 && tempFahrenheit < 68:
		return "moderate"
	default:
		return "hot"
	}
}
