package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	invoicerepository "github.com/lucas/financial-api/internal/cardinvoice/repository"
	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/monthlysummary"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/config"
	"github.com/lucas/financial-api/internal/platform/httpserver"
	"github.com/lucas/financial-api/internal/profile"
	"github.com/lucas/financial-api/internal/savings"
)

type testIdentity struct {
	ID    string
	Token string
}

type responseEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code string `json:"code"`
	} `json:"error"`
}

func TestRelease1AcceptanceAndIsolation(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	supabaseURL := strings.TrimRight(os.Getenv("TEST_SUPABASE_URL"), "/")
	publishableKey := os.Getenv("TEST_SUPABASE_PUBLISHABLE_KEY")
	if databaseURL == "" || supabaseURL == "" || publishableKey == "" {
		t.Skip("TEST_DATABASE_URL, TEST_SUPABASE_URL and TEST_SUPABASE_PUBLISHABLE_KEY are required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	stamp := time.Now().UnixNano()
	first := signUp(t, supabaseURL, publishableKey, fmt.Sprintf("release1-%d-one@example.com", stamp))
	second := signUp(t, supabaseURL, publishableKey, fmt.Sprintf("release1-%d-two@example.com", stamp))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `delete from auth.users where id in ($1,$2)`, first.ID, second.ID)
	})

	planRepository := planning.NewPostgresRepository(pool)
	planService := planning.NewService(planRepository)
	itemRepository := financialitem.NewPostgresRepository(pool)
	itemService := financialitem.NewService(itemRepository, planService)
	savingRepository := savings.NewPostgresRepository(pool)
	savingService := savings.NewService(savingRepository, planService)
	summaryService := monthlysummary.NewService(planService, itemRepository, savingRepository, invoicerepository.NewPostgresRepository(pool))
	server := httpserver.New(config.Config{HTTP: config.HTTPConfig{}}, slog.New(slog.NewTextHandler(io.Discard, nil)), httpserver.Dependencies{
		Database: pool, Authenticator: auth.NewClient(config.SupabaseConfig{URL: supabaseURL, PublishableKey: publishableKey, AuthTimeout: 5 * time.Second}),
		Profiles: profile.NewRepository(pool), Plans: planService, FinancialItems: itemService, Savings: savingService, MonthlySummary: summaryService,
	})
	api := httptest.NewServer(server.Handler)
	defer api.Close()

	firstPlan := createPlanThroughAPI(t, api.URL, first.Token, "Planejamento principal")
	createPlanThroughAPI(t, api.URL, second.Token, "Planejamento isolado")

	salaryID := createItemThroughAPI(t, api.URL, first.Token, map[string]any{
		"name": "Salário", "kind": "recurring_income",
		"period": map[string]any{"start_month": "2026-01", "end_month": "2026-12", "amount_cents": 600000, "recurrence": "monthly", "cash_month_offset": 0},
	})
	createItemThroughAPI(t, api.URL, first.Token, map[string]any{
		"name": "Despesas fixas", "kind": "fixed_expense",
		"period": map[string]any{"start_month": "2026-01", "end_month": "2026-12", "amount_cents": 210000, "recurrence": "monthly", "cash_month_offset": 0},
	})
	createItemThroughAPI(t, api.URL, first.Token, map[string]any{
		"name": "Gasolina", "kind": "projected_variable_expense",
		"period": map[string]any{"start_month": "2026-01", "end_month": "2026-12", "amount_cents": 70000, "recurrence": "monthly", "cash_month_offset": 0},
	})
	secondItemID := createItemThroughAPI(t, api.URL, second.Token, map[string]any{
		"name": "Renda do segundo usuário", "kind": "recurring_income",
		"period": map[string]any{"start_month": "2026-01", "end_month": "2026-12", "amount_cents": 300000, "recurrence": "monthly", "cash_month_offset": 0},
	})

	doAPI(t, api.URL, first.Token, http.MethodPut, "/v1/plans/current/savings", map[string]any{"effective_from": "2026-01", "end_month": "2026-12", "amount_cents": 120000}, http.StatusOK)
	doAPI(t, api.URL, second.Token, http.MethodPut, "/v1/plans/current/savings", map[string]any{"effective_from": "2026-01", "end_month": "2026-12", "amount_cents": 0}, http.StatusOK)

	assertSummary(t, doAPI(t, api.URL, first.Token, http.MethodGet, "/v1/plans/current/months/2026-07/summary?basis=cash", nil, http.StatusOK), 600000, 280000, 120000, 200000, 4)
	assertSummary(t, doAPI(t, api.URL, second.Token, http.MethodGet, "/v1/plans/current/months/2026-07/summary?basis=cash", nil, http.StatusOK), 300000, 0, 0, 300000, 2)

	doAPI(t, api.URL, second.Token, http.MethodGet, "/v1/plans/current/items/"+salaryID, nil, http.StatusNotFound)
	doAPI(t, api.URL, second.Token, http.MethodPatch, "/v1/plans/current/items/"+salaryID, map[string]any{"name": "Tentativa cruzada"}, http.StatusNotFound)
	doAPI(t, api.URL, first.Token, http.MethodGet, "/v1/plans/current/items/"+secondItemID, nil, http.StatusNotFound)

	assertDataAPIEmpty(t, supabaseURL, publishableKey, second.Token, "plans?id=eq."+url.QueryEscape(firstPlan)+"&select=id")
	assertDataAPIEmpty(t, supabaseURL, publishableKey, second.Token, "financial_items?id=eq."+url.QueryEscape(salaryID)+"&select=id")
	assertDataAPIEmpty(t, supabaseURL, publishableKey, second.Token, "financial_item_periods?financial_item_id=eq."+url.QueryEscape(salaryID)+"&select=id")
	assertDataAPIEmpty(t, supabaseURL, publishableKey, second.Token, "saving_periods?plan_id=eq."+url.QueryEscape(firstPlan)+"&select=id")

	doAPI(t, api.URL, first.Token, http.MethodPost, "/v1/plans/current/items/"+salaryID+"/changes", map[string]any{"effective_from": "2026-08", "end_month": "2026-12", "amount_cents": 650000, "cash_month_offset": 0}, http.StatusCreated)
	assertSummary(t, doAPI(t, api.URL, first.Token, http.MethodGet, "/v1/plans/current/months/2026-07/summary?basis=reference", nil, http.StatusOK), 600000, 280000, 120000, 200000, 4)
	assertSummary(t, doAPI(t, api.URL, first.Token, http.MethodGet, "/v1/plans/current/months/2026-08/summary?basis=reference", nil, http.StatusOK), 650000, 280000, 120000, 250000, 4)

	activation := doAPI(t, api.URL, first.Token, http.MethodPost, "/v1/plans/current/activate", nil, http.StatusOK)
	originalID := nestedString(t, activation, "original", "id")
	if nestedString(t, activation, "plan", "status") != "active" || originalID == "" {
		t.Fatalf("invalid activation: %s", activation)
	}
	repeated := doAPI(t, api.URL, first.Token, http.MethodPost, "/v1/plans/current/activate", nil, http.StatusOK)
	if nestedString(t, repeated, "original", "id") != originalID {
		t.Fatalf("activation created another original: first=%s repeated=%s", activation, repeated)
	}
	original := doAPI(t, api.URL, first.Token, http.MethodGet, "/v1/plans/current/original", nil, http.StatusOK)
	if nestedString(t, original, "id") != originalID {
		t.Fatalf("original lookup mismatch: %s", original)
	}
}

