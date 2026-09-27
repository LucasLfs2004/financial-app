package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtrepository "github.com/lucas/financial-api/internal/debt/repository"
	paymentapplication "github.com/lucas/financial-api/internal/paymentmethod/application"
	paymentdomain "github.com/lucas/financial-api/internal/paymentmethod/domain"
	paymentrepository "github.com/lucas/financial-api/internal/paymentmethod/repository"
	"github.com/lucas/financial-api/internal/profile"
)

func TestRelease3DebtPaymentMethodChangesPreserveCompetencies(t *testing.T) {
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
	_, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, ownerID, "r3-payment-"+ownerID+"@example.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `delete from auth.users where id=$1`, ownerID) })

	institutionID := randomTestUUID(t)
	cardID := randomTestUUID(t)
	if _, err = pool.Exec(ctx, `insert into public.financial_institutions(id,user_id,name) values($1,$2,'Banco')`, institutionID, ownerID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `insert into public.credit_cards(id,user_id,institution_id,name) values($1,$2,$3,'Principal')`, cardID, ownerID, institutionID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `insert into public.credit_card_periods(credit_card_id,user_id,start_month,end_month,nominal_due_day,payment_month_offset) values($1,$2,'2026-01-01','2026-03-01',10,1)`, cardID, ownerID); err != nil {
		t.Fatal(err)
	}

	debts := debtapplication.NewService(debtrepository.NewPostgresRepository(pool), profile.NewRepository(pool))
	created, err := debts.Create(ctx, ownerID, debtapplication.CreateInput{
		Name: "Parcelamento", TotalInstallments: 3, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2026-01"), InstallmentAmountCents: 10000,
		PaymentMethod: "credit_card", CreditCardID: &cardID,
	})
	if err != nil {
		t.Fatal(err)
	}

	payments := paymentapplication.NewService(paymentrepository.NewPostgresRepository(pool))
	end := month(t, "2026-03")
	if _, err = payments.Create(ctx, ownerID, created.Debt.ID, paymentapplication.Input{
		EffectiveFrom: month(t, "2026-02"), EndMonth: &end, Method: paymentdomain.Direct,
	}); err != nil {
		t.Fatal(err)
	}

	schedule, err := debts.Schedule(ctx, ownerID, created.Debt.ID, month(t, "2026-01"), month(t, "2026-03"))
	if err != nil {
		t.Fatal(err)
	}
	if len(schedule.Occurrences) != 3 ||
		schedule.Occurrences[0].PaymentMethod != "credit_card" ||
		schedule.Occurrences[0].CashMonth.String() != "2026-02" ||
		schedule.Occurrences[1].PaymentMethod != "direct" ||
		schedule.Occurrences[1].CashMonth.String() != "2026-02" {
		t.Fatalf("schedule=%+v", schedule)
	}
}
