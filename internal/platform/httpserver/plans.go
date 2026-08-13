package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/platform/auth"
)

type PlanService interface {
	Create(context.Context, string, planning.CreateInput) (planning.Plan, error)
	Current(context.Context, string) (planning.Plan, error)
	UpdateCurrent(context.Context, string, planning.UpdateInput) (planning.Plan, error)
}

type createPlanRequest struct {
	Name         string `json:"name"`
	StartMonth   string `json:"start_month"`
	EndMonth     string `json:"end_month"`
	CurrencyCode string `json:"currency_code"`
}

type updatePlanRequest struct {
	Name         *string `json:"name"`
	StartMonth   *string `json:"start_month"`
	EndMonth     *string `json:"end_month"`
	CurrencyCode *string `json:"currency_code"`
}

func createPlanHandler(plans PlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		principal, ok := principalFromRequest(request)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
			return
		}

		var payload createPlanRequest
		if err := decodeJSON(request, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
			return
		}
		startMonth, err := domain.ParseYearMonth(payload.StartMonth)
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
			return
		}
		endMonth, err := domain.ParseYearMonth(payload.EndMonth)
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
			return
		}

		plan, err := plans.Create(request.Context(), principal.UserID, planning.CreateInput{
			Name:         payload.Name,
			StartMonth:   startMonth,
			EndMonth:     endMonth,
			CurrencyCode: payload.CurrencyCode,
		})
		if err != nil {
			writePlanError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": planResponse(plan)})
	}
}

func currentPlanHandler(plans PlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		principal, ok := principalFromRequest(request)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
			return
		}
		plan, err := plans.Current(request.Context(), principal.UserID)
		if err != nil {
			writePlanError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": planResponse(plan)})
	}
}

func updateCurrentPlanHandler(plans PlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		principal, ok := principalFromRequest(request)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
			return
		}
		var payload updatePlanRequest
		if err := decodeJSON(request, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
			return
		}
		input, err := updateInput(payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
			return
		}
		plan, err := plans.UpdateCurrent(request.Context(), principal.UserID, input)
		if err != nil {
			writePlanError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": planResponse(plan)})
	}
}

func updateInput(payload updatePlanRequest) (planning.UpdateInput, error) {
	input := planning.UpdateInput{Name: payload.Name, CurrencyCode: payload.CurrencyCode}
	if payload.StartMonth != nil {
		month, err := domain.ParseYearMonth(*payload.StartMonth)
		if err != nil {
			return planning.UpdateInput{}, err
		}
		input.StartMonth = &month
	}
	if payload.EndMonth != nil {
		month, err := domain.ParseYearMonth(*payload.EndMonth)
		if err != nil {
			return planning.UpdateInput{}, err
		}
		input.EndMonth = &month
	}
	return input, nil
}

func decodeJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain a single JSON object")
	}
	return nil
}

func writePlanError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, planning.ErrValidation):
		writeError(w, http.StatusBadRequest, "validation_error", "Request validation failed")
	case errors.Is(err, planning.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "plan_already_exists", "A current plan already exists")
	case errors.Is(err, planning.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, planning.ErrNotDraft):
		writeError(w, http.StatusConflict, "conflict", "Only draft plans can be changed")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected internal error occurred")
	}
}

func planResponse(plan planning.Plan) map[string]any {
	return map[string]any{
		"id":            plan.ID,
		"name":          plan.Name,
		"status":        plan.Status,
		"start_month":   plan.StartMonth.String(),
		"end_month":     plan.EndMonth.String(),
		"currency_code": plan.CurrencyCode,
		"activated_at":  plan.ActivatedAt,
		"archived_at":   plan.ArchivedAt,
		"created_at":    plan.CreatedAt,
		"updated_at":    plan.UpdatedAt,
	}
}

func principalFromRequest(request *http.Request) (auth.Principal, bool) {
	principal, ok := request.Context().Value(principalContextKey).(auth.Principal)
	return principal, ok && strings.TrimSpace(principal.UserID) != ""
}
