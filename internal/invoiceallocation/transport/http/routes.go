package http

import (
	"context"
	"github.com/lucas/financial-api/internal/invoiceallocation/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/httpapi"
	"net/http"
)

type Service interface {
	Move(context.Context, string, string, planningdomain.YearMonth, string, planningdomain.YearMonth, *string) (domain.Move, error)
	History(context.Context, string, string, planningdomain.YearMonth) ([]domain.Move, error)
}
type Middleware func(http.Handler) http.Handler

func RegisterRoutes(mux *http.ServeMux, auth Middleware, s Service) {
	mux.Handle("POST /v1/plans/current/items/{item_id}/occurrences/{reference_month}/invoice-moves", auth(http.HandlerFunc(moveHandler(s))))
	mux.Handle("GET /v1/plans/current/items/{item_id}/occurrences/{reference_month}/invoice-moves", auth(http.HandlerFunc(historyHandler(s))))
}

type request struct {
	TargetCreditCardID string  `json:"target_credit_card_id"`
	TargetPaymentMonth string  `json:"target_payment_month"`
	Reason             *string `json:"reason"`
}

func moveHandler(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := httpapi.PrincipalFromRequest(r)
		if !ok {
			httpapi.WriteError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		ref, e := planningdomain.ParseYearMonth(r.PathValue("reference_month"))
		if e != nil {
			validation(w)
			return
		}
		var b request
		if httpapi.DecodeJSON(r, &b) != nil {
			validation(w)
			return
		}
		to, e := planningdomain.ParseYearMonth(b.TargetPaymentMonth)
		if e != nil {
			validation(w)
			return
		}
		m, e := s.Move(r.Context(), p.UserID, r.PathValue("item_id"), ref, b.TargetCreditCardID, to, b.Reason)
		if e != nil {
			httpapi.WriteError(w, 409, "invoice_move_rejected", e.Error())
			return
		}
		httpapi.WriteJSON(w, 201, map[string]any{"data": response(m)})
	}
}
func historyHandler(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := httpapi.PrincipalFromRequest(r)
		if !ok {
			httpapi.WriteError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		ref, e := planningdomain.ParseYearMonth(r.PathValue("reference_month"))
		if e != nil {
			validation(w)
			return
		}
		moves, e := s.History(r.Context(), p.UserID, r.PathValue("item_id"), ref)
		if e != nil {
			httpapi.WriteError(w, 500, "internal_error", "An unexpected internal error occurred")
			return
		}
		data := make([]map[string]any, len(moves))
		for i, m := range moves {
			data[i] = response(m)
		}
		httpapi.WriteJSON(w, 200, map[string]any{"data": data})
	}
}
func response(m domain.Move) map[string]any {
	return map[string]any{"id": m.ID, "item_id": m.ItemID, "reference_month": m.ReferenceMonth.String(), "from_credit_card_id": m.FromCardID, "from_payment_month": m.FromPaymentMonth.String(), "to_credit_card_id": m.ToCardID, "to_payment_month": m.ToPaymentMonth.String(), "reason": m.Reason, "recorded_at": m.RecordedAt}
}
func validation(w http.ResponseWriter) {
	httpapi.WriteError(w, 400, "validation_error", "Request validation failed")
}
