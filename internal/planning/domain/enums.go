package domain

import "fmt"

type PlanStatus string

const (
	PlanStatusDraft    PlanStatus = "draft"
	PlanStatusActive   PlanStatus = "active"
	PlanStatusArchived PlanStatus = "archived"
)

func ParsePlanStatus(value string) (PlanStatus, error) {
	status := PlanStatus(value)
	if !status.Valid() {
		return "", invalidEnum("plan status", value)
	}
	return status, nil
}

func (s PlanStatus) Valid() bool {
	return s == PlanStatusDraft || s == PlanStatusActive || s == PlanStatusArchived
}

type FinancialItemKind string

const (
	FinancialItemKindRecurringIncome          FinancialItemKind = "recurring_income"
	FinancialItemKindOneTimeIncome            FinancialItemKind = "one_time_income"
	FinancialItemKindFixedExpense             FinancialItemKind = "fixed_expense"
	FinancialItemKindProjectedVariableExpense FinancialItemKind = "projected_variable_expense"
	FinancialItemKindDebtInstallment          FinancialItemKind = "debt_installment"
)

func ParseFinancialItemKind(value string) (FinancialItemKind, error) {
	kind := FinancialItemKind(value)
	if !kind.Valid() {
		return "", invalidEnum("financial item kind", value)
	}
	return kind, nil
}

func (k FinancialItemKind) Valid() bool {
	switch k {
	case FinancialItemKindRecurringIncome,
		FinancialItemKindOneTimeIncome,
		FinancialItemKindFixedExpense,
		FinancialItemKindProjectedVariableExpense,
		FinancialItemKindDebtInstallment:
		return true
	default:
		return false
	}
}

func (k FinancialItemKind) IsIncome() bool {
	return k == FinancialItemKindRecurringIncome || k == FinancialItemKindOneTimeIncome
}

func (k FinancialItemKind) IsExpense() bool {
	return k == FinancialItemKindFixedExpense ||
		k == FinancialItemKindProjectedVariableExpense ||
		k == FinancialItemKindDebtInstallment
}

type FinancialItemStatus string

const (
	FinancialItemStatusActive   FinancialItemStatus = "active"
	FinancialItemStatusArchived FinancialItemStatus = "archived"
)

func ParseFinancialItemStatus(value string) (FinancialItemStatus, error) {
	status := FinancialItemStatus(value)
	if !status.Valid() {
		return "", invalidEnum("financial item status", value)
	}
	return status, nil
}

func (s FinancialItemStatus) Valid() bool {
	return s == FinancialItemStatusActive || s == FinancialItemStatusArchived
}

type Recurrence string

const (
	RecurrenceMonthly Recurrence = "monthly"
	RecurrenceOnce    Recurrence = "once"
)

func ParseRecurrence(value string) (Recurrence, error) {
	recurrence := Recurrence(value)
	if !recurrence.Valid() {
		return "", invalidEnum("recurrence", value)
	}
	return recurrence, nil
}

func (r Recurrence) Valid() bool {
	return r == RecurrenceMonthly || r == RecurrenceOnce
}

type SummaryBasis string

const (
	SummaryBasisCash      SummaryBasis = "cash"
	SummaryBasisReference SummaryBasis = "reference"
)

func ParseSummaryBasis(value string) (SummaryBasis, error) {
	basis := SummaryBasis(value)
	if !basis.Valid() {
		return "", invalidEnum("summary basis", value)
	}
	return basis, nil
}

func (b SummaryBasis) Valid() bool {
	return b == SummaryBasisCash || b == SummaryBasisReference
}

type SummaryResultKind string

const (
	SummaryResultKindPlannedFree            SummaryResultKind = "planned_free"
	SummaryResultKindPlannedReferenceResult SummaryResultKind = "planned_reference_result"
)

func ParseSummaryResultKind(value string) (SummaryResultKind, error) {
	kind := SummaryResultKind(value)
	if !kind.Valid() {
		return "", invalidEnum("summary result kind", value)
	}
	return kind, nil
}

func (k SummaryResultKind) Valid() bool {
	return k == SummaryResultKindPlannedFree || k == SummaryResultKindPlannedReferenceResult
}

type SummarySourceEffect string

const (
	SummarySourceEffectIncome     SummarySourceEffect = "income"
	SummarySourceEffectCommitment SummarySourceEffect = "commitment"
	SummarySourceEffectSavings    SummarySourceEffect = "savings"
)

func ParseSummarySourceEffect(value string) (SummarySourceEffect, error) {
	effect := SummarySourceEffect(value)
	if !effect.Valid() {
		return "", invalidEnum("summary source effect", value)
	}
	return effect, nil
}

func (e SummarySourceEffect) Valid() bool {
	return e == SummarySourceEffectIncome || e == SummarySourceEffectCommitment || e == SummarySourceEffectSavings
}

type Completeness string

const CompletenessProjected Completeness = "projected"

func ParseCompleteness(value string) (Completeness, error) {
	completeness := Completeness(value)
	if !completeness.Valid() {
		return "", invalidEnum("completeness", value)
	}
	return completeness, nil
}

func (c Completeness) Valid() bool {
	return c == CompletenessProjected
}

func invalidEnum(name, value string) error {
	return fmt.Errorf("%w: %s %q", ErrInvalidEnumValue, name, value)
}
