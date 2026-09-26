package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/httpapi"
)

type Service interface {
	Create(context.Context, string, debtapplication.CreateInput) (debtapplication.View, error)
	List(context.Context, string, debtapplication.Filters) ([]debtapplication.View, error)
	Find(context.Context, string, string, *planningdomain.YearMonth) (debtapplication.View, error)
	Update(context.Context, string, string, debtapplication.UpdateInput) (debtapplication.View, error)
	Change(context.Context, string, string, debtapplication.ChangeInput) (debtapplication.View, error)
	Archive(context.Context, string, string, debtapplication.ArchiveInput) (debtapplication.View, error)
	Schedule(context.Context, string, string, planningdomain.YearMonth, planningdomain.YearMonth) (debtapplication.Schedule, error)
}

type Middleware func(http.Handler) http.Handler

func RegisterRoutes(mux *http.ServeMux, authenticate Middleware, service Service) {
	mux.Handle("POST /v1/debts", authenticate(http.HandlerFunc(createHandler(service))))
	mux.Handle("GET /v1/debts", authenticate(http.HandlerFunc(listHandler(service))))
	mux.Handle("GET /v1/debts/{debt_id}", authenticate(http.HandlerFunc(findHandler(service))))
	mux.Handle("PATCH /v1/debts/{debt_id}", authenticate(http.HandlerFunc(updateHandler(service))))
	mux.Handle("POST /v1/debts/{debt_id}/changes", authenticate(http.HandlerFunc(changeHandler(service))))
	mux.Handle("POST /v1/debts/{debt_id}/archive", authenticate(http.HandlerFunc(archiveHandler(service))))
	mux.Handle("GET /v1/debts/{debt_id}/schedule", authenticate(http.HandlerFunc(scheduleHandler(service))))
}

type optionalInt64 struct {
	Set   bool
	Value *int64
}

func (value *optionalInt64) UnmarshalJSON(data []byte) error {
	value.Set = true
	if string(data) == "null" {
		value.Value = nil
		return nil
	}
	var decoded int64
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	value.Value = &decoded
	return nil
}

type createRequest struct {
	Name                      string        `json:"name"`
	Description               *string       `json:"description"`
	OriginalTotalCents        optionalInt64 `json:"original_total_cents"`
	TotalInstallments         int           `json:"total_installments"`
	FirstProjectedInstallment int           `json:"first_projected_installment"`
	ScheduledStartMonth       string        `json:"scheduled_start_month"`
	InstallmentAmountCents    int64         `json:"installment_amount_cents"`
	CashMonthOffset           int           `json:"cash_month_offset"`
	Context                   *string       `json:"context"`
	PaymentMethod             string        `json:"payment_method"`
	CreditCardID              *string       `json:"credit_card_id"`
}

type updateRequest struct {
	Name        httpapi.OptionalString `json:"name"`
	Description httpapi.OptionalString `json:"description"`
}

type archiveRequest struct {
	Reason *string `json:"reason"`
}

type changeRequest struct {
	EffectiveFrom          string  `json:"effective_from"`
	InstallmentAmountCents int64   `json:"installment_amount_cents"`
	Context                *string `json:"context"`
}

func createHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		var payload createRequest
		if httpapi.DecodeJSON(request, &payload) != nil || !payload.OriginalTotalCents.Set {
			validationError(w)
			return
		}
		start, err := planningdomain.ParseYearMonth(payload.ScheduledStartMonth)
		if err != nil {
			validationError(w)
			return
		}
		view, err := service.Create(request.Context(), ownerID, debtapplication.CreateInput{
			Name: payload.Name, Description: payload.Description,
			OriginalTotalCents:        payload.OriginalTotalCents.Value,
			TotalInstallments:         payload.TotalInstallments,
			FirstProjectedInstallment: payload.FirstProjectedInstallment,
			ScheduledStart:            start, InstallmentAmountCents: payload.InstallmentAmountCents,
			CashMonthOffset: payload.CashMonthOffset, Context: payload.Context,
			PaymentMethod: payload.PaymentMethod, CreditCardID: payload.CreditCardID,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, map[string]any{"data": debtResponse(view)})
	}
}

func listHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		filters := debtapplication.Filters{}
		if raw := request.URL.Query().Get("as_of"); raw != "" {
			month, err := planningdomain.ParseYearMonth(raw)
			if err != nil {
				validationError(w)
				return
			}
			filters.AsOf = &month
		}
		if raw := request.URL.Query().Get("projection_status"); raw != "" {
			status := debtdomain.ProjectionStatus(raw)
			if !status.Valid() {
				validationError(w)
				return
			}
			filters.ProjectionStatus = &status
		}
		if raw := request.URL.Query().Get("status"); raw != "" {
			status, err := planningdomain.ParseFinancialItemStatus(raw)
			if err != nil {
				validationError(w)
				return
			}
			filters.Status = &status
		}
		views, err := service.List(request.Context(), ownerID, filters)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		data := make([]map[string]any, len(views))
		for index, view := range views {
			data[index] = debtResponse(view)
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": data})
	}
}

func findHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		var asOf *planningdomain.YearMonth
		if raw := request.URL.Query().Get("as_of"); raw != "" {
			month, err := planningdomain.ParseYearMonth(raw)
			if err != nil {
				validationError(w)
				return
			}
			asOf = &month
		}
		view, err := service.Find(request.Context(), ownerID, request.PathValue("debt_id"), asOf)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": debtResponse(view)})
	}
}

func updateHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		var payload updateRequest
		if httpapi.DecodeJSON(request, &payload) != nil || (!payload.Name.Set && !payload.Description.Set) ||
			(payload.Name.Set && payload.Name.Value == nil) {
			validationError(w)
			return
		}
		view, err := service.Update(request.Context(), ownerID, request.PathValue("debt_id"), debtapplication.UpdateInput{
			Name: payload.Name.Value, Description: payload.Description.Value,
			DescriptionSet: payload.Description.Set,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": debtResponse(view)})
	}
}

func changeHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		var payload changeRequest
		if httpapi.DecodeJSON(request, &payload) != nil {
			validationError(w)
			return
		}
		effectiveFrom, err := planningdomain.ParseYearMonth(payload.EffectiveFrom)
		if err != nil {
			validationError(w)
			return
		}
		view, err := service.Change(request.Context(), ownerID, request.PathValue("debt_id"), debtapplication.ChangeInput{
			EffectiveFrom: effectiveFrom, InstallmentAmountCents: payload.InstallmentAmountCents,
			Context: payload.Context,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, map[string]any{"data": debtResponse(view)})
	}
}

func archiveHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		payload := archiveRequest{}
		if request.Body != nil && request.ContentLength != 0 {
			if httpapi.DecodeJSON(request, &payload) != nil {
				validationError(w)
				return
			}
		}
		view, err := service.Archive(request.Context(), ownerID, request.PathValue("debt_id"), debtapplication.ArchiveInput{Reason: payload.Reason})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": debtResponse(view)})
	}
}

func scheduleHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		from, fromErr := planningdomain.ParseYearMonth(request.URL.Query().Get("from"))
		to, toErr := planningdomain.ParseYearMonth(request.URL.Query().Get("to"))
		if fromErr != nil || toErr != nil {
			validationError(w)
			return
		}
		schedule, err := service.Schedule(request.Context(), ownerID, request.PathValue("debt_id"), from, to)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		occurrences := make([]map[string]any, len(schedule.Occurrences))
		for index, occurrence := range schedule.Occurrences {
			var invoicePaymentMonth any
			if occurrence.InvoicePaymentMonth != nil {
				invoicePaymentMonth = occurrence.InvoicePaymentMonth.String()
			}
			occurrences[index] = map[string]any{
				"debt_id":               occurrence.Occurrence.DebtID,
				"source_id":             occurrence.Occurrence.SourceID,
				"reference_month":       occurrence.Occurrence.ReferenceMonth.String(),
				"installment_number":    occurrence.Occurrence.InstallmentNumber,
				"installments_total":    occurrence.Occurrence.InstallmentsTotal,
				"amount_cents":          occurrence.Occurrence.Amount.Cents(),
				"debt_occurrence_kind":  occurrence.Occurrence.Kind,
				"payment_method":        occurrence.PaymentMethod,
				"cash_month":            occurrence.CashMonth.String(),
				"credit_card_id":        occurrence.CreditCardID,
				"invoice_payment_month": invoicePaymentMonth,
				"completeness":          occurrence.Completeness,
			}
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
			"debt":        debtResponse(schedule.Debt),
			"range":       map[string]string{"from": schedule.From.String(), "to": schedule.To.String()},
			"occurrences": occurrences,
		}})
	}
}

