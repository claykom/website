package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server ServerConfig
	TLS    TLSConfig
	App    AppConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// TLSConfig holds TLS/HTTPS configuration
type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Environment string
	LogLevel    string
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	port, err := parsePort(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	readTimeout, err := parseDuration(getEnv("READ_TIMEOUT", "15s"))
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := parseDuration(getEnv("WRITE_TIMEOUT", "15s"))
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
	}

	idleTimeout, err := parseDuration(getEnv("IDLE_TIMEOUT", "60s"))
	if err != nil {
		return nil, fmt.Errorf("invalid IDLE_TIMEOUT: %w", err)
	}

	// TLS configuration
	tlsCertFile := getEnv("TLS_CERT_FILE", "")
	tlsKeyFile := getEnv("TLS_KEY_FILE", "")
	tlsEnabled := tlsCertFile != "" && tlsKeyFile != ""

	return &Config{
		Server: ServerConfig{
			Host:         getEnv("HOST", "0.0.0.0"),
			Port:         port,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		TLS: TLSConfig{
			Enabled:  tlsEnabled,
			CertFile: tlsCertFile,
			KeyFile:  tlsKeyFile,
		},
		App: AppConfig{
			Environment: getEnv("ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
	}, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parsePort parses a port string into an integer
func parsePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535")
	}
	return port, nil
}

// parseDuration parses a duration string
func parseDuration(durationStr string) (time.Duration, error) {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, err
	}
	if duration < 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	return duration, nil
}
