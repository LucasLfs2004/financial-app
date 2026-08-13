package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthenticateReturnsSupabaseUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer access-token" {
			t.Fatalf("unexpected authorization header %q", request.Header.Get("Authorization"))
		}
		if request.Header.Get("apikey") != "publishable-key" {
			t.Fatalf("unexpected api key %q", request.Header.Get("apikey"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"user-123","email":"user@example.com"}`))
	}))
	defer server.Close()

	client := &Client{
		userEndpoint:   server.URL,
		publishableKey: "publishable-key",
		httpClient:     &http.Client{Timeout: time.Second},
	}

	principal, err := client.Authenticate(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if principal.UserID != "user-123" || principal.Email != "user@example.com" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
}

func TestAuthenticateRejectsInvalidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := &Client{
		userEndpoint:   server.URL,
		publishableKey: "publishable-key",
		httpClient:     &http.Client{Timeout: time.Second},
	}

	_, err := client.Authenticate(context.Background(), "invalid-token")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestAuthenticateReportsServiceFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		userEndpoint:   server.URL,
		publishableKey: "publishable-key",
		httpClient:     &http.Client{Timeout: time.Second},
	}

	_, err := client.Authenticate(context.Background(), "access-token")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}
