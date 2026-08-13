package domain

import (
	"fmt"
	"strconv"
	"time"
)

const (
	minimumYear = 1
	maximumYear = 9999
)

// YearMonth is a calendar month without a day or timezone.
type YearMonth struct {
	year  int
	month time.Month
}

func NewYearMonth(year int, month time.Month) (YearMonth, error) {
	if year < minimumYear || year > maximumYear || month < time.January || month > time.December {
		return YearMonth{}, fmt.Errorf("%w: year=%d month=%d", ErrInvalidYearMonth, year, month)
	}

	return YearMonth{year: year, month: month}, nil
}

func ParseYearMonth(value string) (YearMonth, error) {
	if len(value) != len("2006-01") || value[4] != '-' {
		return YearMonth{}, fmt.Errorf("%w: %q must use YYYY-MM", ErrInvalidYearMonth, value)
	}

	year, err := strconv.Atoi(value[:4])
	if err != nil {
		return YearMonth{}, fmt.Errorf("%w: invalid year in %q", ErrInvalidYearMonth, value)
	}

	month, err := strconv.Atoi(value[5:])
	if err != nil {
		return YearMonth{}, fmt.Errorf("%w: invalid month in %q", ErrInvalidYearMonth, value)
	}

	return NewYearMonth(year, time.Month(month))
}

func (m YearMonth) Year() int {
	return m.year
}

func (m YearMonth) Month() time.Month {
	return m.month
}

func (m YearMonth) Valid() bool {
	return m.year >= minimumYear && m.year <= maximumYear && m.month >= time.January && m.month <= time.December
}

func (m YearMonth) String() string {
	return fmt.Sprintf("%04d-%02d", m.year, m.month)
}

func (m YearMonth) Time() time.Time {
	return time.Date(m.year, m.month, 1, 0, 0, 0, 0, time.UTC)
}

func (m YearMonth) Compare(other YearMonth) int {
	switch {
	case m.year < other.year || (m.year == other.year && m.month < other.month):
		return -1
	case m.year > other.year || (m.year == other.year && m.month > other.month):
		return 1
	default:
		return 0
	}
}

func (m YearMonth) Before(other YearMonth) bool {
	return m.Compare(other) < 0
}

func (m YearMonth) After(other YearMonth) bool {
	return m.Compare(other) > 0
}

func (m YearMonth) AddMonths(offset int) (YearMonth, error) {
	if !m.Valid() {
		return YearMonth{}, fmt.Errorf("%w: cannot add months to an invalid value", ErrInvalidYearMonth)
	}

	monthIndex := int64(m.year)*12 + int64(m.month-1)
	minimumIndex := int64(minimumYear) * 12
	maximumIndex := int64(maximumYear)*12 + 11
	offset64 := int64(offset)
	if offset64 < minimumIndex-monthIndex || offset64 > maximumIndex-monthIndex {
		return YearMonth{}, fmt.Errorf("%w: adding %d months to %s exceeds the supported range", ErrInvalidYearMonth, offset, m)
	}
	monthIndex += offset64

	year := int(monthIndex / 12)
	month := time.Month(monthIndex%12 + 1)
	return NewYearMonth(year, month)
}
