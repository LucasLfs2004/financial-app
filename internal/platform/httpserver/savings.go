package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/savings"
)

type SavingsService interface {
	Get(context.Context, string) (savings.Configuration, error)
	Put(context.Context, string, savings.PutInput) (savings.Configuration, error)
}
type putSavingsRequest struct {
	EffectiveFrom string  `json:"effective_from"`
	EndMonth      *string `json:"end_month"`
	AmountCents   int64   `json:"amount_cents"`
	Context       *string `json:"context"`
}

func getSavingsHandler(service SavingsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		configuration, err := service.Get(r.Context(), principal.UserID)
		if err != nil {
			writeSavingsError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"data": savingsResponse(configuration)})
	}
}
func putSavingsHandler(service SavingsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		var payload putSavingsRequest
		if decodeJSON(r, &payload) != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		effective, err := domain.ParseYearMonth(payload.EffectiveFrom)
		if err != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		end, err := optionalMonth(payload.EndMonth)
		if err != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		configuration, err := service.Put(r.Context(), principal.UserID, savings.PutInput{EffectiveFrom: effective, EndMonth: end, AmountCents: payload.AmountCents, Context: payload.Context})
		if err != nil {
			writeSavingsError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"data": savingsResponse(configuration)})
	}
}
func writeSavingsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, savings.ErrValidation):
		writeError(w, 400, "validation_error", "Request validation failed")
	case errors.Is(err, planning.ErrNotFound):
		writeError(w, 404, "not_found", "Resource not found")
	case errors.Is(err, savings.ErrOverlap):
		writeError(w, 409, "period_overlap", "The new period overlaps an existing period")
	default:
		writeError(w, 500, "internal_error", "An unexpected internal error occurred")
	}
}
func savingsResponse(configuration savings.Configuration) map[string]any {
	periods := make([]map[string]any, len(configuration.Periods))
	for i, p := range configuration.Periods {
		var end any
		if p.EndMonth != nil {
			end = p.EndMonth.String()
		}
		periods[i] = map[string]any{"id": p.ID, "start_month": p.StartMonth.String(), "end_month": end, "amount_cents": p.AmountCents, "context": p.Context, "created_at": p.CreatedAt}
	}
	return map[string]any{"configured": configuration.Configured, "periods": periods}
}
