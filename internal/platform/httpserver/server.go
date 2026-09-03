package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	institutionhttp "github.com/lucas/financial-api/internal/financialinstitution/transport/http"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/config"
	"github.com/lucas/financial-api/internal/platform/httpapi"
	"github.com/lucas/financial-api/internal/profile"
)

type Database interface {
	Ping(context.Context) error
}

type Authenticator interface {
	Authenticate(context.Context, string) (auth.Principal, error)
}

type ProfileReader interface {
	FindByID(context.Context, string) (profile.Profile, error)
}

type Dependencies struct {
	Database       Database
	Authenticator  Authenticator
	Profiles       ProfileReader
	Plans          PlanService
	FinancialItems FinancialItemService
	Savings        SavingsService
	MonthlySummary MonthlySummaryService
	Institutions   institutionhttp.Service
}

func New(cfg config.Config, logger *slog.Logger, dependencies Dependencies) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ready", readyHandler(dependencies.Database))
	mux.HandleFunc("GET /", rootHandler)
	mux.Handle("GET /v1/me", authenticate(
		dependencies.Authenticator,
		http.HandlerFunc(meHandler(dependencies.Profiles)),
	))
	mux.Handle("POST /v1/plans", authenticate(
		dependencies.Authenticator,
		http.HandlerFunc(createPlanHandler(dependencies.Plans)),
	))
	mux.Handle("GET /v1/plans/current", authenticate(
		dependencies.Authenticator,
		http.HandlerFunc(currentPlanHandler(dependencies.Plans)),
	))
	mux.Handle("PATCH /v1/plans/current", authenticate(
		dependencies.Authenticator,
		http.HandlerFunc(updateCurrentPlanHandler(dependencies.Plans)),
	))
	mux.Handle("POST /v1/plans/current/activate", authenticate(dependencies.Authenticator, http.HandlerFunc(activatePlanHandler(dependencies.Plans))))
	mux.Handle("GET /v1/plans/current/original", authenticate(dependencies.Authenticator, http.HandlerFunc(originalPlanHandler(dependencies.Plans))))
	mux.Handle("POST /v1/plans/current/items", authenticate(dependencies.Authenticator, http.HandlerFunc(createFinancialItemHandler(dependencies.FinancialItems))))
	mux.Handle("GET /v1/plans/current/items", authenticate(dependencies.Authenticator, http.HandlerFunc(listFinancialItemsHandler(dependencies.FinancialItems))))
	mux.Handle("GET /v1/plans/current/items/{item_id}", authenticate(dependencies.Authenticator, http.HandlerFunc(getFinancialItemHandler(dependencies.FinancialItems))))
	mux.Handle("PATCH /v1/plans/current/items/{item_id}", authenticate(dependencies.Authenticator, http.HandlerFunc(updateFinancialItemHandler(dependencies.FinancialItems))))
	mux.Handle("POST /v1/plans/current/items/{item_id}/changes", authenticate(dependencies.Authenticator, http.HandlerFunc(changeFinancialItemHandler(dependencies.FinancialItems))))
	mux.Handle("POST /v1/plans/current/items/{item_id}/archive", authenticate(dependencies.Authenticator, http.HandlerFunc(archiveFinancialItemHandler(dependencies.FinancialItems))))
	mux.Handle("GET /v1/plans/current/savings", authenticate(dependencies.Authenticator, http.HandlerFunc(getSavingsHandler(dependencies.Savings))))
	mux.Handle("PUT /v1/plans/current/savings", authenticate(dependencies.Authenticator, http.HandlerFunc(putSavingsHandler(dependencies.Savings))))
	mux.Handle("GET /v1/plans/current/months/{month}/summary", authenticate(dependencies.Authenticator, http.HandlerFunc(monthlySummaryHandler(dependencies.MonthlySummary))))
	if dependencies.Institutions != nil {
		institutionhttp.RegisterRoutes(mux, func(next http.Handler) http.Handler {
			return authenticate(dependencies.Authenticator, next)
		}, dependencies.Institutions)
	}

	handler := recoveryMiddleware(logger,
		requestIDMiddleware(
			securityHeadersMiddleware(
				loggingMiddleware(logger, mux),
			),
		),
	)

	return &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func readyHandler(database Database) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
			writeError(w, http.StatusServiceUnavailable, "not_ready", "API dependencies are unavailable")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ready",
		})
	}
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name":    "financial-api",
		"message": "API is running",
	})
}

func meHandler(profiles ProfileReader) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		principal, ok := principalFromRequest(request)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
			return
		}

		currentProfile, err := profiles.FindByID(request.Context(), principal.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "Unable to load the authenticated user")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"id":            principal.UserID,
				"email":         principal.Email,
				"display_name":  currentProfile.DisplayName,
				"timezone":      currentProfile.Timezone,
				"currency_code": currentProfile.CurrencyCode,
			},
		})
	}
}

func authenticate(authenticator Authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		token, ok := bearerToken(request.Header.Get("Authorization"))
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "A valid bearer token is required")
			return
		}

		principal, err := authenticator.Authenticate(request.Context(), token)
		if errors.Is(err, auth.ErrUnauthorized) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "A valid bearer token is required")
			return
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "authentication_unavailable", "Authentication is temporarily unavailable")
			return
		}

		ctx := httpapi.WithPrincipal(request.Context(), principal)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	httpapi.WriteJSON(w, status, payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	httpapi.WriteError(w, status, code, message)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *statusRecorder) Write(data []byte) (int, error) {
	if recorder.status == 0 {
		recorder.WriteHeader(http.StatusOK)
	}
	return recorder.ResponseWriter.Write(data)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, request)

		logger.Info("HTTP request",
			"method", request.Method,
			"path", request.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(startedAt).Milliseconds(),
			"request_id", w.Header().Get("X-Request-ID"),
		)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, request)
	})
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, request)
	})
}

func recoveryMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered",
					"error", recovered,
					"stack", string(debug.Stack()),
					"request_id", w.Header().Get("X-Request-ID"),
				)
				writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			}
		}()

		next.ServeHTTP(w, request)
	})
}

func newRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(value[:])
}
