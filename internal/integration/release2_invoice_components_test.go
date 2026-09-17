package integration_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/cardinvoice"
	invoiceapplication "github.com/lucas/financial-api/internal/cardinvoice/application"
	invoicerepository "github.com/lucas/financial-api/internal/cardinvoice/repository"
	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
	cardrepository "github.com/lucas/financial-api/internal/creditcard/repository"
	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutionrepository "github.com/lucas/financial-api/internal/financialinstitution/repository"
	"github.com/lucas/financial-api/internal/financialitem"
	adjustmentapplication "github.com/lucas/financial-api/internal/invoiceadjustment/application"
	adjustmentrepository "github.com/lucas/financial-api/internal/invoiceadjustment/repository"
	paymentapplication "github.com/lucas/financial-api/internal/paymentmethod/application"
	paymentdomain "github.com/lucas/financial-api/internal/paymentmethod/domain"
	paymentrepository "github.com/lucas/financial-api/internal/paymentmethod/repository"
	"github.com/lucas/financial-api/internal/planning"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

const (
	invoiceOwnerOne = "93000000-0000-0000-0000-000000000003"
	invoiceOwnerTwo = "94000000-0000-0000-0000-000000000004"
)

func TestRelease2InvoiceComponentSelection(t *testing.T) {
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
	if err := cleanupInvoiceOwners(ctx, pool); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cleanupInvoiceOwners(context.Background(), pool)
	})
	for id, email := range map[string]string{invoiceOwnerOne: "invoice-owner-3@example.com", invoiceOwnerTwo: "invoice-owner-4@example.com"} {
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

	planRepository := planning.NewPostgresRepository(pool)
	plans := planning.NewService(planRepository)
	for _, owner := range []string{invoiceOwnerOne, invoiceOwnerTwo} {
		if _, err = plans.Create(ctx, owner, planning.CreateInput{
			Name: "Planejamento", StartMonth: month(t, "2026-01"), EndMonth: month(t, "2026-12"), CurrencyCode: "BRL",
		}); err != nil {
			t.Fatal(err)
		}
	}

	institutions := institutionapplication.NewService(institutionrepository.NewPostgresRepository(pool))
	cards := cardapplication.NewService(cardrepository.NewPostgresRepository(pool))
	institution, err := institutions.Create(ctx, invoiceOwnerOne, institutionapplication.CreateInput{Name: "Banco da Fatura"})
	if err != nil {
		t.Fatal(err)
	}
	card, err := cards.Create(ctx, invoiceOwnerOne, cardapplication.CreateInput{
		InstitutionID: institution.ID,
		Name:          "Principal",
		Configuration: carddomain.ConfigurationInput{
			EffectiveFrom: month(t, "2026-01"), NominalDueDay: 6, PaymentMonthOffset: 1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	items := financialitem.NewService(financialitem.NewPostgresRepository(pool), plans)
	item, err := items.Create(ctx, invoiceOwnerOne, financialitem.CreateInput{
		Name: "Gasolina", Kind: planningdomain.FinancialItemKindProjectedVariableExpense,
		Period: financialitem.PeriodInput{
			StartMonth: month(t, "2026-01"), EndMonth: monthPtr(t, "2026-12"),
			AmountCents: 1000, Recurrence: planningdomain.RecurrenceMonthly,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	paymentMethods := paymentapplication.NewService(paymentrepository.NewPostgresRepository(pool), plans)
	if _, err = paymentMethods.Create(ctx, invoiceOwnerOne, item.ID, paymentapplication.Input{
		EffectiveFrom: month(t, "2026-11"), EndMonth: monthPtr(t, "2026-12"),
		Method: paymentdomain.CreditCard, CreditCardID: &card.ID,
	}); err != nil {
		t.Fatal(err)
	}

	adjustments := adjustmentapplication.NewService(adjustmentrepository.NewPostgresRepository(pool), plans)
	_, err = adjustments.Create(ctx, invoiceOwnerOne, adjustmentapplication.Input{
		Name: "Ajuste sem referência", AmountCents: 200, CardID: card.ID, PaymentMonth: month(t, "2026-12"),
	})
	if err != nil {
		t.Fatal(err)
	}

	invoices := invoiceapplication.NewService(invoicerepository.NewPostgresRepository(pool), plans)
	invoice, err := invoices.Project(ctx, invoiceOwnerOne, card.ID, month(t, "2026-12"))
	if err != nil {
		t.Fatal(err)
	}
	if invoice.Total.Cents() != 1200 || len(invoice.Components) != 2 || invoice.Components[1].ReferenceKnown {
		t.Fatalf("initial invoice=%+v", invoice)
	}

	_, err = invoices.Project(ctx, invoiceOwnerTwo, card.ID, month(t, "2026-12"))
	if !errors.Is(err, cardinvoice.ErrCardNotFound) {
		t.Fatalf("cross-owner projection error=%v", err)
	}
}

func cleanupInvoiceOwners(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `delete from public.card_invoice_audit_events where user_id in ($1, $2)`, invoiceOwnerOne, invoiceOwnerTwo); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `delete from auth.users where id in ($1, $2)`, invoiceOwnerOne, invoiceOwnerTwo); err != nil {
		return err
	}
	return nil
}
