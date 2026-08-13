package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment string
	HTTP        HTTPConfig
	Database    DatabaseConfig
	Supabase    SupabaseConfig
}

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	URL            string
	MaxConnections int32
	ConnectTimeout time.Duration
}

type SupabaseConfig struct {
	URL            string
	PublishableKey string
	AuthTimeout    time.Duration
}

func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	port, err := integer(getenv, "PORT", 8080)
	if err != nil {
		return Config{}, err
	}
	if port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be between 1 and 65535")
	}

	maxConnections, err := integer(getenv, "DATABASE_MAX_CONNECTIONS", 10)
	if err != nil {
		return Config{}, err
	}
	if maxConnections < 1 {
		return Config{}, fmt.Errorf("DATABASE_MAX_CONNECTIONS must be greater than zero")
	}

	databaseURL, err := required(getenv, "DATABASE_URL")
	if err != nil {
		return Config{}, err
	}

	supabaseURL, err := required(getenv, "SUPABASE_URL")
	if err != nil {
		return Config{}, err
	}
	if err := validateHTTPURL("SUPABASE_URL", supabaseURL); err != nil {
		return Config{}, err
	}

	publishableKey, err := required(getenv, "SUPABASE_PUBLISHABLE_KEY")
	if err != nil {
		return Config{}, err
	}

	authTimeout, err := duration(getenv, "SUPABASE_AUTH_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment: value(getenv, "APP_ENV", "development"),
		HTTP: HTTPConfig{
			Address:         ":" + strconv.Itoa(port),
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			URL:            databaseURL,
			MaxConnections: int32(maxConnections),
			ConnectTimeout: 10 * time.Second,
		},
		Supabase: SupabaseConfig{
			URL:            strings.TrimRight(supabaseURL, "/"),
			PublishableKey: publishableKey,
			AuthTimeout:    authTimeout,
		},
	}, nil
}

func required(getenv func(string) string, key string) (string, error) {
	result := strings.TrimSpace(getenv(key))
	if result == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return result, nil
}

func value(getenv func(string) string, key, fallback string) string {
	if result := strings.TrimSpace(getenv(key)); result != "" {
		return result
	}
	return fallback
}

func integer(getenv func(string) string, key string, fallback int) (int, error) {
	raw := value(getenv, key, strconv.Itoa(fallback))
	result, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}
	return result, nil
}

func duration(getenv func(string) string, key string, fallback time.Duration) (time.Duration, error) {
	raw := value(getenv, key, fallback.String())
	result, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	if result <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return result, nil
}

func validateHTTPURL(key, raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s must be a valid URL: %w", key, err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%s must use http or https and include a host", key)
	}
	return nil
}
