package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	invoicerepository "github.com/lucas/financial-api/internal/cardinvoice/repository"
	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/monthlysummary"
	"github.com/lucas/financial-api/internal/paymentmethod/application"
	paymentdomain "github.com/lucas/financial-api/internal/paymentmethod/domain"
	paymentrepository "github.com/lucas/financial-api/internal/paymentmethod/repository"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/savings"
)

func TestRelease21RecordsOutlivePlanHorizon(t *testing.T) {
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
	_, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, ownerID, "r21-"+ownerID+"@example.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `delete from auth.users where id=$1`, ownerID) })

	plans := planning.NewService(planning.NewPostgresRepository(pool))
	items := financialitem.NewService(financialitem.NewPostgresRepository(pool), plans)
	savingRepository := savings.NewPostgresRepository(pool)
	savingService := savings.NewService(savingRepository, plans)
	invoiceRepository := invoicerepository.NewPostgresRepository(pool)
	summaries := monthlysummary.NewService(plans, financialitem.NewPostgresRepository(pool), savingRepository, invoiceRepository)
	paymentMethods := application.NewService(paymentrepository.NewPostgresRepository(pool))

	firstPlan, err := plans.Create(ctx, ownerID, planning.CreateInput{
		Name: "Setembro a dezembro", StartMonth: month(t, "2026-09"),
		EndMonth: month(t, "2026-12"), CurrencyCode: "BRL",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = items.Create(ctx, ownerID, financialitem.CreateInput{
		Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome,
		Period: financialitem.PeriodInput{StartMonth: month(t, "2026-09"), EndMonth: monthPtr(t, "2027-09"), AmountCents: 500000, Recurrence: domain.RecurrenceMonthly},
	})
	if err != nil {
		t.Fatal(err)
	}
	insurance, err := items.Create(ctx, ownerID, financialitem.CreateInput{
		Name: "Seguro", Kind: domain.FinancialItemKindFixedExpense,
		Period: financialitem.PeriodInput{StartMonth: month(t, "2026-09"), EndMonth: monthPtr(t, "2027-09"), AmountCents: 18000, Recurrence: domain.RecurrenceMonthly},
	})
	if err != nil {
		t.Fatal(err)
	}
	if insurance.CurrencyCode != "BRL" || insurance.Periods[0].EndMonth == nil || insurance.Periods[0].EndMonth.String() != "2027-09" {
		t.Fatalf("insurance=%+v", insurance)
	}
	_, err = paymentMethods.Create(ctx, ownerID, insurance.ID, application.Input{
		EffectiveFrom: month(t, "2026-09"), EndMonth: monthPtr(t, "2027-09"), Method: paymentdomain.Direct,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = savingService.Put(ctx, ownerID, savings.PutInput{EffectiveFrom: month(t, "2026-09"), EndMonth: monthPtr(t, "2026-12"), AmountCents: 0})
	if err != nil {
		t.Fatal(err)
	}

	december, err := summaries.Get(ctx, ownerID, month(t, "2026-12"), domain.SummaryBasisReference)
	if err != nil || december.IncomeCents != 500000 || december.CommitmentsCents != 18000 || december.ResultCents != 482000 {
		t.Fatalf("december=%+v error=%v", december, err)
	}
	activation, err := plans.Activate(ctx, ownerID)
	if err != nil || activation.Original.SchemaVersion != 4 {
		t.Fatalf("activation=%+v error=%v", activation, err)
	}
	var snapshot struct {
		Items []struct {
			ID      string `json:"id"`
			Periods []struct {
				EndMonth *string `json:"end_month"`
			} `json:"periods"`
		} `json:"items"`
	}
	if err = json.Unmarshal(activation.Original.Plan, &snapshot); err != nil {
		t.Fatal(err)
	}
	foundFullInsurancePeriod := false
	for _, item := range snapshot.Items {
		if item.ID == insurance.ID && len(item.Periods) == 1 && item.Periods[0].EndMonth != nil && *item.Periods[0].EndMonth == "2027-09" {
			foundFullInsurancePeriod = true
		}
	}
	if !foundFullInsurancePeriod {
		t.Fatalf("snapshot did not preserve full insurance period: %+v", snapshot.Items)
	}

	_, err = pool.Exec(ctx, `update public.plans set status='archived', archived_at=now() where id=$1 and user_id=$2`, firstPlan.ID, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = plans.Create(ctx, ownerID, planning.CreateInput{
		Name: "Planejamento 2027", StartMonth: month(t, "2027-01"),
		EndMonth: month(t, "2027-09"), CurrencyCode: "BRL",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = savingService.Put(ctx, ownerID, savings.PutInput{EffectiveFrom: month(t, "2027-01"), EndMonth: monthPtr(t, "2027-09"), AmountCents: 0})
	if err != nil {
		t.Fatal(err)
	}
	september, err := summaries.Get(ctx, ownerID, month(t, "2027-09"), domain.SummaryBasisReference)
	if err != nil || september.CommitmentsCents != 18000 || september.ResultCents != 482000 {
		t.Fatalf("future summary=%+v error=%v", september, err)
	}
	periods, err := paymentMethods.List(ctx, ownerID, insurance.ID)
	if err != nil || len(periods) != 1 || periods[0].EndMonth == nil || periods[0].EndMonth.String() != "2027-09" {
		t.Fatalf("payment periods=%+v error=%v", periods, err)
	}
}
