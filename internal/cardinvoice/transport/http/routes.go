package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/lucas/financial-api/internal/cardinvoice"
	invoiceapplication "github.com/lucas/financial-api/internal/cardinvoice/application"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/httpapi"
)

type Service interface {
	Project(context.Context, string, string, planningdomain.YearMonth) (cardinvoice.Invoice, error)
	List(context.Context, string, string, planningdomain.YearMonth, planningdomain.YearMonth, bool) ([]cardinvoice.Invoice, error)
}

type Middleware func(http.Handler) http.Handler

func RegisterRoutes(mux *http.ServeMux, authenticate Middleware, service Service) {
	mux.Handle("GET /v1/credit-cards/{card_id}/invoices", authenticate(http.HandlerFunc(listHandler(service))))
	mux.Handle("GET /v1/credit-cards/{card_id}/invoices/{payment_month}", authenticate(http.HandlerFunc(detailHandler(service))))
}

func listHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		from, err := planningdomain.ParseYearMonth(request.URL.Query().Get("from"))
		if err != nil {
			validationError(w)
			return
		}
		to, err := planningdomain.ParseYearMonth(request.URL.Query().Get("to"))
		if err != nil {
			validationError(w)
			return
		}
		includeEmpty := false
		if raw := request.URL.Query().Get("include_empty"); raw != "" {
			includeEmpty, err = strconv.ParseBool(raw)
			if err != nil {
				validationError(w)
				return
			}
		}
		invoices, err := service.List(request.Context(), ownerID, request.PathValue("card_id"), from, to, includeEmpty)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		data := make([]map[string]any, len(invoices))
		for index, invoice := range invoices {
			data[index] = invoiceSummaryResponse(invoice)
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{
			"data":  data,
			"range": map[string]string{"from": from.String(), "to": to.String()},
		})
	}
}

func detailHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		paymentMonth, err := planningdomain.ParseYearMonth(request.PathValue("payment_month"))
		if err != nil {
			validationError(w)
			return
		}
		invoice, err := service.Project(request.Context(), ownerID, request.PathValue("card_id"), paymentMonth)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		response := invoiceSummaryResponse(invoice)
		components := make([]map[string]any, len(invoice.Components))
		for index, component := range invoice.Components {
			components[index] = componentResponse(component)
		}
		response["components"] = components
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": response})
	}
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, invoiceapplication.ErrValidation):
		validationError(w)
	case errors.Is(err, cardinvoice.ErrCardNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, invoiceapplication.ErrOutsideOperationalHorizon):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "invoice_outside_operational_horizon", "Invoice month is outside the operational horizon")
	case errors.Is(err, cardinvoice.ErrCardConfigurationMissing):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "card_configuration_missing", "No credit card configuration applies to the invoice month")
	case errors.Is(err, cardinvoice.ErrInconsistentProjection), errors.Is(err, cardinvoice.ErrDuplicateOccurrence), errors.Is(err, cardinvoice.ErrDuplicateAdjustment):
		httpapi.WriteError(w, http.StatusConflict, "invoice_inconsistent", "Invoice components are inconsistent")
	default:
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "An unexpected internal error occurred")
	}
}

func invoiceSummaryResponse(invoice cardinvoice.Invoice) map[string]any {
	var dueDate any
	if invoice.NominalDueDate.Date != nil {
		dueDate = invoice.NominalDueDate.Date.Format(time.DateOnly)
	}
	return map[string]any{
		"card_id": invoice.CardID, "card_name": invoice.CardName,
		"institution":   institutionResponse(invoice.Institution),
		"payment_month": invoice.PaymentMonth.String(), "nominal_due_day": invoice.NominalDueDate.Day,
		"nominal_due_date": dueDate, "nominal_due_date_resolution": invoice.NominalDueDate.Resolution,
		"currency_code": invoice.CurrencyCode, "projected_total_cents": invoice.Total.Cents(),
		"component_count": len(invoice.Components),
	}
}

func institutionResponse(institution cardinvoice.Institution) map[string]any {
	return map[string]any{
		"id": institution.ID, "name": institution.Name, "status": institution.Status,
		"archived_at": institution.ArchivedAt, "created_at": institution.CreatedAt, "updated_at": institution.UpdatedAt,
	}
}

func componentResponse(component cardinvoice.Component) map[string]any {
	var referenceMonth any
	if component.ReferenceMonth != nil {
		referenceMonth = component.ReferenceMonth.String()
	}
	return map[string]any{
		"source_id": component.SourceID, "source_type": component.SourceType,
		"item_id": component.ItemID, "adjustment_id": component.AdjustmentID,
		"name": component.Name, "reference_month": referenceMonth,
		"reference_known": component.ReferenceKnown, "payment_month": component.PaymentMonth.String(),
		"amount_cents": component.Amount.Cents(), "allocation": component.Allocation,
		"debt_id": component.DebtID, "installment_number": component.InstallmentNumber,
		"installments_total":   component.InstallmentsTotal,
		"debt_occurrence_kind": component.DebtOccurrenceKind,
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
