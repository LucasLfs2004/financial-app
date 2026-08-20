package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
)

type FinancialItemService interface {
	Create(context.Context, string, financialitem.CreateInput) (financialitem.Item, error)
	List(context.Context, string, financialitem.Filters) ([]financialitem.Item, error)
	Find(context.Context, string, string) (financialitem.Item, error)
	Update(context.Context, string, string, financialitem.UpdateInput) (financialitem.Item, error)
	Change(context.Context, string, string, financialitem.ChangeInput) (financialitem.Item, error)
	Archive(context.Context, string, string, financialitem.ArchiveInput) (financialitem.Item, error)
}

type periodRequest struct {
	StartMonth      string  `json:"start_month"`
	EndMonth        *string `json:"end_month"`
	AmountCents     int64   `json:"amount_cents"`
	Recurrence      string  `json:"recurrence"`
	CashMonthOffset int     `json:"cash_month_offset"`
	Context         *string `json:"context"`
}
type createItemRequest struct {
	Name        string        `json:"name"`
	Kind        string        `json:"kind"`
	Description *string       `json:"description"`
	Period      periodRequest `json:"period"`
}
type updateItemRequest struct {
	Name        *string        `json:"name"`
	Description optionalString `json:"description"`
}
type optionalString struct {
	Set   bool
	Value *string
}

func (value *optionalString) UnmarshalJSON(data []byte) error {
	value.Set = true
	if string(data) == "null" {
		value.Value = nil
		return nil
	}
	var decoded string
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	value.Value = &decoded
	return nil
}

type changeItemRequest struct {
	EffectiveFrom   string  `json:"effective_from"`
	EndMonth        *string `json:"end_month"`
	AmountCents     int64   `json:"amount_cents"`
	CashMonthOffset int     `json:"cash_month_offset"`
	Context         *string `json:"context"`
}
type archiveItemRequest struct {
	EffectiveFrom *string `json:"effective_from"`
	Reason        *string `json:"reason"`
}

func createFinancialItemHandler(service FinancialItemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		var payload createItemRequest
		if decodeJSON(r, &payload) != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		kind, err := domain.ParseFinancialItemKind(payload.Kind)
		if err != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		period, err := periodInput(payload.Period)
		if err != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		item, err := service.Create(r.Context(), principal.UserID, financialitem.CreateInput{Name: payload.Name, Kind: kind, Description: payload.Description, Period: period})
		if err != nil {
			writeFinancialItemError(w, err)
			return
		}
		writeJSON(w, 201, map[string]any{"data": financialItemResponse(item)})
	}
}

func listFinancialItemsHandler(service FinancialItemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		filters := financialitem.Filters{}
		if raw := r.URL.Query().Get("kind"); raw != "" {
			value, err := domain.ParseFinancialItemKind(raw)
			if err != nil {
				writeError(w, 400, "validation_error", "Request validation failed")
				return
			}
			filters.Kind = &value
		}
		if raw := r.URL.Query().Get("status"); raw != "" {
			value, err := domain.ParseFinancialItemStatus(raw)
			if err != nil {
				writeError(w, 400, "validation_error", "Request validation failed")
				return
			}
			filters.Status = &value
		}
		items, err := service.List(r.Context(), principal.UserID, filters)
		if err != nil {
			writeFinancialItemError(w, err)
			return
		}
		data := make([]map[string]any, len(items))
		for i, item := range items {
			data[i] = financialItemResponse(item)
		}
		writeJSON(w, 200, map[string]any{"data": data})
	}
}

func getFinancialItemHandler(service FinancialItemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		item, err := service.Find(r.Context(), principal.UserID, r.PathValue("item_id"))
		if err != nil {
			writeFinancialItemError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"data": financialItemResponse(item)})
	}
}

