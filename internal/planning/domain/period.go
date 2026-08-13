package domain

import "fmt"

const (
	MinimumCashMonthOffset = 0
	MaximumCashMonthOffset = 12
)

// MonthInterval represents an inclusive interval. An interval without an end
// remains valid until it is explicitly closed or the plan horizon is reached.
type MonthInterval struct {
	start  YearMonth
	end    YearMonth
	hasEnd bool
}

func NewMonthInterval(start, end YearMonth) (MonthInterval, error) {
	if !start.Valid() || !end.Valid() {
		return MonthInterval{}, fmt.Errorf("%w: start and end must be valid months", ErrInvalidMonthInterval)
	}
	if end.Before(start) {
		return MonthInterval{}, fmt.Errorf("%w: end %s is before start %s", ErrInvalidMonthInterval, end, start)
	}

	return MonthInterval{start: start, end: end, hasEnd: true}, nil
}

func NewOpenMonthInterval(start YearMonth) (MonthInterval, error) {
	if !start.Valid() {
		return MonthInterval{}, fmt.Errorf("%w: start must be a valid month", ErrInvalidMonthInterval)
	}

	return MonthInterval{start: start}, nil
}

func (i MonthInterval) Start() YearMonth {
	return i.start
}

func (i MonthInterval) End() (YearMonth, bool) {
	return i.end, i.hasEnd
}

func (i MonthInterval) IsOpen() bool {
	return !i.hasEnd
}

func (i MonthInterval) Valid() bool {
	if !i.start.Valid() {
		return false
	}
	return !i.hasEnd || (i.end.Valid() && !i.end.Before(i.start))
}

func (i MonthInterval) Contains(month YearMonth) bool {
	if month.Before(i.start) {
		return false
	}
	return !i.hasEnd || !month.After(i.end)
}

func (i MonthInterval) Overlaps(other MonthInterval) bool {
	if !i.Valid() || !other.Valid() {
		return false
	}
	if i.hasEnd && i.end.Before(other.start) {
		return false
	}
	if other.hasEnd && other.end.Before(i.start) {
		return false
	}
	return true
}

type FinancialPeriod struct {
	interval        MonthInterval
	amount          Money
	recurrence      Recurrence
	cashMonthOffset int
}

func NewFinancialPeriod(
	interval MonthInterval,
	amount Money,
	recurrence Recurrence,
	cashMonthOffset int,
) (FinancialPeriod, error) {
	if !interval.Valid() {
		return FinancialPeriod{}, ErrInvalidMonthInterval
	}
	if amount.IsNegative() {
		return FinancialPeriod{}, ErrNegativeAmount
	}
	if !recurrence.Valid() {
		return FinancialPeriod{}, fmt.Errorf("%w: %q", ErrInvalidRecurrence, recurrence)
	}
	if cashMonthOffset < MinimumCashMonthOffset || cashMonthOffset > MaximumCashMonthOffset {
		return FinancialPeriod{}, fmt.Errorf("%w: %d", ErrInvalidCashMonthOffset, cashMonthOffset)
	}
	if recurrence == RecurrenceOnce {
		end, hasEnd := interval.End()
		if !hasEnd || end != interval.Start() {
			return FinancialPeriod{}, fmt.Errorf("%w: one-time occurrence must start and end in the same month", ErrInvalidRecurrence)
		}
	}

	return FinancialPeriod{
		interval:        interval,
		amount:          amount,
		recurrence:      recurrence,
		cashMonthOffset: cashMonthOffset,
	}, nil
}

func (p FinancialPeriod) Interval() MonthInterval {
	return p.interval
}

func (p FinancialPeriod) Amount() Money {
	return p.amount
}

func (p FinancialPeriod) Recurrence() Recurrence {
	return p.recurrence
}

func (p FinancialPeriod) CashMonthOffset() int {
	return p.cashMonthOffset
}

func (p FinancialPeriod) CashMonth(referenceMonth YearMonth) (YearMonth, error) {
	return referenceMonth.AddMonths(p.cashMonthOffset)
}

func ValidateFinancialItemRecurrence(kind FinancialItemKind, recurrence Recurrence) error {
	if !kind.Valid() || !recurrence.Valid() {
		return fmt.Errorf("%w: kind=%q recurrence=%q", ErrInvalidRecurrence, kind, recurrence)
	}

	valid := recurrence == RecurrenceMonthly
	if kind == FinancialItemKindOneTimeIncome {
		valid = recurrence == RecurrenceOnce
	}
	if kind == FinancialItemKindProjectedVariableExpense {
		valid = true
	}
	if !valid {
		return fmt.Errorf("%w: kind=%q recurrence=%q", ErrInvalidRecurrence, kind, recurrence)
	}

	return nil
}
