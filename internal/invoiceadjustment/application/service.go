package application

import (
	"context"
	"errors"
	"github.com/lucas/financial-api/internal/invoiceadjustment/domain"
	"github.com/lucas/financial-api/internal/planning"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"strings"
)

var (
	ErrNotFound     = errors.New("invoice adjustment not found")
	ErrCardNotFound = errors.New("credit card not found")
	ErrCardArchived = errors.New("credit card is archived")
	ErrArchived     = errors.New("invoice adjustment is archived")
)

type Input struct {
	Name           string
	AmountCents    int64
	CardID         string
	PaymentMonth   planningdomain.YearMonth
	ReferenceMonth *planningdomain.YearMonth
	Context        *string
}
type Repository interface {
	Create(context.Context, string, string, Input) (domain.Adjustment, error)
	Update(context.Context, string, string, string, Input) (domain.Adjustment, error)
	Archive(context.Context, string, string, string) (domain.Adjustment, error)
	List(context.Context, string, string, string, planningdomain.YearMonth) ([]domain.Adjustment, error)
}
type PlanReader interface {
	Current(context.Context, string) (planning.Plan, error)
}
type Service struct {
	repo  Repository
	plans PlanReader
}

func NewService(repo Repository, plans PlanReader) *Service {
	return &Service{repo: repo, plans: plans}
}
func (s *Service) Create(ctx context.Context, owner string, in Input) (domain.Adjustment, error) {
	if err := s.validate(ctx, owner, in); err != nil {
		return domain.Adjustment{}, err
	}
	p, err := s.plans.Current(ctx, owner)
	if err != nil {
		return domain.Adjustment{}, err
	}
	return s.repo.Create(ctx, owner, p.ID, in)
}
func (s *Service) validate(ctx context.Context, owner string, in Input) error {
	if strings.TrimSpace(owner) == "" {
		return domain.ErrValidation
	}
	if err := domain.Validate(in.Name, in.AmountCents, in.CardID, in.PaymentMonth, in.ReferenceMonth); err != nil {
		return err
	}
	p, err := s.plans.Current(ctx, owner)
	if err != nil {
		return err
	}
	max, e := p.EndMonth.AddMonths(12)
	if e != nil || in.PaymentMonth.Before(p.StartMonth) || in.PaymentMonth.After(max) || (in.ReferenceMonth != nil && (in.ReferenceMonth.Before(p.StartMonth) || in.ReferenceMonth.After(p.EndMonth))) {
		return domain.ErrValidation
	}
	return nil
}
func (s *Service) Update(ctx context.Context, owner, id string, in Input) (domain.Adjustment, error) {
	p, err := s.plans.Current(ctx, owner)
	if err != nil {
		return domain.Adjustment{}, err
	}
	if err = domain.Validate(in.Name, in.AmountCents, in.CardID, in.PaymentMonth, in.ReferenceMonth); err != nil {
		return domain.Adjustment{}, err
	}
	return s.repo.Update(ctx, owner, p.ID, id, in)
}
func (s *Service) Archive(ctx context.Context, owner, id string) (domain.Adjustment, error) {
	p, err := s.plans.Current(ctx, owner)
	if err != nil {
		return domain.Adjustment{}, err
	}
	return s.repo.Archive(ctx, owner, p.ID, id)
}
func (s *Service) List(ctx context.Context, owner, card string, month planningdomain.YearMonth) ([]domain.Adjustment, error) {
	p, err := s.plans.Current(ctx, owner)
	if err != nil {
		return nil, err
	}
	return s.repo.List(ctx, owner, p.ID, card, month)
}
