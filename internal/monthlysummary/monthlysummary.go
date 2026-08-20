package monthlysummary

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/savings"
)

var (
	ErrValidation     = errors.New("invalid monthly summary input")
	ErrOutsideHorizon = errors.New("month outside plan horizon")
	ErrInconsistent   = errors.New("inconsistent monthly summary data")
)

const PlannedSavingsKind = "planned_savings"

type PlanReader interface {
	Current(context.Context, string) (planning.Plan, error)
}

type ItemReader interface {
	List(context.Context, string, string, financialitem.Filters) ([]financialitem.Item, error)
}

type SavingsReader interface {
	Get(context.Context, string, string) (savings.Configuration, error)
}

type Service struct {
	plans   PlanReader
	items   ItemReader
	savings SavingsReader
}

func NewService(plans PlanReader, items ItemReader, savingReader SavingsReader) *Service {
	return &Service{plans: plans, items: items, savings: savingReader}
}

type Input struct {
	Month   domain.YearMonth
	Basis   domain.SummaryBasis
	Plan    planning.Plan
	Items   []financialitem.Item
	Savings savings.Configuration
}

type Breakdown struct {
	RecurringIncomeCents           int64
	OneTimeIncomeCents             int64
	FixedExpensesCents             int64
	ProjectedVariableExpensesCents int64
}

type Source struct {
	SourceID       string
	ItemID         *string
	Name           string
	Kind           string
	Effect         domain.SummarySourceEffect
	ReferenceMonth domain.YearMonth
	CashMonth      domain.YearMonth
	AmountCents    int64
}

type Summary struct {
	Month               domain.YearMonth
	Basis               domain.SummaryBasis
	ResultKind          domain.SummaryResultKind
	CurrencyCode        string
	PlanStatus          domain.PlanStatus
	IncomeCents         int64
	CommitmentsCents    int64
	PlannedSavingsCents int64
	ResultCents         int64
	IsNegative          bool
	Completeness        domain.Completeness
	Breakdown           Breakdown
	Sources             []Source
}

func (service *Service) Get(ctx context.Context, ownerID string, month domain.YearMonth, basis domain.SummaryBasis) (Summary, error) {
	if strings.TrimSpace(ownerID) == "" || !month.Valid() || !basis.Valid() {
		return Summary{}, ErrValidation
	}
	plan, err := service.plans.Current(ctx, ownerID)
	if err != nil {
		return Summary{}, err
	}
	if month.Before(plan.StartMonth) || month.After(plan.EndMonth) {
		return Summary{}, ErrOutsideHorizon
	}
	items, err := service.items.List(ctx, ownerID, plan.ID, financialitem.Filters{})
	if err != nil {
		return Summary{}, err
	}
	configuration, err := service.savings.Get(ctx, ownerID, plan.ID)
	if err != nil {
		return Summary{}, err
	}
	return CalculateMonthlySummary(Input{Month: month, Basis: basis, Plan: plan, Items: items, Savings: configuration})
}

