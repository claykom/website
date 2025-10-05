package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original environment variables
	originalEnv := make(map[string]string)
	envVars := []string{"PORT", "HOST", "READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT", "TLS_CERT_FILE", "TLS_KEY_FILE", "ENV", "LOG_LEVEL"}

	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			originalEnv[env] = val
		}
	}

	// Cleanup function to restore original environment
	cleanup := func() {
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		for env, val := range originalEnv {
			os.Setenv(env, val)
		}
	}
	defer cleanup()

	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(*testing.T, *Config)
	}{
		{
			name:        "default configuration",
			envVars:     map[string]string{},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Host != "0.0.0.0" {
					t.Errorf("Expected default host to be 0.0.0.0, got %s", cfg.Server.Host)
				}
				if cfg.Server.Port != 8080 {
					t.Errorf("Expected default port to be 8080, got %d", cfg.Server.Port)
				}
				if cfg.Server.ReadTimeout != 15*time.Second {
					t.Errorf("Expected default read timeout to be 15s, got %v", cfg.Server.ReadTimeout)
				}
				if cfg.TLS.Enabled {
					t.Error("Expected TLS to be disabled by default")
				}
				if cfg.App.Environment != "development" {
					t.Errorf("Expected default environment to be development, got %s", cfg.App.Environment)
				}
			},
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"PORT":          "3000",
				"HOST":          "127.0.0.1",
				"READ_TIMEOUT":  "30s",
				"WRITE_TIMEOUT": "30s",
				"IDLE_TIMEOUT":  "120s",
				"ENV":           "production",
				"LOG_LEVEL":     "error",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Host != "127.0.0.1" {
					t.Errorf("Expected host to be 127.0.0.1, got %s", cfg.Server.Host)
				}
				if cfg.Server.Port != 3000 {
					t.Errorf("Expected port to be 3000, got %d", cfg.Server.Port)
				}
				if cfg.Server.ReadTimeout != 30*time.Second {
					t.Errorf("Expected read timeout to be 30s, got %v", cfg.Server.ReadTimeout)
				}
				if cfg.App.Environment != "production" {
					t.Errorf("Expected environment to be production, got %s", cfg.App.Environment)
				}
				if cfg.App.LogLevel != "error" {
					t.Errorf("Expected log level to be error, got %s", cfg.App.LogLevel)
				}
			},
		},
		{
			name: "TLS enabled configuration",
			envVars: map[string]string{
				"TLS_CERT_FILE": "/path/to/cert.pem",
				"TLS_KEY_FILE":  "/path/to/key.pem",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if !cfg.TLS.Enabled {
					t.Error("Expected TLS to be enabled")
				}
				if cfg.TLS.CertFile != "/path/to/cert.pem" {
					t.Errorf("Expected cert file to be /path/to/cert.pem, got %s", cfg.TLS.CertFile)
				}
				if cfg.TLS.KeyFile != "/path/to/key.pem" {
					t.Errorf("Expected key file to be /path/to/key.pem, got %s", cfg.TLS.KeyFile)
				}
			},
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"PORT": "invalid",
			},
			expectError: true,
		},
		{
			name: "port out of range",
			envVars: map[string]string{
				"PORT": "99999",
			},
			expectError: true,
		},
		{
			name: "invalid timeout",
			envVars: map[string]string{
				"READ_TIMEOUT": "invalid",
			},
			expectError: true,
		},
		{
			name: "negative timeout",
			envVars: map[string]string{
				"READ_TIMEOUT": "-5s",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			for _, env := range envVars {
				os.Unsetenv(env)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg, err := Load()

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if cfg == nil {
				t.Error("Expected config to be non-nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{"valid port", "8080", 8080, false},
		{"minimum port", "1", 1, false},
		{"maximum port", "65535", 65535, false},
		{"invalid string", "invalid", 0, true},
		{"port too low", "0", 0, true},
		{"port too high", "65536", 0, true},
		{"negative port", "-1", 0, true},
		{"empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePort(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Duration
		expectError bool
	}{
		{"valid seconds", "30s", 30 * time.Second, false},
		{"valid minutes", "5m", 5 * time.Minute, false},
		{"valid hours", "2h", 2 * time.Hour, false},
		{"valid mixed", "1h30m", 90 * time.Minute, false},
		{"invalid format", "invalid", 0, true},
		{"negative duration", "-30s", 0, true},
		{"empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "environment variable not set",
			key:          "TEST_VAR_UNSET",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "empty environment variable",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			defer os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkLoad(b *testing.B) {
	// Set up a valid environment
	os.Setenv("PORT", "8080")
	os.Setenv("HOST", "localhost")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load()
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkParsePort(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parsePort("8080")
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkParseDuration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseDuration("30s")
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}
