package monthlysummary

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lucas/financial-api/internal/cardinvoice"
	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/savings"
)

func TestSelectOccurrencesUsesReferenceAndCashBases(t *testing.T) {
	salary := financialitem.Item{ID: "salary", Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome, Periods: []financialitem.Period{{ID: "salary-period", StartMonth: summaryMonth(t, "2026-01"), EndMonth: summaryMonthPtr(t, "2026-12"), AmountCents: 600000, Recurrence: domain.RecurrenceMonthly, CashMonthOffset: 1}}}

	reference, err := SelectOccurrences(summaryMonth(t, "2026-07"), domain.SummaryBasisReference, []financialitem.Item{salary}, savings.Configuration{})
	if err != nil || len(reference) != 1 || reference[0].ReferenceMonth == nil || reference[0].ReferenceMonth.String() != "2026-07" || reference[0].CashMonth.String() != "2026-08" {
		t.Fatalf("reference=%+v error=%v", reference, err)
	}
	cash, err := SelectOccurrences(summaryMonth(t, "2026-07"), domain.SummaryBasisCash, []financialitem.Item{salary}, savings.Configuration{})
	if err != nil || len(cash) != 1 || cash[0].ReferenceMonth == nil || cash[0].ReferenceMonth.String() != "2026-06" || cash[0].CashMonth.String() != "2026-07" {
		t.Fatalf("cash=%+v error=%v", cash, err)
	}
}

func TestSelectOccurrencesExpandsOnceOnlyInItsReferenceMonth(t *testing.T) {
	bonus := financialitem.Item{ID: "bonus", Name: "Bônus", Kind: domain.FinancialItemKindOneTimeIncome, Periods: []financialitem.Period{{ID: "bonus-period", StartMonth: summaryMonth(t, "2026-04"), EndMonth: summaryMonthPtr(t, "2026-04"), AmountCents: 100000, Recurrence: domain.RecurrenceOnce, CashMonthOffset: 2}}}

	selected, err := SelectOccurrences(summaryMonth(t, "2026-06"), domain.SummaryBasisCash, []financialitem.Item{bonus, bonus}, savings.Configuration{})
	if err != nil || len(selected) != 1 || selected[0].ReferenceMonth == nil || selected[0].ReferenceMonth.String() != "2026-04" {
		t.Fatalf("selected=%+v error=%v", selected, err)
	}
	selected, err = SelectOccurrences(summaryMonth(t, "2026-07"), domain.SummaryBasisCash, []financialitem.Item{bonus}, savings.Configuration{})
	if err != nil || len(selected) != 0 {
		t.Fatalf("selected=%+v error=%v", selected, err)
	}
}

func TestSelectOccurrencesRejectsTwoApplicablePeriodsForSameItem(t *testing.T) {
	month := summaryMonth(t, "2026-07")
	item := financialitem.Item{ID: "salary", Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome, Periods: []financialitem.Period{
		{ID: "period-a", StartMonth: month, EndMonth: &month, AmountCents: 600000, Recurrence: domain.RecurrenceMonthly},
		{ID: "period-b", StartMonth: month, EndMonth: &month, AmountCents: 650000, Recurrence: domain.RecurrenceMonthly},
	}}

	_, err := SelectOccurrences(month, domain.SummaryBasisReference, []financialitem.Item{item}, savings.Configuration{})
	if !errors.Is(err, ErrInconsistent) {
		t.Fatalf("error=%v", err)
	}
}

