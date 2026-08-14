package planning

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lucas/financial-api/internal/planning/domain"
)

var (
	ErrNotFound      = errors.New("plan not found")
	ErrAlreadyExists = errors.New("current plan already exists")
	ErrNotDraft      = errors.New("plan is not a draft")
	ErrValidation    = errors.New("invalid plan input")
)

var currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

type Plan struct {
	ID           string
	UserID       string
	Name         string
	Status       domain.PlanStatus
	StartMonth   domain.YearMonth
	EndMonth     domain.YearMonth
	CurrencyCode string
	ActivatedAt  *time.Time
	ArchivedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateInput struct {
	Name         string
	StartMonth   domain.YearMonth
	EndMonth     domain.YearMonth
	CurrencyCode string
}

type UpdateInput struct {
	Name         *string
	StartMonth   *domain.YearMonth
	EndMonth     *domain.YearMonth
	CurrencyCode *string
}

type Repository interface {
	Create(context.Context, string, CreateInput) (Plan, error)
	FindCurrent(context.Context, string) (Plan, error)
	UpdateDraft(context.Context, string, UpdateInput) (Plan, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) Create(ctx context.Context, ownerID string, input CreateInput) (Plan, error) {
	if strings.TrimSpace(ownerID) == "" {
		return Plan{}, fmt.Errorf("%w: authenticated user is required", ErrValidation)
	}
	if err := validateCreateInput(input); err != nil {
		return Plan{}, err
	}

	plan, err := service.repository.Create(ctx, ownerID, normalizedCreateInput(input))
	if err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (service *Service) Current(ctx context.Context, ownerID string) (Plan, error) {
	if strings.TrimSpace(ownerID) == "" {
		return Plan{}, fmt.Errorf("%w: authenticated user is required", ErrValidation)
	}
	return service.repository.FindCurrent(ctx, ownerID)
}

func (service *Service) UpdateCurrent(ctx context.Context, ownerID string, input UpdateInput) (Plan, error) {
	if strings.TrimSpace(ownerID) == "" {
		return Plan{}, fmt.Errorf("%w: authenticated user is required", ErrValidation)
	}
	if err := validateUpdateInput(input); err != nil {
		return Plan{}, err
	}
	current, err := service.repository.FindCurrent(ctx, ownerID)
	if err != nil {
		return Plan{}, err
	}
	if current.Status != domain.PlanStatusDraft {
		return Plan{}, ErrNotDraft
	}
	if input.StartMonth == nil {
		startMonth := current.StartMonth
		input.StartMonth = &startMonth
	}
	if input.EndMonth == nil {
		endMonth := current.EndMonth
		input.EndMonth = &endMonth
	}
	if input.EndMonth.Before(*input.StartMonth) {
		return Plan{}, fmt.Errorf("%w: start_month must be before or equal to end_month", ErrValidation)
	}

	plan, err := service.repository.UpdateDraft(ctx, ownerID, normalizedUpdateInput(input))
	if err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func validateCreateInput(input CreateInput) error {
	if err := validateName(input.Name); err != nil {
		return err
	}
	if !input.StartMonth.Valid() || !input.EndMonth.Valid() || input.EndMonth.Before(input.StartMonth) {
		return fmt.Errorf("%w: start_month must be before or equal to end_month", ErrValidation)
	}
	if !currencyCodePattern.MatchString(input.CurrencyCode) {
		return fmt.Errorf("%w: currency_code must use three uppercase letters", ErrValidation)
	}
	return nil
}

func validateUpdateInput(input UpdateInput) error {
	if input.Name == nil && input.StartMonth == nil && input.EndMonth == nil && input.CurrencyCode == nil {
		return fmt.Errorf("%w: at least one field is required", ErrValidation)
	}
	if input.Name != nil {
		if err := validateName(*input.Name); err != nil {
			return err
		}
	}
	if input.StartMonth != nil && !input.StartMonth.Valid() {
		return fmt.Errorf("%w: invalid start_month", ErrValidation)
	}
	if input.EndMonth != nil && !input.EndMonth.Valid() {
		return fmt.Errorf("%w: invalid end_month", ErrValidation)
	}
	if input.StartMonth != nil && input.EndMonth != nil && input.EndMonth.Before(*input.StartMonth) {
		return fmt.Errorf("%w: start_month must be before or equal to end_month", ErrValidation)
	}
	if input.CurrencyCode != nil && !currencyCodePattern.MatchString(*input.CurrencyCode) {
		return fmt.Errorf("%w: currency_code must use three uppercase letters", ErrValidation)
	}
	return nil
}

func validateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) > 120 {
		return fmt.Errorf("%w: name must contain between 1 and 120 characters", ErrValidation)
	}
	return nil
}

func normalizedCreateInput(input CreateInput) CreateInput {
	input.Name = strings.TrimSpace(input.Name)
	return input
}

func normalizedUpdateInput(input UpdateInput) UpdateInput {
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		input.Name = &name
	}
	return input
}
