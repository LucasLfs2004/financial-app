package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/httpapi"
)

type serviceStub struct{ created cardapplication.CreateInput }

func (service *serviceStub) Create(_ context.Context, _ string, input cardapplication.CreateInput) (carddomain.Card, error) {
	service.created = input
	return carddomain.Card{ID: "card", Name: input.Name, Status: carddomain.StatusActive}, nil
}
func (*serviceStub) List(context.Context, string, cardapplication.ListFilters) ([]carddomain.Card, error) {
	return []carddomain.Card{}, nil
}
func (*serviceStub) Find(context.Context, string, string) (carddomain.Card, error) {
	return carddomain.Card{}, cardapplication.ErrNotFound
}
func (*serviceStub) Update(context.Context, string, string, cardapplication.UpdateInput) (carddomain.Card, error) {
	return carddomain.Card{}, nil
}
func (*serviceStub) Change(context.Context, string, string, cardapplication.ChangeInput) (carddomain.Card, error) {
	return carddomain.Card{}, nil
}
func (*serviceStub) Archive(context.Context, string, string, cardapplication.ArchiveInput) (carddomain.Card, error) {
	return carddomain.Card{}, nil
}

func TestCreateRoute(t *testing.T) {
	service := &serviceStub{}
	mux := http.NewServeMux()
	RegisterRoutes(mux, func(next http.Handler) http.Handler { return next }, service)
	body := `{"institution_id":"d1000000-0000-0000-0000-000000000001","name":"Principal","configuration":{"effective_from":"2026-01","end_month":null,"nominal_due_day":31,"payment_month_offset":1}}`
	request := httptest.NewRequest(http.MethodPost, "/v1/credit-cards", strings.NewReader(body))
	request = request.WithContext(httpapi.WithPrincipal(request.Context(), auth.Principal{UserID: "owner"}))
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.created.Configuration.NominalDueDay != 31 {
		t.Fatalf("input=%+v", service.created)
	}
}

func TestCreateRouteRequiresExplicitEndMonth(t *testing.T) {
	service := &serviceStub{}
	mux := http.NewServeMux()
	RegisterRoutes(mux, func(next http.Handler) http.Handler { return next }, service)
	body := `{"institution_id":"d1000000-0000-0000-0000-000000000001","name":"Principal","configuration":{"effective_from":"2026-01","nominal_due_day":31,"payment_month_offset":1}}`
	request := httptest.NewRequest(http.MethodPost, "/v1/credit-cards", strings.NewReader(body))
	request = request.WithContext(httpapi.WithPrincipal(request.Context(), auth.Principal{UserID: "owner"}))
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
