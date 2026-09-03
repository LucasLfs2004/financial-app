package savings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/planning/domain"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repository *PostgresRepository) Get(ctx context.Context, ownerID, planID string) (Configuration, error) {
	rows, err := repository.pool.Query(ctx, `select id,start_month,end_month,amount_cents,context,created_at from public.saving_periods where user_id=$1 and plan_id=$2 order by start_month,id`, ownerID, planID)
	if err != nil {
		return Configuration{}, fmt.Errorf("get savings: %w", err)
	}
	defer rows.Close()
	periods := []Period{}
	for rows.Next() {
		var period Period
		var start time.Time
		var end *time.Time
		if err := rows.Scan(&period.ID, &start, &end, &period.AmountCents, &period.Context, &period.CreatedAt); err != nil {
			return Configuration{}, err
		}
		period.StartMonth, _ = domain.NewYearMonth(start.Year(), start.Month())
		if end != nil {
			parsed, _ := domain.NewYearMonth(end.Year(), end.Month())
			period.EndMonth = &parsed
		}
		periods = append(periods, period)
	}
	if err := rows.Err(); err != nil {
		return Configuration{}, err
	}
	return Configuration{Configured: len(periods) > 0, Periods: periods}, nil
}

func (repository *PostgresRepository) Put(ctx context.Context, ownerID, planID string, input PutInput) (Configuration, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return Configuration{}, err
	}
	defer tx.Rollback(ctx)
	var periodID string
	var start time.Time
	err = tx.QueryRow(ctx, `select id,start_month from public.saving_periods where user_id=$1 and plan_id=$2 and start_month<=$3 and (end_month is null or end_month>=$3) for update`, ownerID, planID, input.EffectiveFrom.Time()).Scan(&periodID, &start)
	if err == nil {
		if start.Year() == input.EffectiveFrom.Year() && start.Month() == input.EffectiveFrom.Month() {
			_, err = tx.Exec(ctx, `delete from public.saving_periods where id=$1`, periodID)
		} else {
			previous, _ := input.EffectiveFrom.AddMonths(-1)
			_, err = tx.Exec(ctx, `update public.saving_periods set end_month=$2 where id=$1`, periodID, previous.Time())
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	} else {
		return Configuration{}, fmt.Errorf("lock savings: %w", err)
	}
	if err != nil {
		return Configuration{}, err
	}
	_, err = tx.Exec(ctx, `insert into public.saving_periods (plan_id,user_id,start_month,end_month,amount_cents,context) values ($1,$2,$3,$4,$5,$6)`, planID, ownerID, input.EffectiveFrom.Time(), savingMonthTime(input.EndMonth), input.AmountCents, input.Context)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23P01" {
			return Configuration{}, ErrOverlap
		}
		return Configuration{}, fmt.Errorf("put savings: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Configuration{}, err
	}
	return repository.Get(ctx, ownerID, planID)
}

func savingMonthTime(month *domain.YearMonth) any {
	if month == nil {
		return nil
	}
	return month.Time()
}
