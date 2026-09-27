package integration_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/planning"
)

type snapshotV2Document struct {
	Release2 struct {
		Institutions []struct {
			ID string `json:"id"`
		} `json:"financial_institutions"`
		Cards []struct {
			ID      string `json:"id"`
			Periods []struct {
				ID string `json:"id"`
			} `json:"periods"`
		} `json:"credit_cards"`
		PaymentMethods []struct {
			ID string `json:"id"`
		} `json:"payment_methods"`
		Adjustments []struct {
			ID string `json:"id"`
		} `json:"invoice_adjustments"`
		Moves []json.RawMessage `json:"invoice_moves"`
	} `json:"release_2"`
}

func TestSnapshotV2IsReachableOrderedImmutableAndV1RemainsReadable(t *testing.T) {
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
	email := "snapshot-" + ownerID + "@example.com"
	if _, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, ownerID, email); err != nil {
		t.Fatal(err)
	}

	var planID string
	if err = pool.QueryRow(ctx, `insert into public.plans(user_id,name,start_month,end_month,currency_code) values($1,'Snapshot v2','2026-01-01','2026-12-01','BRL') returning id`, ownerID).Scan(&planID); err != nil {
		t.Fatal(err)
	}
	itemIDs := make([]string, 3)
	for index, definition := range []struct{ name, kind string }{{"Salário", "recurring_income"}, {"Cartão B", "fixed_expense"}, {"Cartão A", "projected_variable_expense"}} {
		if err = pool.QueryRow(ctx, `insert into public.financial_items(user_id,currency_code,name,kind) values($1,'BRL',$2,$3) returning id`, ownerID, definition.name, definition.kind).Scan(&itemIDs[index]); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `insert into public.financial_item_periods(financial_item_id,user_id,start_month,end_month,amount_cents,recurrence) values($1,$2,'2026-01-01','2026-12-01',10000,'monthly')`, itemIDs[index], ownerID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `insert into public.saving_periods(plan_id,user_id,start_month,end_month,amount_cents) values($1,$2,'2026-01-01','2026-12-01',0)`, planID, ownerID); err != nil {
		t.Fatal(err)
	}

	var institutionID string
	if err = pool.QueryRow(ctx, `insert into public.financial_institutions(user_id,name) values($1,'Banco alcançável') returning id`, ownerID).Scan(&institutionID); err != nil {
		t.Fatal(err)
	}
	cardIDs := make([]string, 3)
	for index, name := range []string{"Segundo", "Primeiro", "Não alcançável"} {
		if err = pool.QueryRow(ctx, `insert into public.credit_cards(user_id,institution_id,name) values($1,$2,$3) returning id`, ownerID, institutionID, name).Scan(&cardIDs[index]); err != nil {
			t.Fatal(err)
		}
	}
	periodIDs := make([]string, 2)
	if err = pool.QueryRow(ctx, `insert into public.credit_card_periods(credit_card_id,user_id,start_month,end_month,nominal_due_day,payment_month_offset) values($1,$2,'2026-07-01','2026-12-01',6,1) returning id`, cardIDs[0], ownerID).Scan(&periodIDs[1]); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `insert into public.credit_card_periods(credit_card_id,user_id,start_month,end_month,nominal_due_day,payment_month_offset) values($1,$2,'2026-01-01','2026-06-01',6,1) returning id`, cardIDs[0], ownerID).Scan(&periodIDs[0]); err != nil {
		t.Fatal(err)
	}
	paymentIDs := make([]string, 2)
	for index := range 2 {
		if err = pool.QueryRow(ctx, `insert into public.financial_item_payment_periods(financial_item_id,user_id,start_month,end_month,method,credit_card_id) values($1,$2,'2026-01-01','2026-12-01','credit_card',$3) returning id`, itemIDs[index+1], ownerID, cardIDs[index]).Scan(&paymentIDs[index]); err != nil {
			t.Fatal(err)
		}
	}
	var adjustmentID string
	if err = pool.QueryRow(ctx, `insert into public.card_invoice_adjustments(user_id,currency_code,credit_card_id,payment_month,name,amount_cents) values($1,'BRL',$2,'2026-08-01','Tarifa',500) returning id`, ownerID, cardIDs[1]).Scan(&adjustmentID); err != nil {
		t.Fatal(err)
	}

	repository := planning.NewPostgresRepository(pool)
	first, err := repository.Activate(ctx, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Original.SchemaVersion != 4 {
		t.Fatalf("schema version=%d", first.Original.SchemaVersion)
	}
	var document snapshotV2Document
	if err = json.Unmarshal(first.Original.Plan, &document); err != nil {
		t.Fatal(err)
	}
	expectedCards := append([]string(nil), cardIDs[:2]...)
	slices.Sort(expectedCards)
	if len(document.Release2.Institutions) != 1 || len(document.Release2.Cards) != 2 {
		t.Fatalf("reachable resources institutions=%+v cards=%+v", document.Release2.Institutions, document.Release2.Cards)
	}
	actualCards := []string{document.Release2.Cards[0].ID, document.Release2.Cards[1].ID}
	if !slices.Equal(actualCards, expectedCards) {
		t.Fatalf("cards are not deterministically ordered: got=%v want=%v", actualCards, expectedCards)
	}
	if len(document.Release2.PaymentMethods) != 2 || len(document.Release2.Adjustments) != 1 || document.Release2.Adjustments[0].ID != adjustmentID || len(document.Release2.Moves) != 0 {
		t.Fatalf("release 2 collections=%+v", document.Release2)
	}
	expectedPayments := append([]string(nil), paymentIDs...)
	slices.Sort(expectedPayments)
	if !slices.Equal([]string{document.Release2.PaymentMethods[0].ID, document.Release2.PaymentMethods[1].ID}, expectedPayments) {
		t.Fatalf("payment methods are not ordered: %+v", document.Release2.PaymentMethods)
	}
	for _, card := range document.Release2.Cards {
		if card.ID == cardIDs[0] {
			if len(card.Periods) != 2 || card.Periods[0].ID != periodIDs[0] || card.Periods[1].ID != periodIDs[1] {
				t.Fatalf("card periods are not ordered: %+v", card.Periods)
			}
		}
	}

	second, err := repository.Activate(ctx, ownerID)
	if err != nil || second.Original.ID != first.Original.ID || !bytes.Equal(second.Original.Plan, first.Original.Plan) {
		t.Fatalf("snapshot changed on repeated activation: %+v error=%v", second.Original, err)
	}
	if _, err = pool.Exec(ctx, `update public.plan_snapshots set document='{}'::jsonb where id=$1`, first.Original.ID); err == nil {
		t.Fatal("snapshot update should be rejected")
	}

	v1OwnerID := randomTestUUID(t)
	if _, err = pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, v1OwnerID, "snapshot-v1-"+v1OwnerID+"@example.com"); err != nil {
		t.Fatal(err)
	}
	var v1PlanID string
	if err = pool.QueryRow(ctx, `insert into public.plans(user_id,name,status,start_month,end_month,currency_code,activated_at) values($1,'Snapshot v1','active','2026-01-01','2026-12-01','BRL',now()) returning id`, v1OwnerID).Scan(&v1PlanID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `insert into public.plan_snapshots(plan_id,user_id,kind,schema_version,document) values($1,$2,'original',1,'{"name":"legado"}'::jsonb)`, v1PlanID, v1OwnerID); err != nil {
		t.Fatal(err)
	}
	legacy, err := repository.FindOriginal(ctx, v1OwnerID)
	if err != nil || legacy.SchemaVersion != 1 || !json.Valid(legacy.Plan) {
		t.Fatalf("legacy snapshot=%+v error=%v", legacy, err)
	}
	for _, version := range []int{2, 3} {
		legacyOwnerID := randomTestUUID(t)
		if _, err := pool.Exec(ctx, `insert into auth.users(id,instance_id,aud,role,email,encrypted_password,email_confirmed_at,raw_app_meta_data,raw_user_meta_data,created_at,updated_at) values($1,'00000000-0000-0000-0000-000000000000','authenticated','authenticated',$2,'',now(),'{"provider":"email","providers":["email"]}','{}',now(),now())`, legacyOwnerID, "snapshot-legacy-"+legacyOwnerID+"@example.com"); err != nil {
			t.Fatal(err)
		}
		var legacyPlanID string
		if err := pool.QueryRow(ctx, `insert into public.plans(user_id,name,status,start_month,end_month,currency_code,activated_at) values($1,'Legado','active','2026-01-01','2026-12-01','BRL',now()) returning id`, legacyOwnerID).Scan(&legacyPlanID); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `insert into public.plan_snapshots(plan_id,user_id,kind,schema_version,document) values($1,$2,'original',$3,'{"name":"legado"}'::jsonb)`, legacyPlanID, legacyOwnerID, version); err != nil {
			t.Fatal(err)
		}
		read, err := repository.FindOriginal(ctx, legacyOwnerID)
		if err != nil || read.SchemaVersion != version || !json.Valid(read.Plan) {
			t.Fatalf("legacy v%d snapshot=%+v error=%v", version, read, err)
		}
	}
}

func randomTestUUID(t *testing.T) string {
	t.Helper()
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		t.Fatal(err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(value)
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}
