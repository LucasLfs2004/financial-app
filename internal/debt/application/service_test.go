package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lucas/financial-api/internal/cardinvoice"
	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/profile"
)

type repositoryStub struct {
	created CreateRecord
	debts   []debtdomain.Debt
	payment PaymentContext
	settled SettlementInput
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
func (repository *repositoryStub) Settle(_ context.Context, _, _ string, input SettlementInput) (debtdomain.Debt, error) {
	repository.settled = input
	if repository.err != nil {
		return debtdomain.Debt{}, repository.err
	}
	debt := repository.debts[0]
	debt.Settlement = &debtdomain.EarlySettlement{
		ID: "settlement", ReferenceMonth: input.ReferenceMonth,
		Amount: planningdomain.NewMoney(input.AmountCents), Reason: input.Reason,
	}
	return debt, nil
}
func (repository *repositoryStub) PaymentContext(context.Context, string, string) (PaymentContext, error) {
	return repository.payment, repository.err
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

func TestCreateAcceptsInitialCreditCardPayment(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository, profileStub{profile.Profile{CurrencyCode: "BRL", Timezone: "UTC"}})
	cardID := "card"
	_, err := service.Create(context.Background(), "owner", CreateInput{
		Name: "Dívida", TotalInstallments: 2, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2026-01"), InstallmentAmountCents: 100,
		PaymentMethod: "credit_card", CreditCardID: &cardID,
	})
	if err != nil || repository.created.PaymentMethod != cardinvoice.PaymentMethodCreditCard ||
		repository.created.CreditCardID == nil || *repository.created.CreditCardID != cardID {
		t.Fatalf("created=%+v error=%v", repository.created, err)
	}
}

func TestCreateValidatesPaymentMethodCardPair(t *testing.T) {
	service := NewService(&repositoryStub{}, profileStub{profile.Profile{CurrencyCode: "BRL", Timezone: "UTC"}})
	_, err := service.Create(context.Background(), "owner", CreateInput{
		Name: "Dívida", TotalInstallments: 2, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2026-01"), InstallmentAmountCents: 100,
		PaymentMethod: "credit_card",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
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

func TestScheduleResolvesDirectCashOffset(t *testing.T) {
	debt := debtFixture("owner", "BRL", "debt", "Debt", month(t, "2026-01"), month(t, "2026-02"), 2, 1, 100)
	debt.Periods[0].CashMonthOffset = 2
	service := NewService(&repositoryStub{debts: []debtdomain.Debt{debt}}, profileStub{profile.Profile{Timezone: "UTC"}})
	service.now = func() time.Time { return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC) }
	schedule, err := service.Schedule(context.Background(), "owner", "debt", month(t, "2026-01"), month(t, "2026-02"))
	if err != nil {
		t.Fatal(err)
	}
	if len(schedule.Occurrences) != 2 || schedule.Occurrences[0].PaymentMethod != cardinvoice.PaymentMethodDirect ||
		schedule.Occurrences[0].CashMonth.String() != "2026-03" || schedule.Occurrences[0].InvoicePaymentMonth != nil {
		t.Fatalf("schedule=%+v", schedule)
	}
}

func TestScheduleResolvesCreditCardInvoiceMonth(t *testing.T) {
	debt := debtFixture("owner", "BRL", "debt", "Debt", month(t, "2026-01"), month(t, "2026-02"), 2, 1, 100)
	interval, _ := planningdomain.NewMonthInterval(month(t, "2026-01"), month(t, "2026-02"))
	paymentPeriod, _ := cardinvoice.NewPaymentMethodPeriod("payment", interval, cardinvoice.PaymentMethodCreditCard, "card")
	configuration, _ := cardinvoice.NewCardConfiguration(10, 1)
	configurationPeriod, _ := cardinvoice.NewCardConfigurationPeriod("configuration", interval, configuration)
	service := NewService(&repositoryStub{
		debts: []debtdomain.Debt{debt},
		payment: PaymentContext{
			Periods:            []cardinvoice.PaymentMethodPeriod{paymentPeriod},
			CardConfigurations: map[string][]cardinvoice.CardConfigurationPeriod{"card": {configurationPeriod}},
		},
	}, profileStub{profile.Profile{Timezone: "UTC"}})
	service.now = func() time.Time { return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC) }
	schedule, err := service.Schedule(context.Background(), "owner", "debt", month(t, "2026-01"), month(t, "2026-02"))
	if err != nil {
		t.Fatal(err)
	}
	first := schedule.Occurrences[0]
	if first.PaymentMethod != cardinvoice.PaymentMethodCreditCard || first.CreditCardID == nil || *first.CreditCardID != "card" ||
		first.InvoicePaymentMonth == nil || first.InvoicePaymentMonth.String() != "2026-02" || first.CashMonth.String() != "2026-02" {
		t.Fatalf("occurrence=%+v", first)
	}
}

func TestScheduleOmitsOccurrencesAfterSettlement(t *testing.T) {
	debt := debtFixture("owner", "BRL", "debt", "Debt", month(t, "2026-01"), month(t, "2026-03"), 3, 1, 100)
	debt.Settlement = &debtdomain.EarlySettlement{ID: "settlement", ReferenceMonth: month(t, "2026-02"), Amount: planningdomain.NewMoney(150)}
	service := NewService(&repositoryStub{debts: []debtdomain.Debt{debt}}, profileStub{profile.Profile{Timezone: "UTC"}})
	service.now = func() time.Time { return time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC) }
	schedule, err := service.Schedule(context.Background(), "owner", "debt", month(t, "2026-01"), month(t, "2026-03"))
	if err != nil {
		t.Fatal(err)
	}
	if len(schedule.Occurrences) != 2 || schedule.Occurrences[1].Occurrence.Kind != debtdomain.OccurrenceKindEarlySettlement ||
		schedule.Occurrences[1].Occurrence.Amount.Cents() != 150 {
		t.Fatalf("schedule=%+v", schedule)
	}
}

func TestSettleReprojectsEndReleaseAndSchedule(t *testing.T) {
	debt := debtFixture("owner", "BRL", "debt", "Debt", month(t, "2026-01"), month(t, "2026-04"), 4, 1, 100)
	repository := &repositoryStub{debts: []debtdomain.Debt{debt}}
	service := NewService(repository, profileStub{profile.Profile{Timezone: "UTC"}})
	service.now = func() time.Time { return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC) }
	view, err := service.Settle(context.Background(), "owner", "debt", SettlementInput{
		ReferenceMonth: month(t, "2026-02"), AmountCents: 250,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repository.settled.AmountCents != 250 || view.Projection.EffectiveEnd.String() != "2026-02" ||
		view.Projection.ReleaseFrom.String() != "2026-03" || view.Projection.ReleasedMonthly.Cents() != 100 ||
		len(view.Projection.Occurrences) != 2 || view.Projection.Occurrences[1].Kind != debtdomain.OccurrenceKindEarlySettlement {
		t.Fatalf("view=%+v input=%+v", view, repository.settled)
	}
}

func TestSettleRejectsFinalMonthAndExistingSettlement(t *testing.T) {
	debt := debtFixture("owner", "BRL", "debt", "Debt", month(t, "2026-01"), month(t, "2026-02"), 2, 1, 100)
	service := NewService(&repositoryStub{debts: []debtdomain.Debt{debt}}, profileStub{profile.Profile{Timezone: "UTC"}})
	service.now = func() time.Time { return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC) }
	_, err := service.Settle(context.Background(), "owner", "debt", SettlementInput{
		ReferenceMonth: month(t, "2026-02"), AmountCents: 100,
	})
	if !errors.Is(err, ErrSettlementOutsideRange) {
		t.Fatalf("expected outside range, got %v", err)
	}
	debt.Settlement = &debtdomain.EarlySettlement{ID: "settlement", ReferenceMonth: month(t, "2026-01"), Amount: planningdomain.NewMoney(150)}
	service = NewService(&repositoryStub{debts: []debtdomain.Debt{debt}}, profileStub{profile.Profile{Timezone: "UTC"}})
	_, err = service.Settle(context.Background(), "owner", "debt", SettlementInput{
		ReferenceMonth: month(t, "2026-01"), AmountCents: 150,
	})
	if !errors.Is(err, ErrSettlementExists) {
		t.Fatalf("expected existing settlement, got %v", err)
	}
}

func TestReleasesReturnsDebtAndMonthlyAggregates(t *testing.T) {
	first := debtFixture("owner", "BRL", "a", "A", month(t, "2026-01"), month(t, "2026-02"), 2, 1, 100)
	second := debtFixture("owner", "BRL", "b", "B", month(t, "2026-02"), month(t, "2026-02"), 1, 1, 250)
	service := NewService(&repositoryStub{debts: []debtdomain.Debt{second, first}}, profileStub{})
	result, err := service.Releases(context.Background(), "owner", month(t, "2026-03"), month(t, "2026-03"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Releases) != 2 || result.Releases[0].DebtID != "a" || result.Releases[1].DebtID != "b" ||
		len(result.MonthlyTotals) != 1 || result.MonthlyTotals[0].Amount.Cents() != 350 {
		t.Fatalf("result=%+v", result)
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