func CalculateMonthlySummary(input Input) (Summary, error) {
	if !input.Month.Valid() || !input.Basis.Valid() || input.Month.Before(input.Plan.StartMonth) || input.Month.After(input.Plan.EndMonth) {
		return Summary{}, ErrValidation
	}
	sources, err := SelectOccurrences(input.Month, input.Basis, input.Items, input.Savings)
	if err != nil {
		return Summary{}, err
	}
	result := Summary{Month: input.Month, Basis: input.Basis, CurrencyCode: input.Plan.CurrencyCode, PlanStatus: input.Plan.Status, Completeness: domain.CompletenessProjected, Sources: sources}
	if input.Basis == domain.SummaryBasisCash {
		result.ResultKind = domain.SummaryResultKindPlannedFree
	} else {
		result.ResultKind = domain.SummaryResultKindPlannedReferenceResult
	}
	for _, source := range sources {
		switch source.Kind {
		case string(domain.FinancialItemKindRecurringIncome):
			result.Breakdown.RecurringIncomeCents, err = checkedAdd(result.Breakdown.RecurringIncomeCents, source.AmountCents)
		case string(domain.FinancialItemKindOneTimeIncome):
			result.Breakdown.OneTimeIncomeCents, err = checkedAdd(result.Breakdown.OneTimeIncomeCents, source.AmountCents)
		case string(domain.FinancialItemKindFixedExpense):
			result.Breakdown.FixedExpensesCents, err = checkedAdd(result.Breakdown.FixedExpensesCents, source.AmountCents)
		case string(domain.FinancialItemKindProjectedVariableExpense):
			result.Breakdown.ProjectedVariableExpensesCents, err = checkedAdd(result.Breakdown.ProjectedVariableExpensesCents, source.AmountCents)
		case PlannedSavingsKind:
			result.PlannedSavingsCents, err = checkedAdd(result.PlannedSavingsCents, source.AmountCents)
		default:
			err = ErrInconsistent
		}
		if err != nil {
			return Summary{}, fmt.Errorf("%w: total overflow or unknown source", ErrInconsistent)
		}
	}
	result.IncomeCents, err = checkedAdd(result.Breakdown.RecurringIncomeCents, result.Breakdown.OneTimeIncomeCents)
	if err == nil {
		result.CommitmentsCents, err = checkedAdd(result.Breakdown.FixedExpensesCents, result.Breakdown.ProjectedVariableExpensesCents)
	}
	if err != nil {
		return Summary{}, fmt.Errorf("%w: total overflow", ErrInconsistent)
	}
	result.ResultCents, err = checkedSubtract(result.IncomeCents, result.CommitmentsCents)
	if err == nil {
		result.ResultCents, err = checkedSubtract(result.ResultCents, result.PlannedSavingsCents)
	}
	if err != nil {
		return Summary{}, fmt.Errorf("%w: result overflow", ErrInconsistent)
	}
	result.IsNegative = result.ResultCents < 0
	if err := validateConsistency(result); err != nil {
		return Summary{}, err
	}
	return result, nil
}

