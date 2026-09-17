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

func TestRelease2AcceptanceScenario(t *testing.T) {
	databaseURL, supabaseURL, publishableKey := release2Environment(t)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	stamp := time.Now().UnixNano()
	identity := signUp(t, supabaseURL, publishableKey, fmt.Sprintf("r2-acceptance-%d@example.com", stamp))
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `delete from auth.users where id=$1`, identity.ID) })
	api := newRelease2TestServer(t, pool, supabaseURL, publishableKey, slog.New(slog.NewTextHandler(io.Discard, nil)))

	createPlanThroughAPI(t, api.URL, identity.Token, "Aceite Release 2")
	createItemThroughAPI(t, api.URL, identity.Token, map[string]any{
		"name": "Salário", "kind": "recurring_income",
		"period": map[string]any{"start_month": "2026-01", "end_month": "2026-12", "amount_cents": 600000, "recurrence": "monthly", "cash_month_offset": 0},
	})
	rentID := createItemThroughAPI(t, api.URL, identity.Token, map[string]any{
		"name": "Aluguel", "kind": "fixed_expense",
		"period": map[string]any{"start_month": "2026-01", "end_month": "2026-12", "amount_cents": 100000, "recurrence": "monthly", "cash_month_offset": 0},
	})
	fuelID := createItemThroughAPI(t, api.URL, identity.Token, map[string]any{
		"name": "Gasolina", "kind": "projected_variable_expense",
		"period": map[string]any{"start_month": "2026-01", "end_month": "2026-12", "amount_cents": 70000, "recurrence": "monthly", "cash_month_offset": 0},
	})
	doAPI(t, api.URL, identity.Token, http.MethodPut, "/v1/plans/current/savings", map[string]any{
		"effective_from": "2026-01", "end_month": "2026-12", "amount_cents": 0,
	}, http.StatusOK)

	institutionID := nestedString(t, doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/financial-institutions", map[string]any{"name": "Banco Aceite"}, http.StatusCreated), "id")
	oldCardID := createCardThroughAPI(t, api.URL, identity.Token, institutionID, "Cartão antigo", 6, 1)
	currentCardID := nestedString(t, doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/credit-cards", map[string]any{
		"institution_id": institutionID, "name": "Cartão atual",
		"configuration": map[string]any{"effective_from": "2026-01", "end_month": nil, "nominal_due_day": 6, "payment_month_offset": 1},
	}, http.StatusCreated), "id")
	day31CardID := nestedString(t, doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/credit-cards", map[string]any{
		"institution_id": institutionID, "name": "Cartão dia 31",
		"configuration": map[string]any{"effective_from": "2026-01", "end_month": nil, "nominal_due_day": 31, "payment_month_offset": 1},
	}, http.StatusCreated), "id")

	createPaymentChange(t, api.URL, identity.Token, rentID, "2026-01", "2026-12", currentCardID)
	createPaymentChange(t, api.URL, identity.Token, fuelID, "2026-01", "2026-07", oldCardID)
	createPaymentChange(t, api.URL, identity.Token, fuelID, "2026-08", "2026-12", currentCardID)
	adjustmentID := nestedString(t, doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/credit-cards/"+currentCardID+"/invoices/2026-12/adjustments", map[string]any{
		"name": "Ajuste consolidado", "amount_cents": 18000,
	}, http.StatusCreated), "id")
	if adjustmentID == "" {
		t.Fatal("adjustment id is empty")
	}

	oldAugust := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+oldCardID+"/invoices/2026-08", nil, http.StatusOK))
	if oldAugust.Total != 70000 || oldAugust.ComponentCount != 1 || oldAugust.Components[0].ReferenceMonth != "2026-07" {
		t.Fatalf("future card change rewrote old invoice: %+v", oldAugust)
	}
	december := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+currentCardID+"/invoices/2026-12", nil, http.StatusOK))
	if december.Total != 188000 || december.ComponentCount != 3 {
		t.Fatalf("december invoice=%+v", december)
	}
	reference := decodeRelease2Summary(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/plans/current/months/2026-11/summary?basis=reference", nil, http.StatusOK))
	if reference.Commitments != 170000 || reference.CardAdjustments != 0 {
		t.Fatalf("november reference summary=%+v", reference)
	}
	cash := decodeRelease2Summary(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/plans/current/months/2026-12/summary?basis=cash", nil, http.StatusOK))
	if cash.Commitments != 188000 || cash.CardAdjustments != 18000 || cash.CommitmentSourceTotal() != 188000 {
		t.Fatalf("december cash summary double counted components: %+v", cash)
	}

	move := doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/plans/current/items/"+fuelID+"/occurrences/2026-11/invoice-moves", map[string]any{
		"target_credit_card_id": currentCardID, "target_payment_month": "2027-01", "reason": "Exceção de fechamento",
	}, http.StatusCreated)
	if nestedString(t, move, "from_payment_month") != "2026-12" || nestedString(t, move, "to_payment_month") != "2027-01" {
		t.Fatalf("move=%s", move)
	}
	var history []struct {
		From string `json:"from_payment_month"`
		To   string `json:"to_payment_month"`
	}
	if err := json.Unmarshal(doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/plans/current/items/"+fuelID+"/occurrences/2026-11/invoice-moves", nil, http.StatusOK), &history); err != nil || len(history) != 1 || history[0].From != "2026-12" || history[0].To != "2027-01" {
		t.Fatalf("history=%+v error=%v", history, err)
	}
	decemberAfterMove := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+currentCardID+"/invoices/2026-12", nil, http.StatusOK))
	if decemberAfterMove.Total != 118000 || decemberAfterMove.ComponentCount != 2 {
		t.Fatalf("december after move=%+v", decemberAfterMove)
	}
	januaryOverflow := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+currentCardID+"/invoices/2027-01", nil, http.StatusOK))
	if januaryOverflow.Total != 240000 || januaryOverflow.ComponentCount != 3 || !januaryOverflow.HasComponent("2026-12", 70000) || !januaryOverflow.HasComponent("2026-11", 70000) {
		t.Fatalf("overflow invoice=%+v", januaryOverflow)
	}
	doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/plans/current/months/2027-01/summary?basis=cash", nil, http.StatusUnprocessableEntity)
	invalidDate := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+day31CardID+"/invoices/2027-02", nil, http.StatusOK))
	if invalidDate.DueDay != 31 || invalidDate.DueDate != nil || invalidDate.DueResolution != "invalid_for_month" {
		t.Fatalf("invalid nominal date invoice=%+v", invalidDate)
	}

	activation := doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/plans/current/activate", nil, http.StatusOK)
	var activated struct {
		Original struct {
			SchemaVersion int `json:"schema_version"`
			Plan          struct {
				Release2 struct {
					Cards       []json.RawMessage `json:"credit_cards"`
					Adjustments []json.RawMessage `json:"invoice_adjustments"`
					Moves       []json.RawMessage `json:"invoice_moves"`
				} `json:"release_2"`
			} `json:"plan"`
		} `json:"original"`
	}
	if err := json.Unmarshal(activation, &activated); err != nil {
		t.Fatal(err)
	}
	if activated.Original.SchemaVersion != 2 || len(activated.Original.Plan.Release2.Cards) != 2 || len(activated.Original.Plan.Release2.Adjustments) != 1 || len(activated.Original.Plan.Release2.Moves) != 1 {
		t.Fatalf("activation snapshot=%+v", activated.Original)
	}
}

