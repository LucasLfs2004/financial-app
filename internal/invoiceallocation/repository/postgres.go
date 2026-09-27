package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/invoiceallocation/application"
	"github.com/lucas/financial-api/internal/invoiceallocation/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
	"time"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}
func (r *PostgresRepository) Move(ctx context.Context, ownerID, itemID string, reference planningdomain.YearMonth, targetCard string, targetMonth planningdomain.YearMonth, reason *string) (domain.Move, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Move{}, err
	}
	defer tx.Rollback(ctx)
	var fromCard string
	var fromMonth time.Time
	// Lock the item and its latest move so two moves cannot derive the same origin.
	var kind string
	err = tx.QueryRow(ctx, `select kind from public.financial_items where id=$1 and user_id=$2 for share`, itemID, ownerID).Scan(&kind)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Move{}, application.ErrNotFound
	}
	if err != nil {
		return domain.Move{}, err
	}
	itemKind, parseErr := planningdomain.ParseFinancialItemKind(kind)
	if parseErr != nil || !itemKind.IsExpense() {
		return domain.Move{}, application.ErrNotCardLinked
	}
	err = tx.QueryRow(ctx, `select to_credit_card_id,to_payment_month from public.card_invoice_audit_events where user_id=$1 and financial_item_id=$2 and reference_month=$3 and event_type='occurrence_moved' order by recorded_at desc,id desc limit 1 for update`, ownerID, itemID, reference.Time()).Scan(&fromCard, &fromMonth)
	if errors.Is(err, pgx.ErrNoRows) {
		var method string
		var cardID *string
		err = tx.QueryRow(ctx, `select method,credit_card_id from public.financial_item_payment_periods where financial_item_id=$1 and user_id=$2 and start_month<=$3 and (end_month is null or end_month>=$3)`, itemID, ownerID, reference.Time()).Scan(&method, &cardID)
		if errors.Is(err, pgx.ErrNoRows) || method != "credit_card" || cardID == nil {
			return domain.Move{}, application.ErrNotCardLinked
		}
		if err != nil {
			return domain.Move{}, err
		}
		fromCard = *cardID
		var offset int
		err = tx.QueryRow(ctx, `select payment_month_offset from public.credit_card_periods where credit_card_id=$1 and user_id=$2 and start_month<=$3 and (end_month is null or end_month>=$3)`, fromCard, ownerID, reference.Time()).Scan(&offset)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Move{}, application.ErrNotFound
		}
		if err != nil {
			return domain.Move{}, err
		}
		paymentMonth, monthErr := reference.AddMonths(offset)
		if monthErr != nil {
			return domain.Move{}, monthErr
		}
		fromMonth = paymentMonth.Time()
		if err != nil {
			return domain.Move{}, err
		}
	} else if err != nil {
		return domain.Move{}, err
	}
	if fromCard == targetCard && fromMonth.Equal(targetMonth.Time()) {
		return domain.Move{}, application.ErrSameDestination
	}
	var status string
	err = tx.QueryRow(ctx, `select status from public.credit_cards where id=$1 and user_id=$2 for share`, targetCard, ownerID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Move{}, application.ErrCardNotFound
	}
	if err != nil {
		return domain.Move{}, err
	}
	if status == "archived" {
		return domain.Move{}, application.ErrCardArchived
	}
	originCard, originMonth := fromCard, fromMonth
	var eventID string
	var recordedAt time.Time
	err = tx.QueryRow(ctx, `insert into public.card_invoice_audit_events(user_id,event_type,financial_item_id,reference_month,from_credit_card_id,from_payment_month,to_credit_card_id,to_payment_month,reason) values($1,'occurrence_moved',$2,$3,$4,$5,$6,$7,$8) returning id,recorded_at`, ownerID, itemID, reference.Time(), fromCard, fromMonth, targetCard, targetMonth.Time(), reason).Scan(&eventID, &recordedAt)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23503" {
			return domain.Move{}, application.ErrCardNotFound
		}
		return domain.Move{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Move{}, err
	}
	originYM, _ := planningdomain.NewYearMonth(originMonth.Year(), originMonth.Month())
	return domain.Move{ID: eventID, ItemID: itemID, ReferenceMonth: reference, FromCardID: originCard, FromPaymentMonth: originYM, ToCardID: targetCard, ToPaymentMonth: targetMonth, Reason: reason, RecordedAt: recordedAt}, nil
}
func (r *PostgresRepository) History(ctx context.Context, ownerID, itemID string, reference planningdomain.YearMonth) ([]domain.Move, error) {
	rows, err := r.pool.Query(ctx, `select id,from_credit_card_id,from_payment_month,to_credit_card_id,to_payment_month,reason,recorded_at from public.card_invoice_audit_events where user_id=$1 and financial_item_id=$2 and reference_month=$3 and event_type='occurrence_moved' order by recorded_at,id`, ownerID, itemID, reference.Time())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Move{}
	for rows.Next() {
		var m domain.Move
		var fromM, toM time.Time
		if err := rows.Scan(&m.ID, &m.FromCardID, &fromM, &m.ToCardID, &toM, &m.Reason, &m.RecordedAt); err != nil {
			return nil, err
		}
		m.ItemID = itemID
		m.ReferenceMonth = reference
		m.FromPaymentMonth, _ = planningdomain.NewYearMonth(fromM.Year(), fromM.Month())
		m.ToPaymentMonth, _ = planningdomain.NewYearMonth(toM.Year(), toM.Month())
		result = append(result, m)
	}
	return result, rows.Err()
}
