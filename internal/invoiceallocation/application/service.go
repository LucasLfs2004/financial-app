package application

import (
	"context"
	"errors"
	"github.com/lucas/financial-api/internal/invoiceallocation/domain"
	"github.com/lucas/financial-api/internal/planning"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"strings"
)

var (
	ErrNotFound        = errors.New("occurrence not found")
	ErrNotCardLinked   = errors.New("occurrence is not linked to a card")
	ErrSameDestination = errors.New("occurrence already belongs to destination invoice")
	ErrStale           = errors.New("invoice allocation is stale")
	ErrCardNotFound    = errors.New("target card not found")
	ErrCardArchived    = errors.New("target card is archived")
)

type Repository interface {
	Move(context.Context, string, string, planningdomain.YearMonth, string, planningdomain.YearMonth, *string) (domain.Move, error)
	History(context.Context, string, string, planningdomain.YearMonth) ([]domain.Move, error)
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
func (s *Service) Move(ctx context.Context, ownerID, itemID string, reference planningdomain.YearMonth, targetCard string, targetMonth planningdomain.YearMonth, reason *string) (domain.Move, error) {
	if strings.TrimSpace(ownerID) == "" {
		return domain.Move{}, domain.ErrValidation
	}
	if err := domain.Validate(itemID, reference, targetCard, targetMonth); err != nil {
		return domain.Move{}, err
	}
	plan, err := s.plans.Current(ctx, ownerID)
	if err != nil {
		return domain.Move{}, err
	}
	maxMonth, err := plan.EndMonth.AddMonths(12)
	if err != nil {
		return domain.Move{}, domain.ErrValidation
	}
	if targetMonth.Before(plan.StartMonth) || targetMonth.After(maxMonth) {
		return domain.Move{}, domain.ErrValidation
	}
	return s.repository.Move(ctx, ownerID, itemID, reference, targetCard, targetMonth, reason)
}
func (s *Service) History(ctx context.Context, ownerID, itemID string, reference planningdomain.YearMonth) ([]domain.Move, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(itemID) == "" || !reference.Valid() {
		return nil, domain.ErrValidation
	}
	return s.repository.History(ctx, ownerID, itemID, reference)
}
