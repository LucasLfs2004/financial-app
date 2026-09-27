package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/lucas/financial-api/internal/monthlysummary"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
)

type MonthlySummaryService interface {
	Get(context.Context, string, domain.YearMonth, domain.SummaryBasis) (monthlysummary.Summary, error)
}

func monthlySummaryHandler(service MonthlySummaryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := principalFromRequest(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
			return
		}
		month, err := domain.ParseYearMonth(r.PathValue("month"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
			return
		}
		basis := domain.SummaryBasisCash
		if raw := r.URL.Query().Get("basis"); raw != "" {
			basis, err = domain.ParseSummaryBasis(raw)
			if err != nil {
				writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
				return
			}
		}
		summary, err := service.Get(r.Context(), principal.UserID, month, basis)
		if err != nil {
			writeMonthlySummaryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": monthlySummaryResponse(summary)})
	}
}

func writeMonthlySummaryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, monthlysummary.ErrValidation):
		writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
	case errors.Is(err, monthlysummary.ErrOutsideHorizon):
		writeError(w, http.StatusUnprocessableEntity, "month_outside_horizon", "The requested month is outside the plan horizon")
	case errors.Is(err, planning.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Resource not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected internal error occurred")
	}
}

func monthlySummaryResponse(summary monthlysummary.Summary) map[string]any {
	sources := make([]map[string]any, len(summary.Sources))
	for index, source := range summary.Sources {
		referenceMonth := any(nil)
		if source.ReferenceMonth != nil {
			referenceMonth = source.ReferenceMonth.String()
		}
		response := map[string]any{
			"source_id": source.SourceID, "item_id": source.ItemID, "name": source.Name,
			"kind": source.Kind, "effect": source.Effect, "reference_month": referenceMonth,
			"cash_month": source.CashMonth.String(), "amount_cents": source.AmountCents,
		}
		if source.SourceType != "" {
			response["source_type"] = source.SourceType
		}
		if source.PaymentMethod != nil {
			response["payment_method"] = *source.PaymentMethod
		}
		if source.CreditCardID != nil {
			response["credit_card_id"] = *source.CreditCardID
		}
		if source.InvoicePaymentMonth != nil {
			response["invoice_payment_month"] = source.InvoicePaymentMonth.String()
		}
		if source.InvoiceAllocation != nil {
			response["invoice_allocation"] = *source.InvoiceAllocation
		}
		if source.ReferenceKnown != nil {
			response["reference_known"] = *source.ReferenceKnown
		}
		if source.DebtID != nil {
			response["debt_id"] = *source.DebtID
		}
		if source.InstallmentNumber != nil {
			response["installment_number"] = *source.InstallmentNumber
		}
		if source.InstallmentsTotal != nil {
			response["installments_total"] = *source.InstallmentsTotal
		}
		if source.DebtOccurrenceKind != nil {
			response["debt_occurrence_kind"] = *source.DebtOccurrenceKind
		}
		sources[index] = response
	}
	return map[string]any{
		"month": summary.Month.String(), "basis": summary.Basis, "result_kind": summary.ResultKind,
		"currency_code": summary.CurrencyCode, "plan_status": summary.PlanStatus, "income_cents": summary.IncomeCents,
		"commitments_cents": summary.CommitmentsCents, "planned_savings_cents": summary.PlannedSavingsCents,
		"result_cents": summary.ResultCents, "is_negative": summary.IsNegative, "completeness": summary.Completeness,
		"breakdown": map[string]int64{
			"recurring_income_cents":            summary.Breakdown.RecurringIncomeCents,
			"one_time_income_cents":             summary.Breakdown.OneTimeIncomeCents,
			"fixed_expenses_cents":              summary.Breakdown.FixedExpensesCents,
			"projected_variable_expenses_cents": summary.Breakdown.ProjectedVariableExpensesCents,
			"debt_installments_cents":           summary.Breakdown.DebtInstallmentsCents,
			"card_invoice_adjustments_cents":    summary.Breakdown.CardInvoiceAdjustmentsCents,
		},
		"sources": sources,
	}
}