func SelectOccurrences(month domain.YearMonth, basis domain.SummaryBasis, items []financialitem.Item, configuration savings.Configuration) ([]Source, error) {
	if !month.Valid() || !basis.Valid() {
		return nil, ErrValidation
	}
	sources := make([]Source, 0)
	seenPeriods := make(map[string]struct{})
	selectedItems := make(map[string]struct{})
	for _, item := range items {
		for _, period := range item.Periods {
			if item.ID == "" || period.ID == "" || period.AmountCents < 0 || !item.Kind.Valid() || !period.Recurrence.Valid() || period.CashMonthOffset < domain.MinimumCashMonthOffset || period.CashMonthOffset > domain.MaximumCashMonthOffset {
				return nil, ErrInconsistent
			}
			referenceMonth := month
			if basis == domain.SummaryBasisCash {
				var err error
				referenceMonth, err = month.AddMonths(-period.CashMonthOffset)
				if err != nil {
					continue
				}
			}
			if !periodApplies(period, referenceMonth) {
				continue
			}
			cashMonth, err := referenceMonth.AddMonths(period.CashMonthOffset)
			if err != nil {
				return nil, ErrInconsistent
			}
			periodKey := item.ID + ":" + period.ID
			if _, duplicate := seenPeriods[periodKey]; duplicate {
				continue
			}
			seenPeriods[periodKey] = struct{}{}
			if _, duplicate := selectedItems[item.ID]; duplicate {
				return nil, fmt.Errorf("%w: multiple periods apply to item %s", ErrInconsistent, item.ID)
			}
			selectedItems[item.ID] = struct{}{}
			itemID := item.ID
			effect := domain.SummarySourceEffectCommitment
			if item.Kind.IsIncome() {
				effect = domain.SummarySourceEffectIncome
			}
			sources = append(sources, Source{SourceID: period.ID, ItemID: &itemID, Name: item.Name, Kind: string(item.Kind), Effect: effect, ReferenceMonth: referenceMonth, CashMonth: cashMonth, AmountCents: period.AmountCents})
		}
	}
	savingSelected := false
	for _, period := range configuration.Periods {
		if period.ID == "" || period.AmountCents < 0 {
			return nil, ErrInconsistent
		}
		if month.Before(period.StartMonth) || (period.EndMonth != nil && month.After(*period.EndMonth)) {
			continue
		}
		periodKey := "saving:" + period.ID
		if _, duplicate := seenPeriods[periodKey]; duplicate {
			continue
		}
		seenPeriods[periodKey] = struct{}{}
		if savingSelected {
			return nil, fmt.Errorf("%w: multiple saving periods apply", ErrInconsistent)
		}
		savingSelected = true
		sources = append(sources, Source{SourceID: period.ID, Name: "Valor planejado para guardar", Kind: PlannedSavingsKind, Effect: domain.SummarySourceEffectSavings, ReferenceMonth: month, CashMonth: month, AmountCents: period.AmountCents})
	}
	sort.Slice(sources, func(i, j int) bool {
		left, right := sources[i], sources[j]
		if effectRank(left.Effect) != effectRank(right.Effect) {
			return effectRank(left.Effect) < effectRank(right.Effect)
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		return left.SourceID < right.SourceID
	})
	return sources, nil
}

func periodApplies(period financialitem.Period, month domain.YearMonth) bool {
	if month.Before(period.StartMonth) || (period.EndMonth != nil && month.After(*period.EndMonth)) {
		return false
	}
	return period.Recurrence == domain.RecurrenceMonthly || (period.Recurrence == domain.RecurrenceOnce && month == period.StartMonth)
}

func effectRank(effect domain.SummarySourceEffect) int {
	switch effect {
	case domain.SummarySourceEffectIncome:
		return 0
	case domain.SummarySourceEffectCommitment:
		return 1
	default:
		return 2
	}
}

func validateConsistency(summary Summary) error {
	var income, commitments, saving int64
	var err error
	for _, source := range summary.Sources {
		switch source.Effect {
		case domain.SummarySourceEffectIncome:
			income, err = checkedAdd(income, source.AmountCents)
		case domain.SummarySourceEffectCommitment:
			commitments, err = checkedAdd(commitments, source.AmountCents)
		case domain.SummarySourceEffectSavings:
			saving, err = checkedAdd(saving, source.AmountCents)
		default:
			err = ErrInconsistent
		}
		if err != nil {
			return fmt.Errorf("%w: source totals", ErrInconsistent)
		}
	}
	breakdownIncome, err := checkedAdd(summary.Breakdown.RecurringIncomeCents, summary.Breakdown.OneTimeIncomeCents)
	if err != nil {
		return ErrInconsistent
	}
	breakdownCommitments, err := checkedAdd(summary.Breakdown.FixedExpensesCents, summary.Breakdown.ProjectedVariableExpensesCents)
	if err != nil {
		return ErrInconsistent
	}
	expectedResult, err := checkedSubtract(income, commitments)
	if err == nil {
		expectedResult, err = checkedSubtract(expectedResult, saving)
	}
	if err != nil || income != summary.IncomeCents || income != breakdownIncome || commitments != summary.CommitmentsCents || commitments != breakdownCommitments || saving != summary.PlannedSavingsCents || expectedResult != summary.ResultCents {
		return fmt.Errorf("%w: totals do not match components", ErrInconsistent)
	}
	return nil
}

func checkedAdd(left, right int64) (int64, error) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, ErrInconsistent
	}
	return left + right, nil
}

func checkedSubtract(left, right int64) (int64, error) {
	if right > 0 && left < math.MinInt64+right {
		return 0, ErrInconsistent
	}
	return left - right, nil
}
