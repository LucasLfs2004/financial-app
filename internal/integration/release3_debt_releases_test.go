package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	debtrepository "github.com/lucas/financial-api/internal/debt/repository"
	"github.com/lucas/financial-api/internal/profile"
)

func TestRelease3DebtReleasesAggregateWithoutCreatingSources(t *testing.T) {
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
	_, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, ownerID, "r3-releases-"+ownerID+"@example.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `delete from auth.users where id=$1`, ownerID) })

	service := debtapplication.NewService(debtrepository.NewPostgresRepository(pool), profile.NewRepository(pool))
	if _, err = service.Create(ctx, ownerID, debtapplication.CreateInput{
		Name: "Natural", TotalInstallments: 1, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2090-09"), InstallmentAmountCents: 10000,
	}); err != nil {
		t.Fatal(err)
	}
	early, err := service.Create(ctx, ownerID, debtapplication.CreateInput{
		Name: "Antecipada", TotalInstallments: 3, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2090-09"), InstallmentAmountCents: 20000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Settle(ctx, ownerID, early.Debt.ID, debtapplication.SettlementInput{
		ReferenceMonth: month(t, "2090-09"), AmountCents: 45000,
	}); err != nil {
		t.Fatal(err)
	}

	var itemsBefore, incomesBefore int
	if err = pool.QueryRow(ctx, `select count(*), count(*) filter (where kind in ('recurring_income','one_time_income')) from public.financial_items where user_id=$1`, ownerID).Scan(&itemsBefore, &incomesBefore); err != nil {
		t.Fatal(err)
	}
	projection, err := service.Releases(ctx, ownerID, month(t, "2090-10"), month(t, "2090-10"))
	if err != nil {
		t.Fatal(err)
	}
	reasons := map[debtdomain.ReleaseReason]int{}
	for _, release := range projection.Releases {
		reasons[release.Reason]++
	}
	if len(projection.Releases) != 2 || len(projection.MonthlyTotals) != 1 ||
		projection.MonthlyTotals[0].Amount.Cents() != 30000 ||
		reasons[debtdomain.ReleaseReasonScheduledCompletion] != 1 ||
		reasons[debtdomain.ReleaseReasonEarlySettlement] != 1 {
		t.Fatalf("projection=%+v", projection)
	}
	var itemsAfter, incomesAfter int
	if err = pool.QueryRow(ctx, `select count(*), count(*) filter (where kind in ('recurring_income','one_time_income')) from public.financial_items where user_id=$1`, ownerID).Scan(&itemsAfter, &incomesAfter); err != nil {
		t.Fatal(err)
	}
	if itemsAfter != itemsBefore || incomesBefore != 0 || incomesAfter != 0 {
		t.Fatalf("financial items before=%d/%d after=%d/%d", itemsBefore, incomesBefore, itemsAfter, incomesAfter)
	}
}
