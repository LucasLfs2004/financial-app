package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/cardinvoice"
	invoicerepository "github.com/lucas/financial-api/internal/cardinvoice/repository"
	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
	cardrepository "github.com/lucas/financial-api/internal/creditcard/repository"
	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	debtrepository "github.com/lucas/financial-api/internal/debt/repository"
	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutionrepository "github.com/lucas/financial-api/internal/financialinstitution/repository"
	"github.com/lucas/financial-api/internal/financialitem"
	allocationapplication "github.com/lucas/financial-api/internal/invoiceallocation/application"
	allocationrepository "github.com/lucas/financial-api/internal/invoiceallocation/repository"
	"github.com/lucas/financial-api/internal/monthlysummary"
	"github.com/lucas/financial-api/internal/planning"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/profile"
	"github.com/lucas/financial-api/internal/savings"
)

func TestRelease3DebtsReconcileMonthlySummaryByBasis(t *testing.T) {
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
	_, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, ownerID, "r3-summary-"+ownerID+"@example.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `delete from auth.users where id=$1`, ownerID) })

	planRepository := planning.NewPostgresRepository(pool)
	plans := planning.NewService(planRepository)
	if _, err = plans.Create(ctx, ownerID, planning.CreateInput{
		Name: "Planejamento", StartMonth: month(t, "2090-09"),
		EndMonth: month(t, "2090-12"), CurrencyCode: "BRL",
	}); err != nil {
		t.Fatal(err)
	}
	itemRepository := financialitem.NewPostgresRepository(pool)
	items := financialitem.NewService(itemRepository, plans)
	if _, err = items.Create(ctx, ownerID, financialitem.CreateInput{
		Name: "Salário", Kind: planningdomain.FinancialItemKindRecurringIncome,
		Period: financialitem.PeriodInput{
			StartMonth: month(t, "2090-09"), EndMonth: monthPtr(t, "2090-12"),
			AmountCents: 500000, Recurrence: planningdomain.RecurrenceMonthly,
		},
	}); err != nil {
		t.Fatal(err)
	}

	institutions := institutionapplication.NewService(institutionrepository.NewPostgresRepository(pool))
	cards := cardapplication.NewService(cardrepository.NewPostgresRepository(pool))
	institution, err := institutions.Create(ctx, ownerID, institutionapplication.CreateInput{Name: "Banco"})
	if err != nil {
		t.Fatal(err)
	}
	createCard := func(name string) string {
		card, createErr := cards.Create(ctx, ownerID, cardapplication.CreateInput{
			InstitutionID: institution.ID, Name: name,
			Configuration: carddomain.ConfigurationInput{
				EffectiveFrom: month(t, "2090-09"), NominalDueDay: 6, PaymentMonthOffset: 1,
			},
		})
		if createErr != nil {
			t.Fatal(createErr)
		}
		return card.ID
	}
	cardA := createCard("Principal")
	cardB := createCard("Reserva")

	debts := debtapplication.NewService(debtrepository.NewPostgresRepository(pool), profile.NewRepository(pool))
	direct, err := debts.Create(ctx, ownerID, debtapplication.CreateInput{
		Name: "Direta", TotalInstallments: 3, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2090-09"), InstallmentAmountCents: 10000, CashMonthOffset: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = debts.Settle(ctx, ownerID, direct.Debt.ID, debtapplication.SettlementInput{
		ReferenceMonth: month(t, "2090-10"), AmountCents: 25000,
	}); err != nil {
		t.Fatal(err)
	}
	cardDebt, err := debts.Create(ctx, ownerID, debtapplication.CreateInput{
		Name: "Cartão", TotalInstallments: 3, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2090-09"), InstallmentAmountCents: 20000,
		PaymentMethod: "credit_card", CreditCardID: &cardA,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = debts.Settle(ctx, ownerID, cardDebt.Debt.ID, debtapplication.SettlementInput{
		ReferenceMonth: month(t, "2090-10"), AmountCents: 45000,
	}); err != nil {
		t.Fatal(err)
	}
	moves := allocationapplication.NewService(allocationrepository.NewPostgresRepository(pool), plans)
	if _, err = moves.Move(ctx, ownerID, cardDebt.Debt.ID, month(t, "2090-10"), cardB, month(t, "2090-12"), nil); err != nil {
		t.Fatal(err)
	}

	summaries := monthlysummary.NewService(
		plans, itemRepository, savings.NewPostgresRepository(pool),
		invoicerepository.NewPostgresRepository(pool),
	)
	reference, err := summaries.Get(ctx, ownerID, month(t, "2090-10"), planningdomain.SummaryBasisReference)
	if err != nil || reference.IncomeCents != 500000 || reference.CommitmentsCents != 70000 ||
		reference.Breakdown.DebtInstallmentsCents != 70000 || len(reference.Sources) != 3 {
		t.Fatalf("reference=%+v error=%v", reference, err)
	}
	debtSources := 0
	for _, source := range reference.Sources {
		if source.Kind != string(planningdomain.FinancialItemKindDebtInstallment) {
			continue
		}
		debtSources++
		if source.DebtOccurrenceKind == nil || *source.DebtOccurrenceKind != debtdomain.OccurrenceKindEarlySettlement {
			t.Fatalf("reference debt source=%+v", source)
		}
		if source.DebtID != nil && *source.DebtID == cardDebt.Debt.ID &&
			(source.CreditCardID == nil || *source.CreditCardID != cardB || source.InvoiceAllocation == nil ||
				*source.InvoiceAllocation != cardinvoice.AllocationMovedByUser) {
			t.Fatalf("moved reference source=%+v", source)
		}
	}
	if debtSources != 2 {
		t.Fatalf("reference debt sources=%d", debtSources)
	}

	directCash, err := summaries.Get(ctx, ownerID, month(t, "2090-11"), planningdomain.SummaryBasisCash)
	if err != nil || directCash.Breakdown.DebtInstallmentsCents != 25000 || directCash.CommitmentsCents != 25000 ||
		len(directCash.Sources) != 2 {
		t.Fatalf("direct cash=%+v error=%v", directCash, err)
	}
	cardCash, err := summaries.Get(ctx, ownerID, month(t, "2090-12"), planningdomain.SummaryBasisCash)
	if err != nil || cardCash.Breakdown.DebtInstallmentsCents != 45000 || cardCash.CommitmentsCents != 45000 ||
		len(cardCash.Sources) != 2 {
		t.Fatalf("card cash=%+v error=%v", cardCash, err)
	}

	afterRelease, err := summaries.Get(ctx, ownerID, month(t, "2090-11"), planningdomain.SummaryBasisReference)
	if err != nil || afterRelease.IncomeCents != 500000 || afterRelease.Breakdown.DebtInstallmentsCents != 0 ||
		afterRelease.CommitmentsCents != 0 || len(afterRelease.Sources) != 1 {
		t.Fatalf("after release=%+v error=%v", afterRelease, err)
	}
}
