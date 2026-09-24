package http

import (
	"context"
	"errors"
	"github.com/lucas/financial-api/internal/paymentmethod/application"
	"github.com/lucas/financial-api/internal/paymentmethod/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/httpapi"
	"net/http"
)

type Service interface {
	Create(context.Context, string, string, application.Input) (domain.Period, error)
	List(context.Context, string, string) ([]domain.Period, error)
}
type Middleware func(http.Handler) http.Handler

func RegisterRoutes(mux *http.ServeMux, auth Middleware, service Service) {
	mux.Handle("POST /v1/plans/current/items/{item_id}/payment-changes", auth(http.HandlerFunc(createHandler(service))))
	mux.Handle("GET /v1/plans/current/items/{item_id}/payment-history", auth(http.HandlerFunc(listHandler(service))))
	mux.Handle("POST /v1/financial-items/{item_id}/payment-changes", auth(http.HandlerFunc(createHandler(service))))
	mux.Handle("GET /v1/financial-items/{item_id}/payment-history", auth(http.HandlerFunc(listHandler(service))))
}

type request struct {
	EffectiveFrom string                 `json:"effective_from"`
	EndMonth      httpapi.OptionalString `json:"end_month"`
	Method        domain.Kind            `json:"method"`
	CreditCardID  *string                `json:"credit_card_id"`
	Context       *string                `json:"context"`
}

func createHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := httpapi.PrincipalFromRequest(r)
		if !ok {
			httpapi.WriteError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		var body request
		if httpapi.DecodeJSON(r, &body) != nil {
			validation(w)
			return
		}
		start, e := planningdomain.ParseYearMonth(body.EffectiveFrom)
		if e != nil {
			validation(w)
			return
		}
		var end *planningdomain.YearMonth
		if body.EndMonth.Set && body.EndMonth.Value != nil {
			v, e := planningdomain.ParseYearMonth(*body.EndMonth.Value)
			if e != nil {
				validation(w)
				return
			}
			end = &v
		}
		period, e := service.Create(r.Context(), p.UserID, r.PathValue("item_id"), application.Input{EffectiveFrom: start, EndMonth: end, Method: body.Method, CreditCardID: body.CreditCardID, Context: body.Context})
		if e != nil {
			writeError(w, e)
			return
		}
		httpapi.WriteJSON(w, 201, map[string]any{"data": response(period)})
	}
}
func listHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := httpapi.PrincipalFromRequest(r)
		if !ok {
			httpapi.WriteError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		periods, e := service.List(r.Context(), p.UserID, r.PathValue("item_id"))
		if e != nil {
			writeError(w, e)
			return
		}
		data := make([]map[string]any, len(periods))
		for i, v := range periods {
			data[i] = response(v)
		}
		httpapi.WriteJSON(w, 200, map[string]any{"data": data})
	}
}
func response(p domain.Period) map[string]any {
	var end any
	if p.EndMonth != nil {
		end = p.EndMonth.String()
	}
	return map[string]any{"id": p.ID, "start_month": p.StartMonth.String(), "end_month": end, "method": p.Method, "credit_card_id": p.CreditCardID, "context": p.Context, "recorded_at": p.RecordedAt, "created_at": p.CreatedAt}
}
func validation(w http.ResponseWriter) {
	httpapi.WriteError(w, 400, "validation_error", "Request validation failed")
}
func writeError(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, domain.ErrValidation):
		validation(w)
	case errors.Is(e, application.ErrNotFound), errors.Is(e, application.ErrItemNotFound), errors.Is(e, application.ErrCardNotFound):
		httpapi.WriteError(w, 404, "not_found", "Resource not found")
	case errors.Is(e, domain.ErrOverlap):
		httpapi.WriteError(w, 409, "payment_period_overlap", "The payment period overlaps another period")
	case errors.Is(e, application.ErrCardArchived):
		httpapi.WriteError(w, 409, "card_archived", "The credit card is archived")
	default:
		httpapi.WriteError(w, 500, "internal_error", "An unexpected internal error occurred")
	}
}
