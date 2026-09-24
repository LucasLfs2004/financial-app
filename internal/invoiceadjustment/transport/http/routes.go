package http

import (
	"context"
	"github.com/lucas/financial-api/internal/invoiceadjustment/application"
	"github.com/lucas/financial-api/internal/invoiceadjustment/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/httpapi"
	"net/http"
)

type Service interface {
	Create(context.Context, string, application.Input) (domain.Adjustment, error)
	Update(context.Context, string, string, application.Input) (domain.Adjustment, error)
	Archive(context.Context, string, string) (domain.Adjustment, error)
}
type Middleware func(http.Handler) http.Handler

func RegisterRoutes(mux *http.ServeMux, auth Middleware, s Service) {
	mux.Handle("POST /v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments", auth(http.HandlerFunc(create(s))))
	mux.Handle("PATCH /v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments/{adjustment_id}", auth(http.HandlerFunc(update(s))))
	mux.Handle("POST /v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments/{adjustment_id}/archive", auth(http.HandlerFunc(archive(s))))
}

type request struct {
	Name           string  `json:"name"`
	AmountCents    int64   `json:"amount_cents"`
	ReferenceMonth *string `json:"reference_month"`
	Context        *string `json:"context"`
}

func input(r *http.Request) (application.Input, error) {
	var b request
	if e := httpapi.DecodeJSON(r, &b); e != nil {
		return application.Input{}, e
	}
	pm, e := planningdomain.ParseYearMonth(r.PathValue("payment_month"))
	if e != nil {
		return application.Input{}, e
	}
	var ref *planningdomain.YearMonth
	if b.ReferenceMonth != nil {
		v, e := planningdomain.ParseYearMonth(*b.ReferenceMonth)
		if e != nil {
			return application.Input{}, e
		}
		ref = &v
	}
	return application.Input{Name: b.Name, AmountCents: b.AmountCents, CardID: r.PathValue("card_id"), PaymentMonth: pm, ReferenceMonth: ref, Context: b.Context}, nil
}
func create(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := httpapi.PrincipalFromRequest(r)
		if !ok {
			unauth(w)
			return
		}
		in, e := input(r)
		if e != nil {
			valid(w)
			return
		}
		a, e := s.Create(r.Context(), p.UserID, in)
		if e != nil {
			valid(w)
			return
		}
		httpapi.WriteJSON(w, 201, map[string]any{"data": response(a)})
	}
}
func update(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := httpapi.PrincipalFromRequest(r)
		if !ok {
			unauth(w)
			return
		}
		in, e := input(r)
		if e != nil {
			valid(w)
			return
		}
		a, e := s.Update(r.Context(), p.UserID, r.PathValue("adjustment_id"), in)
		if e != nil {
			httpapi.WriteError(w, 409, "adjustment_rejected", e.Error())
			return
		}
		httpapi.WriteJSON(w, 200, map[string]any{"data": response(a)})
	}
}
func archive(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := httpapi.PrincipalFromRequest(r)
		if !ok {
			unauth(w)
			return
		}
		a, e := s.Archive(r.Context(), p.UserID, r.PathValue("adjustment_id"))
		if e != nil {
			httpapi.WriteError(w, 409, "adjustment_rejected", e.Error())
			return
		}
		httpapi.WriteJSON(w, 200, map[string]any{"data": response(a)})
	}
}
func response(a domain.Adjustment) map[string]any {
	var ref any
	if a.ReferenceMonth != nil {
		ref = a.ReferenceMonth.String()
	}
	return map[string]any{"id": a.ID, "currency_code": a.CurrencyCode, "credit_card_id": a.CardID, "payment_month": a.PaymentMonth.String(), "reference_month": ref, "name": a.Name, "amount_cents": a.AmountCents, "context": a.Context, "status": a.Status, "archived_at": a.ArchivedAt, "created_at": a.CreatedAt, "updated_at": a.UpdatedAt}
}
func unauth(w http.ResponseWriter) {
	httpapi.WriteError(w, 401, "unauthorized", "Authentication is required")
}
func valid(w http.ResponseWriter) {
	httpapi.WriteError(w, 400, "validation_error", "Request validation failed")
}
