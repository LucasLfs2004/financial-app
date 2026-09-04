package integration_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
	cardrepository "github.com/lucas/financial-api/internal/creditcard/repository"
	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutionrepository "github.com/lucas/financial-api/internal/financialinstitution/repository"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

const (
	cardOwnerOne = "d1000000-0000-0000-0000-000000000001"
	cardOwnerTwo = "d2000000-0000-0000-0000-000000000002"
)

func TestRelease2CreditCardLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	_, _ = pool.Exec(ctx, `delete from auth.users where id in ($1, $2)`, cardOwnerOne, cardOwnerTwo)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `delete from auth.users where id in ($1, $2)`, cardOwnerOne, cardOwnerTwo)
	})
	for id, email := range map[string]string{cardOwnerOne: "card-owner-1@example.com", cardOwnerTwo: "card-owner-2@example.com"} {
		_, err = pool.Exec(ctx, `
			insert into auth.users (
				id, instance_id, aud, role, email, encrypted_password,
				email_confirmed_at, raw_app_meta_data, raw_user_meta_data, created_at, updated_at
			) values ($1, '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', $2, '', now(),
				'{"provider":"email","providers":["email"]}', '{}', now(), now())
		`, id, email)
		if err != nil {
			t.Fatal(err)
		}
	}

	institutionService := institutionapplication.NewService(institutionrepository.NewPostgresRepository(pool))
	cardService := cardapplication.NewService(cardrepository.NewPostgresRepository(pool))
	institution, err := institutionService.Create(ctx, cardOwnerOne, institutionapplication.CreateInput{Name: "Banco Cartões"})
	if err != nil {
		t.Fatal(err)
	}
	otherInstitution, err := institutionService.Create(ctx, cardOwnerTwo, institutionapplication.CreateInput{Name: "Outro Banco"})
	if err != nil {
		t.Fatal(err)
	}
	archivedInstitution, err := institutionService.Create(ctx, cardOwnerOne, institutionapplication.CreateInput{Name: "Banco Arquivado"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = institutionService.Archive(ctx, cardOwnerOne, archivedInstitution.ID, institutionapplication.ArchiveInput{}); err != nil {
		t.Fatal(err)
	}

	start, _ := planningdomain.ParseYearMonth("2026-01")
	end, _ := planningdomain.ParseYearMonth("2026-12")
	card, err := cardService.Create(ctx, cardOwnerOne, cardapplication.CreateInput{
		InstitutionID: institution.ID,
		Name:          "  Cartão Principal  ",
		Configuration: carddomain.ConfigurationInput{
			EffectiveFrom: start, EndMonth: &end, NominalDueDay: 31, PaymentMonthOffset: 1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if card.Name != "Cartão Principal" || card.Institution.ID != institution.ID || len(card.Configurations) != 1 {
		t.Fatalf("created card=%+v", card)
	}
	if _, err := cardService.Create(ctx, cardOwnerOne, cardapplication.CreateInput{
		InstitutionID: institution.ID, Name: "cartão principal",
		Configuration: carddomain.ConfigurationInput{EffectiveFrom: start, EndMonth: &end, NominalDueDay: 10, PaymentMonthOffset: 1},
	}); !errors.Is(err, cardapplication.ErrAlreadyExists) {
		t.Fatalf("duplicate error=%v", err)
	}
	if _, err := cardService.Create(ctx, cardOwnerOne, cardapplication.CreateInput{
		InstitutionID: otherInstitution.ID, Name: "Cross owner",
		Configuration: carddomain.ConfigurationInput{EffectiveFrom: start, EndMonth: &end, NominalDueDay: 10, PaymentMonthOffset: 1},
	}); !errors.Is(err, cardapplication.ErrInstitutionNotFound) {
		t.Fatalf("cross-owner institution error=%v", err)
	}
	if _, err := cardService.Create(ctx, cardOwnerOne, cardapplication.CreateInput{
		InstitutionID: archivedInstitution.ID, Name: "Archived institution",
		Configuration: carddomain.ConfigurationInput{EffectiveFrom: start, EndMonth: &end, NominalDueDay: 10, PaymentMonthOffset: 1},
	}); !errors.Is(err, cardapplication.ErrInstitutionArchived) {
		t.Fatalf("archived institution error=%v", err)
	}

	otherCards, err := cardService.List(ctx, cardOwnerTwo, cardapplication.ListFilters{})
	if err != nil || len(otherCards) != 0 {
		t.Fatalf("cross-owner list=%+v error=%v", otherCards, err)
	}
	if _, err := cardService.Find(ctx, cardOwnerTwo, card.ID); !errors.Is(err, cardapplication.ErrNotFound) {
		t.Fatalf("cross-owner find error=%v", err)
	}

	card, err = cardService.Update(ctx, cardOwnerOne, card.ID, cardapplication.UpdateInput{Name: "Cartão Renomeado"})
	if err != nil || card.Name != "Cartão Renomeado" || card.Institution.ID != institution.ID {
		t.Fatalf("updated card=%+v error=%v", card, err)
	}

	changeMonth, _ := planningdomain.ParseYearMonth("2026-08")
	card, err = cardService.Change(ctx, cardOwnerOne, card.ID, cardapplication.ChangeInput{Configuration: carddomain.ConfigurationInput{
		EffectiveFrom: changeMonth, EndMonth: &end, NominalDueDay: 6, PaymentMonthOffset: 2,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(card.Configurations) != 2 || card.Configurations[0].EndMonth == nil || card.Configurations[0].EndMonth.String() != "2026-07" ||
		card.Configurations[1].StartMonth.String() != "2026-08" || card.Configurations[1].NominalDueDay != 6 || card.Configurations[1].PaymentMonthOffset != 2 {
		t.Fatalf("changed configurations=%+v", card.Configurations)
	}

	card, err = cardService.Archive(ctx, cardOwnerOne, card.ID, cardapplication.ArchiveInput{})
	if err != nil || card.Status != carddomain.StatusArchived || card.ArchivedAt == nil || len(card.Configurations) != 2 {
		t.Fatalf("archived card=%+v error=%v", card, err)
	}
	if _, err := cardService.Update(ctx, cardOwnerOne, card.ID, cardapplication.UpdateInput{Name: "Blocked"}); !errors.Is(err, cardapplication.ErrArchived) {
		t.Fatalf("archived update error=%v", err)
	}
	if _, err := cardService.Change(ctx, cardOwnerOne, card.ID, cardapplication.ChangeInput{Configuration: carddomain.ConfigurationInput{
		EffectiveFrom: changeMonth, EndMonth: &end, NominalDueDay: 7, PaymentMonthOffset: 1,
	}}); !errors.Is(err, cardapplication.ErrArchived) {
		t.Fatalf("archived change error=%v", err)
	}
}
