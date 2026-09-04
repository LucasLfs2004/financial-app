package application_test

import (
	"context"
	"testing"

	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

const validID = "d1000000-0000-0000-0000-000000000001"

type repositoryStub struct {
	createdName          string
	createdConfiguration carddomain.ConfigurationInput
}

func (repository *repositoryStub) Create(_ context.Context, _, _, name string, configuration carddomain.ConfigurationInput) (carddomain.Card, error) {
	repository.createdName = name
	repository.createdConfiguration = configuration
	return carddomain.Card{Name: name, Configurations: []carddomain.Configuration{{NominalDueDay: configuration.NominalDueDay}}}, nil
}
func (*repositoryStub) List(context.Context, string, *carddomain.Status) ([]carddomain.Card, error) {
	return nil, nil
}
func (*repositoryStub) Find(context.Context, string, string) (carddomain.Card, error) {
	return carddomain.Card{}, nil
}
func (*repositoryStub) UpdateName(context.Context, string, string, string) (carddomain.Card, error) {
	return carddomain.Card{}, nil
}
func (*repositoryStub) ChangeConfiguration(context.Context, string, string, carddomain.ConfigurationInput) (carddomain.Card, error) {
	return carddomain.Card{}, nil
}
func (*repositoryStub) Archive(context.Context, string, string) (carddomain.Card, error) {
	return carddomain.Card{}, nil
}

func TestCreateNormalizesAndDelegates(t *testing.T) {
	month, _ := planningdomain.ParseYearMonth("2026-01")
	repository := &repositoryStub{}
	service := cardapplication.NewService(repository)
	card, err := service.Create(context.Background(), "owner", cardapplication.CreateInput{
		InstitutionID: validID, Name: "  Cartão principal  ",
		Configuration: carddomain.ConfigurationInput{EffectiveFrom: month, NominalDueDay: 31, PaymentMonthOffset: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if card.Name != "Cartão principal" || repository.createdName != "Cartão principal" {
		t.Fatalf("card=%+v", card)
	}
	if repository.createdConfiguration.PaymentMonthOffset != 2 {
		t.Fatalf("configuration=%+v", repository.createdConfiguration)
	}
}

func TestFindRejectsInvalidCardID(t *testing.T) {
	service := cardapplication.NewService(&repositoryStub{})
	if _, err := service.Find(context.Background(), "owner", "invalid"); err == nil {
		t.Fatal("expected validation error")
	}
}
