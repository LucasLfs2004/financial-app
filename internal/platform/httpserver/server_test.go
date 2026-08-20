package httpserver

import (
	"bytes"
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

	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/monthlysummary"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/config"
	"github.com/lucas/financial-api/internal/profile"
	"github.com/lucas/financial-api/internal/savings"
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

	activationResult planning.Activation
	activationErr    error
	activationOwner  string
	originalResult   planning.Snapshot
	originalErr      error
	originalOwner    string
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
func (plans *fakePlans) Activate(_ context.Context, ownerID string) (planning.Activation, error) {
	plans.activationOwner = ownerID
	return plans.activationResult, plans.activationErr
}
func (plans *fakePlans) Original(_ context.Context, ownerID string) (planning.Snapshot, error) {
	plans.originalOwner = ownerID
	return plans.originalResult, plans.originalErr
}

type fakeFinancialItems struct {
	createOwner  string
	createInput  financialitem.CreateInput
	createResult financialitem.Item
	updateOwner  string
	updateInput  financialitem.UpdateInput
}

func (items *fakeFinancialItems) Create(_ context.Context, ownerID string, input financialitem.CreateInput) (financialitem.Item, error) {
	items.createOwner, items.createInput = ownerID, input
	return items.createResult, nil
}
func (*fakeFinancialItems) List(context.Context, string, financialitem.Filters) ([]financialitem.Item, error) {
	return []financialitem.Item{}, nil
}
func (*fakeFinancialItems) Find(context.Context, string, string) (financialitem.Item, error) {
	return financialitem.Item{}, nil
}
func (items *fakeFinancialItems) Update(_ context.Context, ownerID, _ string, input financialitem.UpdateInput) (financialitem.Item, error) {
	items.updateOwner, items.updateInput = ownerID, input
	return financialitem.Item{ID: "item", PlanID: "plan", Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome, Status: domain.FinancialItemStatusActive, Periods: []financialitem.Period{}}, nil
}
func (*fakeFinancialItems) Change(context.Context, string, string, financialitem.ChangeInput) (financialitem.Item, error) {
	return financialitem.Item{}, nil
}
func (*fakeFinancialItems) Archive(context.Context, string, string, financialitem.ArchiveInput) (financialitem.Item, error) {
	return financialitem.Item{}, nil
}

type fakeSavings struct {
	putOwner  string
	putInput  savings.PutInput
	putResult savings.Configuration
}

type fakeMonthlySummary struct {
	owner  string
	month  domain.YearMonth
	basis  domain.SummaryBasis
	result monthlysummary.Summary
	err    error
}

func (service *fakeMonthlySummary) Get(_ context.Context, ownerID string, month domain.YearMonth, basis domain.SummaryBasis) (monthlysummary.Summary, error) {
	service.owner, service.month, service.basis = ownerID, month, basis
	return service.result, service.err
}

func (*fakeSavings) Get(context.Context, string) (savings.Configuration, error) {
	return savings.Configuration{}, nil
}
func (service *fakeSavings) Put(_ context.Context, ownerID string, input savings.PutInput) (savings.Configuration, error) {
	service.putOwner, service.putInput = ownerID, input
	return service.putResult, nil
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

func TestRequestLogDoesNotExposeAuthorizationOrFinancialQuery(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := loggingMiddleware(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	request := httptest.NewRequest(http.MethodGet, "/v1/plans/current/months/2026-07/summary?amount_cents=600000&name=Salario", nil)
	request.Header.Set("Authorization", "Bearer secret-financial-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	logEntry := output.String()
	for _, sensitive := range []string{"600000", "Salario", "secret-financial-token"} {
		if strings.Contains(logEntry, sensitive) {
			t.Fatalf("request log exposed %q: %s", sensitive, logEntry)
		}
	}
	if !strings.Contains(logEntry, `"path":"/v1/plans/current/months/2026-07/summary"`) {
		t.Fatalf("request path was not logged: %s", logEntry)
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

func TestCreateFinancialItemUsesAuthenticatedUserAndPreservesOffset(t *testing.T) {
	items := &fakeFinancialItems{createResult: financialitem.Item{ID: "item", PlanID: "plan", Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome, Status: domain.FinancialItemStatusActive, Periods: []financialitem.Period{}}}
	server := newTestServerWithServices(authenticatedTestClient(), &fakePlans{}, items, &fakeSavings{})
	request := httptest.NewRequest(http.MethodPost, "/v1/plans/current/items", strings.NewReader(`{"name":"Salário","kind":"recurring_income","period":{"start_month":"2026-01","end_month":"2026-12","amount_cents":600000,"recurrence":"monthly","cash_month_offset":1}}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if items.createOwner != "authenticated-user" || items.createInput.Period.CashMonthOffset != 1 {
		t.Fatalf("owner=%q offset=%d", items.createOwner, items.createInput.Period.CashMonthOffset)
	}
}

func TestCreateFinancialItemRejectsMissingRequiredAmount(t *testing.T) {
	items := &fakeFinancialItems{}
	server := newTestServerWithServices(authenticatedTestClient(), &fakePlans{}, items, &fakeSavings{})
	request := httptest.NewRequest(http.MethodPost, "/v1/plans/current/items", strings.NewReader(`{"name":"Salário","kind":"recurring_income","period":{"start_month":"2026-01","end_month":"2026-12","recurrence":"monthly","cash_month_offset":0}}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || items.createOwner != "" {
		t.Fatalf("status=%d owner=%q body=%s", response.Code, items.createOwner, response.Body.String())
	}
}

func TestPutSavingsPreservesExplicitZero(t *testing.T) {
	service := &fakeSavings{putResult: savings.Configuration{Configured: true, Periods: []savings.Period{}}}
	server := newTestServerWithServices(authenticatedTestClient(), &fakePlans{}, &fakeFinancialItems{}, service)
	request := httptest.NewRequest(http.MethodPut, "/v1/plans/current/savings", strings.NewReader(`{"effective_from":"2026-01","end_month":"2026-12","amount_cents":0}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || service.putOwner != "authenticated-user" || service.putInput.AmountCents != 0 {
		t.Fatalf("status=%d owner=%q amount=%d body=%s", response.Code, service.putOwner, service.putInput.AmountCents, response.Body.String())
	}
}

func TestPutSavingsRejectsMissingRequiredEndMonth(t *testing.T) {
	service := &fakeSavings{}
	server := newTestServerWithServices(authenticatedTestClient(), &fakePlans{}, &fakeFinancialItems{}, service)
	request := httptest.NewRequest(http.MethodPut, "/v1/plans/current/savings", strings.NewReader(`{"effective_from":"2026-01","amount_cents":0}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || service.putOwner != "" {
		t.Fatalf("status=%d owner=%q body=%s", response.Code, service.putOwner, response.Body.String())
	}
}

func TestUpdateFinancialItemCanClearDescription(t *testing.T) {
	items := &fakeFinancialItems{}
	server := newTestServerWithServices(authenticatedTestClient(), &fakePlans{}, items, &fakeSavings{})
	request := httptest.NewRequest(http.MethodPatch, "/v1/plans/current/items/item", strings.NewReader(`{"description":null}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !items.updateInput.DescriptionSet || items.updateInput.Description != nil {
		t.Fatalf("status=%d input=%+v body=%s", response.Code, items.updateInput, response.Body.String())
	}
}

func TestActivatePlanMapsMissingPremisesToConflict(t *testing.T) {
	plans := &fakePlans{activationErr: planning.ErrNotActivatable}
	server := newTestServerWithServices(authenticatedTestClient(), plans, &fakeFinancialItems{}, &fakeSavings{})
	request := httptest.NewRequest(http.MethodPost, "/v1/plans/current/activate", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict || plans.activationOwner != "authenticated-user" {
		t.Fatalf("status=%d owner=%q body=%s", response.Code, plans.activationOwner, response.Body.String())
	}
}

func TestMonthlySummaryDefaultsToCashAndUsesAuthenticatedUser(t *testing.T) {
	month, _ := domain.ParseYearMonth("2026-07")
	service := &fakeMonthlySummary{result: monthlysummary.Summary{Month: month, Basis: domain.SummaryBasisCash, ResultKind: domain.SummaryResultKindPlannedFree, CurrencyCode: "BRL", PlanStatus: domain.PlanStatusDraft, Completeness: domain.CompletenessProjected, Sources: []monthlysummary.Source{}}}
	server := newTestServerWithMonthlySummary(authenticatedTestClient(), service)
	request := httptest.NewRequest(http.MethodGet, "/v1/plans/current/months/2026-07/summary", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || service.owner != "authenticated-user" || service.month.String() != "2026-07" || service.basis != domain.SummaryBasisCash || !strings.Contains(response.Body.String(), `"plan_status":"draft"`) || !strings.Contains(response.Body.String(), `"result_kind":"planned_free"`) {
		t.Fatalf("status=%d owner=%q month=%s basis=%s body=%s", response.Code, service.owner, service.month, service.basis, response.Body.String())
	}
}

func TestMonthlySummaryRejectsInvalidBasis(t *testing.T) {
	service := &fakeMonthlySummary{}
	server := newTestServerWithMonthlySummary(authenticatedTestClient(), service)
	request := httptest.NewRequest(http.MethodGet, "/v1/plans/current/months/2026-07/summary?basis=accrual", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || service.owner != "" {
		t.Fatalf("status=%d owner=%q body=%s", response.Code, service.owner, response.Body.String())
	}
}

func TestMonthlySummaryRejectsInvalidMonth(t *testing.T) {
	service := &fakeMonthlySummary{}
	server := newTestServerWithMonthlySummary(authenticatedTestClient(), service)
	request := httptest.NewRequest(http.MethodGet, "/v1/plans/current/months/2026-13/summary", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || service.owner != "" {
		t.Fatalf("status=%d owner=%q body=%s", response.Code, service.owner, response.Body.String())
	}
}

func TestMonthlySummaryMapsOutsideHorizonToUnprocessableEntity(t *testing.T) {
	service := &fakeMonthlySummary{err: monthlysummary.ErrOutsideHorizon}
	server := newTestServerWithMonthlySummary(authenticatedTestClient(), service)
	request := httptest.NewRequest(http.MethodGet, "/v1/plans/current/months/2027-01/summary?basis=reference", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
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
		Database:       database,
		Authenticator:  authenticator,
		Profiles:       profiles,
		Plans:          plans,
		FinancialItems: nil,
		Savings:        nil,
	})
}

func newTestServerWithServices(authenticator Authenticator, plans PlanService, items FinancialItemService, savingService SavingsService) *http.Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(config.Config{HTTP: config.HTTPConfig{}}, logger, Dependencies{Database: fakeDatabase{}, Authenticator: authenticator, Profiles: &fakeProfiles{}, Plans: plans, FinancialItems: items, Savings: savingService})
}

func newTestServerWithMonthlySummary(authenticator Authenticator, summary MonthlySummaryService) *http.Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(config.Config{HTTP: config.HTTPConfig{}}, logger, Dependencies{Database: fakeDatabase{}, Authenticator: authenticator, Profiles: &fakeProfiles{}, Plans: &fakePlans{}, FinancialItems: &fakeFinancialItems{}, Savings: &fakeSavings{}, MonthlySummary: summary})
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
