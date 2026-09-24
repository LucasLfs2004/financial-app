package application

import (
	"context"
	"errors"
	"strings"

	"github.com/lucas/financial-api/internal/paymentmethod/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

var (
	ErrNotFound     = errors.New("payment method period not found")
	ErrCardNotFound = errors.New("credit card not found")
	ErrCardArchived = errors.New("credit card is archived")
	ErrItemNotFound = errors.New("financial item not found")
)

type Repository interface {
	Create(context.Context, string, string, Input) (domain.Period, error)
	List(context.Context, string, string) ([]domain.Period, error)
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
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
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
	return service.repository.Create(ctx, ownerID, itemID, input)
}

func (service *Service) List(ctx context.Context, ownerID, itemID string) ([]domain.Period, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(itemID) == "" {
		return nil, domain.ErrValidation
	}
	return service.repository.List(ctx, ownerID, itemID)
}
