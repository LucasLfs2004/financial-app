package domain

import (
	"fmt"
	"sort"
	"strings"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

const MaximumScheduleMonths = 120

type OccurrenceKind string

const (
	OccurrenceKindScheduled       OccurrenceKind = "scheduled"
	OccurrenceKindEarlySettlement OccurrenceKind = "early_settlement"
)

func (k OccurrenceKind) Valid() bool {
	return k == OccurrenceKindScheduled || k == OccurrenceKindEarlySettlement
}

type ProjectionStatus string

const (
	ProjectionStatusPlanned      ProjectionStatus = "planned"
	ProjectionStatusActive       ProjectionStatus = "active"
	ProjectionStatusCompleted    ProjectionStatus = "completed"
	ProjectionStatusSettledEarly ProjectionStatus = "settled_early"
	ProjectionStatusArchived     ProjectionStatus = "archived"
)

func (s ProjectionStatus) Valid() bool {
	switch s {
	case ProjectionStatusPlanned,
		ProjectionStatusActive,
		ProjectionStatusCompleted,
		ProjectionStatusSettledEarly,
		ProjectionStatusArchived:
		return true
	default:
		return false
	}
}

type ReleaseReason string

const (
	ReleaseReasonScheduledCompletion ReleaseReason = "scheduled_completion"
	ReleaseReasonEarlySettlement     ReleaseReason = "early_settlement"
)

type InstallmentPeriod struct {
	ID       string
	Interval planning.MonthInterval
	Amount   planning.Money
}

type EarlySettlement struct {
	ID             string
	ReferenceMonth planning.YearMonth
	Amount         planning.Money
}

type Debt struct {
	ID                        string
	Name                      string
	OriginalTotal             *planning.Money
	TotalInstallments         int
	FirstProjectedInstallment int
	ScheduledStart            planning.YearMonth
	ScheduledEnd              planning.YearMonth
	Periods                   []InstallmentPeriod
	Settlement                *EarlySettlement
	Status                    planning.FinancialItemStatus
}

type NewDebtInput struct {
	ID                        string
	Name                      string
	OriginalTotal             *planning.Money
	TotalInstallments         int
	FirstProjectedInstallment int
	ScheduledStart            planning.YearMonth
	Periods                   []InstallmentPeriod
	Settlement                *EarlySettlement
	Status                    planning.FinancialItemStatus
}

// NewDebt validates the complete structural schedule and derives its end. It
// accepts already-persisted temporal periods but never trusts a stored end.
func NewDebt(input NewDebtInput) (Debt, error) {
	if strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.Name) == "" ||
		!input.ScheduledStart.Valid() || !input.Status.Valid() ||
		input.TotalInstallments < 1 || input.FirstProjectedInstallment < 1 ||
		input.FirstProjectedInstallment > input.TotalInstallments {
		return Debt{}, ErrInvalidDebt
	}
	if input.OriginalTotal != nil && input.OriginalTotal.IsNegative() {
		return Debt{}, fmt.Errorf("%w: original total cannot be negative", ErrInvalidDebt)
	}

	remaining := input.TotalInstallments - input.FirstProjectedInstallment + 1
	if remaining > MaximumScheduleMonths {
		return Debt{}, ErrDebtScheduleTooLong
	}
	scheduledEnd, err := input.ScheduledStart.AddMonths(remaining - 1)
	if err != nil {
		return Debt{}, fmt.Errorf("%w: derive scheduled end: %v", ErrInvalidDebt, err)
	}

	periods, err := validateAndSortPeriods(input.Periods, input.ScheduledStart, scheduledEnd)
	if err != nil {
		return Debt{}, err
	}
	if err := validateSettlement(input.Settlement, input.ScheduledStart, scheduledEnd); err != nil {
		return Debt{}, err
	}

	return Debt{
		ID:                        input.ID,
		Name:                      input.Name,
		OriginalTotal:             input.OriginalTotal,
		TotalInstallments:         input.TotalInstallments,
		FirstProjectedInstallment: input.FirstProjectedInstallment,
		ScheduledStart:            input.ScheduledStart,
		ScheduledEnd:              scheduledEnd,
		Periods:                   periods,
		Settlement:                input.Settlement,
		Status:                    input.Status,
	}, nil
}

func validateAndSortPeriods(periods []InstallmentPeriod, start, end planning.YearMonth) ([]InstallmentPeriod, error) {
	if len(periods) == 0 {
		return nil, ErrDebtPeriodGap
	}

	ordered := append([]InstallmentPeriod(nil), periods...)
	sort.Slice(ordered, func(i, j int) bool {
		comparison := ordered[i].Interval.Start().Compare(ordered[j].Interval.Start())
		if comparison == 0 {
			return ordered[i].ID < ordered[j].ID
		}
		return comparison < 0
	})

	for index, period := range ordered {
		periodEnd, hasEnd := period.Interval.End()
		if strings.TrimSpace(period.ID) == "" || !period.Interval.Valid() || !hasEnd || period.Amount.Cents() <= 0 ||
			period.Interval.Start().Before(start) || periodEnd.After(end) {
			return nil, fmt.Errorf("%w: invalid period at index %d", ErrInvalidDebt, index)
		}
		if index == 0 {
			if period.Interval.Start() != start {
				return nil, ErrDebtPeriodGap
			}
			continue
		}

		previousEnd, _ := ordered[index-1].Interval.End()
		if !period.Interval.Start().After(previousEnd) {
			return nil, ErrDebtPeriodOverlap
		}
		expectedStart, err := previousEnd.AddMonths(1)
		if err != nil || period.Interval.Start() != expectedStart {
			return nil, ErrDebtPeriodGap
		}
	}

	lastEnd, _ := ordered[len(ordered)-1].Interval.End()
	if lastEnd != end {
		return nil, ErrDebtPeriodGap
	}
	return ordered, nil
}

func validateSettlement(settlement *EarlySettlement, start, end planning.YearMonth) error {
	if settlement == nil {
		return nil
	}
	if strings.TrimSpace(settlement.ID) == "" || !settlement.ReferenceMonth.Valid() ||
		settlement.ReferenceMonth.Before(start) || !settlement.ReferenceMonth.Before(end) ||
		settlement.Amount.Cents() <= 0 {
		return ErrInvalidSettlement
	}
	return nil
}
