package integration_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtrepository "github.com/lucas/financial-api/internal/debt/repository"
	"github.com/lucas/financial-api/internal/profile"
)

func TestRelease3DebtRegistrationLifecycleAndOwnership(t *testing.T) {
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

	ownerID := randomTestUUID(t)
	otherID := randomTestUUID(t)
	for _, user := range []struct{ id, email string }{{ownerID, "r3-owner-" + ownerID + "@example.com"}, {otherID, "r3-other-" + otherID + "@example.com"}} {
		_, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, user.id, user.email)
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `delete from auth.users where id in ($1, $2)`, ownerID, otherID)
	})

	service := debtapplication.NewService(debtrepository.NewPostgresRepository(pool), profile.NewRepository(pool))
	original := int64(720000)
	created, err := service.Create(ctx, ownerID, debtapplication.CreateInput{
		Name: "Transplante", OriginalTotalCents: &original,
		TotalInstallments: 12, FirstProjectedInstallment: 5,
		ScheduledStart: month(t, "2026-09"), InstallmentAmountCents: 60000, CashMonthOffset: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Debt.CurrencyCode != "BRL" || created.Projection.ScheduledEnd.String() != "2027-04" ||
		len(created.Debt.Periods) != 1 {
		t.Fatalf("created=%+v", created)
	}

	listed, err := service.List(ctx, ownerID, debtapplication.Filters{AsOf: monthPtr(t, "2026-09")})
	if err != nil || len(listed) != 1 || listed[0].Projection.RemainingInstallments != 8 {
		t.Fatalf("listed=%+v error=%v", listed, err)
	}
	if _, err = service.Find(ctx, otherID, created.Debt.ID, monthPtr(t, "2026-09")); !errors.Is(err, debtapplication.ErrNotFound) {
		t.Fatalf("expected ownership-safe not found, got %v", err)
	}
	for _, change := range []struct {
		month  string
		amount int64
	}{
		{month: "2026-09", amount: 61000},
		{month: "2027-01", amount: 65000},
		{month: "2027-04", amount: 70000},
	} {
		changed, changeErr := service.Change(ctx, ownerID, created.Debt.ID, debtapplication.ChangeInput{
			EffectiveFrom: month(t, change.month), InstallmentAmountCents: change.amount,
		})
		if changeErr != nil {
			t.Fatalf("change %s: %v", change.month, changeErr)
		}
		created = changed
	}
	if len(created.Debt.Periods) != 3 || len(created.Projection.Occurrences) != 8 ||
		created.Projection.Occurrences[0].Amount.Cents() != 61000 ||
		created.Projection.Occurrences[4].Amount.Cents() != 65000 ||
		created.Projection.Occurrences[7].Amount.Cents() != 70000 ||
		created.Projection.Occurrences[7].InstallmentNumber != 12 ||
		created.Projection.ScheduledEnd.String() != "2027-04" {
		t.Fatalf("changed schedule=%+v periods=%+v", created.Projection, created.Debt.Periods)
	}
	if _, err = service.Change(ctx, ownerID, created.Debt.ID, debtapplication.ChangeInput{
		EffectiveFrom: month(t, "2027-05"), InstallmentAmountCents: 1,
	}); !errors.Is(err, debtapplication.ErrChangeOutsideSchedule) {
		t.Fatalf("expected outside schedule error, got %v", err)
	}
	schedule, err := service.Schedule(ctx, ownerID, created.Debt.ID, month(t, "2026-12"), month(t, "2027-01"))
	if err != nil || len(schedule.Occurrences) != 2 ||
		schedule.Occurrences[0].Occurrence.InstallmentNumber != 8 ||
		schedule.Occurrences[0].Occurrence.Amount.Cents() != 61000 ||
		schedule.Occurrences[0].CashMonth.String() != "2027-01" ||
		schedule.Occurrences[1].Occurrence.InstallmentNumber != 9 ||
		schedule.Occurrences[1].Occurrence.Amount.Cents() != 65000 ||
		schedule.Occurrences[1].CashMonth.String() != "2027-02" {
		t.Fatalf("schedule=%+v error=%v", schedule, err)
	}

	name := "Transplante atualizado"
	updated, err := service.Update(ctx, ownerID, created.Debt.ID, debtapplication.UpdateInput{Name: &name})
	if err != nil || updated.Debt.Name != name {
		t.Fatalf("updated=%+v error=%v", updated, err)
	}
	archived, err := service.Archive(ctx, ownerID, created.Debt.ID, debtapplication.ArchiveInput{})
	if err != nil || archived.Projection.ProjectionStatus != "archived" {
		t.Fatalf("archived=%+v error=%v", archived, err)
	}
	if _, err = service.Update(ctx, ownerID, created.Debt.ID, debtapplication.UpdateInput{Name: &name}); !errors.Is(err, debtapplication.ErrArchived) {
		t.Fatalf("expected archived error, got %v", err)
	}
}
