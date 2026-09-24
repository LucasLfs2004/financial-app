package financialitem

import (
	"context"
	"errors"
	"testing"

	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
)

type fakePlans struct{ plan planning.Plan }

func (plans fakePlans) Current(context.Context, string) (planning.Plan, error) {
	return plans.plan, nil
}

type fakeRepository struct {
	created  CreateInput
	currency string
	changed  ChangeInput
	findItem Item
}

func (repository *fakeRepository) Create(_ context.Context, _, currency string, input CreateInput) (Item, error) {
	repository.created = input
	repository.currency = currency
	return Item{}, nil
}
func (*fakeRepository) List(context.Context, string, Filters) ([]Item, error) {
	return []Item{}, nil
}
func (repository *fakeRepository) Find(context.Context, string, string) (Item, error) {
	if repository.findItem.Kind.Valid() {
		return repository.findItem, nil
	}
	return Item{Kind: domain.FinancialItemKindRecurringIncome, Periods: []Period{{StartMonth: mustMonth("2026-01"), EndMonth: monthPointer("2026-12"), Recurrence: domain.RecurrenceMonthly}}}, nil
}
func (*fakeRepository) Update(context.Context, string, string, UpdateInput) (Item, error) {
	return Item{}, nil
}
func (repository *fakeRepository) Change(_ context.Context, _, _ string, input ChangeInput) (Item, error) {
	repository.changed = input
	return Item{}, nil
}
func (*fakeRepository) Archive(context.Context, string, string, ArchiveInput) (Item, error) {
	return Item{}, nil
}

func TestCreateAcceptsDeferredRecurringIncome(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, fakePlans{testPlan(t)})
	_, err := service.Create(context.Background(), "owner", CreateInput{Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome, Period: PeriodInput{StartMonth: testMonth(t, "2026-01"), EndMonth: testMonthPtr(t, "2026-12"), AmountCents: 600000, Recurrence: domain.RecurrenceMonthly, CashMonthOffset: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if repository.created.Period.CashMonthOffset != 1 {
		t.Fatal("cash offset was not preserved")
	}
	if repository.currency != "BRL" {
		t.Fatalf("currency=%q", repository.currency)
	}
}
func TestCreateRejectsInvalidKindRecurrence(t *testing.T) {
	service := NewService(&fakeRepository{}, fakePlans{testPlan(t)})
	_, err := service.Create(context.Background(), "owner", CreateInput{Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome, Period: PeriodInput{StartMonth: testMonth(t, "2026-01"), EndMonth: testMonthPtr(t, "2026-01"), AmountCents: 100, Recurrence: domain.RecurrenceOnce}})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error=%v", err)
	}
}
func TestCreateAcceptsPeriodOutsidePlanHorizon(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, fakePlans{testPlan(t)})
	_, err := service.Create(context.Background(), "owner", CreateInput{Name: "Seguro", Kind: domain.FinancialItemKindFixedExpense, Period: PeriodInput{StartMonth: testMonth(t, "2026-09"), EndMonth: testMonthPtr(t, "2027-09"), AmountCents: 100, Recurrence: domain.RecurrenceMonthly}})
	if err != nil {
		t.Fatal(err)
	}
	if repository.created.Period.EndMonth == nil || repository.created.Period.EndMonth.String() != "2027-09" {
		t.Fatalf("end=%v", repository.created.Period.EndMonth)
	}
}

func TestChangeValidatesOriginalOnceRecurrence(t *testing.T) {
	repository := &fakeRepository{findItem: Item{Kind: domain.FinancialItemKindProjectedVariableExpense, Periods: []Period{{StartMonth: testMonth(t, "2026-04"), EndMonth: testMonthPtr(t, "2026-04"), Recurrence: domain.RecurrenceOnce}}}}
	service := NewService(repository, fakePlans{testPlan(t)})
	_, err := service.Change(context.Background(), "owner", "item", ChangeInput{EffectiveFrom: testMonth(t, "2026-04"), AmountCents: 100})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error=%v", err)
	}
}

func testPlan(t *testing.T) planning.Plan {
	return planning.Plan{ID: "plan", StartMonth: testMonth(t, "2026-01"), EndMonth: testMonth(t, "2026-12"), CurrencyCode: "BRL", Status: domain.PlanStatusDraft}
}
func testMonth(t *testing.T, value string) domain.YearMonth {
	t.Helper()
	result, err := domain.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func testMonthPtr(t *testing.T, value string) *domain.YearMonth {
	result := testMonth(t, value)
	return &result
}
func mustMonth(value string) domain.YearMonth {
	result, _ := domain.ParseYearMonth(value)
	return result
}
func monthPointer(value string) *domain.YearMonth {
	result := mustMonth(value)
	return &result
}
