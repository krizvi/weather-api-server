package main

import (
	"context"
	"github.com/krizvi/weather-app-server/internal/handler"
	"github.com/krizvi/weather-app-server/internal/service"
	"github.com/krizvi/weather-app-server/internal/utils"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Config holds configuration for the server including:
// - HTTP server port and timeouts
// - OpenWeather API credentials and endpoint
// - Client timeout for external API calls
type Config struct {
	Port                     string // HTTP server port
	OpenWeatherAPIKey        string // API key for OpenWeather API authentication
	OpenWeatherBaseURL       string // Base URL for OpenWeather API endpoints
	ReadTimeoutSec           int    // Maximum duration for reading request body
	WriteTimeoutSec          int    // Maximum duration for writing response
	IdleTimeoutSec           int    // Maximum duration to wait for the next request when keep-alives are enabled
	ClientTimeoutSec         int    // Timeout for external API client requests
	ServerShutdownTimeoutSec int    // Maximum timeout to allow in-flight requests to complete
}

// loadServerConfig reads configuration from environment variables with the following precedence:
// 1. Required OPENWEATHER_API_KEY must be set
// 2. Optional variables use defaults if not set:
//   - APP_SERVER_PORT (default: 8080)
//   - OPENWEATHER_BASE_URL (default: https://api.openweathermap.org/data/2.5)
//   - APP_SERVER_READ_TIMEOUT_SEC (default: 15)
//   - APP_SERVER_WRITE_TIMEOUT_SEC (default: 15)
//   - APP_SERVER_IDLE_TIMEOUT_SEC (default: 120)
//   - APP_SERVER_CLIENT_TIMEOUT_SEC (default: 10)
//   - APP_SERVER_SHUTDOWN_TIMEOUT_SEC (default: 30)
func loadServerConfig() (*Config, error) {
	apiKey, err := utils.GetEnvAsMustStr("OPENWEATHER_API_KEY", "OPENWEATHER_API_KEY environment variable is required")
	if err != nil {
		return nil, err

	}

	port := utils.GetEnvAsStrWithDefault("APP_SERVER_PORT", "8080")

	baseURL := utils.GetEnvAsStrWithDefault("OPENWEATHER_BASE_URL", "https://api.openweathermap.org/data/2.5")

	ReadTimeoutSec := utils.GetEnvAsIntWithDefault("APP_SERVER_READ_TIMEOUT_SEC", 15)               // don't wait too long for requests
	WriteTimeoutSec := utils.GetEnvAsIntWithDefault("APP_SERVER_WRITE_TIMEOUT_SEC", 15)             // don't hang sending responses
	IdleTimeoutSec := utils.GetEnvAsIntWithDefault("APP_SERVER_IDLE_TIMEOUT_SEC", 120)              // keep connections open for reuse
	ClientTimeoutSec := utils.GetEnvAsIntWithDefault("APP_SERVER_CLIENT_TIMEOUT_SEC", 10)           // timeout for weather API calls
	ServerShutdownTimeoutSec := utils.GetEnvAsIntWithDefault("APP_SERVER_SHUTDOWN_TIMEOUT_SEC", 30) // time to finish requests on shutdown

	return &Config{
		Port:                     port,
		OpenWeatherAPIKey:        apiKey,
		OpenWeatherBaseURL:       baseURL,
		ReadTimeoutSec:           ReadTimeoutSec,
		WriteTimeoutSec:          WriteTimeoutSec,
		IdleTimeoutSec:           IdleTimeoutSec,
		ClientTimeoutSec:         ClientTimeoutSec,
		ServerShutdownTimeoutSec: ServerShutdownTimeoutSec,
	}, nil
}

func main() {
	// Load configuration from environment variables
	config, err := loadServerConfig()
	if err != nil {
		slog.Error("Error", slog.String("Load Config Failed", err.Error()))
		os.Exit(-1)
	}

	// Client timeout (3x request timeout) - safety net if context cancellation fails
	weatherService := service.New(config.OpenWeatherAPIKey, config.OpenWeatherBaseURL, config.ClientTimeoutSec*3)

	// Per-request timeout - normal timeout control
	weatherHandler := handler.New(weatherService, config.ClientTimeoutSec)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/weather", weatherHandler.GetWeather)
	mux.HandleFunc("/health", handler.HealthCheck)

	// Create HTTP server with reasonable timeouts
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  time.Duration(config.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(config.IdleTimeoutSec) * time.Second,
	}

	// Run server in background so main-thread can handle shutdown signals
	go func() {
		log.Printf("Starting server on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup graceful shutdown by listening for interrupt signals (Ctrl+C) or termination requests
	// When signal is received, server stops accepting new connections and waits for existing
	// requests to complete within the timeout period before shutting down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a context with timeout to allow in-flight requests to complete
	// If timeout is reached, remaining connections will be forcefully closed
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.ServerShutdownTimeoutSec)*time.Second)
	defer cancel()

	// Initiate graceful shutdown - waits for existing requests to complete
	// Returns error if shutdown exceeds context timeout
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
