package domain

import (
	"fmt"
	"sort"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

type Occurrence struct {
	DebtID            string
	SourceID          string
	ReferenceMonth    planning.YearMonth
	InstallmentNumber int
	InstallmentsTotal int
	Amount            planning.Money
	Kind              OccurrenceKind
}

type ProjectionInput struct {
	Debt Debt
	From planning.YearMonth
	To   planning.YearMonth
	AsOf planning.YearMonth
}

type Projection struct {
	ScheduledEnd          planning.YearMonth
	EffectiveEnd          planning.YearMonth
	ReleaseFrom           planning.YearMonth
	ReleasedMonthly       planning.Money
	ProjectionStatus      ProjectionStatus
	RemainingInstallments int
	Occurrences           []Occurrence
}

type Release struct {
	DebtID          string
	Name            string
	ScheduledEnd    planning.YearMonth
	EffectiveEnd    planning.YearMonth
	ReleaseFrom     planning.YearMonth
	ReleasedMonthly planning.Money
	Reason          ReleaseReason
}

type MonthlyReleaseTotal struct {
	Month  planning.YearMonth
	Amount planning.Money
}

type ReleaseProjectionInput struct {
	Debts []Debt
	From  planning.YearMonth
	To    planning.YearMonth
}

type ReleaseProjection struct {
	Releases      []Release
	MonthlyTotals []MonthlyReleaseTotal
}

// ProjectDebt expands a validated structural schedule without consulting a
// clock, database or transport concern. The requested range is inclusive.
func ProjectDebt(input ProjectionInput) (Projection, error) {
	if err := validateProjectionRange(input.From, input.To); err != nil {
		return Projection{}, err
	}
	if !input.AsOf.Valid() {
		return Projection{}, ErrInvalidProjectionRange
	}

	debt, err := normalizeDebt(input.Debt)
	if err != nil {
		return Projection{}, err
	}
	effectiveEnd := debt.ScheduledEnd
	releaseReason := ReleaseReasonScheduledCompletion
	if debt.Settlement != nil {
		effectiveEnd = debt.Settlement.ReferenceMonth
		releaseReason = ReleaseReasonEarlySettlement
	}
	releaseFrom, err := effectiveEnd.AddMonths(1)
	if err != nil {
		return Projection{}, fmt.Errorf("%w: derive release month: %v", ErrInvalidDebt, err)
	}

	releasedMonth := debt.ScheduledEnd
	if releaseReason == ReleaseReasonEarlySettlement {
		releasedMonth, err = debt.Settlement.ReferenceMonth.AddMonths(1)
		if err != nil {
			return Projection{}, fmt.Errorf("%w: derive first suppressed month: %v", ErrInvalidDebt, err)
		}
	}
	releasedPeriod, ok := periodAt(debt.Periods, releasedMonth)
	if !ok {
		return Projection{}, ErrDebtPeriodGap
	}

	projection := Projection{
		ScheduledEnd:          debt.ScheduledEnd,
		EffectiveEnd:          effectiveEnd,
		ReleaseFrom:           releaseFrom,
		ReleasedMonthly:       releasedPeriod.Amount,
		ProjectionStatus:      projectionStatus(debt, input.AsOf, effectiveEnd),
		RemainingInstallments: remainingInstallments(debt, input.AsOf, effectiveEnd),
		Occurrences:           make([]Occurrence, 0),
	}

	first := maxMonth(input.From, debt.ScheduledStart)
	last := minMonth(input.To, effectiveEnd)
	if last.Before(first) {
		return projection, nil
	}

	for month := first; !month.After(last); {
		period, exists := periodAt(debt.Periods, month)
		if !exists {
			return Projection{}, ErrDebtPeriodGap
		}
		sourceID := period.ID
		amount := period.Amount
		kind := OccurrenceKindScheduled
		if debt.Settlement != nil && month == debt.Settlement.ReferenceMonth {
			sourceID = debt.Settlement.ID
			amount = debt.Settlement.Amount
			kind = OccurrenceKindEarlySettlement
		}
		projection.Occurrences = append(projection.Occurrences, Occurrence{
			DebtID:            debt.ID,
			SourceID:          sourceID,
			ReferenceMonth:    month,
			InstallmentNumber: debt.FirstProjectedInstallment + monthsBetween(debt.ScheduledStart, month),
			InstallmentsTotal: debt.TotalInstallments,
			Amount:            amount,
			Kind:              kind,
		})

		if month == last {
			break
		}
		month, err = month.AddMonths(1)
		if err != nil {
			return Projection{}, fmt.Errorf("%w: advance occurrence month: %v", ErrInvalidDebt, err)
		}
	}
	return projection, nil
}

// ProjectReleases selects releases inside an inclusive range and aggregates
// them with overflow-safe Money operations.
func ProjectReleases(input ReleaseProjectionInput) (ReleaseProjection, error) {
	if err := validateProjectionRange(input.From, input.To); err != nil {
		return ReleaseProjection{}, err
	}

	result := ReleaseProjection{Releases: make([]Release, 0), MonthlyTotals: make([]MonthlyReleaseTotal, 0)}
	totals := make(map[string]planning.Money)
	months := make(map[string]planning.YearMonth)
	for _, candidate := range input.Debts {
		debt, err := normalizeDebt(candidate)
		if err != nil {
			return ReleaseProjection{}, err
		}
		projection, err := ProjectDebt(ProjectionInput{
			Debt: debt,
			From: debt.ScheduledStart,
			To:   debt.ScheduledEnd,
			AsOf: input.From,
		})
		if err != nil {
			return ReleaseProjection{}, err
		}
		if projection.ReleaseFrom.Before(input.From) || projection.ReleaseFrom.After(input.To) {
			continue
		}
		reason := ReleaseReasonScheduledCompletion
		if debt.Settlement != nil {
			reason = ReleaseReasonEarlySettlement
		}
		result.Releases = append(result.Releases, Release{
			DebtID:          debt.ID,
			Name:            debt.Name,
			ScheduledEnd:    projection.ScheduledEnd,
			EffectiveEnd:    projection.EffectiveEnd,
			ReleaseFrom:     projection.ReleaseFrom,
			ReleasedMonthly: projection.ReleasedMonthly,
			Reason:          reason,
		})

		key := projection.ReleaseFrom.String()
		total, err := totals[key].Add(projection.ReleasedMonthly)
		if err != nil {
			return ReleaseProjection{}, fmt.Errorf("aggregate releases for %s: %w", key, err)
		}
		totals[key] = total
		months[key] = projection.ReleaseFrom
	}

	sort.Slice(result.Releases, func(i, j int) bool {
		comparison := result.Releases[i].ReleaseFrom.Compare(result.Releases[j].ReleaseFrom)
		if comparison == 0 {
			return result.Releases[i].DebtID < result.Releases[j].DebtID
		}
		return comparison < 0
	})
	for key, total := range totals {
		result.MonthlyTotals = append(result.MonthlyTotals, MonthlyReleaseTotal{Month: months[key], Amount: total})
	}
	sort.Slice(result.MonthlyTotals, func(i, j int) bool {
		return result.MonthlyTotals[i].Month.Before(result.MonthlyTotals[j].Month)
	})
	return result, nil
}

func normalizeDebt(candidate Debt) (Debt, error) {
	debt, err := NewDebt(NewDebtInput{
		ID:                        candidate.ID,
		UserID:                    candidate.UserID,
		CurrencyCode:              candidate.CurrencyCode,
		Name:                      candidate.Name,
		Description:               candidate.Description,
		OriginalTotal:             candidate.OriginalTotal,
		TotalInstallments:         candidate.TotalInstallments,
		FirstProjectedInstallment: candidate.FirstProjectedInstallment,
		ScheduledStart:            candidate.ScheduledStart,
		Periods:                   candidate.Periods,
		Settlement:                candidate.Settlement,
		Status:                    candidate.Status,
		ArchivedAt:                candidate.ArchivedAt,
		CreatedAt:                 candidate.CreatedAt,
		UpdatedAt:                 candidate.UpdatedAt,
	})
	if err != nil {
		return Debt{}, err
	}
	if candidate.ScheduledEnd.Valid() && candidate.ScheduledEnd != debt.ScheduledEnd {
		return Debt{}, fmt.Errorf("%w: scheduled end does not match installment structure", ErrInvalidDebt)
	}
	return debt, nil
}

func validateProjectionRange(from, to planning.YearMonth) error {
	if !from.Valid() || !to.Valid() || to.Before(from) {
		return ErrInvalidProjectionRange
	}
	if monthsBetween(from, to)+1 > MaximumScheduleMonths {
		return fmt.Errorf("%w: range exceeds %d months", ErrInvalidProjectionRange, MaximumScheduleMonths)
	}
	return nil
}

func projectionStatus(debt Debt, asOf, effectiveEnd planning.YearMonth) ProjectionStatus {
	if debt.Status == planning.FinancialItemStatusArchived {
		return ProjectionStatusArchived
	}
	if debt.Settlement != nil && !asOf.Before(debt.Settlement.ReferenceMonth) {
		return ProjectionStatusSettledEarly
	}
	if asOf.Before(debt.ScheduledStart) {
		return ProjectionStatusPlanned
	}
	if asOf.After(effectiveEnd) {
		return ProjectionStatusCompleted
	}
	return ProjectionStatusActive
}

func remainingInstallments(debt Debt, asOf, effectiveEnd planning.YearMonth) int {
	if debt.Status == planning.FinancialItemStatusArchived || asOf.After(effectiveEnd) ||
		(debt.Settlement != nil && !asOf.Before(debt.Settlement.ReferenceMonth)) {
		return 0
	}
	first := maxMonth(asOf, debt.ScheduledStart)
	return monthsBetween(first, effectiveEnd) + 1
}

func periodAt(periods []InstallmentPeriod, month planning.YearMonth) (InstallmentPeriod, bool) {
	for _, period := range periods {
		if period.Interval.Contains(month) {
			return period, true
		}
	}
	return InstallmentPeriod{}, false
}

func monthsBetween(from, to planning.YearMonth) int {
	return (to.Year()-from.Year())*12 + int(to.Month()-from.Month())
}

func maxMonth(left, right planning.YearMonth) planning.YearMonth {
	if left.After(right) {
		return left
	}
	return right
}

func minMonth(left, right planning.YearMonth) planning.YearMonth {
	if left.Before(right) {
		return left
	}
	return right
}
