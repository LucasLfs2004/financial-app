package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucas/financial-api/internal/cardinvoice"
	invoiceapplication "github.com/lucas/financial-api/internal/cardinvoice/application"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/httpapi"
)

type serviceStub struct {
	invoice      cardinvoice.Invoice
	invoices     []cardinvoice.Invoice
	err          error
	includeEmpty bool
}

func (stub *serviceStub) Project(context.Context, string, string, planningdomain.YearMonth) (cardinvoice.Invoice, error) {
	return stub.invoice, stub.err
}

func (stub *serviceStub) List(_ context.Context, _, _ string, _, _ planningdomain.YearMonth, includeEmpty bool) ([]cardinvoice.Invoice, error) {
	stub.includeEmpty = includeEmpty
	return stub.invoices, stub.err
}

func TestDetailReturnsInvoiceComposition(t *testing.T) {
	month := mustMonth(t, "2026-12")
	reference := mustMonth(t, "2026-11")
	dueDate := time.Date(2026, time.December, 6, 0, 0, 0, 0, time.UTC)
	itemID := "item"
	service := &serviceStub{invoice: cardinvoice.Invoice{
		CardID: "card", CardName: "Principal", Institution: testInstitution(), PaymentMonth: month,
		CurrencyCode: "BRL", NominalDueDate: cardinvoice.NominalDueDate{Day: 6, Date: &dueDate, Resolution: cardinvoice.DueDateResolutionExact},
		Total: planningdomain.NewMoney(70000), Components: []cardinvoice.Component{{
			SourceID: "period", SourceType: cardinvoice.ComponentTypeFinancialItemOccurrence,
			ItemID: &itemID, Name: "Gasolina", ReferenceMonth: &reference, ReferenceKnown: true,
			PaymentMonth: month, CardID: "card", Amount: planningdomain.NewMoney(70000), Allocation: cardinvoice.AllocationCalculatedFromReference,
		}},
	}}
	mux := http.NewServeMux()
	RegisterRoutes(mux, authenticated, service)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/credit-cards/card/invoices/2026-12", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	data := body["data"].(map[string]any)
	if data["projected_total_cents"] != float64(70000) || data["nominal_due_date"] != "2026-12-06" || len(data["components"].([]any)) != 1 {
		t.Fatalf("body=%v", body)
	}
}

func TestListParsesInclusiveRangeAndIncludeEmpty(t *testing.T) {
	service := &serviceStub{invoices: []cardinvoice.Invoice{testInvoice(t)}}
	mux := http.NewServeMux()
	RegisterRoutes(mux, authenticated, service)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/credit-cards/card/invoices?from=2026-11&to=2026-12&include_empty=true", nil))
	if response.Code != http.StatusOK || !service.includeEmpty {
		t.Fatalf("status=%d includeEmpty=%v body=%s", response.Code, service.includeEmpty, response.Body.String())
	}
}

func TestRoutesMapValidationHorizonAndOwnershipErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "validation", err: invoiceapplication.ErrValidation, status: http.StatusBadRequest},
		{name: "horizon", err: invoiceapplication.ErrOutsideOperationalHorizon, status: http.StatusUnprocessableEntity},
		{name: "ownership", err: cardinvoice.ErrCardNotFound, status: http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			RegisterRoutes(mux, authenticated, &serviceStub{err: test.err})
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/credit-cards/card/invoices/2026-12", nil))
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		ctx := httpapi.WithPrincipal(request.Context(), auth.Principal{UserID: "owner"})
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

func testInvoice(t *testing.T) cardinvoice.Invoice {
	t.Helper()
	month := mustMonth(t, "2026-12")
	return cardinvoice.Invoice{CardID: "card", CardName: "Principal", Institution: testInstitution(), PaymentMonth: month, CurrencyCode: "BRL", NominalDueDate: cardinvoice.NominalDueDate{Day: 6, Resolution: cardinvoice.DueDateResolutionExact}}
}

func testInstitution() cardinvoice.Institution {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	return cardinvoice.Institution{ID: "institution", Name: "Banco", Status: cardinvoice.FinancialResourceStatusActive, CreatedAt: now, UpdatedAt: now}
}

func mustMonth(t *testing.T, value string) planningdomain.YearMonth {
	t.Helper()
	month, err := planningdomain.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return month
}
