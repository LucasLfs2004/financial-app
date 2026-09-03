package financialitem

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

func (repository *PostgresRepository) Create(ctx context.Context, ownerID, planID string, input CreateInput) (Item, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return Item{}, fmt.Errorf("begin create item: %w", err)
	}
	defer tx.Rollback(ctx)
	var itemID string
	err = tx.QueryRow(ctx, `insert into public.financial_items (plan_id,user_id,name,kind,description) values ($1,$2,$3,$4,$5) returning id`, planID, ownerID, input.Name, input.Kind, input.Description).Scan(&itemID)
	if err != nil {
		return Item{}, mapError("create item", err)
	}
	_, err = tx.Exec(ctx, `insert into public.financial_item_periods (financial_item_id,plan_id,user_id,start_month,end_month,amount_cents,recurrence,cash_month_offset,context) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, itemID, planID, ownerID, input.Period.StartMonth.Time(), monthTime(input.Period.EndMonth), input.Period.AmountCents, input.Period.Recurrence, input.Period.CashMonthOffset, input.Period.Context)
	if err != nil {
		return Item{}, mapError("create item period", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Item{}, fmt.Errorf("commit create item: %w", err)
	}
	return repository.Find(ctx, ownerID, planID, itemID)
}

func (repository *PostgresRepository) List(ctx context.Context, ownerID, planID string, filters Filters) ([]Item, error) {
	rows, err := repository.pool.Query(ctx, `
		select
			i.id, i.plan_id, i.name, i.kind, i.description, i.status,
			i.archived_at, i.created_at, i.updated_at,
			p.id, p.start_month, p.end_month, p.amount_cents, p.recurrence,
			p.cash_month_offset, p.context, p.recorded_at, p.created_at
		from public.financial_items i
		left join public.financial_item_periods p
			on p.financial_item_id = i.id
			and p.plan_id = i.plan_id
			and p.user_id = i.user_id
		where i.user_id = $1
			and i.plan_id = $2
			and ($3::financial_item_kind is null or i.kind = $3)
			and ($4::financial_item_status is null or i.status = $4)
		order by i.created_at, i.id, p.start_month, p.id
	`, ownerID, planID, enumValue(filters.Kind), enumValue(filters.Status))
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := []Item{}
	itemIndexes := make(map[string]int)
	for rows.Next() {
		var item Item
		var kind, status string
		var periodID *string
		var periodStart, periodEnd *time.Time
		var periodAmount *int64
		var recurrence *string
		var cashMonthOffset *int
		var periodContext *string
		var recordedAt, periodCreatedAt *time.Time

		if err := rows.Scan(
			&item.ID, &item.PlanID, &item.Name, &kind, &item.Description, &status,
			&item.ArchivedAt, &item.CreatedAt, &item.UpdatedAt,
			&periodID, &periodStart, &periodEnd, &periodAmount, &recurrence,
			&cashMonthOffset, &periodContext, &recordedAt, &periodCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan listed item: %w", err)
		}

		item.Kind, err = domain.ParseFinancialItemKind(kind)
		if err != nil {
			return nil, fmt.Errorf("parse listed item kind: %w", err)
		}
		item.Status, err = domain.ParseFinancialItemStatus(status)
		if err != nil {
			return nil, fmt.Errorf("parse listed item status: %w", err)
		}
		item.Periods = []Period{}

		index, exists := itemIndexes[item.ID]
		if !exists {
			index = len(items)
			itemIndexes[item.ID] = index
			items = append(items, item)
		}

		if periodID == nil {
			continue
		}
		if periodStart == nil || periodAmount == nil || recurrence == nil || cashMonthOffset == nil || recordedAt == nil || periodCreatedAt == nil {
			return nil, fmt.Errorf("scan listed item: incomplete period %s", *periodID)
		}

		startMonth, err := domain.NewYearMonth(periodStart.Year(), periodStart.Month())
		if err != nil {
			return nil, fmt.Errorf("parse listed period start month: %w", err)
		}
		parsedRecurrence, err := domain.ParseRecurrence(*recurrence)
		if err != nil {
			return nil, fmt.Errorf("parse listed period recurrence: %w", err)
		}

		items[index].Periods = append(items[index].Periods, Period{
			ID:              *periodID,
			StartMonth:      startMonth,
			EndMonth:        parsedMonth(periodEnd),
			AmountCents:     *periodAmount,
			Recurrence:      parsedRecurrence,
			CashMonthOffset: *cashMonthOffset,
			Context:         periodContext,
			RecordedAt:      *recordedAt,
			CreatedAt:       *periodCreatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	return items, nil
}

func (repository *PostgresRepository) Find(ctx context.Context, ownerID, planID, itemID string) (Item, error) {
	item, err := findItem(ctx, repository.pool, ownerID, planID, itemID)
	if err != nil {
		return Item{}, err
	}
	item.Periods, err = repository.periods(ctx, repository.pool, ownerID, planID, itemID)
	return item, err
}

func (repository *PostgresRepository) Update(ctx context.Context, ownerID, planID, itemID string, input UpdateInput) (Item, error) {
	result, err := repository.pool.Exec(ctx, `update public.financial_items set name=coalesce($4,name), description=case when $5 then $6 else description end where id=$1 and plan_id=$2 and user_id=$3`, itemID, planID, ownerID, nullable(input.Name), input.DescriptionSet, nullable(input.Description))
	if err != nil {
		return Item{}, mapError("update item", err)
	}
	if result.RowsAffected() == 0 {
		return Item{}, ErrNotFound
	}
	return repository.Find(ctx, ownerID, planID, itemID)
}

func (repository *PostgresRepository) Change(ctx context.Context, ownerID, planID, itemID string, input ChangeInput) (Item, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Item{}, fmt.Errorf("begin change item: %w", err)
	}
	defer tx.Rollback(ctx)
	var periodID string
	var start time.Time
	var recurrence string
	err = tx.QueryRow(ctx, `select id,start_month,recurrence from public.financial_item_periods where financial_item_id=$1 and plan_id=$2 and user_id=$3 and start_month<=$4 and (end_month is null or end_month>=$4) for update`, itemID, planID, ownerID, input.EffectiveFrom.Time()).Scan(&periodID, &start, &recurrence)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, fmt.Errorf("lock applicable period: %w", err)
	}
	if sameMonth(start, input.EffectiveFrom) {
		_, err = tx.Exec(ctx, `delete from public.financial_item_periods where id=$1`, periodID)
	} else {
		previous, previousErr := input.EffectiveFrom.AddMonths(-1)
		if previousErr != nil {
			return Item{}, ErrValidation
		}
		_, err = tx.Exec(ctx, `update public.financial_item_periods set end_month=$2 where id=$1`, periodID, previous.Time())
	}
	if err != nil {
		return Item{}, fmt.Errorf("close applicable period: %w", err)
	}
	_, err = tx.Exec(ctx, `insert into public.financial_item_periods (financial_item_id,plan_id,user_id,start_month,end_month,amount_cents,recurrence,cash_month_offset,context,recorded_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,now())`, itemID, planID, ownerID, input.EffectiveFrom.Time(), monthTime(input.EndMonth), input.AmountCents, recurrence, input.CashMonthOffset, input.Context)
	if err != nil {
		return Item{}, mapError("create changed period", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Item{}, fmt.Errorf("commit item change: %w", err)
	}
	return repository.Find(ctx, ownerID, planID, itemID)
}

func (repository *PostgresRepository) Archive(ctx context.Context, ownerID, planID, itemID string, input ArchiveInput) (Item, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return Item{}, fmt.Errorf("begin archive item: %w", err)
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `update public.financial_items set status='archived',archived_at=now() where id=$1 and plan_id=$2 and user_id=$3 and status='active'`, itemID, planID, ownerID)
	if err != nil {
		return Item{}, fmt.Errorf("archive item: %w", err)
	}
	if result.RowsAffected() == 0 {
		return Item{}, ErrNotFound
	}
	if input.EffectiveFrom != nil {
		_, err = tx.Exec(ctx, `delete from public.financial_item_periods where financial_item_id=$1 and plan_id=$2 and user_id=$3 and start_month >= $4`, itemID, planID, ownerID, input.EffectiveFrom.Time())
		if err == nil {
			previous, previousErr := input.EffectiveFrom.AddMonths(-1)
			if previousErr != nil {
				return Item{}, ErrValidation
			}
			_, err = tx.Exec(ctx, `update public.financial_item_periods set end_month=$4 where financial_item_id=$1 and plan_id=$2 and user_id=$3 and start_month < $5 and (end_month is null or end_month >= $5)`, itemID, planID, ownerID, previous.Time(), input.EffectiveFrom.Time())
		}
		if err != nil {
			return Item{}, fmt.Errorf("close archived item periods: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Item{}, fmt.Errorf("commit archive item: %w", err)
	}
	return repository.Find(ctx, ownerID, planID, itemID)
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func findItem(ctx context.Context, database queryer, ownerID, planID, itemID string) (Item, error) {
	item, err := scanItem(database.QueryRow(ctx, `select id,plan_id,name,kind,description,status,archived_at,created_at,updated_at from public.financial_items where id=$1 and plan_id=$2 and user_id=$3`, itemID, planID, ownerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, fmt.Errorf("find item: %w", err)
	}
	return item, nil
}

func (repository *PostgresRepository) periods(ctx context.Context, database queryer, ownerID, planID, itemID string) ([]Period, error) {
	rows, err := database.Query(ctx, `select id,start_month,end_month,amount_cents,recurrence,cash_month_offset,context,recorded_at,created_at from public.financial_item_periods where financial_item_id=$1 and plan_id=$2 and user_id=$3 order by start_month,id`, itemID, planID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list item periods: %w", err)
	}
	defer rows.Close()
	periods := []Period{}
	for rows.Next() {
		var period Period
		var start time.Time
		var end *time.Time
		var recurrence string
		if err := rows.Scan(&period.ID, &start, &end, &period.AmountCents, &recurrence, &period.CashMonthOffset, &period.Context, &period.RecordedAt, &period.CreatedAt); err != nil {
			return nil, err
		}
		period.StartMonth, _ = domain.NewYearMonth(start.Year(), start.Month())
		period.EndMonth = parsedMonth(end)
		period.Recurrence, _ = domain.ParseRecurrence(recurrence)
		periods = append(periods, period)
	}
	return periods, rows.Err()
}

func scanItem(row pgx.Row) (Item, error) {
	var item Item
	var kind, status string
	if err := row.Scan(&item.ID, &item.PlanID, &item.Name, &kind, &item.Description, &status, &item.ArchivedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return Item{}, err
	}
	item.Kind, _ = domain.ParseFinancialItemKind(kind)
	item.Status, _ = domain.ParseFinancialItemStatus(status)
	item.Periods = []Period{}
	return item, nil
}

func mapError(operation string, err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23P01" {
		return ErrPeriodOverlap
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func monthTime(month *domain.YearMonth) any {
	if month == nil {
		return nil
	}
	return month.Time()
}
func parsedMonth(value *time.Time) *domain.YearMonth {
	if value == nil {
		return nil
	}
	month, _ := domain.NewYearMonth(value.Year(), value.Month())
	return &month
}
func sameMonth(value time.Time, month domain.YearMonth) bool {
	return value.Year() == month.Year() && value.Month() == month.Month()
}
func nullable[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}
func enumValue[T ~string](value *T) any {
	if value == nil {
		return nil
	}
	return string(*value)
}
