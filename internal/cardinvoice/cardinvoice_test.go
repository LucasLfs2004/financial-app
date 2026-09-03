package cardinvoice

import (
	"errors"
	"math"
	"testing"
	"time"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

func TestCardConfigurationDerivesPaymentMonth(t *testing.T) {
	configuration, err := NewCardConfiguration(6, 1)
	if err != nil {
		t.Fatal(err)
	}

	paymentMonth, err := configuration.PaymentMonth(testMonth(t, "2026-12"))
	if err != nil {
		t.Fatal(err)
	}
	if paymentMonth.String() != "2027-01" {
		t.Fatalf("expected 2027-01, got %s", paymentMonth)
	}
}

func TestCardConfigurationAcceptsOffsetBoundaries(t *testing.T) {
	for _, offset := range []int{0, 12} {
		if _, err := NewCardConfiguration(1, offset); err != nil {
			t.Fatalf("expected offset %d to be valid: %v", offset, err)
		}
	}
	for _, offset := range []int{-1, 13} {
		if _, err := NewCardConfiguration(1, offset); !errors.Is(err, ErrInvalidConfiguration) {
			t.Fatalf("expected invalid configuration for offset %d, got %v", offset, err)
		}
	}
}

func TestResolveNominalDueDateDoesNotInventDate(t *testing.T) {
	dueDate, err := ResolveNominalDueDate(testMonth(t, "2027-02"), 31)
	if err != nil {
		t.Fatal(err)
	}
	if dueDate.Date != nil || dueDate.Resolution != DueDateResolutionInvalidForMonth || dueDate.Day != 31 {
		t.Fatalf("unexpected unresolved nominal date: %+v", dueDate)
	}
}

func TestResolveNominalDueDateHandlesLeapYear(t *testing.T) {
	dueDate, err := ResolveNominalDueDate(testMonth(t, "2028-02"), 29)
	if err != nil {
		t.Fatal(err)
	}
	if dueDate.Date == nil || dueDate.Date.Format("2006-01-02") != "2028-02-29" || dueDate.Resolution != DueDateResolutionExact {
		t.Fatalf("unexpected leap year nominal date: %+v", dueDate)
	}
}

func TestSelectCardConfigurationUsesApplicablePeriod(t *testing.T) {
	first, _ := NewCardConfiguration(6, 1)
	second, _ := NewCardConfiguration(10, 0)
	periods := []CardConfigurationPeriod{
		mustConfigurationPeriod(t, "first", "2026-01", "2026-07", first),
		mustConfigurationPeriod(t, "second", "2026-08", "2026-12", second),
	}

	selected, err := SelectCardConfiguration(testMonth(t, "2026-08"), periods)
	if err != nil {
		t.Fatal(err)
	}
	if selected.NominalDueDay() != 10 || selected.PaymentMonthOffset() != 0 {
		t.Fatalf("unexpected selected configuration: %+v", selected)
	}
}

func TestSelectCardConfigurationRejectsOverlap(t *testing.T) {
	configuration, _ := NewCardConfiguration(6, 1)
	periods := []CardConfigurationPeriod{
		mustConfigurationPeriod(t, "one", "2026-01", "2026-08", configuration),
		mustConfigurationPeriod(t, "two", "2026-08", "2026-12", configuration),
	}
	_, err := SelectCardConfiguration(testMonth(t, "2026-08"), periods)
	if !errors.Is(err, ErrPeriodOverlap) {
		t.Fatalf("expected overlap error, got %v", err)
	}
}

func TestSelectCardConfigurationRejectsInvalidPeriodData(t *testing.T) {
	interval, _ := planning.NewMonthInterval(testMonth(t, "2026-01"), testMonth(t, "2026-12"))
	_, err := SelectCardConfiguration(testMonth(t, "2026-08"), []CardConfigurationPeriod{{ID: "invalid", Interval: interval}})
	if !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("expected invalid configuration, got %v", err)
	}
}

func TestSelectPaymentMethodDefaultsToDirectAndSelectsCard(t *testing.T) {
	month := testMonth(t, "2026-08")
	defaultSelection, err := SelectPaymentMethod(month, nil)
	if err != nil {
		t.Fatal(err)
	}
	if defaultSelection.Method != PaymentMethodDirect || defaultSelection.Explicit {
		t.Fatalf("unexpected default selection: %+v", defaultSelection)
	}

	interval, _ := planning.NewMonthInterval(month, testMonth(t, "2026-12"))
	period, err := NewPaymentMethodPeriod("payment", interval, PaymentMethodCreditCard, "card")
	if err != nil {
		t.Fatal(err)
	}
	selection, err := SelectPaymentMethod(month, []PaymentMethodPeriod{period})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Method != PaymentMethodCreditCard || selection.CardID != "card" || !selection.Explicit {
		t.Fatalf("unexpected card selection: %+v", selection)
	}
}