func debtResponse(view debtapplication.View) map[string]any {
	debt := view.Debt
	projection := view.Projection
	var originalTotal any
	if debt.OriginalTotal != nil {
		originalTotal = debt.OriginalTotal.Cents()
	}
	var settlement any
	if debt.Settlement != nil {
		settlement = map[string]any{
			"id": debt.Settlement.ID, "debt_id": debt.ID,
			"reference_month": debt.Settlement.ReferenceMonth.String(),
			"amount_cents":    debt.Settlement.Amount.Cents(), "reason": debt.Settlement.Reason,
			"recorded_by": debt.Settlement.RecordedBy, "recorded_at": debt.Settlement.RecordedAt,
			"created_at": debt.Settlement.CreatedAt,
		}
	}
	return map[string]any{
		"id": debt.ID, "currency_code": debt.CurrencyCode, "name": debt.Name,
		"description": debt.Description, "original_total_cents": originalTotal,
		"total_installments":          debt.TotalInstallments,
		"first_projected_installment": debt.FirstProjectedInstallment,
		"scheduled_start_month":       debt.ScheduledStart.String(),
		"scheduled_end_month":         projection.ScheduledEnd.String(),
		"effective_end_month":         projection.EffectiveEnd.String(),
		"release_from_month":          projection.ReleaseFrom.String(),
		"released_monthly_cents":      projection.ReleasedMonthly.Cents(),
		"as_of_month":                 view.AsOf.String(), "projection_status": projection.ProjectionStatus,
		"remaining_installments": projection.RemainingInstallments,
		"status":                 debt.Status, "settlement": settlement,
		"created_at": debt.CreatedAt, "updated_at": debt.UpdatedAt,
	}
}

func ownerFromRequest(request *http.Request) (string, bool) {
	principal, ok := httpapi.PrincipalFromRequest(request)
	return principal.UserID, ok
}

func unauthorized(w http.ResponseWriter) {
	httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
}

func validationError(w http.ResponseWriter) {
	httpapi.WriteError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, debtapplication.ErrValidation):
		validationError(w)
	case errors.Is(err, debtapplication.ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, debtapplication.ErrArchived):
		httpapi.WriteError(w, http.StatusConflict, "debt_archived", "The debt is archived")
	case errors.Is(err, debtapplication.ErrPaymentMethodUnsupported):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "invalid_payment_method", "The payment method is incompatible with this resource")
	case errors.Is(err, debtapplication.ErrChangeOutsideSchedule), errors.Is(err, debtapplication.ErrChangeAfterSettlement):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "invalid_debt_schedule", "The debt change is outside the mutable schedule")
	case errors.Is(err, debtdomain.ErrDebtScheduleTooLong):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "debt_schedule_too_long", "The debt schedule exceeds 120 months")
	case errors.Is(err, debtdomain.ErrDebtPeriodGap):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "debt_period_gap", "The debt schedule contains a period gap")
	case errors.Is(err, debtdomain.ErrInvalidDebt):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "invalid_debt_schedule", "The debt schedule is invalid")
	case errors.Is(err, debtdomain.ErrInvalidProjectionRange):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "invalid_debt_schedule", "The debt projection range is invalid")
	case errors.Is(err, debtapplication.ErrProjectionInconsistent):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "debt_projection_inconsistent", "The debt projection is inconsistent")
	default:
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "An unexpected internal error occurred")
	}
}
