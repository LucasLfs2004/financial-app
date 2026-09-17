package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lucas/financial-api/internal/cardinvoice"
	"github.com/lucas/financial-api/internal/planning"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

var (
	ErrValidation                = errors.New("invalid card invoice input")
	ErrOutsideOperationalHorizon = errors.New("invoice outside operational horizon")
)

const MaximumInvoiceRangeMonths = 24

type PlanReader interface {
	Current(context.Context, string) (planning.Plan, error)
}

type Repository interface {
	LoadProjectionData(context.Context, string, string) (cardinvoice.ProjectionData, error)
}

type Service struct {
	repository Repository
	plans      PlanReader
}

func NewService(repository Repository, plans PlanReader) *Service {
	return &Service{repository: repository, plans: plans}
}

func (service *Service) Project(ctx context.Context, ownerID, cardID string, paymentMonth planningdomain.YearMonth) (cardinvoice.Invoice, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(cardID) == "" || !paymentMonth.Valid() {
		return cardinvoice.Invoice{}, ErrValidation
	}
	plan, err := service.plans.Current(ctx, ownerID)
	if err != nil {
		return cardinvoice.Invoice{}, err
	}
	if err := validateOperationalMonth(plan, paymentMonth); err != nil {
		return cardinvoice.Invoice{}, ErrOutsideOperationalHorizon
	}
	data, err := service.repository.LoadProjectionData(ctx, ownerID, plan.ID)
	if err != nil {
		return cardinvoice.Invoice{}, err
	}
	return selectInvoice(plan, data, cardID, paymentMonth)
}

func (service *Service) List(ctx context.Context, ownerID, cardID string, from, to planningdomain.YearMonth, includeEmpty bool) ([]cardinvoice.Invoice, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(cardID) == "" || !from.Valid() || !to.Valid() || to.Before(from) {
		return nil, ErrValidation
	}
	months, err := inclusiveMonths(from, to)
	if err != nil || len(months) > MaximumInvoiceRangeMonths {
		return nil, ErrValidation
	}
	plan, err := service.plans.Current(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if err := validateOperationalMonth(plan, from); err != nil {
		return nil, err
	}
	if err := validateOperationalMonth(plan, to); err != nil {
		return nil, err
	}
	data, err := service.repository.LoadProjectionData(ctx, ownerID, plan.ID)
	if err != nil {
		return nil, err
	}
	invoices := make([]cardinvoice.Invoice, 0, len(months))
	for _, month := range months {
		invoice, err := selectInvoice(plan, data, cardID, month)
		if err != nil {
			return nil, err
		}
		if includeEmpty || len(invoice.Components) > 0 {
			invoices = append(invoices, invoice)
		}
	}
	return invoices, nil
}

func selectInvoice(plan planning.Plan, data cardinvoice.ProjectionData, cardID string, paymentMonth planningdomain.YearMonth) (cardinvoice.Invoice, error) {
	return cardinvoice.SelectInvoiceComponents(cardinvoice.SelectionInput{
		PlanStart: plan.StartMonth, PlanEnd: plan.EndMonth, CardID: cardID,
		PaymentMonth: paymentMonth, CurrencyCode: plan.CurrencyCode,
		Cards: data.Cards, Items: data.Items, Adjustments: data.Adjustments, Moves: data.Moves,
	})
}

func validateOperationalMonth(plan planning.Plan, month planningdomain.YearMonth) error {
	operationalEnd, err := plan.EndMonth.AddMonths(planningdomain.MaximumCashMonthOffset)
	if err != nil {
		return fmt.Errorf("%w: calculate operational horizon", ErrValidation)
	}
	if month.Before(plan.StartMonth) || month.After(operationalEnd) {
		return ErrOutsideOperationalHorizon
	}
	return nil
}

func inclusiveMonths(from, to planningdomain.YearMonth) ([]planningdomain.YearMonth, error) {
	months := make([]planningdomain.YearMonth, 0)
	for month := from; !month.After(to); {
		months = append(months, month)
		if len(months) > MaximumInvoiceRangeMonths {
			return months, nil
		}
		next, err := month.AddMonths(1)
		if err != nil {
			return nil, err
		}
		month = next
	}
	return months, nil
}
