package domain

import (
	"errors"
	"math"
	"testing"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

func TestProjectDebtDerivesRemainingScheduleAcrossYear(t *testing.T) {
	debt := testDebt(t, debtFixture{
		id: "debt", total: 12, first: 5, start: "2026-09",
		periods: []periodFixture{{id: "period", start: "2026-09", end: "2027-04", amount: 60000}},
	})

	projection, err := ProjectDebt(ProjectionInput{
		Debt: debt, From: month(t, "2026-09"), To: month(t, "2027-04"), AsOf: month(t, "2026-09"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if projection.ScheduledEnd.String() != "2027-04" || projection.EffectiveEnd.String() != "2027-04" ||
		projection.ReleaseFrom.String() != "2027-05" || projection.ReleasedMonthly.Cents() != 60000 {
		t.Fatalf("unexpected projection summary: %+v", projection)
	}
	if projection.ProjectionStatus != ProjectionStatusActive || projection.RemainingInstallments != 8 {
		t.Fatalf("status=%s remaining=%d", projection.ProjectionStatus, projection.RemainingInstallments)
	}
	if len(projection.Occurrences) != 8 {
		t.Fatalf("occurrences=%d", len(projection.Occurrences))
	}
	first, last := projection.Occurrences[0], projection.Occurrences[len(projection.Occurrences)-1]
	if first.InstallmentNumber != 5 || first.ReferenceMonth.String() != "2026-09" ||
		last.InstallmentNumber != 12 || last.ReferenceMonth.String() != "2027-04" {
		t.Fatalf("first=%+v last=%+v", first, last)
	}
}

func TestProjectDebtSelectsTemporalValueAndKeepsNumbering(t *testing.T) {
	debt := testDebt(t, debtFixture{
		id: "debt", total: 12, first: 5, start: "2026-09",
		periods: []periodFixture{
			{id: "old", start: "2026-09", end: "2026-12", amount: 60000},
			{id: "new", start: "2027-01", end: "2027-04", amount: 65000},
		},
	})
	projection, err := ProjectDebt(ProjectionInput{
		Debt: debt, From: month(t, "2026-12"), To: month(t, "2027-01"), AsOf: month(t, "2026-12"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Occurrences) != 2 || projection.Occurrences[0].Amount.Cents() != 60000 ||
		projection.Occurrences[0].InstallmentNumber != 8 || projection.Occurrences[1].Amount.Cents() != 65000 ||
		projection.Occurrences[1].InstallmentNumber != 9 || projection.ReleasedMonthly.Cents() != 65000 {
		t.Fatalf("projection=%+v", projection)
	}
}

func TestProjectDebtAppliesSubstitutiveEarlySettlement(t *testing.T) {
	debt := testDebt(t, debtFixture{
		id: "debt", total: 8, first: 5, start: "2026-09",
		periods:    []periodFixture{{id: "period", start: "2026-09", end: "2026-12", amount: 60000}},
		settlement: &settlementFixture{id: "settlement", month: "2026-10", amount: 150000},
	})
	projection, err := ProjectDebt(ProjectionInput{
		Debt: debt, From: month(t, "2026-09"), To: month(t, "2026-12"), AsOf: month(t, "2026-10"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if projection.EffectiveEnd.String() != "2026-10" || projection.ReleaseFrom.String() != "2026-11" ||
		projection.ReleasedMonthly.Cents() != 60000 || projection.ProjectionStatus != ProjectionStatusSettledEarly ||
		projection.RemainingInstallments != 0 {
		t.Fatalf("projection=%+v", projection)
	}
	if len(projection.Occurrences) != 2 {
		t.Fatalf("occurrences=%+v", projection.Occurrences)
	}
	settlement := projection.Occurrences[1]
	if settlement.SourceID != "settlement" || settlement.Kind != OccurrenceKindEarlySettlement ||
		settlement.Amount.Cents() != 150000 || settlement.InstallmentNumber != 6 {
		t.Fatalf("settlement occurrence=%+v", settlement)
	}
}

func TestProjectDebtCalculatesStatusesAndRemainingInstallments(t *testing.T) {
	base := debtFixture{
		id: "debt", total: 3, first: 1, start: "2026-09",
		periods: []periodFixture{{id: "period", start: "2026-09", end: "2026-11", amount: 100}},
	}
	tests := []struct {
		name      string
		asOf      string
		archived  bool
		status    ProjectionStatus
		remaining int
	}{
		{name: "planned", asOf: "2026-08", status: ProjectionStatusPlanned, remaining: 3},
		{name: "active last month", asOf: "2026-11", status: ProjectionStatusActive, remaining: 1},
		{name: "completed", asOf: "2026-12", status: ProjectionStatusCompleted, remaining: 0},
		{name: "archived", asOf: "2026-10", archived: true, status: ProjectionStatusArchived, remaining: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := base
			fixture.archived = test.archived
			projection, err := ProjectDebt(ProjectionInput{
				Debt: testDebt(t, fixture), From: month(t, "2026-09"), To: month(t, "2026-11"), AsOf: month(t, test.asOf),
			})
			if err != nil {
				t.Fatal(err)
			}
			if projection.ProjectionStatus != test.status || projection.RemainingInstallments != test.remaining {
				t.Fatalf("status=%s remaining=%d", projection.ProjectionStatus, projection.RemainingInstallments)
			}
		})
	}
}

func TestNewDebtRejectsInvalidStructures(t *testing.T) {
	tests := []struct {
		name    string
		fixture debtFixture
		want    error
	}{
		{
			name: "initial installment after total",
			fixture: debtFixture{id: "debt", total: 2, first: 3, start: "2026-01",
				periods: []periodFixture{{id: "period", start: "2026-01", end: "2026-01", amount: 1}}},
			want: ErrInvalidDebt,
		},
		{
			name: "schedule over limit",
			fixture: debtFixture{id: "debt", total: 121, first: 1, start: "2026-01",
				periods: []periodFixture{{id: "period", start: "2026-01", end: "2036-01", amount: 1}}},
			want: ErrDebtScheduleTooLong,
		},
		{
			name: "period gap",
			fixture: debtFixture{id: "debt", total: 3, first: 1, start: "2026-01", periods: []periodFixture{
				{id: "first", start: "2026-01", end: "2026-01", amount: 1},
				{id: "last", start: "2026-03", end: "2026-03", amount: 1},
			}},
			want: ErrDebtPeriodGap,
		},
		{
			name: "period overlap",
			fixture: debtFixture{id: "debt", total: 3, first: 1, start: "2026-01", periods: []periodFixture{
				{id: "first", start: "2026-01", end: "2026-02", amount: 1},
				{id: "last", start: "2026-02", end: "2026-03", amount: 1},
			}},
			want: ErrDebtPeriodOverlap,
		},
		{
			name: "settlement on last installment",
			fixture: debtFixture{id: "debt", total: 2, first: 1, start: "2026-01",
				periods:    []periodFixture{{id: "period", start: "2026-01", end: "2026-02", amount: 1}},
				settlement: &settlementFixture{id: "settlement", month: "2026-02", amount: 2}},
			want: ErrInvalidSettlement,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := buildDebt(t, test.fixture)
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestProjectDebtRejectsRangeOver120Months(t *testing.T) {
	debt := testDebt(t, debtFixture{
		id: "debt", total: 1, first: 1, start: "2026-01",
		periods: []periodFixture{{id: "period", start: "2026-01", end: "2026-01", amount: 1}},
	})
	_, err := ProjectDebt(ProjectionInput{
		Debt: debt, From: month(t, "2026-01"), To: month(t, "2036-01"), AsOf: month(t, "2026-01"),
	})
	if !errors.Is(err, ErrInvalidProjectionRange) {
		t.Fatalf("expected invalid range, got %v", err)
	}
}

func TestProjectReleasesSortsAndAggregatesSafely(t *testing.T) {
	late := testDebt(t, debtFixture{
		id: "z", total: 2, first: 1, start: "2026-11",
		periods: []periodFixture{{id: "z-period", start: "2026-11", end: "2026-12", amount: 60000}},
	})
	early := testDebt(t, debtFixture{
		id: "a", total: 3, first: 1, start: "2026-10",
		periods:    []periodFixture{{id: "a-period", start: "2026-10", end: "2026-12", amount: 40000}},
		settlement: &settlementFixture{id: "a-settlement", month: "2026-11", amount: 70000},
	})
	result, err := ProjectReleases(ReleaseProjectionInput{
		Debts: []Debt{late, early}, From: month(t, "2026-12"), To: month(t, "2027-01"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Releases) != 2 || result.Releases[0].DebtID != "a" || result.Releases[1].DebtID != "z" {
		t.Fatalf("releases=%+v", result.Releases)
	}
	if len(result.MonthlyTotals) != 2 || result.MonthlyTotals[0].Month.String() != "2026-12" ||
		result.MonthlyTotals[0].Amount.Cents() != 40000 || result.MonthlyTotals[1].Month.String() != "2027-01" ||
		result.MonthlyTotals[1].Amount.Cents() != 60000 {
		t.Fatalf("totals=%+v", result.MonthlyTotals)
	}
}

func TestProjectReleasesPropagatesMoneyOverflow(t *testing.T) {
	first := testDebt(t, debtFixture{
		id: "a", total: 1, first: 1, start: "2026-01",
		periods: []periodFixture{{id: "a-period", start: "2026-01", end: "2026-01", amount: math.MaxInt64}},
	})
	second := testDebt(t, debtFixture{
		id: "b", total: 1, first: 1, start: "2026-01",
		periods: []periodFixture{{id: "b-period", start: "2026-01", end: "2026-01", amount: 1}},
	})
	_, err := ProjectReleases(ReleaseProjectionInput{
		Debts: []Debt{first, second}, From: month(t, "2026-02"), To: month(t, "2026-02"),
	})
	if !errors.Is(err, planning.ErrMoneyOverflow) {
		t.Fatalf("expected money overflow, got %v", err)
	}
}

type periodFixture struct {
	id         string
	start, end string
	amount     int64
}

type settlementFixture struct {
	id, month string
	amount    int64
}

type debtFixture struct {
	id         string
	total      int
	first      int
	start      string
	periods    []periodFixture
	settlement *settlementFixture
	archived   bool
}

func testDebt(t *testing.T, fixture debtFixture) Debt {
	t.Helper()
	debt, err := buildDebt(t, fixture)
	if err != nil {
		t.Fatal(err)
	}
	return debt
}

func buildDebt(t *testing.T, fixture debtFixture) (Debt, error) {
	t.Helper()
	periods := make([]InstallmentPeriod, 0, len(fixture.periods))
	for _, value := range fixture.periods {
		interval, err := planning.NewMonthInterval(month(t, value.start), month(t, value.end))
		if err != nil {
			t.Fatal(err)
		}
		periods = append(periods, InstallmentPeriod{ID: value.id, Interval: interval, Amount: planning.NewMoney(value.amount)})
	}
	var settlement *EarlySettlement
	if fixture.settlement != nil {
		settlement = &EarlySettlement{
			ID: fixture.settlement.id, ReferenceMonth: month(t, fixture.settlement.month), Amount: planning.NewMoney(fixture.settlement.amount),
		}
	}
	status := planning.FinancialItemStatusActive
	if fixture.archived {
		status = planning.FinancialItemStatusArchived
	}
	return NewDebt(NewDebtInput{
		ID: fixture.id, Name: "Dívida", TotalInstallments: fixture.total,
		FirstProjectedInstallment: fixture.first, ScheduledStart: month(t, fixture.start),
		Periods: periods, Settlement: settlement, Status: status,
	})
}

func month(t *testing.T, value string) planning.YearMonth {
	t.Helper()
	month, err := planning.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return month
}
