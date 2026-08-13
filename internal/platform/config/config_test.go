package config

import (
	"strings"
	"testing"
)

func TestLoadUsesRequiredServicesAndDefaults(t *testing.T) {
	values := map[string]string{
		"DATABASE_URL":             "postgresql://postgres:postgres@localhost:54322/postgres",
		"SUPABASE_URL":             "http://127.0.0.1:54321/",
		"SUPABASE_PUBLISHABLE_KEY": "test-key",
	}

	cfg, err := load(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.HTTP.Address != ":8080" {
		t.Fatalf("expected default address :8080, got %q", cfg.HTTP.Address)
	}
	if cfg.Database.MaxConnections != 10 {
		t.Fatalf("expected 10 database connections, got %d", cfg.Database.MaxConnections)
	}
	if cfg.Supabase.URL != "http://127.0.0.1:54321" {
		t.Fatalf("expected trailing slash removed, got %q", cfg.Supabase.URL)
	}
}

func TestLoadRejectsMissingRequiredValue(t *testing.T) {
	_, err := load(func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL is required") {
		t.Fatalf("expected missing database URL error, got %v", err)
	}
}

func TestLoadRejectsInvalidSupabaseURL(t *testing.T) {
	values := map[string]string{
		"DATABASE_URL":             "postgresql://postgres:postgres@localhost:54322/postgres",
		"SUPABASE_URL":             "supabase.local",
		"SUPABASE_PUBLISHABLE_KEY": "test-key",
	}

	_, err := load(func(key string) string { return values[key] })
	if err == nil || !strings.Contains(err.Error(), "SUPABASE_URL") {
		t.Fatalf("expected invalid Supabase URL error, got %v", err)
	}
}
