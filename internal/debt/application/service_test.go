package application

import (
	"context"
	"errors"
	"testing"
	"time"

	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/profile"
)

type repositoryStub struct {
	created CreateRecord
	debts   []debtdomain.Debt
	err     error
}

func (repository *repositoryStub) Create(_ context.Context, ownerID, currency string, input CreateRecord) (debtdomain.Debt, error) {
	repository.created = input
	if repository.err != nil {
		return debtdomain.Debt{}, repository.err
	}
	return debtFixture(ownerID, currency, "debt", input.Name, input.ScheduledStart, input.ScheduledEnd, input.TotalInstallments, input.FirstProjectedInstallment, input.InstallmentAmountCents), nil
}
func (repository *repositoryStub) List(context.Context, string, *planningdomain.FinancialItemStatus) ([]debtdomain.Debt, error) {
	return repository.debts, repository.err
}
func (repository *repositoryStub) Find(context.Context, string, string) (debtdomain.Debt, error) {
	if repository.err != nil {
		return debtdomain.Debt{}, repository.err
	}
	return repository.debts[0], nil
}
func (repository *repositoryStub) Update(context.Context, string, string, UpdateInput) (debtdomain.Debt, error) {
	if repository.err != nil {
		return debtdomain.Debt{}, repository.err
	}
	return repository.debts[0], nil
}
func (repository *repositoryStub) Change(context.Context, string, string, ChangeInput) (debtdomain.Debt, error) {
	if repository.err != nil {
		return debtdomain.Debt{}, repository.err
	}
	return repository.debts[0], nil
}
func (repository *repositoryStub) Archive(context.Context, string, string) (debtdomain.Debt, error) {
	if repository.err != nil {
		return debtdomain.Debt{}, repository.err
	}
	return repository.debts[0], nil
}

type profileStub struct{ value profile.Profile }

func (stub profileStub) FindByID(context.Context, string) (profile.Profile, error) {
	return stub.value, nil
}

func TestCreateValidatesAndDelegatesCompleteSchedule(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository, profileStub{profile.Profile{CurrencyCode: "BRL", Timezone: "America/Sao_Paulo"}})
	service.now = func() time.Time { return time.Date(2026, time.September, 1, 2, 0, 0, 0, time.UTC) }
	original := int64(720000)
	view, err := service.Create(context.Background(), "owner", CreateInput{
		Name: "  Transplante  ", OriginalTotalCents: &original,
		TotalInstallments: 12, FirstProjectedInstallment: 5,
		ScheduledStart: month(t, "2026-09"), InstallmentAmountCents: 60000,
		PaymentMethod: "direct",
	})
	if err != nil {
		t.Fatal(err)
	}
	if repository.created.Name != "Transplante" || repository.created.ScheduledEnd.String() != "2027-04" {
		t.Fatalf("created=%+v", repository.created)
	}
	if view.Projection.RemainingInstallments != 8 || view.Projection.ReleaseFrom.String() != "2027-05" {
		t.Fatalf("view=%+v", view)
	}
}

func TestCreateRejectsUnsupportedCardUntilPaymentIntegration(t *testing.T) {
	service := NewService(&repositoryStub{}, profileStub{profile.Profile{CurrencyCode: "BRL", Timezone: "UTC"}})
	_, err := service.Create(context.Background(), "owner", CreateInput{
		Name: "Dívida", TotalInstallments: 2, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2026-01"), InstallmentAmountCents: 100,
		PaymentMethod: "credit_card",
	})
	if !errors.Is(err, ErrPaymentMethodUnsupported) {
		t.Fatalf("expected unsupported payment method, got %v", err)
	}
}

func TestListFiltersProjectionStatusAndSortsByEffectiveEnd(t *testing.T) {
	early := debtFixture("owner", "BRL", "early", "Early", month(t, "2026-01"), month(t, "2026-02"), 2, 1, 100)
	late := debtFixture("owner", "BRL", "late", "Late", month(t, "2026-01"), month(t, "2026-03"), 3, 1, 100)
	repository := &repositoryStub{debts: []debtdomain.Debt{late, early}}
	service := NewService(repository, profileStub{profile.Profile{Timezone: "UTC"}})
	status := debtdomain.ProjectionStatusCompleted
	views, err := service.List(context.Background(), "owner", Filters{
		AsOf: ptrMonth(t, "2026-04"), ProjectionStatus: &status,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 2 || views[0].Debt.ID != "early" || views[1].Debt.ID != "late" {
		t.Fatalf("views=%+v", views)
	}
}

func TestChangeRejectsMonthAfterSettlement(t *testing.T) {
	debt := debtFixture("owner", "BRL", "debt", "Debt", month(t, "2026-01"), month(t, "2026-03"), 3, 1, 100)
	debt.Settlement = &debtdomain.EarlySettlement{ID: "settlement", ReferenceMonth: month(t, "2026-02"), Amount: planningdomain.NewMoney(150)}
	service := NewService(&repositoryStub{debts: []debtdomain.Debt{debt}}, profileStub{profile.Profile{Timezone: "UTC"}})
	_, err := service.Change(context.Background(), "owner", "debt", ChangeInput{
		EffectiveFrom: month(t, "2026-03"), InstallmentAmountCents: 120,
	})
	if !errors.Is(err, ErrChangeAfterSettlement) {
		t.Fatalf("expected change after settlement error, got %v", err)
	}
}

func debtFixture(ownerID, currency, id, name string, start, end planningdomain.YearMonth, total, first int, amount int64) debtdomain.Debt {
	interval, _ := planningdomain.NewMonthInterval(start, end)
	debt, _ := debtdomain.NewDebt(debtdomain.NewDebtInput{
		ID: id, UserID: ownerID, CurrencyCode: currency, Name: name,
		TotalInstallments: total, FirstProjectedInstallment: first, ScheduledStart: start,
		Periods: []debtdomain.InstallmentPeriod{{ID: id + "-period", Interval: interval, Amount: planningdomain.NewMoney(amount)}},
		Status:  planningdomain.FinancialItemStatusActive,
	})
	return debt
}

func month(t *testing.T, value string) planningdomain.YearMonth {
	t.Helper()
	month, err := planningdomain.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return month
}

func ptrMonth(t *testing.T, value string) *planningdomain.YearMonth {
	month := month(t, value)
	return &month
}