type release2Invoice struct {
	Total          int64   `json:"projected_total_cents"`
	ComponentCount int     `json:"component_count"`
	DueDay         int     `json:"nominal_due_day"`
	DueDate        *string `json:"nominal_due_date"`
	DueResolution  string  `json:"nominal_due_date_resolution"`
	Components     []struct {
		ReferenceMonth string `json:"reference_month"`
		Amount         int64  `json:"amount_cents"`
	} `json:"components"`
}

func (invoice release2Invoice) HasComponent(reference string, amount int64) bool {
	for _, component := range invoice.Components {
		if component.ReferenceMonth == reference && component.Amount == amount {
			return true
		}
	}
	return false
}

func decodeInvoice(t *testing.T, data json.RawMessage) release2Invoice {
	t.Helper()
	var invoice release2Invoice
	if err := json.Unmarshal(data, &invoice); err != nil {
		t.Fatal(err)
	}
	return invoice
}

type release2Summary struct {
	Commitments int64 `json:"commitments_cents"`
	Breakdown   struct {
		CardAdjustments int64 `json:"card_invoice_adjustments_cents"`
	} `json:"breakdown"`
	Sources []struct {
		Effect string `json:"effect"`
		Amount int64  `json:"amount_cents"`
	} `json:"sources"`
	CardAdjustments int64
}

func decodeRelease2Summary(t *testing.T, data json.RawMessage) release2Summary {
	t.Helper()
	var summary release2Summary
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatal(err)
	}
	summary.CardAdjustments = summary.Breakdown.CardAdjustments
	return summary
}

func (summary release2Summary) CommitmentSourceTotal() int64 {
	var total int64
	for _, source := range summary.Sources {
		if source.Effect == "commitment" {
			total += source.Amount
		}
	}
	return total
}

func createPaymentChange(t *testing.T, baseURL, token, itemID, start, end, cardID string) {
	t.Helper()
	doAPI(t, baseURL, token, http.MethodPost, "/v1/plans/current/items/"+itemID+"/payment-changes", map[string]any{
		"effective_from": start, "end_month": end, "method": "credit_card", "credit_card_id": cardID,
	}, http.StatusCreated)
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
