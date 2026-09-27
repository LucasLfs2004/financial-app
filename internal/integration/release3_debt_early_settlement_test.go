package integration_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	debtrepository "github.com/lucas/financial-api/internal/debt/repository"
	"github.com/lucas/financial-api/internal/profile"
)

func TestRelease3EarlySettlementIsSubstitutiveAndUnique(t *testing.T) {
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
	_, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, ownerID, "r3-settlement-"+ownerID+"@example.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `delete from auth.users where id=$1`, ownerID) })

	service := debtapplication.NewService(debtrepository.NewPostgresRepository(pool), profile.NewRepository(pool))
	created, err := service.Create(ctx, ownerID, debtapplication.CreateInput{
		Name: "Acordo", TotalInstallments: 4, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2090-09"), InstallmentAmountCents: 60000,
	})
	if err != nil {
		t.Fatal(err)
	}
	reason := "Liquidação negociada"
	settled, err := service.Settle(ctx, ownerID, created.Debt.ID, debtapplication.SettlementInput{
		ReferenceMonth: month(t, "2090-10"), AmountCents: 150000, Reason: &reason,
	})
	if err != nil {
		t.Fatal(err)
	}
	if settled.Debt.Settlement == nil || settled.Debt.Settlement.RecordedBy != ownerID ||
		settled.Projection.EffectiveEnd.String() != "2090-10" ||
		settled.Projection.ReleaseFrom.String() != "2090-11" ||
		settled.Projection.ReleasedMonthly.Cents() != 60000 ||
		len(settled.Projection.Occurrences) != 2 ||
		settled.Projection.Occurrences[1].Kind != debtdomain.OccurrenceKindEarlySettlement {
		t.Fatalf("settled=%+v", settled)
	}
	schedule, err := service.Schedule(ctx, ownerID, created.Debt.ID, month(t, "2090-09"), month(t, "2090-12"))
	if err != nil || len(schedule.Occurrences) != 2 || schedule.Occurrences[1].PaymentMethod != "direct" ||
		schedule.Occurrences[1].Occurrence.Amount.Cents() != 150000 {
		t.Fatalf("schedule=%+v error=%v", schedule, err)
	}
	if _, err = service.Settle(ctx, ownerID, created.Debt.ID, debtapplication.SettlementInput{
		ReferenceMonth: month(t, "2090-11"), AmountCents: 120000,
	}); !errors.Is(err, debtapplication.ErrSettlementExists) {
		t.Fatalf("expected settlement conflict, got %v", err)
	}
	found, err := service.Settlement(ctx, ownerID, created.Debt.ID)
	if err != nil || found.ID != settled.Debt.Settlement.ID || found.Reason == nil || *found.Reason != reason {
		t.Fatalf("found=%+v error=%v", found, err)
	}
}
