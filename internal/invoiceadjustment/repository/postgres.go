package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/invoiceadjustment/application"
	"github.com/lucas/financial-api/internal/invoiceadjustment/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"time"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(p *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool: p} }
func (r *PostgresRepository) Create(ctx context.Context, owner string, in application.Input) (domain.Adjustment, error) {
	var cardStatus string
	if err := r.pool.QueryRow(ctx, `select status from public.credit_cards where id=$1 and user_id=$2`, in.CardID, owner).Scan(&cardStatus); errors.Is(err, pgx.ErrNoRows) {
		return domain.Adjustment{}, application.ErrCardNotFound
	} else if err != nil {
		return domain.Adjustment{}, err
	} else if cardStatus == "archived" {
		return domain.Adjustment{}, application.ErrCardArchived
	}
	var a domain.Adjustment
	err := r.pool.QueryRow(ctx, `insert into public.card_invoice_adjustments(user_id,currency_code,credit_card_id,payment_month,reference_month,name,amount_cents,context) values($1,$2,$3,$4,$5,$6,$7,$8) returning id,created_at,updated_at`, owner, in.CurrencyCode, in.CardID, in.PaymentMonth.Time(), month(in.ReferenceMonth), in.Name, in.AmountCents, in.Context).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Adjustment{}, application.ErrNotFound
	}
	if err != nil {
		return domain.Adjustment{}, err
	}
	a.UserID = owner
	a.CurrencyCode = in.CurrencyCode
	a.CardID = in.CardID
	a.Name = in.Name
	a.AmountCents = in.AmountCents
	a.PaymentMonth = in.PaymentMonth
	a.ReferenceMonth = in.ReferenceMonth
	a.Context = in.Context
	a.Status = domain.Active
	return a, nil
}
func (r *PostgresRepository) Update(ctx context.Context, owner, id string, in application.Input) (domain.Adjustment, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return domain.Adjustment{}, e
	}
	defer tx.Rollback(ctx)
	var status string
	e = tx.QueryRow(ctx, `select status from public.card_invoice_adjustments where id=$1 and user_id=$2 for update`, id, owner).Scan(&status)
	if errors.Is(e, pgx.ErrNoRows) {
		return domain.Adjustment{}, application.ErrNotFound
	}
	if e != nil {
		return domain.Adjustment{}, e
	}
	if status == "archived" {
		return domain.Adjustment{}, application.ErrArchived
	}
	_, e = tx.Exec(ctx, `insert into public.card_invoice_audit_events(user_id,event_type,adjustment_id,before_document,after_document) values($1,'adjustment_changed',$2,'{}'::jsonb,jsonb_build_object('name',$3,'amount_cents',$4))`, owner, id, in.Name, in.AmountCents)
	if e != nil {
		return domain.Adjustment{}, e
	}
	_, e = tx.Exec(ctx, `update public.card_invoice_adjustments set credit_card_id=$3,payment_month=$4,reference_month=$5,name=$6,amount_cents=$7,context=$8 where id=$1 and user_id=$2`, id, owner, in.CardID, in.PaymentMonth.Time(), month(in.ReferenceMonth), in.Name, in.AmountCents, in.Context)
	if e != nil {
		return domain.Adjustment{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return domain.Adjustment{}, e
	}
	return r.find(ctx, owner, id)
}
func (r *PostgresRepository) Archive(ctx context.Context, owner, id string) (domain.Adjustment, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return domain.Adjustment{}, e
	}
	defer tx.Rollback(ctx)
	var a domain.Adjustment
	var pm, rm *time.Time
	e = tx.QueryRow(ctx, `update public.card_invoice_adjustments set status='archived',archived_at=now() where id=$1 and user_id=$2 and status='active' returning id,currency_code,credit_card_id,payment_month,reference_month,name,amount_cents,context,status,archived_at,created_at,updated_at`, id, owner).Scan(&a.ID, &a.CurrencyCode, &a.CardID, &pm, &rm, &a.Name, &a.AmountCents, &a.Context, &a.Status, &a.ArchivedAt, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(e, pgx.ErrNoRows) {
		return domain.Adjustment{}, application.ErrNotFound
	}
	if e != nil {
		return domain.Adjustment{}, e
	}
	_, e = tx.Exec(ctx, `insert into public.card_invoice_audit_events(user_id,event_type,adjustment_id) values($1,'adjustment_archived',$2)`, owner, id)
	if e != nil {
		return domain.Adjustment{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return domain.Adjustment{}, e
	}
	a.UserID = owner
	if pm != nil {
		a.PaymentMonth, _ = planningdomain.NewYearMonth(pm.Year(), pm.Month())
	}
	if rm != nil {
		v, _ := planningdomain.NewYearMonth(rm.Year(), rm.Month())
		a.ReferenceMonth = &v
	}
	return a, nil
}
func (r *PostgresRepository) List(ctx context.Context, owner, card string, month planningdomain.YearMonth) ([]domain.Adjustment, error) {
	rows, e := r.pool.Query(ctx, `select id,currency_code,credit_card_id,payment_month,reference_month,name,amount_cents,context,status,archived_at,created_at,updated_at from public.card_invoice_adjustments where user_id=$1 and credit_card_id=$2 and payment_month=$3 and status='active' order by id`, owner, card, month.Time())
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Adjustment{}
	for rows.Next() {
		var a domain.Adjustment
		var pm, rm *time.Time
		if e = rows.Scan(&a.ID, &a.CurrencyCode, &a.CardID, &pm, &rm, &a.Name, &a.AmountCents, &a.Context, &a.Status, &a.ArchivedAt, &a.CreatedAt, &a.UpdatedAt); e != nil {
			return nil, e
		}
		a.UserID = owner
		if pm != nil {
			a.PaymentMonth, _ = planningdomain.NewYearMonth(pm.Year(), pm.Month())
		}
		if rm != nil {
			v, _ := planningdomain.NewYearMonth(rm.Year(), rm.Month())
			a.ReferenceMonth = &v
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
func (r *PostgresRepository) find(ctx context.Context, owner, id string) (domain.Adjustment, error) {
	var a domain.Adjustment
	var pm, rm *time.Time
	err := r.pool.QueryRow(ctx, `select id,currency_code,credit_card_id,payment_month,reference_month,name,amount_cents,context,status,archived_at,created_at,updated_at from public.card_invoice_adjustments where id=$1 and user_id=$2`, id, owner).Scan(&a.ID, &a.CurrencyCode, &a.CardID, &pm, &rm, &a.Name, &a.AmountCents, &a.Context, &a.Status, &a.ArchivedAt, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Adjustment{}, application.ErrNotFound
	}
	if pm != nil {
		a.PaymentMonth, _ = planningdomain.NewYearMonth(pm.Year(), pm.Month())
	}
	if rm != nil {
		v, _ := planningdomain.NewYearMonth(rm.Year(), rm.Month())
		a.ReferenceMonth = &v
	}
	a.UserID = owner
	return a, err
}
func month(m *planningdomain.YearMonth) any {
	if m == nil {
		return nil
	}
	return m.Time()
}
