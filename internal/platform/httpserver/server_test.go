package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/config"
	"github.com/lucas/financial-api/internal/profile"
)

type fakeDatabase struct {
	err error
}

func (database fakeDatabase) Ping(context.Context) error {
	return database.err
}

type fakeAuthenticator struct {
	principal auth.Principal
	err       error
	token     string
}

func (authenticator *fakeAuthenticator) Authenticate(_ context.Context, token string) (auth.Principal, error) {
	authenticator.token = token
	return authenticator.principal, authenticator.err
}

type fakeProfiles struct {
	result        profile.Profile
	err           error
	requestedUser string
}

func (profiles *fakeProfiles) FindByID(_ context.Context, userID string) (profile.Profile, error) {
	profiles.requestedUser = userID
	return profiles.result, profiles.err
}

func TestMeUsesAuthenticatedUserIdentity(t *testing.T) {
	authenticator := &fakeAuthenticator{
		principal: auth.Principal{
			UserID: "authenticated-user",
			Email:  "user@example.com",
		},
	}
	profiles := &fakeProfiles{
		result: profile.Profile{
			ID:           "authenticated-user",
			Timezone:     "America/Sao_Paulo",
			CurrencyCode: "BRL",
		},
	}
	server := newTestServer(fakeDatabase{}, authenticator, profiles)

	request := httptest.NewRequest(http.MethodGet, "/v1/me?user_id=another-user", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if authenticator.token != "valid-token" {
		t.Fatalf("expected bearer token, got %q", authenticator.token)
	}
	if profiles.requestedUser != "authenticated-user" {
		t.Fatalf("profile lookup used %q", profiles.requestedUser)
	}

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func TestMeRequiresAuthentication(t *testing.T) {
	server := newTestServer(fakeDatabase{}, &fakeAuthenticator{}, &fakeProfiles{})
	request := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
}

func TestReadyReportsDatabaseFailure(t *testing.T) {
	server := newTestServer(fakeDatabase{err: errors.New("database unavailable")}, &fakeAuthenticator{}, &fakeProfiles{})
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", response.Code)
	}
}

func newTestServer(database Database, authenticator Authenticator, profiles ProfileReader) *http.Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(config.Config{
		HTTP: config.HTTPConfig{},
	}, logger, Dependencies{
		Database:      database,
		Authenticator: authenticator,
		Profiles:      profiles,
	})
}
