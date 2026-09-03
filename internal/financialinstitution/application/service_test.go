package application

import (
	"context"
	"errors"
	"testing"

	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
)

const testInstitutionID = "a1100000-0000-0000-0000-000000000001"

type fakeRepository struct {
	createdName string
	listStatus  *institutiondomain.Status
	updatedName string
	archivedID  string
	err         error
}

func (repository *fakeRepository) Create(_ context.Context, _ string, name string) (institutiondomain.Institution, error) {
	repository.createdName = name
	return institutiondomain.Institution{ID: testInstitutionID, Name: name, Status: institutiondomain.StatusActive}, repository.err
}

func (repository *fakeRepository) List(_ context.Context, _ string, status *institutiondomain.Status) ([]institutiondomain.Institution, error) {
	repository.listStatus = status
	return []institutiondomain.Institution{}, repository.err
}

func (repository *fakeRepository) UpdateName(_ context.Context, _, _ string, name string) (institutiondomain.Institution, error) {
	repository.updatedName = name
	return institutiondomain.Institution{ID: testInstitutionID, Name: name, Status: institutiondomain.StatusActive}, repository.err
}

func (repository *fakeRepository) Archive(_ context.Context, _, institutionID string) (institutiondomain.Institution, error) {
	repository.archivedID = institutionID
	return institutiondomain.Institution{ID: institutionID, Status: institutiondomain.StatusArchived}, repository.err
}

func TestServiceNormalizesCreateAndUpdate(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	created, err := service.Create(context.Background(), "owner", CreateInput{Name: "  Banco Alfa  "})
	if err != nil || created.Name != "Banco Alfa" || repository.createdName != "Banco Alfa" {
		t.Fatalf("created=%+v name=%q error=%v", created, repository.createdName, err)
	}

	updated, err := service.Update(context.Background(), "owner", testInstitutionID, UpdateInput{Name: "  Banco Beta  "})
	if err != nil || updated.Name != "Banco Beta" || repository.updatedName != "Banco Beta" {
		t.Fatalf("updated=%+v name=%q error=%v", updated, repository.updatedName, err)
	}
}

func TestServiceValidatesIdentityAndFilters(t *testing.T) {
	service := NewService(&fakeRepository{})
	if _, err := service.Create(context.Background(), "", CreateInput{Name: "Banco"}); !errors.Is(err, institutiondomain.ErrValidation) {
		t.Fatalf("create error=%v", err)
	}
	if _, err := service.Update(context.Background(), "owner", "invalid", UpdateInput{Name: "Banco"}); !errors.Is(err, institutiondomain.ErrValidation) {
		t.Fatalf("update error=%v", err)
	}
	invalidStatus := institutiondomain.Status("disabled")
	if _, err := service.List(context.Background(), "owner", ListFilters{Status: &invalidStatus}); !errors.Is(err, institutiondomain.ErrValidation) {
		t.Fatalf("list error=%v", err)
	}
}

func TestServiceArchivesAfterValidatingReason(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	reason := "cadastro antigo"

	institution, err := service.Archive(context.Background(), "owner", testInstitutionID, ArchiveInput{Reason: &reason})
	if err != nil || institution.Status != institutiondomain.StatusArchived || repository.archivedID != testInstitutionID {
		t.Fatalf("institution=%+v archived=%q error=%v", institution, repository.archivedID, err)
	}
}

func TestServicePropagatesRepositoryErrors(t *testing.T) {
	repository := &fakeRepository{err: ErrAlreadyExists}
	service := NewService(repository)
	if _, err := service.Create(context.Background(), "owner", CreateInput{Name: "Banco"}); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("error=%v", err)
	}
}
