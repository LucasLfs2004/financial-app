package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/profile"
)

const (
	maximumNameLength        = 120
	maximumDescriptionLength = 1000
	maximumContextLength     = 500
	maximumReasonLength      = 500
)

var (
	ErrValidation               = errors.New("invalid debt input")
	ErrNotFound                 = errors.New("debt not found")
	ErrArchived                 = errors.New("debt is archived")
	ErrPaymentMethodUnsupported = errors.New("debt payment method is not available yet")
)

type Repository interface {
	Create(context.Context, string, string, CreateRecord) (debtdomain.Debt, error)
	List(context.Context, string, *planningdomain.FinancialItemStatus) ([]debtdomain.Debt, error)
	Find(context.Context, string, string) (debtdomain.Debt, error)
	Update(context.Context, string, string, UpdateInput) (debtdomain.Debt, error)
	Archive(context.Context, string, string) (debtdomain.Debt, error)
}

type ProfileReader interface {
	FindByID(context.Context, string) (profile.Profile, error)
}

type Service struct {
	repository Repository
	profiles   ProfileReader
	now        func() time.Time
}

func NewService(repository Repository, profiles ProfileReader) *Service {
	return &Service{repository: repository, profiles: profiles, now: time.Now}
}

type CreateInput struct {
	Name                      string
	Description               *string
	OriginalTotalCents        *int64
	TotalInstallments         int
	FirstProjectedInstallment int
	ScheduledStart            planningdomain.YearMonth
	InstallmentAmountCents    int64
	CashMonthOffset           int
	Context                   *string
	PaymentMethod             string
	CreditCardID              *string
}

type CreateRecord struct {
	Name                      string
	Description               *string
	OriginalTotalCents        *int64
	TotalInstallments         int
	FirstProjectedInstallment int
	ScheduledStart            planningdomain.YearMonth
	ScheduledEnd              planningdomain.YearMonth
	InstallmentAmountCents    int64
	CashMonthOffset           int
	Context                   *string
}

type Filters struct {
	AsOf             *planningdomain.YearMonth
	ProjectionStatus *debtdomain.ProjectionStatus
	Status           *planningdomain.FinancialItemStatus
}

type UpdateInput struct {
	Name           *string
	Description    *string
	DescriptionSet bool
}

type ArchiveInput struct{ Reason *string }

type View struct {
	Debt       debtdomain.Debt
	Projection debtdomain.Projection
	AsOf       planningdomain.YearMonth
}

func (service *Service) Create(ctx context.Context, ownerID string, input CreateInput) (View, error) {
	if strings.TrimSpace(ownerID) == "" {
		return View{}, ErrValidation
	}
	currentProfile, err := service.profiles.FindByID(ctx, ownerID)
	if err != nil {
		return View{}, fmt.Errorf("load debt owner profile: %w", err)
	}
	record, err := validateCreate(input)
	if err != nil {
		return View{}, err
	}
	debt, err := service.repository.Create(ctx, ownerID, currentProfile.CurrencyCode, record)
	if err != nil {
		return View{}, err
	}
	return projectView(debt, currentMonth(service.now(), currentProfile.Timezone))
}

func (service *Service) List(ctx context.Context, ownerID string, filters Filters) ([]View, error) {
	if strings.TrimSpace(ownerID) == "" ||
		(filters.AsOf != nil && !filters.AsOf.Valid()) ||
		(filters.ProjectionStatus != nil && !filters.ProjectionStatus.Valid()) ||
		(filters.Status != nil && !filters.Status.Valid()) {
		return nil, ErrValidation
	}
	asOf, err := service.resolveAsOf(ctx, ownerID, filters.AsOf)
	if err != nil {
		return nil, err
	}
	debts, err := service.repository.List(ctx, ownerID, filters.Status)
	if err != nil {
		return nil, err
	}
	views := make([]View, 0, len(debts))
	for _, debt := range debts {
		view, projectErr := projectView(debt, asOf)
		if projectErr != nil {
			return nil, projectErr
		}
		if filters.ProjectionStatus == nil || view.Projection.ProjectionStatus == *filters.ProjectionStatus {
			views = append(views, view)
		}
	}
	sort.Slice(views, func(i, j int) bool {
		comparison := views[i].Projection.EffectiveEnd.Compare(views[j].Projection.EffectiveEnd)
		if comparison == 0 {
			return views[i].Debt.ID < views[j].Debt.ID
		}
		return comparison < 0
	})
	return views, nil
}

func (service *Service) Find(ctx context.Context, ownerID, debtID string, asOfInput *planningdomain.YearMonth) (View, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(debtID) == "" ||
		(asOfInput != nil && !asOfInput.Valid()) {
		return View{}, ErrValidation
	}
	asOf, err := service.resolveAsOf(ctx, ownerID, asOfInput)
	if err != nil {
		return View{}, err
	}
	debt, err := service.repository.Find(ctx, ownerID, debtID)
	if err != nil {
		return View{}, err
	}
	return projectView(debt, asOf)
}

