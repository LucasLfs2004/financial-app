package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/cardinvoice"
	debtapplication "github.com/lucas/financial-api/internal/debt/application"
	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repository *PostgresRepository) Create(ctx context.Context, ownerID, currencyCode string, input debtapplication.CreateRecord) (debtdomain.Debt, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return debtdomain.Debt{}, fmt.Errorf("begin debt creation: %w", err)
	}
	defer tx.Rollback(ctx)

	var debtID string
	err = tx.QueryRow(ctx, `
		insert into public.financial_items (
			user_id, currency_code, name, kind, description
		) values ($1, $2, $3, 'debt_installment', $4)
		returning id
	`, ownerID, currencyCode, input.Name, input.Description).Scan(&debtID)
	if err != nil {
		return debtdomain.Debt{}, mapError("create debt financial item", err)
	}

	_, err = tx.Exec(ctx, `
		insert into public.financial_item_periods (
			financial_item_id, user_id, start_month, end_month,
			amount_cents, recurrence, cash_month_offset, context
		) values ($1, $2, $3, $4, $5, 'monthly', $6, $7)
	`, debtID, ownerID, input.ScheduledStart.Time(), input.ScheduledEnd.Time(), input.InstallmentAmountCents, input.CashMonthOffset, input.Context)
	if err != nil {
		return debtdomain.Debt{}, mapError("create initial debt period", err)
	}

	_, err = tx.Exec(ctx, `
		insert into public.debts (
			financial_item_id, user_id, original_total_cents,
			total_installments, first_projected_installment,
			scheduled_start_month, scheduled_end_month
		) values ($1, $2, $3, $4, $5, $6, $7)
	`, debtID, ownerID, nullable(input.OriginalTotalCents), input.TotalInstallments, input.FirstProjectedInstallment, input.ScheduledStart.Time(), input.ScheduledEnd.Time())
	if err != nil {
		return debtdomain.Debt{}, mapError("create debt metadata", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return debtdomain.Debt{}, mapError("commit debt creation", err)
	}
	return repository.Find(ctx, ownerID, debtID)
}

