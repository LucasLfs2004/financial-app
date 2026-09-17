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
	operationalEnd, err := plan.EndMonth.AddMonths(planningdomain.MaximumCashMonthOffset)
	if err != nil {
		return cardinvoice.Invoice{}, fmt.Errorf("%w: calculate operational horizon", ErrValidation)
	}
	if paymentMonth.Before(plan.StartMonth) || paymentMonth.After(operationalEnd) {
		return cardinvoice.Invoice{}, ErrOutsideOperationalHorizon
	}
	data, err := service.repository.LoadProjectionData(ctx, ownerID, plan.ID)
	if err != nil {
		return cardinvoice.Invoice{}, err
	}
	return cardinvoice.SelectInvoiceComponents(cardinvoice.SelectionInput{
		PlanStart: plan.StartMonth, PlanEnd: plan.EndMonth, CardID: cardID,
		PaymentMonth: paymentMonth, CurrencyCode: plan.CurrencyCode,
		Cards: data.Cards, Items: data.Items, Adjustments: data.Adjustments, Moves: data.Moves,
	})
}
