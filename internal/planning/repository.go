package planning

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

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repository *PostgresRepository) Create(ctx context.Context, ownerID string, input CreateInput) (Plan, error) {
	const query = `
		insert into public.plans (user_id, name, start_month, end_month, currency_code)
		values ($1, $2, $3, $4, $5)
		returning id, user_id, name, status, start_month, end_month, currency_code,
			activated_at, archived_at, created_at, updated_at
	`

	plan, err := scanPlan(repository.pool.QueryRow(ctx, query,
		ownerID,
		input.Name,
		input.StartMonth.Time(),
		input.EndMonth.Time(),
		input.CurrencyCode,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return Plan{}, ErrAlreadyExists
		}
		return Plan{}, fmt.Errorf("create plan: %w", err)
	}
	return plan, nil
}

func (repository *PostgresRepository) FindCurrent(ctx context.Context, ownerID string) (Plan, error) {
	const query = `
		select id, user_id, name, status, start_month, end_month, currency_code,
			activated_at, archived_at, created_at, updated_at
		from public.plans
		where user_id = $1 and status <> 'archived'
		limit 1
	`

	plan, err := scanPlan(repository.pool.QueryRow(ctx, query, ownerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, ErrNotFound
	}
	if err != nil {
		return Plan{}, fmt.Errorf("find current plan: %w", err)
	}
	return plan, nil
}

func (repository *PostgresRepository) UpdateDraft(ctx context.Context, ownerID string, input UpdateInput) (Plan, error) {
	const query = `
		update public.plans
		set
			name = coalesce($2, name),
			start_month = coalesce($3, start_month),
			end_month = coalesce($4, end_month),
			currency_code = coalesce($5, currency_code)
		where user_id = $1 and status = 'draft'
		returning id, user_id, name, status, start_month, end_month, currency_code,
			activated_at, archived_at, created_at, updated_at
	`

	var startMonth any
	if input.StartMonth != nil {
		startMonth = input.StartMonth.Time()
	}
	var endMonth any
	if input.EndMonth != nil {
		endMonth = input.EndMonth.Time()
	}
	var name any
	if input.Name != nil {
		name = *input.Name
	}
	var currencyCode any
	if input.CurrencyCode != nil {
		currencyCode = *input.CurrencyCode
	}

	plan, err := scanPlan(repository.pool.QueryRow(ctx, query, ownerID, name, startMonth, endMonth, currencyCode))
	if errors.Is(err, pgx.ErrNoRows) {
		current, findErr := repository.FindCurrent(ctx, ownerID)
		if errors.Is(findErr, ErrNotFound) {
			return Plan{}, ErrNotFound
		}
		if findErr != nil {
			return Plan{}, findErr
		}
		if current.Status != domain.PlanStatusDraft {
			return Plan{}, ErrNotDraft
		}
		return Plan{}, ErrNotFound
	}
	if err != nil {
		return Plan{}, fmt.Errorf("update draft plan: %w", err)
	}
	return plan, nil
}

func scanPlan(row pgx.Row) (Plan, error) {
	var result Plan
	var status string
	var startMonth time.Time
	var endMonth time.Time
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.Name,
		&status,
		&startMonth,
		&endMonth,
		&result.CurrencyCode,
		&result.ActivatedAt,
		&result.ArchivedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return Plan{}, err
	}

	parsedStatus, err := domain.ParsePlanStatus(status)
	if err != nil {
		return Plan{}, fmt.Errorf("parse plan status: %w", err)
	}
	parsedStartMonth, err := domain.NewYearMonth(startMonth.Year(), startMonth.Month())
	if err != nil {
		return Plan{}, fmt.Errorf("parse start month: %w", err)
	}
	parsedEndMonth, err := domain.NewYearMonth(endMonth.Year(), endMonth.Month())
	if err != nil {
		return Plan{}, fmt.Errorf("parse end month: %w", err)
	}
	result.Status = parsedStatus
	result.StartMonth = parsedStartMonth
	result.EndMonth = parsedEndMonth
	return result, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
