package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRelease3AcceptanceScenario(t *testing.T) {
	databaseURL, supabaseURL, key := release2Environment(t)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	identity := signUp(t, supabaseURL, key, fmt.Sprintf("r3-acceptance-%d@example.com", time.Now().UnixNano()))
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `delete from auth.users where id=$1`, identity.ID) })
	api := newRelease2TestServer(t, pool, supabaseURL, key, slog.New(slog.NewTextHandler(io.Discard, nil)))
	doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/plans", map[string]any{
		"name": "Aceite R3", "start_month": "2026-09", "end_month": "2027-06", "currency_code": "BRL",
	}, http.StatusCreated)
	createItemThroughAPI(t, api.URL, identity.Token, map[string]any{
		"name": "Renda", "kind": "recurring_income", "period": map[string]any{
			"start_month": "2026-09", "end_month": "2027-06", "amount_cents": 500000, "recurrence": "monthly", "cash_month_offset": 0,
		},
	})
	doAPI(t, api.URL, identity.Token, http.MethodPut, "/v1/plans/current/savings", map[string]any{
		"effective_from": "2026-09", "end_month": "2027-06", "amount_cents": 0,
	}, http.StatusOK)
	institution := nestedString(t, doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/financial-institutions", map[string]any{"name": "Banco"}, http.StatusCreated), "id")
	card := nestedString(t, doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/credit-cards", map[string]any{
		"institution_id": institution, "name": "Cartão", "configuration": map[string]any{
			"effective_from": "2026-09", "end_month": "2027-06", "nominal_due_day": 6, "payment_month_offset": 1,
		},
	}, http.StatusCreated), "id")
	debtID := nestedString(t, doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/debts", map[string]any{
		"name": "Transplante", "original_total_cents": 720000, "total_installments": 12,
		"first_projected_installment": 5, "scheduled_start_month": "2026-09",
		"installment_amount_cents": 60000,
	}, http.StatusCreated), "id")
	var schedule struct {
		Debt struct {
			ScheduledEnd string `json:"scheduled_end_month"`
		} `json:"debt"`
		Occurrences []struct {
			Month  string `json:"reference_month"`
			Number int    `json:"installment_number"`
			Amount int64  `json:"amount_cents"`
			Kind   string `json:"debt_occurrence_kind"`
		} `json:"occurrences"`
	}
	readSchedule := func() {
		t.Helper()
		if err := json.Unmarshal(doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/debts/"+debtID+"/schedule?from=2026-09&to=2027-04", nil, http.StatusOK), &schedule); err != nil {
			t.Fatal(err)
		}
	}
	readSchedule()
	if schedule.Debt.ScheduledEnd != "2027-04" || len(schedule.Occurrences) != 8 || schedule.Occurrences[0].Number != 5 || schedule.Occurrences[7].Number != 12 || schedule.Occurrences[7].Month != "2027-04" {
		t.Fatalf("initial schedule=%+v", schedule)
	}
	release := doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/debt-releases?from=2027-05&to=2027-05", nil, http.StatusOK)
	if !bytes.Contains(release, []byte(debtID)) || !bytes.Contains(release, []byte(`"release_from_month":"2027-05"`)) {
		t.Fatalf("release=%s", release)
	}
	releasedMonth := decodeRelease2Summary(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/plans/current/months/2027-05/summary?basis=reference", nil, http.StatusOK))
	if releasedMonth.Commitments != 0 || releasedMonth.CommitmentSourceTotal() != 0 {
		t.Fatalf("release became a financial source: %+v", releasedMonth)
	}
	ref := decodeRelease2Summary(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/plans/current/months/2026-09/summary?basis=reference", nil, http.StatusOK))
	if ref.Commitments != 60000 || ref.CommitmentSourceTotal() != 60000 {
		t.Fatalf("reference summary=%+v", ref)
	}
	doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/debts/"+debtID+"/changes", map[string]any{
		"effective_from": "2026-11", "installment_amount_cents": 65000,
	}, http.StatusCreated)
	readSchedule()
	if schedule.Occurrences[0].Amount != 60000 || schedule.Occurrences[2].Amount != 65000 || schedule.Occurrences[2].Number != 7 {
		t.Fatalf("change rewrote past or numbering: %+v", schedule.Occurrences)
	}
	doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/plans/current/items/"+debtID+"/payment-changes", map[string]any{
		"effective_from": "2026-11", "end_month": "2027-04", "method": "credit_card", "credit_card_id": card,
	}, http.StatusCreated)
	invoice := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+card+"/invoices/2026-12", nil, http.StatusOK))
	if invoice.Total != 65000 || !invoice.HasComponent("2026-11", 65000) {
		t.Fatalf("invoice after card switch=%+v", invoice)
	}
	cash := decodeRelease2Summary(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/plans/current/months/2026-12/summary?basis=cash", nil, http.StatusOK))
	if cash.Commitments != 65000 || cash.CommitmentSourceTotal() != 65000 {
		t.Fatalf("cash did not reconcile with card invoice: %+v", cash)
	}
	doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/plans/current/items/"+debtID+"/occurrences/2026-11/invoice-moves", map[string]any{
		"target_credit_card_id": card, "target_payment_month": "2027-02", "reason": "Exceção",
	}, http.StatusCreated)
	oldInvoice := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+card+"/invoices/2026-12", nil, http.StatusOK))
	newInvoice := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+card+"/invoices/2027-02", nil, http.StatusOK))
	if oldInvoice.HasComponent("2026-11", 65000) || !newInvoice.HasComponent("2026-11", 65000) {
		t.Fatalf("moved occurrence duplicated: old=%+v new=%+v", oldInvoice, newInvoice)
	}
	doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/debts/"+debtID+"/early-settlement", map[string]any{
		"reference_month": "2027-01", "amount_cents": 150000,
	}, http.StatusCreated)
	readSchedule()
	if len(schedule.Occurrences) != 5 || schedule.Occurrences[4].Month != "2027-01" || schedule.Occurrences[4].Amount != 150000 || schedule.Occurrences[4].Kind != "early_settlement" {
		t.Fatalf("settled schedule=%+v", schedule)
	}
	settledInvoice := decodeInvoice(t, doAPI(t, api.URL, identity.Token, http.MethodGet, "/v1/credit-cards/"+card+"/invoices/2027-02", nil, http.StatusOK))
	if !settledInvoice.HasComponent("2027-01", 150000) || !settledInvoice.HasComponent("2026-11", 65000) {
		t.Fatalf("settlement overflow invoice=%+v", settledInvoice)
	}
	activation := doAPI(t, api.URL, identity.Token, http.MethodPost, "/v1/plans/current/activate", nil, http.StatusOK)
	var snapshot struct {
		Original struct {
			SchemaVersion int `json:"schema_version"`
			Plan          struct {
				Release3 struct {
					Debts []json.RawMessage `json:"debts"`
				} `json:"release_3"`
			} `json:"plan"`
		} `json:"original"`
	}
	if err := json.Unmarshal(activation, &snapshot); err != nil || snapshot.Original.SchemaVersion != 4 || len(snapshot.Original.Plan.Release3.Debts) != 1 {
		t.Fatalf("activation=%s error=%v", activation, err)
	}
	if !bytes.Contains(snapshot.Original.Plan.Release3.Debts[0], []byte(debtID)) || !bytes.Contains(snapshot.Original.Plan.Release3.Debts[0], []byte("150000")) {
		t.Fatalf("snapshot omitted debt settlement: %s", snapshot.Original.Plan.Release3.Debts[0])
	}
}

