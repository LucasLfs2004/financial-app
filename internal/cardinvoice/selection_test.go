package cardinvoice

import (
	"errors"
	"testing"
	"time"

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
