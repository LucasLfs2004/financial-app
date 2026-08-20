package savings

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
)

var (
	ErrValidation = errors.New("invalid savings input")
	ErrOverlap    = errors.New("saving period overlap")
)

type Period struct {
	ID          string
	StartMonth  domain.YearMonth
	EndMonth    *domain.YearMonth
	AmountCents int64
	Context     *string
	CreatedAt   time.Time
}
type Configuration struct {
	Configured bool
	Periods    []Period
}
type PutInput struct {
	EffectiveFrom domain.YearMonth
	EndMonth      *domain.YearMonth
	AmountCents   int64
	Context       *string
}

type Repository interface {
	Get(context.Context, string, string) (Configuration, error)
	Put(context.Context, string, string, PutInput) (Configuration, error)
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

func (service *Service) Get(ctx context.Context, ownerID string) (Configuration, error) {
	plan, err := service.current(ctx, ownerID)
	if err != nil {
		return Configuration{}, err
	}
	return service.repository.Get(ctx, ownerID, plan.ID)
}

func (service *Service) Put(ctx context.Context, ownerID string, input PutInput) (Configuration, error) {
	plan, err := service.current(ctx, ownerID)
	if err != nil {
		return Configuration{}, err
	}
	if !input.EffectiveFrom.Valid() || input.EffectiveFrom.Before(plan.StartMonth) || input.EffectiveFrom.After(plan.EndMonth) || input.AmountCents < 0 || input.Context != nil && len(*input.Context) > 500 {
		return Configuration{}, ErrValidation
	}
	if input.EndMonth != nil && (input.EndMonth.Before(input.EffectiveFrom) || input.EndMonth.After(plan.EndMonth)) {
		return Configuration{}, ErrValidation
	}
	return service.repository.Put(ctx, ownerID, plan.ID, input)
}

func (service *Service) current(ctx context.Context, ownerID string) (planning.Plan, error) {
	if strings.TrimSpace(ownerID) == "" {
		return planning.Plan{}, ErrValidation
	}
	return service.plans.Current(ctx, ownerID)
}
