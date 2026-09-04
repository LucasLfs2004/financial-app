package http

import (
	"context"
	"errors"
	"net/http"

	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/httpapi"
)

type Service interface {
	Create(context.Context, string, cardapplication.CreateInput) (carddomain.Card, error)
	List(context.Context, string, cardapplication.ListFilters) ([]carddomain.Card, error)
	Find(context.Context, string, string) (carddomain.Card, error)
	Update(context.Context, string, string, cardapplication.UpdateInput) (carddomain.Card, error)
	Change(context.Context, string, string, cardapplication.ChangeInput) (carddomain.Card, error)
	Archive(context.Context, string, string, cardapplication.ArchiveInput) (carddomain.Card, error)
}

type Middleware func(http.Handler) http.Handler

func RegisterRoutes(mux *http.ServeMux, authenticate Middleware, service Service) {
	mux.Handle("POST /v1/credit-cards", authenticate(http.HandlerFunc(createHandler(service))))
	mux.Handle("GET /v1/credit-cards", authenticate(http.HandlerFunc(listHandler(service))))
	mux.Handle("GET /v1/credit-cards/{card_id}", authenticate(http.HandlerFunc(findHandler(service))))
	mux.Handle("PATCH /v1/credit-cards/{card_id}", authenticate(http.HandlerFunc(updateHandler(service))))
	mux.Handle("POST /v1/credit-cards/{card_id}/changes", authenticate(http.HandlerFunc(changeHandler(service))))
	mux.Handle("POST /v1/credit-cards/{card_id}/archive", authenticate(http.HandlerFunc(archiveHandler(service))))
}

type configurationRequest struct {
	EffectiveFrom      string                 `json:"effective_from"`
	EndMonth           httpapi.OptionalString `json:"end_month"`
	NominalDueDay      *int                   `json:"nominal_due_day"`
	PaymentMonthOffset *int                   `json:"payment_month_offset"`
	Context            *string                `json:"context"`
}

type createRequest struct {
	InstitutionID string               `json:"institution_id"`
	Name          string               `json:"name"`
	Configuration configurationRequest `json:"configuration"`
}

type updateRequest struct {
	Name string `json:"name"`
}
type archiveRequest struct {
	Reason *string `json:"reason"`
}

func createHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		var payload createRequest
		if httpapi.DecodeJSON(request, &payload) != nil {
			validationError(w)
			return
		}
		configuration, err := parseConfiguration(payload.Configuration)
		if err != nil {
			validationError(w)
			return
		}
		card, err := service.Create(request.Context(), ownerID, cardapplication.CreateInput{
			InstitutionID: payload.InstitutionID, Name: payload.Name, Configuration: configuration,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, map[string]any{"data": cardResponse(card)})
	}
}

func listHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		filters := cardapplication.ListFilters{}
		if rawStatus := request.URL.Query().Get("status"); rawStatus != "" {
			status, err := institutiondomain.ParseStatus(rawStatus)
			if err != nil {
				validationError(w)
				return
			}
			filters.Status = &status
		}
		cards, err := service.List(request.Context(), ownerID, filters)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		data := make([]map[string]any, len(cards))
		for index, card := range cards {
			data[index] = cardResponse(card)
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
		card, err := service.Find(request.Context(), ownerID, request.PathValue("card_id"))
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": cardResponse(card)})
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
		if httpapi.DecodeJSON(request, &payload) != nil {
			validationError(w)
			return
		}
		card, err := service.Update(request.Context(), ownerID, request.PathValue("card_id"), cardapplication.UpdateInput{Name: payload.Name})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": cardResponse(card)})
	}
}

func changeHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			unauthorized(w)
			return
		}
		var payload configurationRequest
		if httpapi.DecodeJSON(request, &payload) != nil {
			validationError(w)
			return
		}
		configuration, err := parseConfiguration(payload)
		if err != nil {
			validationError(w)
			return
		}
		card, err := service.Change(request.Context(), ownerID, request.PathValue("card_id"), cardapplication.ChangeInput{Configuration: configuration})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, map[string]any{"data": cardResponse(card)})
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
		card, err := service.Archive(request.Context(), ownerID, request.PathValue("card_id"), cardapplication.ArchiveInput{Reason: payload.Reason})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": cardResponse(card)})
	}
}

func parseConfiguration(payload configurationRequest) (carddomain.ConfigurationInput, error) {
	if !payload.EndMonth.Set || payload.NominalDueDay == nil || payload.PaymentMonthOffset == nil {
		return carddomain.ConfigurationInput{}, carddomain.ErrValidation
	}
	effectiveFrom, err := planningdomain.ParseYearMonth(payload.EffectiveFrom)
	if err != nil {
		return carddomain.ConfigurationInput{}, err
	}
	var endMonth *planningdomain.YearMonth
	if payload.EndMonth.Value != nil {
		parsed, parseErr := planningdomain.ParseYearMonth(*payload.EndMonth.Value)
		if parseErr != nil {
			return carddomain.ConfigurationInput{}, parseErr
		}
		endMonth = &parsed
	}
	return carddomain.ConfigurationInput{
		EffectiveFrom: effectiveFrom, EndMonth: endMonth,
		NominalDueDay: *payload.NominalDueDay, PaymentMonthOffset: *payload.PaymentMonthOffset,
		Context: payload.Context,
	}, nil
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
	case errors.Is(err, carddomain.ErrValidation):
		validationError(w)
	case errors.Is(err, cardapplication.ErrNotFound), errors.Is(err, cardapplication.ErrInstitutionNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, cardapplication.ErrAlreadyExists):
		httpapi.WriteError(w, http.StatusConflict, "conflict", "The operation conflicts with the current resource state")
	case errors.Is(err, cardapplication.ErrArchived):
		httpapi.WriteError(w, http.StatusConflict, "card_archived", "The credit card is archived")
	case errors.Is(err, cardapplication.ErrInstitutionArchived):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "institution_archived", "The financial institution is archived")
	case errors.Is(err, cardapplication.ErrPeriodOverlap):
		httpapi.WriteError(w, http.StatusConflict, "period_overlap", "The configuration period overlaps another period")
	case errors.Is(err, cardapplication.ErrConfigurationMissing):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "card_configuration_missing", "No credit card configuration applies to the effective month")
	default:
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "An unexpected internal error occurred")
	}
}

func cardResponse(card carddomain.Card) map[string]any {
	configurations := make([]map[string]any, len(card.Configurations))
	for index, configuration := range card.Configurations {
		var endMonth any
		if configuration.EndMonth != nil {
			endMonth = configuration.EndMonth.String()
		}
		configurations[index] = map[string]any{
			"id": configuration.ID, "start_month": configuration.StartMonth.String(), "end_month": endMonth,
			"nominal_due_day": configuration.NominalDueDay, "payment_month_offset": configuration.PaymentMonthOffset,
			"context": configuration.Context, "recorded_at": configuration.RecordedAt, "created_at": configuration.CreatedAt,
		}
	}
	return map[string]any{
		"id": card.ID,
		"institution": map[string]any{
			"id": card.Institution.ID, "name": card.Institution.Name, "status": card.Institution.Status,
			"archived_at": card.Institution.ArchivedAt, "created_at": card.Institution.CreatedAt, "updated_at": card.Institution.UpdatedAt,
		},
		"name": card.Name, "status": card.Status, "archived_at": card.ArchivedAt,
		"configurations": configurations, "created_at": card.CreatedAt, "updated_at": card.UpdatedAt,
	}
}
