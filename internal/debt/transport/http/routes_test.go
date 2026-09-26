package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/httpapi"
)

type serviceStub struct {
	created debtapplication.CreateInput
	changed debtapplication.ChangeInput
}

func (service *serviceStub) Create(_ context.Context, _ string, input debtapplication.CreateInput) (debtapplication.View, error) {
	service.created = input
	return responseView(tMonth("2026-09")), nil
}
func (*serviceStub) List(context.Context, string, debtapplication.Filters) ([]debtapplication.View, error) {
	return []debtapplication.View{}, nil
}
func (*serviceStub) Find(context.Context, string, string, *planningdomain.YearMonth) (debtapplication.View, error) {
	return debtapplication.View{}, debtapplication.ErrNotFound
}
func (*serviceStub) Update(context.Context, string, string, debtapplication.UpdateInput) (debtapplication.View, error) {
	return responseView(tMonth("2026-09")), nil
}
func (service *serviceStub) Change(_ context.Context, _, _ string, input debtapplication.ChangeInput) (debtapplication.View, error) {
	service.changed = input
	return responseView(tMonth("2026-09")), nil
}
func (*serviceStub) Archive(context.Context, string, string, debtapplication.ArchiveInput) (debtapplication.View, error) {
	return responseView(tMonth("2026-09")), nil
}

func TestCreateDebtRoute(t *testing.T) {
	service := &serviceStub{}
	mux := http.NewServeMux()
	RegisterRoutes(mux, func(next http.Handler) http.Handler { return next }, service)
	body := `{"name":"Transplante","original_total_cents":720000,"total_installments":12,"first_projected_installment":5,"scheduled_start_month":"2026-09","installment_amount_cents":60000}`
	request := httptest.NewRequest(http.MethodPost, "/v1/debts", strings.NewReader(body))
	request = request.WithContext(httpapi.WithPrincipal(request.Context(), auth.Principal{UserID: "owner"}))
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.created.OriginalTotalCents == nil || *service.created.OriginalTotalCents != 720000 {
		t.Fatalf("created=%+v", service.created)
	}
}

func TestCreateDebtRouteRequiresOriginalTotalField(t *testing.T) {
	service := &serviceStub{}
	mux := http.NewServeMux()
	RegisterRoutes(mux, func(next http.Handler) http.Handler { return next }, service)
	body := `{"name":"Dívida","total_installments":1,"first_projected_installment":1,"scheduled_start_month":"2026-09","installment_amount_cents":100}`
	request := httptest.NewRequest(http.MethodPost, "/v1/debts", strings.NewReader(body))
	request = request.WithContext(httpapi.WithPrincipal(request.Context(), auth.Principal{UserID: "owner"}))
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestListDebtRouteRejectsInvalidProjectionStatus(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, func(next http.Handler) http.Handler { return next }, &serviceStub{})
	request := httptest.NewRequest(http.MethodGet, "/v1/debts?projection_status=unknown", nil)
	request = request.WithContext(httpapi.WithPrincipal(request.Context(), auth.Principal{UserID: "owner"}))
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestChangeDebtAmountRoute(t *testing.T) {
	service := &serviceStub{}
	mux := http.NewServeMux()
	RegisterRoutes(mux, func(next http.Handler) http.Handler { return next }, service)
	body := `{"effective_from":"2027-01","installment_amount_cents":65000,"context":"Reajuste"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/debts/debt/changes", strings.NewReader(body))
	request = request.WithContext(httpapi.WithPrincipal(request.Context(), auth.Principal{UserID: "owner"}))
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated || service.changed.EffectiveFrom.String() != "2027-01" ||
		service.changed.InstallmentAmountCents != 65000 {
		t.Fatalf("status=%d input=%+v body=%s", recorder.Code, service.changed, recorder.Body.String())
	}
}

func responseView(asOf planningdomain.YearMonth) debtapplication.View {
	end := tMonth("2027-04")
	interval, _ := planningdomain.NewMonthInterval(tMonth("2026-09"), end)
	debt, _ := debtdomain.NewDebt(debtdomain.NewDebtInput{
		ID: "debt", CurrencyCode: "BRL", Name: "Transplante",
		TotalInstallments: 12, FirstProjectedInstallment: 5, ScheduledStart: tMonth("2026-09"),
		Periods: []debtdomain.InstallmentPeriod{{ID: "period", Interval: interval, Amount: planningdomain.NewMoney(60000)}},
		Status:  planningdomain.FinancialItemStatusActive,
	})
	projection, _ := debtdomain.ProjectDebt(debtdomain.ProjectionInput{Debt: debt, From: debt.ScheduledStart, To: debt.ScheduledEnd, AsOf: asOf})
	return debtapplication.View{Debt: debt, Projection: projection, AsOf: asOf}
}

func tMonth(value string) planningdomain.YearMonth {
	month, _ := planningdomain.ParseYearMonth(value)
	return month
}
