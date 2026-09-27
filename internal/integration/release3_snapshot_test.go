package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtrepository "github.com/lucas/financial-api/internal/debt/repository"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/profile"
)

func TestRelease3SnapshotV4CapturesReachableDebtAndKeepsOriginal(t *testing.T) {
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
	for _, id := range []string{ownerID, otherID} {
		if _, err := pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, id, "r3-snapshot-"+id+"@example.com"); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `delete from auth.users where id in ($1,$2)`, ownerID, otherID)
	})

	plans := planning.NewPostgresRepository(pool)
	if _, err := plans.Create(ctx, ownerID, planning.CreateInput{Name: "R3", StartMonth: month(t, "2090-09"), EndMonth: month(t, "2090-12"), CurrencyCode: "BRL"}); err != nil {
		t.Fatal(err)
	}
	var planID, incomeID string
	if err := pool.QueryRow(ctx, `select id from public.plans where user_id=$1`, ownerID).Scan(&planID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `insert into public.financial_items(user_id,currency_code,name,kind) values($1,'BRL','Salário','recurring_income') returning id`, ownerID).Scan(&incomeID); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{
		`insert into public.financial_item_periods(financial_item_id,user_id,start_month,end_month,amount_cents,recurrence) values($1,$2,'2090-09-01','2090-12-01',300000,'monthly')`,
		`insert into public.saving_periods(plan_id,user_id,start_month,end_month,amount_cents) values($1,$2,'2090-09-01','2090-12-01',0)`,
	} {
		first := incomeID
		if bytes.Contains([]byte(query), []byte("saving_periods")) {
			first = planID
		}
		if _, err := pool.Exec(ctx, query, first, ownerID); err != nil {
			t.Fatal(err)
		}
	}

	var institutionID, cardID string
	if err := pool.QueryRow(ctx, `insert into public.financial_institutions(user_id,name) values($1,'Banco') returning id`, ownerID).Scan(&institutionID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `insert into public.credit_cards(user_id,institution_id,name) values($1,$2,'Cartão') returning id`, ownerID, institutionID).Scan(&cardID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `insert into public.credit_card_periods(credit_card_id,user_id,start_month,end_month,nominal_due_day,payment_month_offset) values($1,$2,'2090-09-01',null,6,1)`, cardID, ownerID); err != nil {
		t.Fatal(err)
	}

	debts := debtapplication.NewService(debtrepository.NewPostgresRepository(pool), profile.NewRepository(pool))
	created, err := debts.Create(ctx, ownerID, debtapplication.CreateInput{
		Name: "Parcelamento", TotalInstallments: 4, FirstProjectedInstallment: 1,
		ScheduledStart: month(t, "2090-09"), InstallmentAmountCents: 60000,
		PaymentMethod: "credit_card", CreditCardID: &cardID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := debts.Change(ctx, ownerID, created.Debt.ID, debtapplication.ChangeInput{EffectiveFrom: month(t, "2090-10"), InstallmentAmountCents: 65000}); err != nil {
		t.Fatal(err)
	}
	settled, err := debts.Settle(ctx, ownerID, created.Debt.ID, debtapplication.SettlementInput{ReferenceMonth: month(t, "2090-11"), AmountCents: 120000})
	if err != nil {
		t.Fatal(err)
	}
	outside, err := debts.Create(ctx, ownerID, debtapplication.CreateInput{Name: "Fora da janela", TotalInstallments: 2, FirstProjectedInstallment: 1, ScheduledStart: month(t, "2091-05"), InstallmentAmountCents: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := debts.Create(ctx, otherID, debtapplication.CreateInput{Name: "Outro usuário", TotalInstallments: 2, FirstProjectedInstallment: 1, ScheduledStart: month(t, "2090-09"), InstallmentAmountCents: 1000}); err != nil {
		t.Fatal(err)
	}

	first, err := plans.Activate(ctx, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Original.SchemaVersion != 4 {
		t.Fatalf("schema version=%d", first.Original.SchemaVersion)
	}
	var document struct {
		Release2 struct {
			Cards []struct {
				ID string `json:"id"`
			} `json:"credit_cards"`
		} `json:"release_2"`
		Release3 struct {
			Debts []struct {
				ID      string `json:"financial_item_id"`
				Periods []struct {
					StartMonth string `json:"start_month"`
					Amount     int64  `json:"amount_cents"`
				} `json:"periods"`
				PaymentMethods []json.RawMessage `json:"payment_methods"`
				Settlement     *struct {
					ID string `json:"id"`
				} `json:"early_settlement"`
			} `json:"debts"`
		} `json:"release_3"`
	}
	if err := json.Unmarshal(first.Original.Plan, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Release3.Debts) != 1 || document.Release3.Debts[0].ID != created.Debt.ID ||
		len(document.Release3.Debts[0].Periods) != 2 || len(document.Release3.Debts[0].PaymentMethods) != 1 ||
		document.Release3.Debts[0].Settlement == nil || document.Release3.Debts[0].Settlement.ID != settled.Debt.Settlement.ID ||
		len(document.Release2.Cards) != 1 || document.Release2.Cards[0].ID != cardID {
		t.Fatalf("snapshot debt/resources=%+v outside=%s", document, outside.Debt.ID)
	}
	if document.Release3.Debts[0].Periods[0].StartMonth != "2090-09-01" ||
		document.Release3.Debts[0].Periods[0].Amount != 60000 ||
		document.Release3.Debts[0].Periods[1].StartMonth != "2090-10-01" ||
		document.Release3.Debts[0].Periods[1].Amount != 65000 {
		t.Fatalf("debt periods are not ordered: %+v", document.Release3.Debts[0].Periods)
	}
	second, err := plans.Activate(ctx, ownerID)
	if err != nil || second.Original.ID != first.Original.ID || !bytes.Equal(first.Original.Plan, second.Original.Plan) {
		t.Fatalf("repeated activation changed original: %v", err)
	}
}
