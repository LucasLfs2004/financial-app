package planning

import (
	"context"
	"errors"
	"testing"

	"github.com/lucas/financial-api/internal/planning/domain"
)

type fakeRepository struct {
	createdInput CreateInput
	createdOwner string
	createResult Plan
	createErr    error

	currentResult Plan
	currentOwner  string
	currentErr    error

	updatedInput UpdateInput
	updatedOwner string
	updateResult Plan
	updateErr    error
}

func (repository *fakeRepository) Create(_ context.Context, ownerID string, input CreateInput) (Plan, error) {
	repository.createdOwner = ownerID
	repository.createdInput = input
	return repository.createResult, repository.createErr
}

func (repository *fakeRepository) FindCurrent(_ context.Context, ownerID string) (Plan, error) {
	repository.currentOwner = ownerID
	return repository.currentResult, repository.currentErr
}

func (repository *fakeRepository) UpdateDraft(_ context.Context, ownerID string, input UpdateInput) (Plan, error) {
	repository.updatedOwner = ownerID
	repository.updatedInput = input
	return repository.updateResult, repository.updateErr
}
func (repository *fakeRepository) Activate(context.Context, string) (Activation, error) {
	return Activation{}, nil
}
func (repository *fakeRepository) FindOriginal(context.Context, string) (Snapshot, error) {
	return Snapshot{}, nil
}

func TestCreateUsesAuthenticatedOwnerAndNormalizesName(t *testing.T) {
	repository := &fakeRepository{createResult: samplePlan(t)}
	service := NewService(repository)

	_, err := service.Create(context.Background(), "owner-1", CreateInput{
		Name:         "  Planejamento 2026  ",
		StartMonth:   mustMonth(t, "2026-01"),
		EndMonth:     mustMonth(t, "2026-12"),
		CurrencyCode: "BRL",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if repository.createdOwner != "owner-1" {
		t.Fatalf("owner = %q, want authenticated owner", repository.createdOwner)
	}
	if repository.createdInput.Name != "Planejamento 2026" {
		t.Fatalf("name = %q, want normalized name", repository.createdInput.Name)
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.Create(context.Background(), "owner-1", CreateInput{
		Name:         "",
		StartMonth:   mustMonth(t, "2026-12"),
		EndMonth:     mustMonth(t, "2026-01"),
		CurrencyCode: "brl",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}
}

func TestUpdatePreservesUnspecifiedHorizonValue(t *testing.T) {
	repository := &fakeRepository{
		currentResult: samplePlan(t),
		updateResult:  samplePlan(t),
	}
	service := NewService(repository)
	name := "Novo nome"
	_, err := service.UpdateCurrent(context.Background(), "owner-1", UpdateInput{Name: &name})
	if err != nil {
		t.Fatalf("UpdateCurrent returned error: %v", err)
	}
	if repository.currentOwner != "owner-1" || repository.updatedOwner != "owner-1" {
		t.Fatal("update did not use authenticated owner for both lookups")
	}
	if repository.updatedInput.StartMonth == nil || repository.updatedInput.EndMonth == nil {
		t.Fatal("update did not complete horizon from current plan")
	}
	if repository.updatedInput.StartMonth.String() != "2026-01" || repository.updatedInput.EndMonth.String() != "2026-12" {
		t.Fatalf("unexpected completed horizon: %s to %s", repository.updatedInput.StartMonth, repository.updatedInput.EndMonth)
	}
}

func TestUpdateRejectsActivePlan(t *testing.T) {
	active := samplePlan(t)
	active.Status = domain.PlanStatusActive
	service := NewService(&fakeRepository{currentResult: active})
	name := "Novo nome"
	_, err := service.UpdateCurrent(context.Background(), "owner-1", UpdateInput{Name: &name})
	if !errors.Is(err, ErrNotDraft) {
		t.Fatalf("error = %v, want ErrNotDraft", err)
	}
}

func TestUpdateRejectsInvalidMergedHorizon(t *testing.T) {
	repository := &fakeRepository{currentResult: samplePlan(t)}
	service := NewService(repository)
	startMonth := mustMonth(t, "2027-01")
	_, err := service.UpdateCurrent(context.Background(), "owner-1", UpdateInput{StartMonth: &startMonth})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}
}

func samplePlan(t *testing.T) Plan {
	t.Helper()
	return Plan{
		ID:           "plan-1",
		UserID:       "owner-1",
		Name:         "Planejamento",
		Status:       domain.PlanStatusDraft,
		StartMonth:   mustMonth(t, "2026-01"),
		EndMonth:     mustMonth(t, "2026-12"),
		CurrencyCode: "BRL",
	}
}

func mustMonth(t *testing.T, value string) domain.YearMonth {
	t.Helper()
	month, err := domain.ParseYearMonth(value)
	if err != nil {
		t.Fatalf("parse month %q: %v", value, err)
	}
	return month
}
