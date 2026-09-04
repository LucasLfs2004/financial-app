package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/paymentmethod/application"
	"github.com/lucas/financial-api/internal/paymentmethod/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, ownerID, planID, itemID string, input application.Input) (domain.Period, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Period{}, fmt.Errorf("begin payment method change: %w", err)
	}
	defer tx.Rollback(ctx)
	var kind string
	if err = tx.QueryRow(ctx, `select kind from public.financial_items where id=$1 and plan_id=$2 and user_id=$3 for share`, itemID, planID, ownerID).Scan(&kind); errors.Is(err, pgx.ErrNoRows) {
		return domain.Period{}, application.ErrItemNotFound
	} else if err != nil {
		return domain.Period{}, fmt.Errorf("lock financial item: %w", err)
	}
	if kind == "recurring_income" || kind == "one_time_income" {
		return domain.Period{}, domain.ErrValidation
	}
	if input.CreditCardID != nil {
		var status string
		err = tx.QueryRow(ctx, `select status from public.credit_cards where id=$1 and user_id=$2 for share`, *input.CreditCardID, ownerID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Period{}, application.ErrCardNotFound
		}
		if err != nil {
			return domain.Period{}, fmt.Errorf("check payment card: %w", err)
		}
		if status == "archived" {
			return domain.Period{}, application.ErrCardArchived
		}
	}
	// Serialize temporal changes for this item and close the applicable
	// explicit period before inserting the new one.
	var previousID string
	var previousStart time.Time
	err = tx.QueryRow(ctx, `select id,start_month from public.financial_item_payment_periods where financial_item_id=$1 and plan_id=$2 and user_id=$3 and start_month <= $4 and (end_month is null or end_month >= $4) for update`, itemID, planID, ownerID, input.EffectiveFrom.Time()).Scan(&previousID, &previousStart)
	if err == nil {
		if previousStart.Year() == input.EffectiveFrom.Year() && previousStart.Month() == input.EffectiveFrom.Month() {
			if _, err = tx.Exec(ctx, `delete from public.financial_item_payment_periods where id=$1`, previousID); err != nil {
				return domain.Period{}, fmt.Errorf("replace payment method period: %w", err)
			}
		} else {
			previousMonth, monthErr := input.EffectiveFrom.AddMonths(-1)
			if monthErr != nil {
				return domain.Period{}, monthErr
			}
			if _, err = tx.Exec(ctx, `update public.financial_item_payment_periods set end_month=$2 where id=$1`, previousID, previousMonth.Time()); err != nil {
				return domain.Period{}, fmt.Errorf("close payment method period: %w", err)
			}
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Period{}, fmt.Errorf("lock applicable payment method period: %w", err)
	}
	var id string
	err = tx.QueryRow(ctx, `insert into public.financial_item_payment_periods (financial_item_id,plan_id,user_id,start_month,end_month,method,credit_card_id,context) values ($1,$2,$3,$4,$5,$6,$7,$8) returning id`, itemID, planID, ownerID, input.EffectiveFrom.Time(), monthTime(input.EndMonth), input.Method, input.CreditCardID, input.Context).Scan(&id)
	if err != nil {
		if isOverlap(err) {
			return domain.Period{}, domain.ErrOverlap
		}
		return domain.Period{}, fmt.Errorf("create payment method period: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Period{}, fmt.Errorf("commit payment method change: %w", err)
	}
	return r.find(ctx, ownerID, planID, id)
}

func (r *PostgresRepository) List(ctx context.Context, ownerID, planID, itemID string) ([]domain.Period, error) {
	rows, err := r.pool.Query(ctx, `select id,start_month,end_month,method,credit_card_id,context,recorded_at,created_at from public.financial_item_payment_periods where user_id=$1 and plan_id=$2 and financial_item_id=$3 order by start_month,id`, ownerID, planID, itemID)
	if err != nil {
		return nil, fmt.Errorf("list payment methods: %w", err)
	}
	defer rows.Close()
	result := []domain.Period{}
	for rows.Next() {
		var p domain.Period
		var start time.Time
		var end *time.Time
		var method string
		if err := rows.Scan(&p.ID, &start, &end, &method, &p.CreditCardID, &p.Context, &p.RecordedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.StartMonth, err = planningdomain.NewYearMonth(start.Year(), start.Month())
		if err != nil {
			return nil, err
		}
		if end != nil {
			value, e := planningdomain.NewYearMonth(end.Year(), end.Month())
			if e != nil {
				return nil, e
			}
			p.EndMonth = &value
		}
		p.Method = domain.Kind(method)
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresRepository) find(ctx context.Context, ownerID, planID, id string) (domain.Period, error) {
	var p domain.Period
	var start time.Time
	var end *time.Time
	var method string
	err := r.pool.QueryRow(ctx, `select id,start_month,end_month,method,credit_card_id,context,recorded_at,created_at from public.financial_item_payment_periods where id=$1 and user_id=$2 and plan_id=$3`, id, ownerID, planID).Scan(&p.ID, &start, &end, &method, &p.CreditCardID, &p.Context, &p.RecordedAt, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Period{}, application.ErrNotFound
	}
	if err != nil {
		return domain.Period{}, err
	}
	p.StartMonth, _ = planningdomain.NewYearMonth(start.Year(), start.Month())
	if end != nil {
		v, _ := planningdomain.NewYearMonth(end.Year(), end.Month())
		p.EndMonth = &v
	}
	p.Method = domain.Kind(method)
	return p, nil
}
func monthTime(month *planningdomain.YearMonth) any {
	if month == nil {
		return nil
	}
	return month.Time()
}
func isOverlap(err error) bool { var e *pgconn.PgError; return errors.As(err, &e) && e.Code == "23P01" }