func signUp(t *testing.T, baseURL, key, email string) testIdentity {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"email": email, "password": "Release1-test-123!"})
	request, err := http.NewRequest(http.MethodPost, baseURL+"/auth/v1/signup", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("apikey", key)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil || response.StatusCode != http.StatusOK || body.AccessToken == "" || body.User.ID == "" {
		t.Fatalf("signup status=%d body=%+v error=%v", response.StatusCode, body, err)
	}
	return testIdentity{ID: body.User.ID, Token: body.AccessToken}
}

func createPlanThroughAPI(t *testing.T, baseURL, token, name string) string {
	t.Helper()
	data := doAPI(t, baseURL, token, http.MethodPost, "/v1/plans", map[string]any{"name": name, "start_month": "2026-01", "end_month": "2026-12", "currency_code": "BRL"}, http.StatusCreated)
	return nestedString(t, data, "id")
}

func createItemThroughAPI(t *testing.T, baseURL, token string, payload map[string]any) string {
	t.Helper()
	return nestedString(t, doAPI(t, baseURL, token, http.MethodPost, "/v1/plans/current/items", payload, http.StatusCreated), "id")
}

func doAPI(t *testing.T, baseURL, token, method, path string, payload any, expectedStatus int) json.RawMessage {
	t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var envelope responseEnvelope
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode %s %s: %v", method, path, err)
	}
	if response.StatusCode != expectedStatus {
		t.Fatalf("%s %s status=%d expected=%d error=%+v data=%s", method, path, response.StatusCode, expectedStatus, envelope.Error, envelope.Data)
	}
	return envelope.Data
}

func assertSummary(t *testing.T, data json.RawMessage, income, commitments, saving, result int64, sources int) {
	t.Helper()
	var summary struct {
		Income      int64             `json:"income_cents"`
		Commitments int64             `json:"commitments_cents"`
		Saving      int64             `json:"planned_savings_cents"`
		Result      int64             `json:"result_cents"`
		Sources     []json.RawMessage `json:"sources"`
	}
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Income != income || summary.Commitments != commitments || summary.Saving != saving || summary.Result != result || len(summary.Sources) != sources {
		t.Fatalf("summary=%+v expected income=%d commitments=%d saving=%d result=%d sources=%d", summary, income, commitments, saving, result, sources)
	}
}

func nestedString(t *testing.T, data json.RawMessage, path ...string) string {
	t.Helper()
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	current := value
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("%s is not an object in %s", key, data)
		}
		current = object[key]
	}
	result, _ := current.(string)
	return result
}

func assertDataAPIEmpty(t *testing.T, baseURL, key, token, query string) {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, baseURL+"/rest/v1/"+query, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("apikey", key)
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil || response.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "[]" {
		t.Fatalf("Data API query=%s status=%d body=%s error=%v", query, response.StatusCode, body, err)
	}
}