func TestRelease3EndToEndIsolation(t *testing.T) {
	databaseURL, supabaseURL, key := release2Environment(t)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	stamp := time.Now().UnixNano()
	first := signUp(t, supabaseURL, key, fmt.Sprintf("r3-security-%d-one@example.com", stamp))
	second := signUp(t, supabaseURL, key, fmt.Sprintf("r3-security-%d-two@example.com", stamp))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `delete from auth.users where id in ($1,$2)`, first.ID, second.ID)
	})
	var logs bytes.Buffer
	api := newRelease2TestServer(t, pool, supabaseURL, key, slog.New(slog.NewJSONHandler(&logs, nil)))
	createPlanThroughAPI(t, api.URL, first.Token, "Plano R3 A")
	createPlanThroughAPI(t, api.URL, second.Token, "Plano R3 B")
	firstInstitution := nestedString(t, doAPI(t, api.URL, first.Token, http.MethodPost, "/v1/financial-institutions", map[string]any{"name": "Banco secreto A"}, http.StatusCreated), "id")
	firstCard := createCardThroughAPI(t, api.URL, first.Token, firstInstitution, "Cartão secreto A", 6, 1)
	firstDebt := nestedString(t, doAPI(t, api.URL, first.Token, http.MethodPost, "/v1/debts", map[string]any{
		"name": "Dívida secreta A", "original_total_cents": 120000, "total_installments": 3,
		"first_projected_installment": 1, "scheduled_start_month": "2026-09",
		"installment_amount_cents": 40000, "payment_method": "credit_card", "credit_card_id": firstCard,
	}, http.StatusCreated), "id")
	secondDebt := nestedString(t, doAPI(t, api.URL, second.Token, http.MethodPost, "/v1/debts", map[string]any{
		"name": "Dívida B", "original_total_cents": nil, "total_installments": 3,
		"first_projected_installment": 1, "scheduled_start_month": "2026-09", "installment_amount_cents": 30000,
	}, http.StatusCreated), "id")
	for _, path := range []string{
		"/v1/debts/" + firstDebt,
		"/v1/debts/" + firstDebt + "/schedule?from=2026-09&to=2026-11",
		"/v1/debts/" + firstDebt + "/early-settlement",
	} {
		doAPI(t, api.URL, second.Token, http.MethodGet, path, nil, http.StatusNotFound)
	}
	doAPI(t, api.URL, second.Token, http.MethodPatch, "/v1/debts/"+firstDebt, map[string]any{"name": "Cruzada"}, http.StatusNotFound)
	doAPI(t, api.URL, second.Token, http.MethodPost, "/v1/debts/"+firstDebt+"/changes", map[string]any{"effective_from": "2026-10", "installment_amount_cents": 1}, http.StatusNotFound)
	doAPI(t, api.URL, second.Token, http.MethodPost, "/v1/debts/"+firstDebt+"/early-settlement", map[string]any{"reference_month": "2026-10", "amount_cents": 1}, http.StatusNotFound)
	doAPI(t, api.URL, second.Token, http.MethodPost, "/v1/plans/current/items/"+secondDebt+"/payment-changes", map[string]any{
		"effective_from": "2026-09", "end_month": "2026-11", "method": "credit_card", "credit_card_id": firstCard,
	}, http.StatusNotFound)
	assertDataAPIEmpty(t, supabaseURL, key, second.Token, "debts?select=financial_item_id&financial_item_id=eq."+url.QueryEscape(firstDebt))
	assertDataAPIEmpty(t, supabaseURL, key, second.Token, "financial_item_periods?select=id&financial_item_id=eq."+url.QueryEscape(firstDebt))
	assertDataAPIEmpty(t, supabaseURL, key, second.Token, "financial_item_payment_periods?select=id&financial_item_id=eq."+url.QueryEscape(firstDebt))
	settlement := doAPI(t, api.URL, first.Token, http.MethodPost, "/v1/debts/"+firstDebt+"/early-settlement", map[string]any{"reference_month": "2026-10", "amount_cents": 50000}, http.StatusCreated)
	if nestedString(t, settlement, "debt_id") != firstDebt {
		t.Fatalf("settlement=%s", settlement)
	}
	assertDataAPIEmpty(t, supabaseURL, key, second.Token, "debt_early_settlements?select=id&financial_item_id=eq."+url.QueryEscape(firstDebt))
	for _, secret := range []string{first.Token, second.Token, "Dívida secreta A", "Banco secreto A", "Cartão secreto A", "120000"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("logs leaked protected data: %s", secret)
		}
	}
}
