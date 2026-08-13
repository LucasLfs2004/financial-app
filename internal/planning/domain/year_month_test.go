package domain

import (
	"errors"
	"testing"
	"time"
)

func TestParseYearMonth(t *testing.T) {
	month, err := ParseYearMonth("2026-07")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if month.Year() != 2026 || month.Month() != time.July || month.String() != "2026-07" {
		t.Fatalf("unexpected month: %+v", month)
	}
	if got := month.Time(); !got.Equal(time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected normalized time: %v", got)
	}
}

func TestParseYearMonthRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"", "2026-1", "2026-13", "2026-00", "0000-01", "10000-01", "abcd-01", "2026-aa", "2026-01x"} {
		t.Run(value, func(t *testing.T) {
			_, err := ParseYearMonth(value)
			if !errors.Is(err, ErrInvalidYearMonth) {
				t.Fatalf("ParseYearMonth(%q) error = %v, want ErrInvalidYearMonth", value, err)
			}
		})
	}
}

func TestNewYearMonthRejectsInvalidParts(t *testing.T) {
	for _, input := range []struct {
		year  int
		month time.Month
	}{{0, time.January}, {10000, time.January}, {2026, 0}, {2026, 13}} {
		_, err := NewYearMonth(input.year, input.month)
		if !errors.Is(err, ErrInvalidYearMonth) {
			t.Fatalf("NewYearMonth(%d, %d) error = %v", input.year, input.month, err)
		}
	}
}

func TestYearMonthAddAndCompare(t *testing.T) {
	november := mustYearMonth(t, "2026-11")
	december, err := november.AddMonths(1)
	if err != nil || december.String() != "2026-12" {
		t.Fatalf("adding one month: got %s, error %v", december, err)
	}
	january, err := november.AddMonths(2)
	if err != nil || january.String() != "2027-01" {
		t.Fatalf("crossing year: got %s, error %v", january, err)
	}
	if !november.Before(january) || !january.After(november) || november.Compare(november) != 0 {
		t.Fatal("comparison methods returned inconsistent results")
	}
}

func TestYearMonthAddRejectsSupportedRangeOverflow(t *testing.T) {
	maximum := mustYearMonth(t, "9999-12")
	if _, err := maximum.AddMonths(1); !errors.Is(err, ErrInvalidYearMonth) {
		t.Fatalf("got error %v, want ErrInvalidYearMonth", err)
	}
	minimum := mustYearMonth(t, "0001-01")
	if _, err := minimum.AddMonths(-1); !errors.Is(err, ErrInvalidYearMonth) {
		t.Fatalf("got error %v, want ErrInvalidYearMonth", err)
	}
	maximumInt := int(^uint(0) >> 1)
	if _, err := minimum.AddMonths(maximumInt); !errors.Is(err, ErrInvalidYearMonth) {
		t.Fatalf("large offset error = %v, want ErrInvalidYearMonth", err)
	}
	if _, err := (YearMonth{}).AddMonths(1); !errors.Is(err, ErrInvalidYearMonth) {
		t.Fatalf("zero value error = %v, want ErrInvalidYearMonth", err)
	}
}

func mustYearMonth(t *testing.T, value string) YearMonth {
	t.Helper()
	month, err := ParseYearMonth(value)
	if err != nil {
		t.Fatalf("ParseYearMonth(%q): %v", value, err)
	}
	return month
}
