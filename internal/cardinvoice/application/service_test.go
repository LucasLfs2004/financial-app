package application

import (
	"context"
	"errors"
	"testing"

	"github.com/lucas/financial-api/internal/cardinvoice"
	"github.com/lucas/financial-api/internal/planning"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

type planReaderStub struct {
	plan planning.Plan
	err  error
}

func (stub planReaderStub) Current(context.Context, string) (planning.Plan, error) {
	return stub.plan, stub.err
}

type repositoryStub struct {
	data   cardinvoice.ProjectionData
	err    error
	called bool
}

func (stub *repositoryStub) LoadProjectionData(context.Context, string, string) (cardinvoice.ProjectionData, error) {
	stub.called = true
	return stub.data, stub.err
}

func TestProjectRejectsMonthOutsideOperationalHorizonBeforeLoadingData(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository, planReaderStub{plan: testPlan(t)})
	_, err := service.Project(context.Background(), "owner", "card", month(t, "2028-01"))
	if !errors.Is(err, ErrOutsideOperationalHorizon) {
		t.Fatalf("expected outside horizon, got %v", err)
	}
	if repository.called {
		t.Fatal("repository should not be called for an invalid horizon")
	}
}

func TestProjectDelegatesPureComponentSelection(t *testing.T) {
	configuration, _ := cardinvoice.NewCardConfiguration(31, 1)
	interval, _ := planningdomain.NewOpenMonthInterval(month(t, "2026-01"))
	configurationPeriod, _ := cardinvoice.NewCardConfigurationPeriod("configuration", interval, configuration)
	repository := &repositoryStub{data: cardinvoice.ProjectionData{Cards: []cardinvoice.Card{{ID: "card", Name: "Principal", Configurations: []cardinvoice.CardConfigurationPeriod{configurationPeriod}}}}}
	service := NewService(repository, planReaderStub{plan: testPlan(t)})
	invoice, err := service.Project(context.Background(), "owner", "card", month(t, "2027-01"))
	if err != nil {
		t.Fatal(err)
	}
	if invoice.CardID != "card" || invoice.PaymentMonth.String() != "2027-01" || invoice.NominalDueDate.Resolution != cardinvoice.DueDateResolutionExact {
		t.Fatalf("invoice=%+v", invoice)
	}
}

func testPlan(t *testing.T) planning.Plan {
	t.Helper()
	return planning.Plan{ID: "plan", StartMonth: month(t, "2026-01"), EndMonth: month(t, "2026-12"), CurrencyCode: "BRL"}
}

func month(t *testing.T, value string) planningdomain.YearMonth {
	t.Helper()
	month, err := planningdomain.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return month
}
