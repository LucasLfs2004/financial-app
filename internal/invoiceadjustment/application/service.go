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
	CurrencyCode   string
	AmountCents    int64
	CardID         string
	PaymentMonth   planningdomain.YearMonth
	ReferenceMonth *planningdomain.YearMonth
	Context        *string
}
type Repository interface {
	Create(context.Context, string, Input) (domain.Adjustment, error)
	Update(context.Context, string, string, Input) (domain.Adjustment, error)
	Archive(context.Context, string, string) (domain.Adjustment, error)
	List(context.Context, string, string, planningdomain.YearMonth) ([]domain.Adjustment, error)
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
	plan, err := s.plans.Current(ctx, owner)
	if err != nil {
		return domain.Adjustment{}, err
	}
	in.CurrencyCode = plan.CurrencyCode
	return s.repo.Create(ctx, owner, in)
}
func (s *Service) validate(ctx context.Context, owner string, in Input) error {
	if strings.TrimSpace(owner) == "" {
		return domain.ErrValidation
	}
	if err := domain.Validate(in.Name, in.AmountCents, in.CardID, in.PaymentMonth, in.ReferenceMonth); err != nil {
		return err
	}
	return nil
}
func (s *Service) Update(ctx context.Context, owner, id string, in Input) (domain.Adjustment, error) {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(id) == "" {
		return domain.Adjustment{}, domain.ErrValidation
	}
	if err := domain.Validate(in.Name, in.AmountCents, in.CardID, in.PaymentMonth, in.ReferenceMonth); err != nil {
		return domain.Adjustment{}, err
	}
	return s.repo.Update(ctx, owner, id, in)
}
func (s *Service) Archive(ctx context.Context, owner, id string) (domain.Adjustment, error) {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(id) == "" {
		return domain.Adjustment{}, domain.ErrValidation
	}
	return s.repo.Archive(ctx, owner, id)
}
func (s *Service) List(ctx context.Context, owner, card string, month planningdomain.YearMonth) ([]domain.Adjustment, error) {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(card) == "" || !month.Valid() {
		return nil, domain.ErrValidation
	}
	return s.repo.List(ctx, owner, card, month)
}
