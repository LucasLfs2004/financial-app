package financialitem

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
)

var (
	ErrNotFound      = errors.New("financial item not found")
	ErrValidation    = errors.New("invalid financial item input")
	ErrPeriodOverlap = errors.New("financial period overlap")
)

type Period struct {
	ID              string
	StartMonth      domain.YearMonth
	EndMonth        *domain.YearMonth
	AmountCents     int64
	Recurrence      domain.Recurrence
	CashMonthOffset int
	Context         *string
	RecordedAt      time.Time
	CreatedAt       time.Time
}

type Item struct {
	ID           string
	CurrencyCode string
	Name         string
	Kind         domain.FinancialItemKind
	Description  *string
	Status       domain.FinancialItemStatus
	ArchivedAt   *time.Time
	Periods      []Period
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PeriodInput struct {
	StartMonth      domain.YearMonth
	EndMonth        *domain.YearMonth
	AmountCents     int64
	Recurrence      domain.Recurrence
	CashMonthOffset int
	Context         *string
}

type CreateInput struct {
	Name        string
	Kind        domain.FinancialItemKind
	Description *string
	Period      PeriodInput
}

type UpdateInput struct {
	Name           *string
	Description    *string
	DescriptionSet bool
}

type ChangeInput struct {
	EffectiveFrom   domain.YearMonth
	EndMonth        *domain.YearMonth
	AmountCents     int64
	CashMonthOffset int
	Context         *string
}

type ArchiveInput struct {
	EffectiveFrom *domain.YearMonth
	Reason        *string
}

type Filters struct {
	Kind         *domain.FinancialItemKind
	Status       *domain.FinancialItemStatus
	CurrencyCode *string
}

type Repository interface {
	Create(context.Context, string, string, CreateInput) (Item, error)
	List(context.Context, string, Filters) ([]Item, error)
	Find(context.Context, string, string) (Item, error)
	Update(context.Context, string, string, UpdateInput) (Item, error)
	Change(context.Context, string, string, ChangeInput) (Item, error)
	Archive(context.Context, string, string, ArchiveInput) (Item, error)
}

type PlanReader interface {
	Current(context.Context, string) (planning.Plan, error)
}

type Service struct {
	repository Repository
	plans      PlanReader
}

func NewService(repository Repository, plans PlanReader) *Service {
	return &Service{repository: repository, plans: plans}
}

func (service *Service) Create(ctx context.Context, ownerID string, input CreateInput) (Item, error) {
	plan, err := service.currentPlan(ctx, ownerID)
	if err != nil {
		return Item{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 120 || !input.Kind.Valid() || input.Kind == domain.FinancialItemKindDebtInstallment {
		return Item{}, ErrValidation
	}
	if err := validateDescription(input.Description, 1000); err != nil {
		return Item{}, err
	}
	if err := validatePeriod(input.Period); err != nil {
		return Item{}, err
	}
	if err := domain.ValidateFinancialItemRecurrence(input.Kind, input.Period.Recurrence); err != nil {
		return Item{}, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return service.repository.Create(ctx, ownerID, plan.CurrencyCode, input)
}

func (service *Service) List(ctx context.Context, ownerID string, filters Filters) ([]Item, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, ErrValidation
	}
	if (filters.Kind != nil && !filters.Kind.Valid()) || (filters.Status != nil && !filters.Status.Valid()) {
		return nil, ErrValidation
	}
	return service.repository.List(ctx, ownerID, filters)
}

func (service *Service) Find(ctx context.Context, ownerID, itemID string) (Item, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(itemID) == "" {
		return Item{}, ErrValidation
	}
	return service.repository.Find(ctx, ownerID, itemID)
}

func (service *Service) Update(ctx context.Context, ownerID, itemID string, input UpdateInput) (Item, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(itemID) == "" {
		return Item{}, ErrValidation
	}
	if input.Name == nil && !input.DescriptionSet {
		return Item{}, ErrValidation
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" || len(name) > 120 {
			return Item{}, ErrValidation
		}
		input.Name = &name
	}
	if input.DescriptionSet {
		if err := validateDescription(input.Description, 1000); err != nil {
			return Item{}, err
		}
	}
	return service.repository.Update(ctx, ownerID, itemID, input)
}

func (service *Service) Change(ctx context.Context, ownerID, itemID string, input ChangeInput) (Item, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(itemID) == "" {
		return Item{}, ErrValidation
	}
	item, err := service.repository.Find(ctx, ownerID, itemID)
	if err != nil {
		return Item{}, err
	}
	var recurrence domain.Recurrence
	for _, existing := range item.Periods {
		if !input.EffectiveFrom.Before(existing.StartMonth) && (existing.EndMonth == nil || !input.EffectiveFrom.After(*existing.EndMonth)) {
			recurrence = existing.Recurrence
			break
		}
	}
	if !recurrence.Valid() {
		return Item{}, ErrNotFound
	}
	period := PeriodInput{StartMonth: input.EffectiveFrom, EndMonth: input.EndMonth, AmountCents: input.AmountCents, Recurrence: recurrence, CashMonthOffset: input.CashMonthOffset, Context: input.Context}
	if err := validatePeriod(period); err != nil {
		return Item{}, err
	}
	return service.repository.Change(ctx, ownerID, itemID, input)
}

func (service *Service) Archive(ctx context.Context, ownerID, itemID string, input ArchiveInput) (Item, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(itemID) == "" {
		return Item{}, ErrValidation
	}
	if input.EffectiveFrom != nil && !input.EffectiveFrom.Valid() {
		return Item{}, ErrValidation
	}
	if err := validateDescription(input.Reason, 500); err != nil {
		return Item{}, err
	}
	return service.repository.Archive(ctx, ownerID, itemID, input)
}

func (service *Service) currentPlan(ctx context.Context, ownerID string) (planning.Plan, error) {
	if strings.TrimSpace(ownerID) == "" {
		return planning.Plan{}, ErrValidation
	}
	return service.plans.Current(ctx, ownerID)
}

func validatePeriod(input PeriodInput) error {
	if !input.StartMonth.Valid() || input.AmountCents < 0 || input.CashMonthOffset < 0 || input.CashMonthOffset > 12 || !input.Recurrence.Valid() {
		return ErrValidation
	}
	if input.EndMonth != nil && input.EndMonth.Before(input.StartMonth) {
		return ErrValidation
	}
	if input.Recurrence == domain.RecurrenceOnce && (input.EndMonth == nil || *input.EndMonth != input.StartMonth) {
		return ErrValidation
	}
	return validateDescription(input.Context, 500)
}

func validateDescription(value *string, limit int) error {
	if value != nil && len(*value) > limit {
		return ErrValidation
	}
	return nil
}
