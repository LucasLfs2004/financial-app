package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutionrepository "github.com/lucas/financial-api/internal/financialinstitution/repository"
	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/monthlysummary"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/config"
	"github.com/lucas/financial-api/internal/platform/database"
	"github.com/lucas/financial-api/internal/platform/httpserver"
	"github.com/lucas/financial-api/internal/profile"
	"github.com/lucas/financial-api/internal/savings"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	startupContext, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()

	databasePool, err := database.Open(startupContext, cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer databasePool.Close()

	authClient := auth.NewClient(cfg.Supabase)
	profileRepository := profile.NewRepository(databasePool)
	planRepository := planning.NewPostgresRepository(databasePool)
	planService := planning.NewService(planRepository)
	financialItemRepository := financialitem.NewPostgresRepository(databasePool)
	financialItemService := financialitem.NewService(financialItemRepository, planService)
	savingsRepository := savings.NewPostgresRepository(databasePool)
	savingsService := savings.NewService(savingsRepository, planService)
	monthlySummaryService := monthlysummary.NewService(planService, financialItemRepository, savingsRepository)
	institutionRepository := institutionrepository.NewPostgresRepository(databasePool)
	institutionService := institutionapplication.NewService(institutionRepository)

	server := httpserver.New(cfg, logger, httpserver.Dependencies{
		Database:       databasePool,
		Authenticator:  authClient,
		Profiles:       profileRepository,
		Plans:          planService,
		FinancialItems: financialItemService,
		Savings:        savingsService,
		MonthlySummary: monthlySummaryService,
		Institutions:   institutionService,
	})

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting API", "address", cfg.HTTP.Address, "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("API stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case signal := <-shutdownSignal:
		logger.Info("shutdown signal received", "signal", signal.String())
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("API stopped")
}
