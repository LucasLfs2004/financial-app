package domain

import (
	"errors"
	"testing"
)

func TestMonthIntervalIsInclusive(t *testing.T) {
	start := mustYearMonth(t, "2026-01")
	end := mustYearMonth(t, "2026-03")
	interval, err := NewMonthInterval(start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, value := range []string{"2026-01", "2026-02", "2026-03"} {
		if !interval.Contains(mustYearMonth(t, value)) {
			t.Fatalf("interval should contain %s", value)
		}
	}
	if interval.Contains(mustYearMonth(t, "2025-12")) || interval.Contains(mustYearMonth(t, "2026-04")) {
		t.Fatal("interval contains a month outside its bounds")
	}
}

func TestMonthIntervalsOverlap(t *testing.T) {
	january := mustYearMonth(t, "2026-01")
	march := mustYearMonth(t, "2026-03")
	april := mustYearMonth(t, "2026-04")

	first, _ := NewMonthInterval(january, march)
	touching, _ := NewMonthInterval(march, april)
	separate, _ := NewMonthInterval(april, april)
	open, err := NewOpenMonthInterval(march)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !first.Overlaps(touching) {
		t.Fatal("inclusive intervals sharing a month must overlap")
	}
	if first.Overlaps(separate) {
		t.Fatal("separate intervals cannot overlap")
	}
	if !open.Overlaps(separate) || !separate.Overlaps(open) {
		t.Fatal("open interval should overlap later intervals symmetrically")
	}
}

func TestMonthIntervalRejectsInvalidBounds(t *testing.T) {
	start := mustYearMonth(t, "2026-03")
	end := mustYearMonth(t, "2026-02")
	if _, err := NewMonthInterval(start, end); !errors.Is(err, ErrInvalidMonthInterval) {
		t.Fatalf("got error %v, want ErrInvalidMonthInterval", err)
	}
	if _, err := NewOpenMonthInterval(YearMonth{}); !errors.Is(err, ErrInvalidMonthInterval) {
		t.Fatalf("got error %v, want ErrInvalidMonthInterval", err)
	}
}

func TestFinancialPeriodValidation(t *testing.T) {
	january := mustYearMonth(t, "2026-01")
	march := mustYearMonth(t, "2026-03")
	closed, _ := NewMonthInterval(january, march)
	oneMonth, _ := NewMonthInterval(january, january)
	open, _ := NewOpenMonthInterval(january)

	tests := []struct {
		name       string
		interval   MonthInterval
		amount     Money
		recurrence Recurrence
		offset     int
		wantError  error
	}{
		{"monthly open period", open, NewMoney(100), RecurrenceMonthly, 1, nil},
		{"one-time period", oneMonth, NewMoney(0), RecurrenceOnce, 12, nil},
		{"invalid interval", MonthInterval{}, NewMoney(1), RecurrenceMonthly, 0, ErrInvalidMonthInterval},
		{"negative amount", closed, NewMoney(-1), RecurrenceMonthly, 0, ErrNegativeAmount},
		{"invalid recurrence", closed, NewMoney(1), Recurrence("weekly"), 0, ErrInvalidRecurrence},
		{"one-time open period", open, NewMoney(1), RecurrenceOnce, 0, ErrInvalidRecurrence},
		{"one-time multi-month period", closed, NewMoney(1), RecurrenceOnce, 0, ErrInvalidRecurrence},
		{"negative offset", closed, NewMoney(1), RecurrenceMonthly, -1, ErrInvalidCashMonthOffset},
		{"offset above maximum", closed, NewMoney(1), RecurrenceMonthly, 13, ErrInvalidCashMonthOffset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			period, err := NewFinancialPeriod(tt.interval, tt.amount, tt.recurrence, tt.offset)
			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Fatalf("got error %v, want %v", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			cashMonth, err := period.CashMonth(january)
			if err != nil {
				t.Fatalf("calculate cash month: %v", err)
			}
			wantCashMonth, err := january.AddMonths(tt.offset)
			if err != nil {
				t.Fatalf("prepare expected cash month: %v", err)
			}
			if cashMonth != wantCashMonth {
				t.Fatalf("cash month = %s, want %s", cashMonth, wantCashMonth)
			}
		})
	}
}

func TestFinancialItemRecurrenceCompatibility(t *testing.T) {
	tests := []struct {
		kind       FinancialItemKind
		recurrence Recurrence
		valid      bool
	}{
		{FinancialItemKindRecurringIncome, RecurrenceMonthly, true},
		{FinancialItemKindRecurringIncome, RecurrenceOnce, false},
		{FinancialItemKindOneTimeIncome, RecurrenceOnce, true},
		{FinancialItemKindOneTimeIncome, RecurrenceMonthly, false},
		{FinancialItemKindFixedExpense, RecurrenceMonthly, true},
		{FinancialItemKindFixedExpense, RecurrenceOnce, false},
		{FinancialItemKindProjectedVariableExpense, RecurrenceMonthly, true},
		{FinancialItemKindProjectedVariableExpense, RecurrenceOnce, true},
	}

	for _, tt := range tests {
		err := ValidateFinancialItemRecurrence(tt.kind, tt.recurrence)
		if tt.valid && err != nil {
			t.Fatalf("%s with %s should be valid: %v", tt.kind, tt.recurrence, err)
		}
		if !tt.valid && !errors.Is(err, ErrInvalidRecurrence) {
			t.Fatalf("%s with %s error = %v, want ErrInvalidRecurrence", tt.kind, tt.recurrence, err)
		}
	}
}