func updateFinancialItemHandler(service FinancialItemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		var payload updateItemRequest
		if decodeJSON(r, &payload) != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		item, err := service.Update(r.Context(), principal.UserID, r.PathValue("item_id"), financialitem.UpdateInput{Name: payload.Name, Description: payload.Description.Value, DescriptionSet: payload.Description.Set})
		if err != nil {
			writeFinancialItemError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"data": financialItemResponse(item)})
	}
}

func changeFinancialItemHandler(service FinancialItemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		var payload changeItemRequest
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
		item, err := service.Change(r.Context(), principal.UserID, r.PathValue("item_id"), financialitem.ChangeInput{EffectiveFrom: effective, EndMonth: end, AmountCents: payload.AmountCents, CashMonthOffset: payload.CashMonthOffset, Context: payload.Context})
		if err != nil {
			writeFinancialItemError(w, err)
			return
		}
		writeJSON(w, 201, map[string]any{"data": financialItemResponse(item)})
	}
}

func archiveFinancialItemHandler(service FinancialItemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, 401, "unauthorized", "Authentication is required")
			return
		}
		payload := archiveItemRequest{}
		if r.Body != nil && r.ContentLength != 0 && decodeJSON(r, &payload) != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		effective, err := optionalMonth(payload.EffectiveFrom)
		if err != nil {
			writeError(w, 400, "validation_error", "Request validation failed")
			return
		}
		item, err := service.Archive(r.Context(), principal.UserID, r.PathValue("item_id"), financialitem.ArchiveInput{EffectiveFrom: effective, Reason: payload.Reason})
		if err != nil {
			writeFinancialItemError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"data": financialItemResponse(item)})
	}
}

func periodInput(payload periodRequest) (financialitem.PeriodInput, error) {
	start, err := domain.ParseYearMonth(payload.StartMonth)
	if err != nil {
		return financialitem.PeriodInput{}, err
	}
	end, err := optionalMonth(payload.EndMonth)
	if err != nil {
		return financialitem.PeriodInput{}, err
	}
	recurrence, err := domain.ParseRecurrence(payload.Recurrence)
	if err != nil {
		return financialitem.PeriodInput{}, err
	}
	return financialitem.PeriodInput{StartMonth: start, EndMonth: end, AmountCents: payload.AmountCents, Recurrence: recurrence, CashMonthOffset: payload.CashMonthOffset, Context: payload.Context}, nil
}
func optionalMonth(raw *string) (*domain.YearMonth, error) {
	if raw == nil {
		return nil, nil
	}
	month, err := domain.ParseYearMonth(*raw)
	if err != nil {
		return nil, err
	}
	return &month, nil
}

func writeFinancialItemError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, financialitem.ErrValidation):
		writeError(w, 400, "validation_error", "Request validation failed")
	case errors.Is(err, financialitem.ErrNotFound), errors.Is(err, planning.ErrNotFound):
		writeError(w, 404, "not_found", "Resource not found")
	case errors.Is(err, financialitem.ErrPeriodOverlap):
		writeError(w, 409, "period_overlap", "The new period overlaps an existing period")
	default:
		writeError(w, 500, "internal_error", "An unexpected internal error occurred")
	}
}

func financialItemResponse(item financialitem.Item) map[string]any {
	periods := make([]map[string]any, len(item.Periods))
	for i, p := range item.Periods {
		var end any
		if p.EndMonth != nil {
			end = p.EndMonth.String()
		}
		periods[i] = map[string]any{"id": p.ID, "start_month": p.StartMonth.String(), "end_month": end, "amount_cents": p.AmountCents, "recurrence": p.Recurrence, "cash_month_offset": p.CashMonthOffset, "context": p.Context, "recorded_at": p.RecordedAt, "created_at": p.CreatedAt}
	}
	return map[string]any{"id": item.ID, "plan_id": item.PlanID, "name": item.Name, "kind": item.Kind, "description": item.Description, "status": item.Status, "archived_at": item.ArchivedAt, "periods": periods, "created_at": item.CreatedAt, "updated_at": item.UpdatedAt}
}
