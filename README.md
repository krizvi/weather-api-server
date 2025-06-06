# Weather API Server

Simple Go web server that fetches weather data using OpenWeatherMap API. Built for Jack Henry's Go Developer take-home coding exercise.

## The Assignment

```
Write an http server that uses the Open Weather API that exposes an endpoint that takes in lat/long coordinates. 
This endpoint should return what the weather condition is outside in that area (snow, rain, etc), whether it's hot, 
cold, or moderate outside (use your own discretion on what temperature equates to each type).
```

## What I Built

- HTTP server with `/weather` endpoint
- Takes lat/lon coordinates as query params
- Returns weather condition and temperature category
- Uses only Go standard library (no external deps)
- Proper error handling and input validation

## API Usage

Endpoint: `GET /weather?lat={latitude}&lon={longitude}`

Example Request:
```bash
curl "http://localhost:8080/weather?lat=40.7128&lon=-74.0060"
```

Response:
```json
{
  "ObservationTime": "2025-06-05 20:23:23 EDT",
  "Country": "US",
  "City": "New York",
  "Condition": "Clear",
  "TemperatureCategory": "moderate"
}
```

Temperature Categories (my discretion):
- Cold: Below 50°F
- Moderate: 50°F to 67°F
- Hot: 68°F and above

## Setup & Run

1. Get API key from https://openweathermap.org/api
2. Set env var: `export OPENWEATHER_API_KEY="your-key-here"`
3. Run: `go run main.go`
4. Test: `curl "http://localhost:8080/weather?lat=40.7128&lon=-74.0060"`

## Project Structure

```
├── main.go                    # Server setup and config
├── internal/
│   ├── handler/
│   │   ├── weather.go         # HTTP handlers
│   │   └── weather_test.go    # Handler tests (mocked)
│   ├── service/
│   │   ├── weather_service.go # OpenWeatherMap API client
│   │   └── weather_server_test.go # Service tests (real API)
│   └── utils/
│       └── env.go            # Environment variable helpers
└── README.md
```

## Testing

```bash
# Test handlers (fast, mocked)
go test -count=1 ./internal/handler -v

# Test service (real API calls - needs OPENWEATHER_API_KEY)
go test -count=1 ./internal/service -v

# All tests
go test -count=1 ./... -v
```

## Design Decisions

- I used an interface for the weather service so I can easily test the handler with mock data instead of hitting the real API
- I chose Fahrenheit for temperature because I'm more comfortable with it than Celsius  
- I set context timeouts to 10 seconds for requests and 30 seconds for the HTTP client as a safety backup
- The server shuts down gracefully by waiting for any ongoing requests to finish before stopping
- I only used Go's (Go 1.24) built-in standard library without any external packages (please observe the go.mod and you will find zero dependencies)

## Load Testing

Quick test with 100 requests:
```bash
for i in {1..100}; do curl -s "http://localhost:8080/weather?lat=40.7128&lon=-74.0060" & done; wait
```

Note: You should hit OpenWeatherMap's rate limits (60 calls/minute on free tier).

---

Submitted by: Khalid Rizvi (khalidrizvi@icloud.com)  
Date: June 5, 2025  
Go Version: 1.24