func (service *Service) Update(ctx context.Context, ownerID, debtID string, input UpdateInput) (View, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(debtID) == "" ||
		(input.Name == nil && !input.DescriptionSet) {
		return View{}, ErrValidation
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" || utf8.RuneCountInString(name) > maximumNameLength {
			return View{}, ErrValidation
		}
		input.Name = &name
	}
	if input.DescriptionSet && !validOptionalText(input.Description, maximumDescriptionLength) {
		return View{}, ErrValidation
	}
	debt, err := service.repository.Update(ctx, ownerID, debtID, input)
	if err != nil {
		return View{}, err
	}
	asOf, err := service.resolveAsOf(ctx, ownerID, nil)
	if err != nil {
		return View{}, err
	}
	return projectView(debt, asOf)
}

func (service *Service) Archive(ctx context.Context, ownerID, debtID string, input ArchiveInput) (View, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(debtID) == "" ||
		!validOptionalText(input.Reason, maximumReasonLength) {
		return View{}, ErrValidation
	}
	debt, err := service.repository.Archive(ctx, ownerID, debtID)
	if err != nil {
		return View{}, err
	}
	asOf, err := service.resolveAsOf(ctx, ownerID, nil)
	if err != nil {
		return View{}, err
	}
	return projectView(debt, asOf)
}

func validateCreate(input CreateInput) (CreateRecord, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || utf8.RuneCountInString(input.Name) > maximumNameLength ||
		!validOptionalText(input.Description, maximumDescriptionLength) ||
		!validOptionalText(input.Context, maximumContextLength) ||
		input.InstallmentAmountCents <= 0 || input.CashMonthOffset < 0 || input.CashMonthOffset > 12 {
		return CreateRecord{}, ErrValidation
	}
	if input.PaymentMethod != "" && input.PaymentMethod != "direct" {
		return CreateRecord{}, ErrPaymentMethodUnsupported
	}
	if input.CreditCardID != nil {
		return CreateRecord{}, ErrPaymentMethodUnsupported
	}
	var originalTotal *planningdomain.Money
	if input.OriginalTotalCents != nil {
		value := planningdomain.NewMoney(*input.OriginalTotalCents)
		originalTotal = &value
	}
	remaining := input.TotalInstallments - input.FirstProjectedInstallment + 1
	if remaining < 1 {
		return CreateRecord{}, debtdomain.ErrInvalidDebt
	}
	scheduledEnd, err := input.ScheduledStart.AddMonths(remaining - 1)
	if err != nil {
		return CreateRecord{}, debtdomain.ErrInvalidDebt
	}
	interval, err := planningdomain.NewMonthInterval(input.ScheduledStart, scheduledEnd)
	if err != nil {
		return CreateRecord{}, debtdomain.ErrInvalidDebt
	}
	_, err = debtdomain.NewDebt(debtdomain.NewDebtInput{
		ID: "pending", Name: input.Name, Description: input.Description,
		OriginalTotal: originalTotal, TotalInstallments: input.TotalInstallments,
		FirstProjectedInstallment: input.FirstProjectedInstallment,
		ScheduledStart:            input.ScheduledStart,
		Periods: []debtdomain.InstallmentPeriod{{
			ID: "pending", Interval: interval,
			Amount: planningdomain.NewMoney(input.InstallmentAmountCents),
		}},
		Status: planningdomain.FinancialItemStatusActive,
	})
	if err != nil {
		return CreateRecord{}, err
	}
	return CreateRecord{
		Name: input.Name, Description: input.Description, OriginalTotalCents: input.OriginalTotalCents,
		TotalInstallments: input.TotalInstallments, FirstProjectedInstallment: input.FirstProjectedInstallment,
		ScheduledStart: input.ScheduledStart, ScheduledEnd: scheduledEnd,
		InstallmentAmountCents: input.InstallmentAmountCents, CashMonthOffset: input.CashMonthOffset,
		Context: input.Context,
	}, nil
}

func projectView(debt debtdomain.Debt, asOf planningdomain.YearMonth) (View, error) {
	projection, err := debtdomain.ProjectDebt(debtdomain.ProjectionInput{
		Debt: debt, From: debt.ScheduledStart, To: debt.ScheduledEnd, AsOf: asOf,
	})
	if err != nil {
		return View{}, err
	}
	return View{Debt: debt, Projection: projection, AsOf: asOf}, nil
}

func (service *Service) resolveAsOf(ctx context.Context, ownerID string, provided *planningdomain.YearMonth) (planningdomain.YearMonth, error) {
	if provided != nil {
		return *provided, nil
	}
	currentProfile, err := service.profiles.FindByID(ctx, ownerID)
	if err != nil {
		return planningdomain.YearMonth{}, fmt.Errorf("load debt owner profile: %w", err)
	}
	return currentMonth(service.now(), currentProfile.Timezone), nil
}

func currentMonth(now time.Time, timezone string) planningdomain.YearMonth {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location = time.UTC
	}
	localized := now.In(location)
	month, _ := planningdomain.NewYearMonth(localized.Year(), localized.Month())
	return month
}

func validOptionalText(value *string, maximum int) bool {
	return value == nil || utf8.RuneCountInString(*value) <= maximum
}
