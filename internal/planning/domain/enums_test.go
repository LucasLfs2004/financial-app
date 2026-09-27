package domain

import (
	"errors"
	"testing"
)

func TestEnumParsers(t *testing.T) {
	tests := []struct {
		name  string
		valid func() error
		bad   func() error
	}{
		{
			name:  "plan status",
			valid: func() error { _, err := ParsePlanStatus("draft"); return err },
			bad:   func() error { _, err := ParsePlanStatus("closed"); return err },
		},
		{
			name:  "financial item kind",
			valid: func() error { _, err := ParseFinancialItemKind("fixed_expense"); return err },
			bad:   func() error { _, err := ParseFinancialItemKind("purchase"); return err },
		},
		{
			name:  "financial item status",
			valid: func() error { _, err := ParseFinancialItemStatus("active"); return err },
			bad:   func() error { _, err := ParseFinancialItemStatus("deleted"); return err },
		},
		{
			name:  "recurrence",
			valid: func() error { _, err := ParseRecurrence("monthly"); return err },
			bad:   func() error { _, err := ParseRecurrence("daily"); return err },
		},
		{
			name:  "summary basis",
			valid: func() error { _, err := ParseSummaryBasis("cash"); return err },
			bad:   func() error { _, err := ParseSummaryBasis("accrual"); return err },
		},
		{
			name:  "summary result kind",
			valid: func() error { _, err := ParseSummaryResultKind("planned_free"); return err },
			bad:   func() error { _, err := ParseSummaryResultKind("balance"); return err },
		},
		{
			name:  "summary source effect",
			valid: func() error { _, err := ParseSummarySourceEffect("income"); return err },
			bad:   func() error { _, err := ParseSummarySourceEffect("debit"); return err },
		},
		{
			name:  "completeness",
			valid: func() error { _, err := ParseCompleteness("projected"); return err },
			bad:   func() error { _, err := ParseCompleteness("partial"); return err },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.valid(); err != nil {
				t.Fatalf("valid value returned error: %v", err)
			}
			if err := tt.bad(); !errors.Is(err, ErrInvalidEnumValue) {
				t.Fatalf("invalid value error = %v, want ErrInvalidEnumValue", err)
			}
		})
	}
}

func TestFinancialItemKindClassification(t *testing.T) {
	if !FinancialItemKindRecurringIncome.IsIncome() || FinancialItemKindRecurringIncome.IsExpense() {
		t.Fatal("recurring income classification is invalid")
	}
	if !FinancialItemKindFixedExpense.IsExpense() || FinancialItemKindFixedExpense.IsIncome() {
		t.Fatal("fixed expense classification is invalid")
	}
	if !FinancialItemKindDebtInstallment.IsExpense() || FinancialItemKindDebtInstallment.IsIncome() {
		t.Fatal("debt installment classification is invalid")
	}
}