func (repository *PostgresRepository) List(ctx context.Context, ownerID string, status *planningdomain.FinancialItemStatus) ([]debtdomain.Debt, error) {
	rows, err := repository.pool.Query(ctx, debtSelect+`
		where d.user_id = $1
			and ($2::financial_item_status is null or item.status = $2)
		order by d.scheduled_end_month, d.financial_item_id
	`, ownerID, enumValue(status))
	if err != nil {
		return nil, fmt.Errorf("list debts: %w", err)
	}
	defer rows.Close()

	debts := []debtdomain.Debt{}
	for rows.Next() {
		debt, scanErr := scanDebt(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan listed debt: %w", scanErr)
		}
		debts = append(debts, debt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list debts: %w", err)
	}
	return debts, nil
}

func (repository *PostgresRepository) Find(ctx context.Context, ownerID, debtID string) (debtdomain.Debt, error) {
	debt, err := scanDebt(repository.pool.QueryRow(ctx, debtSelect+`
		where d.user_id = $1 and d.financial_item_id = $2
	`, ownerID, debtID))
	if errors.Is(err, pgx.ErrNoRows) {
		return debtdomain.Debt{}, debtapplication.ErrNotFound
	}
	if err != nil {
		return debtdomain.Debt{}, fmt.Errorf("find debt: %w", err)
	}
	return debt, nil
}

func (repository *PostgresRepository) Update(ctx context.Context, ownerID, debtID string, input debtapplication.UpdateInput) (debtdomain.Debt, error) {
	result, err := repository.pool.Exec(ctx, `
		update public.financial_items as item
		set name = coalesce($3, item.name),
			description = case when $4 then $5 else item.description end
		where item.id = $1 and item.user_id = $2 and item.status = 'active'
			and exists (
				select 1 from public.debts debt
				where debt.financial_item_id = item.id and debt.user_id = item.user_id
			)
	`, debtID, ownerID, nullable(input.Name), input.DescriptionSet, nullable(input.Description))
	if err != nil {
		return debtdomain.Debt{}, mapError("update debt", err)
	}
	if result.RowsAffected() == 0 {
		return debtdomain.Debt{}, repository.missingOrArchived(ctx, ownerID, debtID)
	}
	return repository.Find(ctx, ownerID, debtID)
}

func (repository *PostgresRepository) Change(ctx context.Context, ownerID, debtID string, input debtapplication.ChangeInput) (debtdomain.Debt, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return debtdomain.Debt{}, fmt.Errorf("begin debt amount change: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	var scheduledStart, scheduledEnd time.Time
	var settlementMonth *time.Time
	err = tx.QueryRow(ctx, `
		select item.status, debt.scheduled_start_month, debt.scheduled_end_month,
			(select settlement.reference_month
			 from public.debt_early_settlements settlement
			 where settlement.financial_item_id = debt.financial_item_id
			   and settlement.user_id = debt.user_id)
		from public.debts debt
		join public.financial_items item
			on item.id = debt.financial_item_id and item.user_id = debt.user_id
		where debt.financial_item_id = $1 and debt.user_id = $2
		for update of debt, item
	`, debtID, ownerID).Scan(&status, &scheduledStart, &scheduledEnd, &settlementMonth)
	if errors.Is(err, pgx.ErrNoRows) {
		return debtdomain.Debt{}, debtapplication.ErrNotFound
	}
	if err != nil {
		return debtdomain.Debt{}, fmt.Errorf("lock debt for amount change: %w", err)
	}
	if status == string(planningdomain.FinancialItemStatusArchived) {
		return debtdomain.Debt{}, debtapplication.ErrArchived
	}
	effectiveTime := input.EffectiveFrom.Time()
	if effectiveTime.Before(scheduledStart) || effectiveTime.After(scheduledEnd) {
		return debtdomain.Debt{}, debtapplication.ErrChangeOutsideSchedule
	}
	if settlementMonth != nil && effectiveTime.After(*settlementMonth) {
		return debtdomain.Debt{}, debtapplication.ErrChangeAfterSettlement
	}

	var applicablePeriodID string
	var applicableStart time.Time
	var cashMonthOffset int
	err = tx.QueryRow(ctx, `
		select id, start_month, cash_month_offset
		from public.financial_item_periods
		where financial_item_id = $1 and user_id = $2
			and start_month <= $3 and end_month >= $3
		for update
	`, debtID, ownerID, effectiveTime).Scan(&applicablePeriodID, &applicableStart, &cashMonthOffset)
	if errors.Is(err, pgx.ErrNoRows) {
		return debtdomain.Debt{}, debtdomain.ErrDebtPeriodGap
	}
	if err != nil {
		return debtdomain.Debt{}, fmt.Errorf("lock applicable debt period: %w", err)
	}

	if sameMonth(applicableStart, input.EffectiveFrom) {
		_, err = tx.Exec(ctx, `
			delete from public.financial_item_periods
			where financial_item_id = $1 and user_id = $2 and start_month >= $3
		`, debtID, ownerID, effectiveTime)
	} else {
		previousMonth, previousErr := input.EffectiveFrom.AddMonths(-1)
		if previousErr != nil {
			return debtdomain.Debt{}, debtapplication.ErrValidation
		}
		_, err = tx.Exec(ctx, `
			update public.financial_item_periods set end_month = $2
			where id = $1
		`, applicablePeriodID, previousMonth.Time())
		if err == nil {
			_, err = tx.Exec(ctx, `
				delete from public.financial_item_periods
				where financial_item_id = $1 and user_id = $2 and start_month >= $3
			`, debtID, ownerID, effectiveTime)
		}
	}
	if err != nil {
		return debtdomain.Debt{}, mapError("replace future debt periods", err)
	}

	_, err = tx.Exec(ctx, `
		insert into public.financial_item_periods (
			financial_item_id, user_id, start_month, end_month,
			amount_cents, recurrence, cash_month_offset, context, recorded_at
		) values ($1, $2, $3, $4, $5, 'monthly', $6, $7, now())
	`, debtID, ownerID, effectiveTime, scheduledEnd, input.InstallmentAmountCents, cashMonthOffset, input.Context)
	if err != nil {
		return debtdomain.Debt{}, mapError("create changed debt period", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return debtdomain.Debt{}, mapError("commit debt amount change", err)
	}
	return repository.Find(ctx, ownerID, debtID)
}

func (repository *PostgresRepository) Archive(ctx context.Context, ownerID, debtID string) (debtdomain.Debt, error) {
	result, err := repository.pool.Exec(ctx, `
		update public.financial_items as item
		set status = 'archived', archived_at = now()
		where item.id = $1 and item.user_id = $2 and item.status = 'active'
			and exists (
				select 1 from public.debts debt
				where debt.financial_item_id = item.id and debt.user_id = item.user_id
			)
	`, debtID, ownerID)
	if err != nil {
		return debtdomain.Debt{}, fmt.Errorf("archive debt: %w", err)
	}
	if result.RowsAffected() == 0 {
		return debtdomain.Debt{}, repository.missingOrArchived(ctx, ownerID, debtID)
	}
	return repository.Find(ctx, ownerID, debtID)
}

func (repository *PostgresRepository) PaymentContext(ctx context.Context, ownerID, debtID string) (debtapplication.PaymentContext, error) {
	rows, err := repository.pool.Query(ctx, `
		select id, start_month, end_month, method, credit_card_id
		from public.financial_item_payment_periods
		where financial_item_id = $1 and user_id = $2
		order by start_month, id
	`, debtID, ownerID)
	if err != nil {
		return debtapplication.PaymentContext{}, fmt.Errorf("list debt payment periods: %w", err)
	}
	periods := []cardinvoice.PaymentMethodPeriod{}
	for rows.Next() {
		var id, rawMethod string
		var start time.Time
		var end *time.Time
		var cardID *string
		if err := rows.Scan(&id, &start, &end, &rawMethod, &cardID); err != nil {
			rows.Close()
			return debtapplication.PaymentContext{}, fmt.Errorf("scan debt payment period: %w", err)
		}
		interval, intervalErr := monthInterval(start, end)
		if intervalErr != nil {
			rows.Close()
			return debtapplication.PaymentContext{}, intervalErr
		}
		method, parseErr := cardinvoice.ParsePaymentMethod(rawMethod)
		if parseErr != nil {
			rows.Close()
			return debtapplication.PaymentContext{}, parseErr
		}
		selectedCardID := ""
		if cardID != nil {
			selectedCardID = *cardID
		}
		period, periodErr := cardinvoice.NewPaymentMethodPeriod(id, interval, method, selectedCardID)
		if periodErr != nil {
			rows.Close()
			return debtapplication.PaymentContext{}, periodErr
		}
		periods = append(periods, period)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return debtapplication.PaymentContext{}, fmt.Errorf("list debt payment periods: %w", err)
	}
	rows.Close()

	cardRows, err := repository.pool.Query(ctx, `
		select configuration.id, configuration.credit_card_id,
			configuration.start_month, configuration.end_month,
			configuration.nominal_due_day, configuration.payment_month_offset
		from public.credit_card_periods configuration
		where configuration.user_id = $1
			and configuration.credit_card_id in (
				select distinct payment.credit_card_id
				from public.financial_item_payment_periods payment
				where payment.financial_item_id = $2 and payment.user_id = $1
					and payment.credit_card_id is not null
			)
		order by configuration.credit_card_id, configuration.start_month, configuration.id
	`, ownerID, debtID)
	if err != nil {
		return debtapplication.PaymentContext{}, fmt.Errorf("list debt card configurations: %w", err)
	}
	defer cardRows.Close()
	configurations := make(map[string][]cardinvoice.CardConfigurationPeriod)
	for cardRows.Next() {
		var id, cardID string
		var start time.Time
		var end *time.Time
		var dueDay, offset int
		if err := cardRows.Scan(&id, &cardID, &start, &end, &dueDay, &offset); err != nil {
			return debtapplication.PaymentContext{}, fmt.Errorf("scan debt card configuration: %w", err)
		}
		interval, intervalErr := monthInterval(start, end)
		if intervalErr != nil {
			return debtapplication.PaymentContext{}, intervalErr
		}
		configuration, configurationErr := cardinvoice.NewCardConfiguration(dueDay, offset)
		if configurationErr != nil {
			return debtapplication.PaymentContext{}, configurationErr
		}
		period, periodErr := cardinvoice.NewCardConfigurationPeriod(id, interval, configuration)
		if periodErr != nil {
			return debtapplication.PaymentContext{}, periodErr
		}
		configurations[cardID] = append(configurations[cardID], period)
	}
	if err := cardRows.Err(); err != nil {
		return debtapplication.PaymentContext{}, fmt.Errorf("list debt card configurations: %w", err)
	}
	return debtapplication.PaymentContext{Periods: periods, CardConfigurations: configurations}, nil
}

func (repository *PostgresRepository) missingOrArchived(ctx context.Context, ownerID, debtID string) error {
	var status string
	err := repository.pool.QueryRow(ctx, `
		select item.status
		from public.financial_items item
		join public.debts debt
			on debt.financial_item_id = item.id and debt.user_id = item.user_id
		where item.id = $1 and item.user_id = $2
	`, debtID, ownerID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return debtapplication.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("find debt state: %w", err)
	}
	if status == string(planningdomain.FinancialItemStatusArchived) {
		return debtapplication.ErrArchived
	}
	return debtapplication.ErrNotFound
}

const debtSelect = `
	select
		d.financial_item_id, d.user_id, item.currency_code, item.name,
		item.description, d.original_total_cents, d.total_installments,
		d.first_projected_installment, d.scheduled_start_month,
		d.scheduled_end_month, item.status, item.archived_at,
		item.created_at, item.updated_at,
		coalesce((
			select jsonb_agg(jsonb_build_object(
				'id', period.id,
				'start_month', to_char(period.start_month, 'YYYY-MM'),
				'end_month', to_char(period.end_month, 'YYYY-MM'),
				'amount_cents', period.amount_cents,
				'cash_month_offset', period.cash_month_offset,
				'context', period.context
			) order by period.start_month, period.id)
			from public.financial_item_periods period
			where period.financial_item_id = d.financial_item_id
				and period.user_id = d.user_id
		), '[]'::jsonb),
		coalesce((
			select jsonb_build_object(
				'id', settlement.id,
				'reference_month', to_char(settlement.reference_month, 'YYYY-MM'),
				'amount_cents', settlement.amount_cents,
				'reason', settlement.reason,
				'recorded_by', settlement.recorded_by,
				'recorded_at', settlement.recorded_at,
				'created_at', settlement.created_at
			)
			from public.debt_early_settlements settlement
			where settlement.financial_item_id = d.financial_item_id
				and settlement.user_id = d.user_id
		), 'null'::jsonb)
	from public.debts d
	join public.financial_items item
		on item.id = d.financial_item_id and item.user_id = d.user_id
`

type periodJSON struct {
	ID              string  `json:"id"`
	StartMonth      string  `json:"start_month"`
	EndMonth        string  `json:"end_month"`
	AmountCents     int64   `json:"amount_cents"`
	CashMonthOffset int     `json:"cash_month_offset"`
	Context         *string `json:"context"`
}

type settlementJSON struct {
	ID             string    `json:"id"`
	ReferenceMonth string    `json:"reference_month"`
	AmountCents    int64     `json:"amount_cents"`
	Reason         *string   `json:"reason"`
	RecordedBy     string    `json:"recorded_by"`
	RecordedAt     time.Time `json:"recorded_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type rowScanner interface{ Scan(...any) error }

func scanDebt(row rowScanner) (debtdomain.Debt, error) {
	var rawStatus string
	var originalTotal *int64
	var scheduledStart, scheduledEnd time.Time
	var rawPeriods, rawSettlement []byte
	var result debtdomain.Debt
	if err := row.Scan(
		&result.ID, &result.UserID, &result.CurrencyCode, &result.Name,
		&result.Description, &originalTotal, &result.TotalInstallments,
		&result.FirstProjectedInstallment, &scheduledStart, &scheduledEnd,
		&rawStatus, &result.ArchivedAt, &result.CreatedAt, &result.UpdatedAt,
		&rawPeriods, &rawSettlement,
	); err != nil {
		return debtdomain.Debt{}, err
	}
	status, err := planningdomain.ParseFinancialItemStatus(rawStatus)
	if err != nil {
		return debtdomain.Debt{}, err
	}
	start, err := planningdomain.NewYearMonth(scheduledStart.Year(), scheduledStart.Month())
	if err != nil {
		return debtdomain.Debt{}, err
	}
	storedEnd, err := planningdomain.NewYearMonth(scheduledEnd.Year(), scheduledEnd.Month())
	if err != nil {
		return debtdomain.Debt{}, err
	}

	var encodedPeriods []periodJSON
	if err := json.Unmarshal(rawPeriods, &encodedPeriods); err != nil {
		return debtdomain.Debt{}, fmt.Errorf("decode debt periods: %w", err)
	}
	periods := make([]debtdomain.InstallmentPeriod, 0, len(encodedPeriods))
	for _, encoded := range encodedPeriods {
		periodStart, parseErr := planningdomain.ParseYearMonth(encoded.StartMonth)
		if parseErr != nil {
			return debtdomain.Debt{}, parseErr
		}
		periodEnd, parseErr := planningdomain.ParseYearMonth(encoded.EndMonth)
		if parseErr != nil {
			return debtdomain.Debt{}, parseErr
		}
		interval, intervalErr := planningdomain.NewMonthInterval(periodStart, periodEnd)
		if intervalErr != nil {
			return debtdomain.Debt{}, intervalErr
		}
		periods = append(periods, debtdomain.InstallmentPeriod{
			ID: encoded.ID, Interval: interval, Amount: planningdomain.NewMoney(encoded.AmountCents),
			CashMonthOffset: encoded.CashMonthOffset, Context: encoded.Context,
		})
	}

	var settlement *debtdomain.EarlySettlement
	if string(rawSettlement) != "null" {
		var encoded settlementJSON
		if err := json.Unmarshal(rawSettlement, &encoded); err != nil {
			return debtdomain.Debt{}, fmt.Errorf("decode debt settlement: %w", err)
		}
		month, parseErr := planningdomain.ParseYearMonth(encoded.ReferenceMonth)
		if parseErr != nil {
			return debtdomain.Debt{}, parseErr
		}
		settlement = &debtdomain.EarlySettlement{
			ID: encoded.ID, ReferenceMonth: month, Amount: planningdomain.NewMoney(encoded.AmountCents),
			Reason: encoded.Reason, RecordedBy: encoded.RecordedBy,
			RecordedAt: encoded.RecordedAt, CreatedAt: encoded.CreatedAt,
		}
	}

	var originalTotalMoney *planningdomain.Money
	if originalTotal != nil {
		value := planningdomain.NewMoney(*originalTotal)
		originalTotalMoney = &value
	}
	debt, err := debtdomain.NewDebt(debtdomain.NewDebtInput{
		ID: result.ID, UserID: result.UserID, CurrencyCode: result.CurrencyCode,
		Name: result.Name, Description: result.Description, OriginalTotal: originalTotalMoney,
		TotalInstallments:         result.TotalInstallments,
		FirstProjectedInstallment: result.FirstProjectedInstallment,
		ScheduledStart:            start, Periods: periods, Settlement: settlement, Status: status,
		ArchivedAt: result.ArchivedAt, CreatedAt: result.CreatedAt, UpdatedAt: result.UpdatedAt,
	})
	if err != nil {
		return debtdomain.Debt{}, err
	}
	if debt.ScheduledEnd != storedEnd {
		return debtdomain.Debt{}, fmt.Errorf("%w: stored debt end does not match structure", debtdomain.ErrInvalidDebt)
	}
	return debt, nil
}

func mapError(operation string, err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23514", "23P01":
			return fmt.Errorf("%w: %s", debtdomain.ErrInvalidDebt, postgresError.Message)
		case "23503":
			return debtapplication.ErrNotFound
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
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

func sameMonth(value time.Time, month planningdomain.YearMonth) bool {
	return value.Year() == month.Year() && value.Month() == month.Month()
}

func monthInterval(start time.Time, end *time.Time) (planningdomain.MonthInterval, error) {
	startMonth, err := planningdomain.NewYearMonth(start.Year(), start.Month())
	if err != nil {
		return planningdomain.MonthInterval{}, err
	}
	if end == nil {
		return planningdomain.NewOpenMonthInterval(startMonth)
	}
	endMonth, err := planningdomain.NewYearMonth(end.Year(), end.Month())
	if err != nil {
		return planningdomain.MonthInterval{}, err
	}
	return planningdomain.NewMonthInterval(startMonth, endMonth)
}