func TestCalculateMonthlySummaryReturnsNegativeConsistentBreakdown(t *testing.T) {
	month := summaryMonth(t, "2026-07")
	input := Input{
		Month: month, Basis: domain.SummaryBasisReference, Plan: summaryPlan(t),
		Items: []financialitem.Item{
			{ID: "salary", Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome, Periods: []financialitem.Period{{ID: "salary-period", StartMonth: month, EndMonth: &month, AmountCents: 600000, Recurrence: domain.RecurrenceMonthly}}},
			{ID: "rent", Name: "Aluguel", Kind: domain.FinancialItemKindFixedExpense, Periods: []financialitem.Period{{ID: "rent-period", StartMonth: month, EndMonth: &month, AmountCents: 700000, Recurrence: domain.RecurrenceMonthly}}},
			{ID: "fuel", Name: "Gasolina", Kind: domain.FinancialItemKindProjectedVariableExpense, Periods: []financialitem.Period{{ID: "fuel-period", StartMonth: month, EndMonth: &month, AmountCents: 70000, Recurrence: domain.RecurrenceMonthly}}},
		},
		Savings: savings.Configuration{Configured: true, Periods: []savings.Period{{ID: "saving-period", StartMonth: month, EndMonth: &month, AmountCents: 120000}}},
	}

	result, err := CalculateMonthlySummary(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.IncomeCents != 600000 || result.CommitmentsCents != 770000 || result.PlannedSavingsCents != 120000 || result.ResultCents != -290000 || !result.IsNegative {
		t.Fatalf("unexpected totals: %+v", result)
	}
	if result.Breakdown.FixedExpensesCents+result.Breakdown.ProjectedVariableExpensesCents != result.CommitmentsCents || len(result.Sources) != 4 {
		t.Fatalf("inconsistent breakdown or sources: %+v", result)
	}
}

func TestCalculateMonthlySummaryKeepsExplicitZeroAsSource(t *testing.T) {
	month := summaryMonth(t, "2026-07")
	result, err := CalculateMonthlySummary(Input{Month: month, Basis: domain.SummaryBasisCash, Plan: summaryPlan(t), Savings: savings.Configuration{Configured: true, Periods: []savings.Period{{ID: "saving-period", StartMonth: month, EndMonth: &month, AmountCents: 0}}}})
	if err != nil || result.PlannedSavingsCents != 0 || len(result.Sources) != 1 || result.Sources[0].Kind != PlannedSavingsKind {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestCalculateMonthlySummaryIntegratesCardComponentsByBasisWithoutDoubleCounting(t *testing.T) {
	june := summaryMonth(t, "2026-06")
	july := summaryMonth(t, "2026-07")
	item := financialitem.Item{
		ID: "fuel", Name: "Gasolina", Kind: domain.FinancialItemKindProjectedVariableExpense,
		Periods: []financialitem.Period{{ID: "fuel-period", StartMonth: june, EndMonth: &june, AmountCents: 70000, Recurrence: domain.RecurrenceMonthly}},
	}
	invoiceData := summaryInvoiceData(t, june, july)
	invoiceData.Adjustments = []cardinvoice.Adjustment{
		{ID: "adjustment", Name: "Tarifa", CardID: "card", PaymentMonth: july, ReferenceMonth: &june, Amount: domain.NewMoney(500), Status: cardinvoice.AdjustmentStatusActive},
		{ID: "unknown", Name: "Ajuste sem competência", CardID: "card", PaymentMonth: july, Amount: domain.NewMoney(200), Status: cardinvoice.AdjustmentStatusActive},
	}

	reference, err := CalculateMonthlySummary(Input{Month: june, Basis: domain.SummaryBasisReference, Plan: summaryPlan(t), Items: []financialitem.Item{item}, InvoiceData: invoiceData})
	if err != nil {
		t.Fatal(err)
	}
	if reference.CommitmentsCents != 70500 || reference.Breakdown.ProjectedVariableExpensesCents != 70000 || reference.Breakdown.CardInvoiceAdjustmentsCents != 500 || len(reference.Sources) != 2 {
		t.Fatalf("reference summary=%+v", reference)
	}

	cash, err := CalculateMonthlySummary(Input{Month: july, Basis: domain.SummaryBasisCash, Plan: summaryPlan(t), Items: []financialitem.Item{item}, InvoiceData: invoiceData})
	if err != nil {
		t.Fatal(err)
	}
	if cash.CommitmentsCents != 70700 || cash.Breakdown.ProjectedVariableExpensesCents != 70000 || cash.Breakdown.CardInvoiceAdjustmentsCents != 700 || len(cash.Sources) != 3 {
		t.Fatalf("cash summary=%+v", cash)
	}
	for _, source := range cash.Sources {
		if source.PaymentMethod == nil || *source.PaymentMethod != cardinvoice.PaymentMethodCreditCard || source.InvoicePaymentMonth == nil || *source.InvoicePaymentMonth != july {
			t.Fatalf("missing card metadata: %+v", source)
		}
	}
}

func TestCalculateMonthlySummaryUsesMovedInvoiceOnlyInDestinationCashMonth(t *testing.T) {
	june := summaryMonth(t, "2026-06")
	july := summaryMonth(t, "2026-07")
	august := summaryMonth(t, "2026-08")
	item := financialitem.Item{ID: "fuel", Name: "Gasolina", Kind: domain.FinancialItemKindProjectedVariableExpense, Periods: []financialitem.Period{{ID: "fuel-period", StartMonth: june, EndMonth: &june, AmountCents: 70000, Recurrence: domain.RecurrenceMonthly}}}
	invoiceData := summaryInvoiceData(t, june, july)
	invoiceData.Moves = []cardinvoice.OccurrenceMove{{ID: "move", ItemID: "fuel", ReferenceMonth: june, ToCardID: "card", ToPaymentMonth: august, RecordedAt: time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)}}

	julySummary, err := CalculateMonthlySummary(Input{Month: july, Basis: domain.SummaryBasisCash, Plan: summaryPlan(t), Items: []financialitem.Item{item}, InvoiceData: invoiceData})
	if err != nil || julySummary.CommitmentsCents != 0 {
		t.Fatalf("july=%+v error=%v", julySummary, err)
	}
	augustSummary, err := CalculateMonthlySummary(Input{Month: august, Basis: domain.SummaryBasisCash, Plan: summaryPlan(t), Items: []financialitem.Item{item}, InvoiceData: invoiceData})
	if err != nil || augustSummary.CommitmentsCents != 70000 || len(augustSummary.Sources) != 1 || augustSummary.Sources[0].InvoiceAllocation == nil || *augustSummary.Sources[0].InvoiceAllocation != cardinvoice.AllocationMovedByUser {
		t.Fatalf("august=%+v error=%v", augustSummary, err)
	}
}

func summaryInvoiceData(t *testing.T, referenceMonth, paymentMonth domain.YearMonth) cardinvoice.ProjectionData {
	t.Helper()
	planInterval, err := domain.NewMonthInterval(summaryMonth(t, "2026-01"), summaryMonth(t, "2026-12"))
	if err != nil {
		t.Fatal(err)
	}
	itemInterval, err := domain.NewMonthInterval(referenceMonth, referenceMonth)
	if err != nil {
		t.Fatal(err)
	}
	configuration, err := cardinvoice.NewCardConfiguration(10, 1)
	if err != nil {
		t.Fatal(err)
	}
	configurationPeriod, err := cardinvoice.NewCardConfigurationPeriod("card-period", planInterval, configuration)
	if err != nil {
		t.Fatal(err)
	}
	paymentPeriod, err := cardinvoice.NewPaymentMethodPeriod("payment-period", itemInterval, cardinvoice.PaymentMethodCreditCard, "card")
	if err != nil {
		t.Fatal(err)
	}
	if calculated, calculationErr := configuration.PaymentMonth(referenceMonth); calculationErr != nil || calculated != paymentMonth {
		t.Fatalf("invalid test payment month=%s error=%v", calculated.String(), calculationErr)
	}
	return cardinvoice.ProjectionData{
		Cards: []cardinvoice.Card{{ID: "card", Name: "Principal", Configurations: []cardinvoice.CardConfigurationPeriod{configurationPeriod}}},
		Items: []cardinvoice.FinancialItem{{
			ID: "fuel", Name: "Gasolina", Kind: domain.FinancialItemKindProjectedVariableExpense,
			Periods:        []cardinvoice.FinancialItemPeriod{{ID: "fuel-period", Interval: itemInterval, Amount: domain.NewMoney(70000), Recurrence: domain.RecurrenceMonthly}},
			PaymentPeriods: []cardinvoice.PaymentMethodPeriod{paymentPeriod},
		}},
	}
}

type fakePlans struct{ plan planning.Plan }

func (plans fakePlans) Current(context.Context, string) (planning.Plan, error) {
	return plans.plan, nil
}

type fakeItems struct{ items []financialitem.Item }

func (items fakeItems) List(context.Context, string, string, financialitem.Filters) ([]financialitem.Item, error) {
	return items.items, nil
}

type fakeSavings struct{ configuration savings.Configuration }

func (saving fakeSavings) Get(context.Context, string, string) (savings.Configuration, error) {
	return saving.configuration, nil
}

type fakeInvoiceData struct{ data cardinvoice.ProjectionData }

func (invoices fakeInvoiceData) LoadProjectionData(context.Context, string, string) (cardinvoice.ProjectionData, error) {
	return invoices.data, nil
}

func TestServiceRejectsMonthOutsideHorizonBeforeLoadingPremises(t *testing.T) {
	service := NewService(fakePlans{summaryPlan(t)}, fakeItems{}, fakeSavings{}, fakeInvoiceData{})
	_, err := service.Get(context.Background(), "owner", summaryMonth(t, "2027-01"), domain.SummaryBasisCash)
	if !errors.Is(err, ErrOutsideHorizon) {
		t.Fatalf("error=%v", err)
	}
}

func summaryPlan(t *testing.T) planning.Plan {
	return planning.Plan{ID: "plan", StartMonth: summaryMonth(t, "2026-01"), EndMonth: summaryMonth(t, "2026-12"), CurrencyCode: "BRL", Status: domain.PlanStatusDraft}
}

func summaryMonth(t *testing.T, value string) domain.YearMonth {
	t.Helper()
	month, err := domain.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return month
}

func summaryMonthPtr(t *testing.T, value string) *domain.YearMonth {
	month := summaryMonth(t, value)
	return &month
}
