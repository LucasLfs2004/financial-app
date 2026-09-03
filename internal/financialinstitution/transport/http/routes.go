package http

import (
	"context"
	"errors"
	"net/http"

	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
	"github.com/lucas/financial-api/internal/platform/httpapi"
)

type Service interface {
	Create(context.Context, string, institutionapplication.CreateInput) (institutiondomain.Institution, error)
	List(context.Context, string, institutionapplication.ListFilters) ([]institutiondomain.Institution, error)
	Update(context.Context, string, string, institutionapplication.UpdateInput) (institutiondomain.Institution, error)
	Archive(context.Context, string, string, institutionapplication.ArchiveInput) (institutiondomain.Institution, error)
}

type Middleware func(http.Handler) http.Handler

func RegisterRoutes(mux *http.ServeMux, authenticate Middleware, service Service) {
	mux.Handle("POST /v1/financial-institutions", authenticate(http.HandlerFunc(createHandler(service))))
	mux.Handle("GET /v1/financial-institutions", authenticate(http.HandlerFunc(listHandler(service))))
	mux.Handle("PATCH /v1/financial-institutions/{institution_id}", authenticate(http.HandlerFunc(updateHandler(service))))
	mux.Handle("POST /v1/financial-institutions/{institution_id}/archive", authenticate(http.HandlerFunc(archiveHandler(service))))
}

type createRequest struct {
	Name string `json:"name"`
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
			returnUnauthorized(w)
			return
		}
		var payload createRequest
		if err := httpapi.DecodeJSON(request, &payload); err != nil {
			returnValidationError(w)
			return
		}
		institution, err := service.Create(request.Context(), ownerID, institutionapplication.CreateInput{Name: payload.Name})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, map[string]any{"data": institutionResponse(institution)})
	}
}

func listHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			returnUnauthorized(w)
			return
		}
		filters := institutionapplication.ListFilters{}
		if rawStatus := request.URL.Query().Get("status"); rawStatus != "" {
			status, err := institutiondomain.ParseStatus(rawStatus)
			if err != nil {
				returnValidationError(w)
				return
			}
			filters.Status = &status
		}
		institutions, err := service.List(request.Context(), ownerID, filters)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		data := make([]map[string]any, len(institutions))
		for index, institution := range institutions {
			data[index] = institutionResponse(institution)
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": data})
	}
}

func updateHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			returnUnauthorized(w)
			return
		}
		var payload updateRequest
		if err := httpapi.DecodeJSON(request, &payload); err != nil {
			returnValidationError(w)
			return
		}
		institution, err := service.Update(request.Context(), ownerID, request.PathValue("institution_id"), institutionapplication.UpdateInput{Name: payload.Name})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": institutionResponse(institution)})
	}
}

func archiveHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ownerID, ok := ownerFromRequest(request)
		if !ok {
			returnUnauthorized(w)
			return
		}
		payload := archiveRequest{}
		if request.Body != nil && request.ContentLength != 0 {
			if err := httpapi.DecodeJSON(request, &payload); err != nil {
				returnValidationError(w)
				return
			}
		}
		institution, err := service.Archive(request.Context(), ownerID, request.PathValue("institution_id"), institutionapplication.ArchiveInput{Reason: payload.Reason})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": institutionResponse(institution)})
	}
}

func ownerFromRequest(request *http.Request) (string, bool) {
	principal, ok := httpapi.PrincipalFromRequest(request)
	return principal.UserID, ok
}

func returnUnauthorized(w http.ResponseWriter) {
	httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
}

func returnValidationError(w http.ResponseWriter) {
	httpapi.WriteError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, institutiondomain.ErrValidation):
		returnValidationError(w)
	case errors.Is(err, institutionapplication.ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, institutionapplication.ErrAlreadyExists):
		httpapi.WriteError(w, http.StatusConflict, "conflict", "The operation conflicts with the current resource state")
	case errors.Is(err, institutionapplication.ErrArchived):
		httpapi.WriteError(w, http.StatusConflict, "institution_archived", "The financial institution is archived")
	case errors.Is(err, institutionapplication.ErrHasActiveCards):
		httpapi.WriteError(w, http.StatusConflict, "institution_has_active_cards", "Archive active cards before archiving the institution")
	default:
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "An unexpected internal error occurred")
	}
}

func institutionResponse(institution institutiondomain.Institution) map[string]any {
	return map[string]any{
		"id":          institution.ID,
		"name":        institution.Name,
		"status":      institution.Status,
		"archived_at": institution.ArchivedAt,
		"created_at":  institution.CreatedAt,
		"updated_at":  institution.UpdatedAt,
	}
}