func TestPaymentMethodValidatesCardCompatibility(t *testing.T) {
	interval, _ := planning.NewMonthInterval(testMonth(t, "2026-01"), testMonth(t, "2026-12"))
	if _, err := NewPaymentMethodPeriod("direct", interval, PaymentMethodDirect, "card"); !errors.Is(err, ErrInvalidPaymentMethod) {
		t.Fatalf("expected direct method with card to fail, got %v", err)
	}
	if _, err := NewPaymentMethodPeriod("card", interval, PaymentMethodCreditCard, ""); !errors.Is(err, ErrInvalidPaymentMethod) {
		t.Fatalf("expected credit card method without card to fail, got %v", err)
	}
}

func TestProjectInvoiceCombinesOccurrencesAndAdjustments(t *testing.T) {
	paymentMonth := testMonth(t, "2026-12")
	referenceMonth := testMonth(t, "2026-11")
	invoice, err := ProjectInvoice(ProjectionInput{
		CardID: "card", CardName: "Principal", PaymentMonth: paymentMonth, CurrencyCode: "BRL", NominalDueDay: 6,
		Occurrences: []Occurrence{
			{SourceID: "period-rent", ItemID: "rent", Name: "Aluguel", ReferenceMonth: referenceMonth, DefaultCardID: "card", DefaultPaymentMonth: paymentMonth, Amount: planning.NewMoney(100000)},
			{SourceID: "period-fuel", ItemID: "fuel", Name: "Gasolina", ReferenceMonth: referenceMonth, DefaultCardID: "card", DefaultPaymentMonth: paymentMonth, Amount: planning.NewMoney(70000)},
		},
		Adjustments: []Adjustment{{ID: "adjustment", Name: "Ajuste", CardID: "card", PaymentMonth: paymentMonth, Amount: planning.NewMoney(18000), Status: AdjustmentStatusActive}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if invoice.Total.Cents() != 188000 || len(invoice.Components) != 3 {
		t.Fatalf("unexpected invoice: total=%d components=%d", invoice.Total.Cents(), len(invoice.Components))
	}
	if invoice.Components[0].Name != "Aluguel" || invoice.Components[1].Name != "Gasolina" || invoice.Components[2].Name != "Ajuste" {
		t.Fatalf("components are not deterministically ordered: %+v", invoice.Components)
	}
	if invoice.Components[2].ReferenceKnown || invoice.Components[2].Allocation != AllocationSelectedInvoice {
		t.Fatalf("unexpected adjustment component: %+v", invoice.Components[2])
	}
}

func TestProjectInvoiceAppliesLatestMove(t *testing.T) {
	referenceMonth := testMonth(t, "2026-11")
	december := testMonth(t, "2026-12")
	january := testMonth(t, "2027-01")
	baseTime := time.Date(2026, time.November, 10, 12, 0, 0, 0, time.UTC)
	occurrence := Occurrence{SourceID: "period", ItemID: "fuel", Name: "Gasolina", ReferenceMonth: referenceMonth, DefaultCardID: "card", DefaultPaymentMonth: december, Amount: planning.NewMoney(70000)}
	moves := []OccurrenceMove{
		{ID: "move-1", ItemID: "fuel", ReferenceMonth: referenceMonth, ToCardID: "card", ToPaymentMonth: january, RecordedAt: baseTime},
		{ID: "move-2", ItemID: "fuel", ReferenceMonth: referenceMonth, ToCardID: "other-card", ToPaymentMonth: january, RecordedAt: baseTime.Add(time.Hour)},
	}

	oldInvoice, err := ProjectInvoice(ProjectionInput{CardID: "card", PaymentMonth: december, CurrencyCode: "BRL", NominalDueDay: 6, Occurrences: []Occurrence{occurrence}, Moves: moves})
	if err != nil {
		t.Fatal(err)
	}
	if len(oldInvoice.Components) != 0 {
		t.Fatalf("expected moved occurrence to leave old invoice: %+v", oldInvoice.Components)
	}

	newInvoice, err := ProjectInvoice(ProjectionInput{CardID: "other-card", PaymentMonth: january, CurrencyCode: "BRL", NominalDueDay: 6, Occurrences: []Occurrence{occurrence}, Moves: moves})
	if err != nil {
		t.Fatal(err)
	}
	if len(newInvoice.Components) != 1 || newInvoice.Components[0].Allocation != AllocationMovedByUser {
		t.Fatalf("expected latest move in destination invoice: %+v", newInvoice.Components)
	}
}

func TestProjectInvoiceRejectsDuplicateOccurrence(t *testing.T) {
	month := testMonth(t, "2026-11")
	occurrence := Occurrence{SourceID: "period", ItemID: "fuel", Name: "Gasolina", ReferenceMonth: month, DefaultCardID: "card", DefaultPaymentMonth: month, Amount: planning.NewMoney(1)}
	_, err := ProjectInvoice(ProjectionInput{CardID: "card", PaymentMonth: month, CurrencyCode: "BRL", NominalDueDay: 6, Occurrences: []Occurrence{occurrence, occurrence}})
	if !errors.Is(err, ErrDuplicateOccurrence) {
		t.Fatalf("expected duplicate occurrence, got %v", err)
	}
}

func TestProjectInvoiceRejectsNegativeAmount(t *testing.T) {
	month := testMonth(t, "2026-11")
	_, err := ProjectInvoice(ProjectionInput{
		CardID: "card", PaymentMonth: month, CurrencyCode: "BRL", NominalDueDay: 6,
		Adjustments: []Adjustment{{ID: "adjustment", Name: "Ajuste", CardID: "card", PaymentMonth: month, Amount: planning.NewMoney(-1), Status: AdjustmentStatusActive}},
	})
	if !errors.Is(err, ErrNegativeAmount) {
		t.Fatalf("expected negative amount error, got %v", err)
	}
}

func TestProjectInvoiceDetectsOverflow(t *testing.T) {
	month := testMonth(t, "2026-11")
	_, err := ProjectInvoice(ProjectionInput{
		CardID: "card", PaymentMonth: month, CurrencyCode: "BRL", NominalDueDay: 6,
		Adjustments: []Adjustment{
			{ID: "one", Name: "A", CardID: "card", PaymentMonth: month, Amount: planning.NewMoney(math.MaxInt64), Status: AdjustmentStatusActive},
			{ID: "two", Name: "B", CardID: "card", PaymentMonth: month, Amount: planning.NewMoney(1), Status: AdjustmentStatusActive},
		},
	})
	if !errors.Is(err, ErrInconsistentProjection) {
		t.Fatalf("expected overflow error, got %v", err)
	}
}

func TestProjectInvoiceRejectsInvalidCurrency(t *testing.T) {
	_, err := ProjectInvoice(ProjectionInput{CardID: "card", PaymentMonth: testMonth(t, "2026-11"), CurrencyCode: "brl", NominalDueDay: 6})
	if !errors.Is(err, ErrInvalidProjectionInput) {
		t.Fatalf("expected invalid projection input, got %v", err)
	}
}

func TestEnumsExposeOnlyContractValues(t *testing.T) {
	if !FinancialResourceStatusActive.Valid() || !PaymentMethodCreditCard.Valid() || !ComponentTypeInvoiceAdjustment.Valid() || !AllocationMovedByUser.Valid() || !DueDateResolutionExact.Valid() || !AuditEventOccurrenceMoved.Valid() {
		t.Fatal("expected contract enum values to be valid")
	}
	if FinancialResourceStatus("deleted").Valid() || PaymentMethod("pix").Valid() || ComponentType("purchase").Valid() || AllocationOrigin("automatic").Valid() {
		t.Fatal("unexpected enum value accepted")
	}
	if _, err := ParsePaymentMethod("pix"); !errors.Is(err, ErrInvalidEnumValue) {
		t.Fatalf("expected invalid enum error, got %v", err)
	}
}

func testMonth(t *testing.T, value string) planning.YearMonth {
	t.Helper()
	month, err := planning.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return month
}

func mustConfigurationPeriod(t *testing.T, id, start, end string, configuration CardConfiguration) CardConfigurationPeriod {
	t.Helper()
	interval, err := planning.NewMonthInterval(testMonth(t, start), testMonth(t, end))
	if err != nil {
		t.Fatal(err)
	}
	period, err := NewCardConfigurationPeriod(id, interval, configuration)
	if err != nil {
		t.Fatal(err)
	}
	return period
}
