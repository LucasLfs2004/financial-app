package integration_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/savings"
)

const (
	ownerOne = "70000000-0000-0000-0000-000000000007"
	ownerTwo = "80000000-0000-0000-0000-000000000008"
)

func TestMilestonesCAndD(t *testing.T) {
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
	_, _ = pool.Exec(ctx, `delete from auth.users where id in ($1,$2)`, ownerOne, ownerTwo)
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `delete from auth.users where id in ($1,$2)`, ownerOne, ownerTwo) })
	for id, email := range map[string]string{ownerOne: "milestone-owner-1@example.com", ownerTwo: "milestone-owner-2@example.com"} {
		_, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, id, email)
		if err != nil {
			t.Fatal(err)
		}
	}
	planRepository := planning.NewPostgresRepository(pool)
	plans := planning.NewService(planRepository)
	itemRepository := financialitem.NewPostgresRepository(pool)
	items := financialitem.NewService(itemRepository, plans)
	savingRepository := savings.NewPostgresRepository(pool)
	savingService := savings.NewService(savingRepository, plans)
	for _, owner := range []string{ownerOne, ownerTwo} {
		_, err = plans.Create(ctx, owner, planning.CreateInput{Name: "Planejamento", StartMonth: month(t, "2026-01"), EndMonth: month(t, "2026-12"), CurrencyCode: "BRL"})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = plans.Activate(ctx, ownerTwo); !errors.Is(err, planning.ErrNotActivatable) {
		t.Fatalf("activation without premises error=%v", err)
	}

	configuration, err := savingService.Get(ctx, ownerOne)
	if err != nil || configuration.Configured {
		t.Fatalf("expected absent savings, got %+v %v", configuration, err)
	}
	salary, err := items.Create(ctx, ownerOne, financialitem.CreateInput{Name: "Salário", Kind: domain.FinancialItemKindRecurringIncome, Period: financialitem.PeriodInput{StartMonth: month(t, "2026-01"), EndMonth: monthPtr(t, "2026-12"), AmountCents: 600000, Recurrence: domain.RecurrenceMonthly, CashMonthOffset: 1}})
	if err != nil {
		t.Fatal(err)
	}
	// Commitments may exceed income. The negative monthly total is calculated in T13.
	_, err = items.Create(ctx, ownerOne, financialitem.CreateInput{Name: "Aluguel", Kind: domain.FinancialItemKindFixedExpense, Period: financialitem.PeriodInput{StartMonth: month(t, "2026-01"), AmountCents: 700000, Recurrence: domain.RecurrenceMonthly}})
	if err != nil {
		t.Fatal(err)
	}
	variable, err := items.Create(ctx, ownerOne, financialitem.CreateInput{Name: "Gasolina", Kind: domain.FinancialItemKindProjectedVariableExpense, Period: financialitem.PeriodInput{StartMonth: month(t, "2026-01"), AmountCents: 70000, Recurrence: domain.RecurrenceMonthly}})
	if err != nil {
		t.Fatal(err)
	}
	salary, err = items.Change(ctx, ownerOne, salary.ID, financialitem.ChangeInput{EffectiveFrom: month(t, "2026-08"), EndMonth: monthPtr(t, "2026-12"), AmountCents: 650000, CashMonthOffset: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(salary.Periods) != 2 || salary.Periods[0].EndMonth == nil || salary.Periods[0].EndMonth.String() != "2026-07" || salary.Periods[1].StartMonth.String() != "2026-08" {
		t.Fatalf("history was not preserved: %+v", salary.Periods)
	}
	variable, err = items.Archive(ctx, ownerOne, variable.ID, financialitem.ArchiveInput{EffectiveFrom: monthPtr(t, "2026-07")})
	if err != nil {
		t.Fatal(err)
	}
	if variable.Status != domain.FinancialItemStatusArchived || variable.Periods[0].EndMonth == nil || variable.Periods[0].EndMonth.String() != "2026-06" {
		t.Fatalf("archive did not close history: %+v", variable)
	}
	configuration, err = savingService.Put(ctx, ownerOne, savings.PutInput{EffectiveFrom: month(t, "2026-01"), EndMonth: monthPtr(t, "2026-06"), AmountCents: 0})
	if err != nil || !configuration.Configured || configuration.Periods[0].AmountCents != 0 {
		t.Fatalf("explicit zero failed: %+v %v", configuration, err)
	}
	configuration, err = savingService.Put(ctx, ownerOne, savings.PutInput{EffectiveFrom: month(t, "2026-07"), EndMonth: monthPtr(t, "2026-12"), AmountCents: 120000})
	if err != nil || len(configuration.Periods) != 2 {
		t.Fatalf("future savings failed: %+v %v", configuration, err)
	}
	if _, err = items.Find(ctx, ownerTwo, salary.ID); !errors.Is(err, financialitem.ErrNotFound) {
		t.Fatalf("cross-owner lookup error=%v", err)
	}

	results := make([]planning.Activation, 2)
	errs := make([]error, 2)
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := range 2 {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			results[index], errs[index] = plans.Activate(ctx, ownerOne)
		}(i)
	}
	close(start)
	group.Wait()
	for _, activationErr := range errs {
		if activationErr != nil {
			t.Fatalf("concurrent activation: %v", activationErr)
		}
	}
	if results[0].Original.ID != results[1].Original.ID || results[0].Plan.Status != domain.PlanStatusActive || results[1].Plan.Status != domain.PlanStatusActive {
		t.Fatalf("activation is not idempotent: %+v", results)
	}
	var snapshots int
	if err = pool.QueryRow(ctx, `select count(*) from public.plan_snapshots where user_id=$1`, ownerOne).Scan(&snapshots); err != nil || snapshots != 1 {
		t.Fatalf("snapshot count=%d err=%v", snapshots, err)
	}
	original, err := plans.Original(ctx, ownerOne)
	if err != nil || original.ID != results[0].Original.ID || len(original.Plan) == 0 {
		t.Fatalf("original lookup failed: %+v %v", original, err)
	}
	_, err = items.Change(ctx, ownerOne, salary.ID, financialitem.ChangeInput{EffectiveFrom: month(t, "2026-10"), EndMonth: monthPtr(t, "2026-12"), AmountCents: 680000, CashMonthOffset: 1})
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := plans.Original(ctx, ownerOne)
	if err != nil || !bytes.Equal(original.Plan, unchanged.Plan) {
		t.Fatalf("original snapshot changed after premise update: %v", err)
	}
}

func month(t *testing.T, value string) domain.YearMonth {
	t.Helper()
	result, err := domain.ParseYearMonth(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func monthPtr(t *testing.T, value string) *domain.YearMonth {
	t.Helper()
	result := month(t, value)
	return &result
}
