package application

import (
	"context"
	"errors"
	"strings"

	"github.com/lucas/financial-api/internal/paymentmethod/domain"
	"github.com/lucas/financial-api/internal/planning"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

var (
	ErrNotFound     = errors.New("payment method period not found")
	ErrCardNotFound = errors.New("credit card not found")
	ErrCardArchived = errors.New("credit card is archived")
	ErrItemNotFound = errors.New("financial item not found")
)

type Repository interface {
	Create(context.Context, string, string, string, Input) (domain.Period, error)
	List(context.Context, string, string, string) ([]domain.Period, error)
}

type PlanReader interface {
	Current(context.Context, string) (planning.Plan, error)
}

type Input struct {
	EffectiveFrom planningdomain.YearMonth
	EndMonth      *planningdomain.YearMonth
	Method        domain.Kind
	CreditCardID  *string
	Context       *string
}

type Service struct {
	repository Repository
	plans      PlanReader
}

func NewService(repository Repository, plans PlanReader) *Service {
	return &Service{repository: repository, plans: plans}
}

func (service *Service) Create(ctx context.Context, ownerID, itemID string, input Input) (domain.Period, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(itemID) == "" || !input.EffectiveFrom.Valid() || !input.Method.Valid() {
		return domain.Period{}, domain.ErrValidation
	}
	if err := domain.Validate(input.Method, input.CreditCardID); err != nil {
		return domain.Period{}, err
	}
	if input.EndMonth != nil && (!input.EndMonth.Valid() || input.EndMonth.Before(input.EffectiveFrom)) {
		return domain.Period{}, domain.ErrValidation
	}
	if input.Context != nil && len(*input.Context) > 500 {
		return domain.Period{}, domain.ErrValidation
	}
	plan, err := service.plans.Current(ctx, ownerID)
	if err != nil {
		return domain.Period{}, err
	}
	if input.EffectiveFrom.Before(plan.StartMonth) || (input.EndMonth != nil && input.EndMonth.After(plan.EndMonth)) {
		return domain.Period{}, domain.ErrValidation
	}
	return service.repository.Create(ctx, ownerID, plan.ID, itemID, input)
}

func (service *Service) List(ctx context.Context, ownerID, itemID string) ([]domain.Period, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(itemID) == "" {
		return nil, domain.ErrValidation
	}
	plan, err := service.plans.Current(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return service.repository.List(ctx, ownerID, plan.ID, itemID)
}
