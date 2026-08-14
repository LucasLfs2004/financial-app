package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
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

type fakePlans struct {
	createResult planning.Plan
	createErr    error
	createOwner  string
	createInput  planning.CreateInput

	currentResult planning.Plan
	currentErr    error
	currentOwner  string

	updateResult planning.Plan
	updateErr    error
	updateOwner  string
	updateInput  planning.UpdateInput
}

func (plans *fakePlans) Create(_ context.Context, ownerID string, input planning.CreateInput) (planning.Plan, error) {
	plans.createOwner = ownerID
	plans.createInput = input
	return plans.createResult, plans.createErr
}

func (plans *fakePlans) Current(_ context.Context, ownerID string) (planning.Plan, error) {
	plans.currentOwner = ownerID
	return plans.currentResult, plans.currentErr
}

func (plans *fakePlans) UpdateCurrent(_ context.Context, ownerID string, input planning.UpdateInput) (planning.Plan, error) {
	plans.updateOwner = ownerID
	plans.updateInput = input
	return plans.updateResult, plans.updateErr
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

func TestCreatePlanUsesAuthenticatedUserIdentity(t *testing.T) {
	plans := &fakePlans{createResult: testPlan(t)}
	server := newTestServerWithPlans(fakeDatabase{}, authenticatedTestClient(), &fakeProfiles{}, plans)
	request := httptest.NewRequest(http.MethodPost, "/v1/plans?user_id=another-user", strings.NewReader(`{
		"name":"Planejamento 2026",
		"start_month":"2026-01",
		"end_month":"2026-12",
		"currency_code":"BRL"
	}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if plans.createOwner != "authenticated-user" {
		t.Fatalf("plan creation owner = %q", plans.createOwner)
	}
	if plans.createInput.StartMonth.String() != "2026-01" || plans.createInput.EndMonth.String() != "2026-12" {
		t.Fatalf("unexpected horizon: %s to %s", plans.createInput.StartMonth, plans.createInput.EndMonth)
	}
}

func TestCurrentPlanUsesAuthenticatedUserIdentity(t *testing.T) {
	plans := &fakePlans{currentResult: testPlan(t)}
	server := newTestServerWithPlans(fakeDatabase{}, authenticatedTestClient(), &fakeProfiles{}, plans)
	request := httptest.NewRequest(http.MethodGet, "/v1/plans/current?user_id=another-user", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if plans.currentOwner != "authenticated-user" {
		t.Fatalf("current plan owner = %q", plans.currentOwner)
	}
}

func TestUpdatePlanRejectsInvalidMonth(t *testing.T) {
	plans := &fakePlans{updateResult: testPlan(t)}
	server := newTestServerWithPlans(fakeDatabase{}, authenticatedTestClient(), &fakeProfiles{}, plans)
	request := httptest.NewRequest(http.MethodPatch, "/v1/plans/current", strings.NewReader(`{"start_month":"2026-13"}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
	if plans.updateOwner != "" {
		t.Fatalf("service should not receive invalid request, got owner %q", plans.updateOwner)
	}
}

func TestUpdatePlanMapsActivePlanToConflict(t *testing.T) {
	plans := &fakePlans{updateErr: planning.ErrNotDraft}
	server := newTestServerWithPlans(fakeDatabase{}, authenticatedTestClient(), &fakeProfiles{}, plans)
	request := httptest.NewRequest(http.MethodPatch, "/v1/plans/current", strings.NewReader(`{"name":"Novo nome"}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", response.Code, response.Body.String())
	}
	if plans.updateOwner != "authenticated-user" {
		t.Fatalf("update owner = %q", plans.updateOwner)
	}
}

func newTestServer(database Database, authenticator Authenticator, profiles ProfileReader) *http.Server {
	return newTestServerWithPlans(database, authenticator, profiles, &fakePlans{})
}

func newTestServerWithPlans(database Database, authenticator Authenticator, profiles ProfileReader, plans PlanService) *http.Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(config.Config{
		HTTP: config.HTTPConfig{},
	}, logger, Dependencies{
		Database:      database,
		Authenticator: authenticator,
		Profiles:      profiles,
		Plans:         plans,
	})
}

func authenticatedTestClient() *fakeAuthenticator {
	return &fakeAuthenticator{principal: auth.Principal{UserID: "authenticated-user", Email: "user@example.com"}}
}

func testPlan(t *testing.T) planning.Plan {
	t.Helper()
	startMonth, err := domain.ParseYearMonth("2026-01")
	if err != nil {
		t.Fatal(err)
	}
	endMonth, err := domain.ParseYearMonth("2026-12")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	return planning.Plan{
		ID:           "00000000-0000-0000-0000-000000000001",
		UserID:       "authenticated-user",
		Name:         "Planejamento 2026",
		Status:       domain.PlanStatusDraft,
		StartMonth:   startMonth,
		EndMonth:     endMonth,
		CurrencyCode: "BRL",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
