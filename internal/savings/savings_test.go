package savings

import (
	"context"
	"errors"
	"testing"

	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
)

type fakePlans struct{ plan planning.Plan }

func (plans fakePlans) Current(context.Context, string) (planning.Plan, error) {
	return plans.plan, nil
}

type fakeRepository struct {
	input         PutInput
	configuration Configuration
}

func (repository *fakeRepository) Get(context.Context, string, string) (Configuration, error) {
	return repository.configuration, nil
}
func (repository *fakeRepository) Put(_ context.Context, _, _ string, input PutInput) (Configuration, error) {
	repository.input = input
	return Configuration{Configured: true}, nil
}

func TestPutAcceptsExplicitZero(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, fakePlans{savingsPlan(t)})
	result, err := service.Put(context.Background(), "owner", PutInput{EffectiveFrom: savingsMonth(t, "2026-01"), AmountCents: 0})
	if err != nil || !result.Configured || repository.input.AmountCents != 0 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}
func TestGetPreservesAbsentConfiguration(t *testing.T) {
	service := NewService(&fakeRepository{configuration: Configuration{Configured: false, Periods: []Period{}}}, fakePlans{savingsPlan(t)})
	result, err := service.Get(context.Background(), "owner")
	if err != nil || result.Configured {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}
func TestPutRejectsOutsideHorizon(t *testing.T) {
	service := NewService(&fakeRepository{}, fakePlans{savingsPlan(t)})
	_, err := service.Put(context.Background(), "owner", PutInput{EffectiveFrom: savingsMonth(t, "2027-01"), AmountCents: 1})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error=%v", err)
	}
}
func savingsPlan(t *testing.T) planning.Plan {
	return planning.Plan{ID: "plan", StartMonth: savingsMonth(t, "2026-01"), EndMonth: savingsMonth(t, "2026-12"), Status: domain.PlanStatusDraft}
}
func savingsMonth(t *testing.T, value string) domain.YearMonth {
	t.Helper()
	result, err := domain.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
