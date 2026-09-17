package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	invoiceapplication "github.com/lucas/financial-api/internal/cardinvoice/application"
	invoicerepository "github.com/lucas/financial-api/internal/cardinvoice/repository"
	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	cardrepository "github.com/lucas/financial-api/internal/creditcard/repository"
	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutionrepository "github.com/lucas/financial-api/internal/financialinstitution/repository"
	"github.com/lucas/financial-api/internal/financialitem"
	adjustmentapplication "github.com/lucas/financial-api/internal/invoiceadjustment/application"
	adjustmentrepository "github.com/lucas/financial-api/internal/invoiceadjustment/repository"
	moveapplication "github.com/lucas/financial-api/internal/invoiceallocation/application"
	moverepository "github.com/lucas/financial-api/internal/invoiceallocation/repository"
	"github.com/lucas/financial-api/internal/monthlysummary"
	paymentapplication "github.com/lucas/financial-api/internal/paymentmethod/application"
	paymentrepository "github.com/lucas/financial-api/internal/paymentmethod/repository"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/config"
	"github.com/lucas/financial-api/internal/platform/httpserver"
	"github.com/lucas/financial-api/internal/profile"
	"github.com/lucas/financial-api/internal/savings"
)

func TestRelease2EndToEndIsolation(t *testing.T) {
	databaseURL, supabaseURL, publishableKey := release2Environment(t)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	stamp := time.Now().UnixNano()
	first := signUp(t, supabaseURL, publishableKey, fmt.Sprintf("r2-security-%d-one@example.com", stamp))
	second := signUp(t, supabaseURL, publishableKey, fmt.Sprintf("r2-security-%d-two@example.com", stamp))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `delete from auth.users where id in ($1,$2)`, first.ID, second.ID)
	})

	var logs bytes.Buffer
	api := newRelease2TestServer(t, pool, supabaseURL, publishableKey, slog.New(slog.NewJSONHandler(&logs, nil)))

	createPlanThroughAPI(t, api.URL, first.Token, "Plano confidencial alfa")
	createPlanThroughAPI(t, api.URL, second.Token, "Plano confidencial beta")
	firstInstitution := nestedString(t, doAPI(t, api.URL, first.Token, http.MethodPost, "/v1/financial-institutions", map[string]any{"name": "Banco sigiloso alfa"}, http.StatusCreated), "id")
	secondInstitution := nestedString(t, doAPI(t, api.URL, second.Token, http.MethodPost, "/v1/financial-institutions", map[string]any{"name": "Banco sigiloso beta"}, http.StatusCreated), "id")
	firstCard := createCardThroughAPI(t, api.URL, first.Token, firstInstitution, "Cartão sigiloso alfa", 6, 1)
	secondCard := createCardThroughAPI(t, api.URL, second.Token, secondInstitution, "Cartão sigiloso beta", 6, 1)
	secondItem := createItemThroughAPI(t, api.URL, second.Token, map[string]any{
		"name": "Despesa sigilosa beta", "kind": "fixed_expense",
		"period": map[string]any{"start_month": "2026-01", "end_month": "2026-12", "amount_cents": 987654, "recurrence": "monthly", "cash_month_offset": 0},
	})
	doAPI(t, api.URL, second.Token, http.MethodPost, "/v1/plans/current/items/"+secondItem+"/payment-changes", map[string]any{
		"effective_from": "2026-01", "end_month": "2026-12", "method": "credit_card", "credit_card_id": secondCard,
	}, http.StatusCreated)

	doAPI(t, api.URL, second.Token, http.MethodGet, "/v1/credit-cards/"+firstCard, nil, http.StatusNotFound)
	doAPI(t, api.URL, second.Token, http.MethodGet, "/v1/credit-cards/"+firstCard+"/invoices/2026-02", nil, http.StatusNotFound)
	doAPI(t, api.URL, second.Token, http.MethodPatch, "/v1/financial-institutions/"+firstInstitution, map[string]any{"name": "Tentativa cruzada"}, http.StatusNotFound)
	doAPI(t, api.URL, second.Token, http.MethodPost, "/v1/plans/current/items/"+secondItem+"/payment-changes", map[string]any{
		"effective_from": "2026-01", "end_month": "2026-12", "method": "credit_card", "credit_card_id": firstCard,
	}, http.StatusNotFound)
	doAPI(t, api.URL, second.Token, http.MethodPost, "/v1/plans/current/items/"+secondItem+"/occurrences/2026-01/invoice-moves", map[string]any{
		"target_credit_card_id": firstCard, "target_payment_month": "2026-02", "reason": "Tentativa cruzada",
	}, http.StatusConflict)

	assertDataAPIEmpty(t, supabaseURL, publishableKey, second.Token, "financial_institutions?select=id&id=eq."+url.QueryEscape(firstInstitution))
	assertDataAPIEmpty(t, supabaseURL, publishableKey, second.Token, "credit_cards?select=id&id=eq."+url.QueryEscape(firstCard))
	assertDataAPIEmpty(t, supabaseURL, publishableKey, first.Token, "credit_cards?select=id&id=eq."+url.QueryEscape(secondCard))

	logText := logs.String()
	for _, secret := range []string{"Banco sigiloso alfa", "Banco sigiloso beta", "Cartão sigiloso alfa", "Cartão sigiloso beta", "Despesa sigilosa beta", "987654", first.Token, second.Token} {
		if strings.Contains(logText, secret) {
			t.Fatalf("logs leaked financial or authentication data: %q", secret)
		}
	}
}

