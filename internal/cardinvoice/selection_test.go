package cardinvoice

import (
	"errors"
	"testing"
	"time"

	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planning "github.com/lucas/financial-api/internal/planning/domain"
)

func TestSelectInvoiceComponentsResolvesTemporalPaymentAndLatestMove(t *testing.T) {
	planStart := testMonth(t, "2026-01")
	planEnd := testMonth(t, "2026-12")
	december := testMonth(t, "2026-12")
	january := testMonth(t, "2027-01")
	cards := []Card{
		testCard(t, "card-a", "Principal", "2026-01", "2027-12", 6, 1),
		testCard(t, "card-b", "Reserva", "2026-01", "2027-12", 10, 1),
	}
	item := testSelectionItem(t, "item", "Gasolina", "2026-11", "2026-12", planning.RecurrenceMonthly, 70000, "card-a")
	moves := []OccurrenceMove{
		{ID: "move-1", ItemID: "item", ReferenceMonth: december, ToCardID: "card-b", ToPaymentMonth: january, RecordedAt: time.Date(2026, 12, 1, 10, 0, 0, 0, time.UTC)},
		{ID: "move-2", ItemID: "item", ReferenceMonth: december, ToCardID: "card-a", ToPaymentMonth: january, RecordedAt: time.Date(2026, 12, 1, 11, 0, 0, 0, time.UTC)},
	}

	invoice, err := SelectInvoiceComponents(SelectionInput{
		PlanStart: planStart, PlanEnd: planEnd, CardID: "card-a", PaymentMonth: january,
		CurrencyCode: "BRL", Cards: cards, Items: []FinancialItem{item}, Moves: moves,
	})
	if err != nil {
		t.Fatal(err)
	}
	if invoice.Total.Cents() != 70000 || len(invoice.Components) != 1 {
		t.Fatalf("invoice total=%d components=%+v", invoice.Total.Cents(), invoice.Components)
	}
	if invoice.Components[0].ReferenceMonth.String() != "2026-12" || invoice.Components[0].Allocation != AllocationMovedByUser {
		t.Fatalf("moved component=%+v", invoice.Components[0])
	}
}

