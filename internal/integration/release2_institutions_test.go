package integration_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
	institutionrepository "github.com/lucas/financial-api/internal/financialinstitution/repository"
)

const (
	institutionOwnerOne = "c1000000-0000-0000-0000-000000000001"
	institutionOwnerTwo = "c2000000-0000-0000-0000-000000000002"
)

func TestRelease2FinancialInstitutionLifecycle(t *testing.T) {
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

	_, _ = pool.Exec(ctx, `delete from auth.users where id in ($1, $2)`, institutionOwnerOne, institutionOwnerTwo)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `delete from auth.users where id in ($1, $2)`, institutionOwnerOne, institutionOwnerTwo)
	})
	for id, email := range map[string]string{
		institutionOwnerOne: "institution-owner-1@example.com",
		institutionOwnerTwo: "institution-owner-2@example.com",
	} {
		_, err = pool.Exec(ctx, `
			insert into auth.users (
				id, instance_id, aud, role, email, encrypted_password,
				email_confirmed_at, raw_app_meta_data, raw_user_meta_data,
				created_at, updated_at
			)
			values (
				$1, '00000000-0000-0000-0000-000000000000',
				'authenticated', 'authenticated', $2, '', now(),
				'{"provider":"email","providers":["email"]}', '{}', now(), now()
			)
		`, id, email)
		if err != nil {
			t.Fatal(err)
		}
	}

	repository := institutionrepository.NewPostgresRepository(pool)
	service := institutionapplication.NewService(repository)
	institution, err := service.Create(ctx, institutionOwnerOne, institutionapplication.CreateInput{Name: "  Banco Alfa  "})
	if err != nil || institution.Name != "Banco Alfa" || institution.Status != institutiondomain.StatusActive {
		t.Fatalf("institution=%+v error=%v", institution, err)
	}

	if _, err := service.Create(ctx, institutionOwnerOne, institutionapplication.CreateInput{Name: " banco ALFA "}); !errors.Is(err, institutionapplication.ErrAlreadyExists) {
		t.Fatalf("duplicate error=%v", err)
	}
	otherInstitutions, err := service.List(ctx, institutionOwnerTwo, institutionapplication.ListFilters{})
	if err != nil || len(otherInstitutions) != 0 {
		t.Fatalf("cross-owner list=%+v error=%v", otherInstitutions, err)
	}
	if _, err := service.Update(ctx, institutionOwnerTwo, institution.ID, institutionapplication.UpdateInput{Name: "Tentativa cruzada"}); !errors.Is(err, institutionapplication.ErrNotFound) {
		t.Fatalf("cross-owner update error=%v", err)
	}

	institution, err = service.Update(ctx, institutionOwnerOne, institution.ID, institutionapplication.UpdateInput{Name: "Banco Principal"})
	if err != nil || institution.Name != "Banco Principal" {
		t.Fatalf("institution=%+v error=%v", institution, err)
	}

	var cardID string
	err = pool.QueryRow(ctx, `
		insert into public.credit_cards (user_id, institution_id, name)
		values ($1, $2, 'Cartão principal')
		returning id
	`, institutionOwnerOne, institution.ID).Scan(&cardID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Archive(ctx, institutionOwnerOne, institution.ID, institutionapplication.ArchiveInput{}); !errors.Is(err, institutionapplication.ErrHasActiveCards) {
		t.Fatalf("archive with active card error=%v", err)
	}
	if _, err := pool.Exec(ctx, `update public.credit_cards set status = 'archived', archived_at = now() where id = $1`, cardID); err != nil {
		t.Fatal(err)
	}

	institution, err = service.Archive(ctx, institutionOwnerOne, institution.ID, institutionapplication.ArchiveInput{})
	if err != nil || institution.Status != institutiondomain.StatusArchived || institution.ArchivedAt == nil {
		t.Fatalf("institution=%+v error=%v", institution, err)
	}
	active := institutiondomain.StatusActive
	activeInstitutions, err := service.List(ctx, institutionOwnerOne, institutionapplication.ListFilters{Status: &active})
	if err != nil || len(activeInstitutions) != 0 {
		t.Fatalf("active institutions=%+v error=%v", activeInstitutions, err)
	}
}