func release2Environment(t *testing.T) (string, string, string) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	supabaseURL := strings.TrimRight(os.Getenv("TEST_SUPABASE_URL"), "/")
	publishableKey := os.Getenv("TEST_SUPABASE_PUBLISHABLE_KEY")
	if databaseURL == "" || supabaseURL == "" || publishableKey == "" {
		t.Skip("TEST_DATABASE_URL, TEST_SUPABASE_URL and TEST_SUPABASE_PUBLISHABLE_KEY are required")
	}
	return databaseURL, supabaseURL, publishableKey
}

func newRelease2TestServer(t *testing.T, pool *pgxpool.Pool, supabaseURL, publishableKey string, logger *slog.Logger) *httptest.Server {
	t.Helper()
	planRepository := planning.NewPostgresRepository(pool)
	plans := planning.NewService(planRepository)
	itemRepository := financialitem.NewPostgresRepository(pool)
	items := financialitem.NewService(itemRepository, plans)
	savingRepository := savings.NewPostgresRepository(pool)
	savingService := savings.NewService(savingRepository, plans)
	invoiceRepository := invoicerepository.NewPostgresRepository(pool)
	institutions := institutionapplication.NewService(institutionrepository.NewPostgresRepository(pool))
	cards := cardapplication.NewService(cardrepository.NewPostgresRepository(pool))
	paymentMethods := paymentapplication.NewService(paymentrepository.NewPostgresRepository(pool), plans)
	moves := moveapplication.NewService(moverepository.NewPostgresRepository(pool), plans)
	adjustments := adjustmentapplication.NewService(adjustmentrepository.NewPostgresRepository(pool), plans)
	invoices := invoiceapplication.NewService(invoiceRepository, plans)
	summaries := monthlysummary.NewService(plans, itemRepository, savingRepository, invoiceRepository)
	server := httpserver.New(config.Config{HTTP: config.HTTPConfig{}}, logger, httpserver.Dependencies{
		Database: pool, Authenticator: auth.NewClient(config.SupabaseConfig{URL: supabaseURL, PublishableKey: publishableKey, AuthTimeout: 5 * time.Second}),
		Profiles: profile.NewRepository(pool), Plans: plans, FinancialItems: items, Savings: savingService, MonthlySummary: summaries,
		Institutions: institutions, CreditCards: cards, PaymentMethods: paymentMethods, InvoiceMoves: moves,
		InvoiceAdjustments: adjustments, CardInvoices: invoices,
	})
	api := httptest.NewServer(server.Handler)
	t.Cleanup(api.Close)
	return api
}

func createCardThroughAPI(t *testing.T, baseURL, token, institutionID, name string, dueDay, offset int) string {
	t.Helper()
	return nestedString(t, doAPI(t, baseURL, token, http.MethodPost, "/v1/credit-cards", map[string]any{
		"institution_id": institutionID, "name": name,
		"configuration": map[string]any{"effective_from": "2026-01", "end_month": "2026-12", "nominal_due_day": dueDay, "payment_month_offset": offset},
	}, http.StatusCreated), "id")
}