func TestSelectInvoiceComponentsUsesDirectFallbackAndKnownOrUnknownAdjustments(t *testing.T) {
	planStart := testMonth(t, "2026-01")
	planEnd := testMonth(t, "2026-12")
	paymentMonth := testMonth(t, "2026-06")
	reference := testMonth(t, "2026-05")
	item := testSelectionItem(t, "direct-item", "Direto", "2026-05", "2026-05", planning.RecurrenceMonthly, 5000, "")
	invoice, err := SelectInvoiceComponents(SelectionInput{
		PlanStart: planStart, PlanEnd: planEnd, CardID: "card", PaymentMonth: paymentMonth, CurrencyCode: "BRL",
		Cards: []Card{testCard(t, "card", "Principal", "2026-01", "2026-12", 6, 1)}, Items: []FinancialItem{item},
		Adjustments: []Adjustment{
			{ID: "known", Name: "Conhecido", CardID: "card", PaymentMonth: paymentMonth, ReferenceMonth: &reference, Amount: planning.NewMoney(100), Status: AdjustmentStatusActive},
			{ID: "unknown", Name: "Desconhecido", CardID: "card", PaymentMonth: paymentMonth, Amount: planning.NewMoney(200), Status: AdjustmentStatusActive},
			{ID: "archived", Name: "Arquivado", CardID: "card", PaymentMonth: paymentMonth, Amount: planning.NewMoney(999), Status: AdjustmentStatusArchived},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if invoice.Total.Cents() != 300 || len(invoice.Components) != 2 || !invoice.Components[0].ReferenceKnown || invoice.Components[1].ReferenceKnown {
		t.Fatalf("invoice=%+v", invoice)
	}
}

func TestSelectInvoiceComponentsRejectsOverlappingFinancialPeriods(t *testing.T) {
	item := testSelectionItem(t, "item", "Despesa", "2026-01", "2026-12", planning.RecurrenceMonthly, 100, "card")
	item.Periods = append(item.Periods, item.Periods[0])
	_, err := SelectInvoiceComponents(SelectionInput{
		PlanStart: testMonth(t, "2026-01"), PlanEnd: testMonth(t, "2026-12"), CardID: "card",
		PaymentMonth: testMonth(t, "2026-02"), CurrencyCode: "BRL",
		Cards: []Card{testCard(t, "card", "Principal", "2026-01", "2026-12", 6, 1)}, Items: []FinancialItem{item},
	})
	if !errors.Is(err, ErrDuplicateOccurrence) {
		t.Fatalf("expected duplicate occurrence, got %v", err)
	}
}

func TestSelectInvoiceComponentsUsesProjectedDebtOccurrencesAndMetadata(t *testing.T) {
	jan := testMonth(t, "2026-01")
	feb := testMonth(t, "2026-02")
	mar := testMonth(t, "2026-03")
	janInterval, _ := planning.NewMonthInterval(jan, jan)
	futureInterval, _ := planning.NewMonthInterval(feb, testMonth(t, "2026-04"))
	cardAPayment, _ := NewPaymentMethodPeriod("payment-a", janInterval, PaymentMethodCreditCard, "card-a")
	cardBPayment, _ := NewPaymentMethodPeriod("payment-b", futureInterval, PaymentMethodCreditCard, "card-b")
	debtItem := FinancialItem{
		ID: "debt", Name: "Parcelamento", Kind: planning.FinancialItemKindDebtInstallment,
		PaymentPeriods: []PaymentMethodPeriod{cardAPayment, cardBPayment},
	}
	debtOccurrences := []DebtOccurrence{
		{DebtID: "debt", SourceID: "period", Name: "Parcelamento", ReferenceMonth: jan, InstallmentNumber: 5, InstallmentsTotal: 12, Amount: planning.NewMoney(60000), Kind: debtdomain.OccurrenceKindScheduled},
		{DebtID: "debt", SourceID: "settlement", Name: "Parcelamento", ReferenceMonth: feb, InstallmentNumber: 6, InstallmentsTotal: 12, Amount: planning.NewMoney(150000), Kind: debtdomain.OccurrenceKindEarlySettlement},
	}
	cards := []Card{
		testCard(t, "card-a", "Principal", "2026-01", "2026-12", 6, 1),
		testCard(t, "card-b", "Reserva", "2026-01", "2026-12", 10, 1),
	}

	regular, err := SelectInvoiceComponents(SelectionInput{
		PlanStart: jan, PlanEnd: testMonth(t, "2026-12"), CardID: "card-a", PaymentMonth: feb,
		CurrencyCode: "BRL", Cards: cards, Items: []FinancialItem{debtItem}, DebtOccurrences: debtOccurrences,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(regular.Components) != 1 || regular.Components[0].DebtID == nil || *regular.Components[0].DebtID != "debt" ||
		regular.Components[0].InstallmentNumber == nil || *regular.Components[0].InstallmentNumber != 5 ||
		regular.Components[0].DebtOccurrenceKind == nil || *regular.Components[0].DebtOccurrenceKind != debtdomain.OccurrenceKindScheduled {
		t.Fatalf("regular=%+v", regular)
	}

	settlement, err := SelectInvoiceComponents(SelectionInput{
		PlanStart: jan, PlanEnd: testMonth(t, "2026-12"), CardID: "card-b", PaymentMonth: mar,
		CurrencyCode: "BRL", Cards: cards, Items: []FinancialItem{debtItem}, DebtOccurrences: debtOccurrences,
	})
	if err != nil || len(settlement.Components) != 1 || settlement.Total.Cents() != 150000 ||
		settlement.Components[0].DebtOccurrenceKind == nil || *settlement.Components[0].DebtOccurrenceKind != debtdomain.OccurrenceKindEarlySettlement {
		t.Fatalf("settlement=%+v error=%v", settlement, err)
	}

	afterSettlement, err := SelectInvoiceComponents(SelectionInput{
		PlanStart: jan, PlanEnd: testMonth(t, "2026-12"), CardID: "card-b", PaymentMonth: testMonth(t, "2026-04"),
		CurrencyCode: "BRL", Cards: cards, Items: []FinancialItem{debtItem}, DebtOccurrences: debtOccurrences,
	})
	if err != nil || len(afterSettlement.Components) != 0 {
		t.Fatalf("after settlement=%+v error=%v", afterSettlement, err)
	}
}

func TestSelectInvoiceComponentsRejectsDuplicateDebtOccurrence(t *testing.T) {
	month := testMonth(t, "2026-01")
	interval, _ := planning.NewMonthInterval(month, month)
	payment, _ := NewPaymentMethodPeriod("payment", interval, PaymentMethodCreditCard, "card")
	item := FinancialItem{ID: "debt", Name: "Dívida", Kind: planning.FinancialItemKindDebtInstallment, PaymentPeriods: []PaymentMethodPeriod{payment}}
	occurrence := DebtOccurrence{
		DebtID: "debt", SourceID: "period", Name: "Dívida", ReferenceMonth: month,
		InstallmentNumber: 1, InstallmentsTotal: 2, Amount: planning.NewMoney(100),
		Kind: debtdomain.OccurrenceKindScheduled,
	}
	_, err := SelectInvoiceComponents(SelectionInput{
		PlanStart: month, PlanEnd: month, CardID: "card", PaymentMonth: month,
		CurrencyCode: "BRL", Cards: []Card{testCard(t, "card", "Principal", "2026-01", "2026-01", 6, 0)},
		Items: []FinancialItem{item}, DebtOccurrences: []DebtOccurrence{occurrence, occurrence},
	})
	if !errors.Is(err, ErrDuplicateOccurrence) {
		t.Fatalf("expected duplicate debt occurrence, got %v", err)
	}
}

func testCard(t *testing.T, id, name, start, end string, dueDay, offset int) Card {
	t.Helper()
	configuration, err := NewCardConfiguration(dueDay, offset)
	if err != nil {
		t.Fatal(err)
	}
	interval, err := planning.NewMonthInterval(testMonth(t, start), testMonth(t, end))
	if err != nil {
		t.Fatal(err)
	}
	period, err := NewCardConfigurationPeriod(id+"-period", interval, configuration)
	if err != nil {
		t.Fatal(err)
	}
	return Card{ID: id, Name: name, Configurations: []CardConfigurationPeriod{period}}
}

func testSelectionItem(t *testing.T, id, name, start, end string, recurrence planning.Recurrence, amount int64, cardID string) FinancialItem {
	t.Helper()
	interval, err := planning.NewMonthInterval(testMonth(t, start), testMonth(t, end))
	if err != nil {
		t.Fatal(err)
	}
	item := FinancialItem{ID: id, Name: name, Kind: planning.FinancialItemKindProjectedVariableExpense, Periods: []FinancialItemPeriod{{ID: id + "-period", Interval: interval, Amount: planning.NewMoney(amount), Recurrence: recurrence}}}
	if cardID != "" {
		payment, err := NewPaymentMethodPeriod(id+"-payment", interval, PaymentMethodCreditCard, cardID)
		if err != nil {
			t.Fatal(err)
		}
		item.PaymentPeriods = []PaymentMethodPeriod{payment}
	}
	return item
}
