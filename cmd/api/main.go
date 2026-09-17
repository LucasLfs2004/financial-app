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

	invoiceapplication "github.com/lucas/financial-api/internal/cardinvoice/application"
	invoicerepository "github.com/lucas/financial-api/internal/cardinvoice/repository"
	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	cardrepository "github.com/lucas/financial-api/internal/creditcard/repository"
	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutionrepository "github.com/lucas/financial-api/internal/financialinstitution/repository"
	"github.com/lucas/financial-api/internal/financialitem"
	invoiceadjustmentapplication "github.com/lucas/financial-api/internal/invoiceadjustment/application"
	invoiceadjustmentrepository "github.com/lucas/financial-api/internal/invoiceadjustment/repository"
	invoiceallocationapplication "github.com/lucas/financial-api/internal/invoiceallocation/application"
	invoiceallocationrepository "github.com/lucas/financial-api/internal/invoiceallocation/repository"
	"github.com/lucas/financial-api/internal/monthlysummary"
	paymentmethodapplication "github.com/lucas/financial-api/internal/paymentmethod/application"
	paymentmethodrepository "github.com/lucas/financial-api/internal/paymentmethod/repository"
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
	invoiceRepository := invoicerepository.NewPostgresRepository(databasePool)
	monthlySummaryService := monthlysummary.NewService(planService, financialItemRepository, savingsRepository, invoiceRepository)
	institutionRepository := institutionrepository.NewPostgresRepository(databasePool)
	institutionService := institutionapplication.NewService(institutionRepository)
	cardRepository := cardrepository.NewPostgresRepository(databasePool)
	cardService := cardapplication.NewService(cardRepository)
	paymentMethodRepository := paymentmethodrepository.NewPostgresRepository(databasePool)
	paymentMethodService := paymentmethodapplication.NewService(paymentMethodRepository, planService)
	invoiceMoveRepository := invoiceallocationrepository.NewPostgresRepository(databasePool)
	invoiceMoveService := invoiceallocationapplication.NewService(invoiceMoveRepository, planService)
	invoiceAdjustmentRepository := invoiceadjustmentrepository.NewPostgresRepository(databasePool)
	invoiceAdjustmentService := invoiceadjustmentapplication.NewService(invoiceAdjustmentRepository, planService)
	invoiceService := invoiceapplication.NewService(invoiceRepository, planService)

	server := httpserver.New(cfg, logger, httpserver.Dependencies{
		Database:           databasePool,
		Authenticator:      authClient,
		Profiles:           profileRepository,
		Plans:              planService,
		FinancialItems:     financialItemService,
		Savings:            savingsService,
		MonthlySummary:     monthlySummaryService,
		Institutions:       institutionService,
		CreditCards:        cardService,
		PaymentMethods:     paymentMethodService,
		InvoiceMoves:       invoiceMoveService,
		InvoiceAdjustments: invoiceAdjustmentService,
		CardInvoices:       invoiceService,
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
