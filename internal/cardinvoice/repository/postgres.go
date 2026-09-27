package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/cardinvoice"
	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// LoadProjectionData uses a repeatable-read snapshot so a projection never
// combines financial premises with card changes committed midway through the
// read.
func (repository *PostgresRepository) LoadProjectionData(ctx context.Context, ownerID, currencyCode string) (cardinvoice.ProjectionData, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return cardinvoice.ProjectionData{}, fmt.Errorf("begin card invoice projection read: %w", err)
	}
	defer tx.Rollback(ctx)

	cards, err := loadCards(ctx, tx, ownerID)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	items, itemIndexes, err := loadItems(ctx, tx, ownerID, currencyCode)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	if err := loadPaymentPeriods(ctx, tx, ownerID, currencyCode, items, itemIndexes); err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	debtOccurrences, err := loadDebtOccurrences(ctx, tx, ownerID, currencyCode)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	moves, err := loadMoves(ctx, tx, ownerID, currencyCode)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	adjustments, err := loadAdjustments(ctx, tx, ownerID, currencyCode)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return cardinvoice.ProjectionData{}, fmt.Errorf("commit card invoice projection read: %w", err)
	}
	return cardinvoice.ProjectionData{
		Cards: cards, Items: items, DebtOccurrences: debtOccurrences,
		Moves: moves, Adjustments: adjustments,
	}, nil
}

func loadCards(ctx context.Context, tx pgx.Tx, ownerID string) ([]cardinvoice.Card, error) {
	rows, err := tx.Query(ctx, `
		select c.id, c.name, i.id, i.name, i.status, i.archived_at, i.created_at, i.updated_at,
			p.id, p.start_month, p.end_month,
			p.nominal_due_day, p.payment_month_offset
		from public.credit_cards c
		join public.financial_institutions i
			on i.id = c.institution_id and i.user_id = c.user_id
		left join public.credit_card_periods p
			on p.credit_card_id = c.id and p.user_id = c.user_id
		where c.user_id = $1
		order by c.created_at, c.id, p.start_month, p.id
	`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("load invoice cards: %w", err)
	}
	defer rows.Close()

	cards := []cardinvoice.Card{}
	indexes := make(map[string]int)
	for rows.Next() {
		var id, name string
		var institution cardinvoice.Institution
		var institutionStatus string
		var periodID *string
		var start, end *time.Time
		var dueDay, offset *int
		if err := rows.Scan(
			&id, &name, &institution.ID, &institution.Name, &institutionStatus,
			&institution.ArchivedAt, &institution.CreatedAt, &institution.UpdatedAt,
			&periodID, &start, &end, &dueDay, &offset,
		); err != nil {
			return nil, fmt.Errorf("scan invoice card: %w", err)
		}
		institution.Status, err = cardinvoice.ParseFinancialResourceStatus(institutionStatus)
		if err != nil {
			return nil, fmt.Errorf("parse invoice institution %s status: %w", institution.ID, err)
		}
		index, exists := indexes[id]
		if !exists {
			index = len(cards)
			indexes[id] = index
			cards = append(cards, cardinvoice.Card{ID: id, Name: name, Institution: institution, Configurations: []cardinvoice.CardConfigurationPeriod{}})
		}
		if periodID == nil {
			continue
		}
		if start == nil || dueDay == nil || offset == nil {
			return nil, fmt.Errorf("load invoice cards: incomplete configuration %s", *periodID)
		}
		interval, err := monthInterval(*start, end)
		if err != nil {
			return nil, fmt.Errorf("parse card configuration %s: %w", *periodID, err)
		}
		configuration, err := cardinvoice.NewCardConfiguration(*dueDay, *offset)
		if err != nil {
			return nil, fmt.Errorf("parse card configuration %s: %w", *periodID, err)
		}
		period, err := cardinvoice.NewCardConfigurationPeriod(*periodID, interval, configuration)
		if err != nil {
			return nil, fmt.Errorf("parse card configuration %s: %w", *periodID, err)
		}
		cards[index].Configurations = append(cards[index].Configurations, period)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load invoice cards: %w", err)
	}
	return cards, nil
}

func loadItems(ctx context.Context, tx pgx.Tx, ownerID, currencyCode string) ([]cardinvoice.FinancialItem, map[string]int, error) {
	rows, err := tx.Query(ctx, `
		select i.id, i.name, i.kind, p.id, p.start_month, p.end_month,
			p.amount_cents, p.recurrence
		from public.financial_items i
		join public.financial_item_periods p
			on p.financial_item_id = i.id and p.user_id = i.user_id
		where i.user_id = $1 and i.currency_code = $2
			and i.kind in ('fixed_expense', 'projected_variable_expense', 'debt_installment')
		order by i.created_at, i.id, p.start_month, p.id
	`, ownerID, currencyCode)
	if err != nil {
		return nil, nil, fmt.Errorf("load invoice financial items: %w", err)
	}
	defer rows.Close()

	items := []cardinvoice.FinancialItem{}
	indexes := make(map[string]int)
	for rows.Next() {
		var itemID, name, kindValue, periodID, recurrenceValue string
		var start time.Time
		var end *time.Time
		var amount int64
		if err := rows.Scan(&itemID, &name, &kindValue, &periodID, &start, &end, &amount, &recurrenceValue); err != nil {
			return nil, nil, fmt.Errorf("scan invoice financial item: %w", err)
		}
		kind, err := planningdomain.ParseFinancialItemKind(kindValue)
		if err != nil {
			return nil, nil, fmt.Errorf("parse invoice item %s kind: %w", itemID, err)
		}
		recurrence, err := planningdomain.ParseRecurrence(recurrenceValue)
		if err != nil {
			return nil, nil, fmt.Errorf("parse invoice item %s recurrence: %w", itemID, err)
		}
		interval, err := monthInterval(start, end)
		if err != nil {
			return nil, nil, fmt.Errorf("parse invoice item period %s: %w", periodID, err)
		}
		index, exists := indexes[itemID]
		if !exists {
			index = len(items)
			indexes[itemID] = index
			items = append(items, cardinvoice.FinancialItem{ID: itemID, Name: name, Kind: kind, Periods: []cardinvoice.FinancialItemPeriod{}, PaymentPeriods: []cardinvoice.PaymentMethodPeriod{}})
		}
		items[index].Periods = append(items[index].Periods, cardinvoice.FinancialItemPeriod{
			ID: periodID, Interval: interval, Amount: planningdomain.NewMoney(amount), Recurrence: recurrence,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("load invoice financial items: %w", err)
	}
	return items, indexes, nil
}

type debtProjectionBuilder struct {
	ID                        string
	Name                      string
	TotalInstallments         int
	FirstProjectedInstallment int
	ScheduledStart            planningdomain.YearMonth
	StoredEnd                 planningdomain.YearMonth
	Periods                   []debtdomain.InstallmentPeriod
	Settlement                *debtdomain.EarlySettlement
}

// loadDebtOccurrences reads every debt and its temporal periods in one query
// inside the surrounding repeatable-read transaction, then delegates all
// numbering and settlement cutoffs to the pure debt projector.
func loadDebtOccurrences(ctx context.Context, tx pgx.Tx, ownerID, currencyCode string) ([]cardinvoice.DebtOccurrence, error) {
	rows, err := tx.Query(ctx, `
		select debt.financial_item_id, item.name,
			debt.total_installments, debt.first_projected_installment,
			debt.scheduled_start_month, debt.scheduled_end_month,
			period.id, period.start_month, period.end_month,
			period.amount_cents, period.cash_month_offset,
			settlement.id, settlement.reference_month, settlement.amount_cents
		from public.debts debt
		join public.financial_items item
			on item.id = debt.financial_item_id and item.user_id = debt.user_id
		join public.financial_item_periods period
			on period.financial_item_id = debt.financial_item_id
			and period.user_id = debt.user_id
		left join public.debt_early_settlements settlement
			on settlement.financial_item_id = debt.financial_item_id
			and settlement.user_id = debt.user_id
		where debt.user_id = $1 and item.currency_code = $2
			and item.status = 'active'
		order by item.created_at, debt.financial_item_id,
			period.start_month, period.id
	`, ownerID, currencyCode)
	if err != nil {
		return nil, fmt.Errorf("load invoice debt occurrences: %w", err)
	}
	defer rows.Close()

	builders := make(map[string]*debtProjectionBuilder)
	order := make([]string, 0)
	for rows.Next() {
		var debtID, name, periodID string
		var total, first, cashOffset int
		var scheduledStart, scheduledEnd, periodStart, periodEnd time.Time
		var amount int64
		var settlementID *string
		var settlementMonth *time.Time
		var settlementAmount *int64
		if err := rows.Scan(
			&debtID, &name, &total, &first, &scheduledStart, &scheduledEnd,
			&periodID, &periodStart, &periodEnd, &amount, &cashOffset,
			&settlementID, &settlementMonth, &settlementAmount,
		); err != nil {
			return nil, fmt.Errorf("scan invoice debt occurrence premise: %w", err)
		}
		builder, exists := builders[debtID]
		if !exists {
			start, parseErr := yearMonth(scheduledStart)
			if parseErr != nil {
				return nil, fmt.Errorf("parse debt %s start: %w", debtID, parseErr)
			}
			end, parseErr := yearMonth(scheduledEnd)
			if parseErr != nil {
				return nil, fmt.Errorf("parse debt %s end: %w", debtID, parseErr)
			}
			builder = &debtProjectionBuilder{
				ID: debtID, Name: name, TotalInstallments: total,
				FirstProjectedInstallment: first, ScheduledStart: start,
				StoredEnd: end, Periods: []debtdomain.InstallmentPeriod{},
			}
			if settlementID != nil && settlementMonth != nil && settlementAmount != nil {
				month, monthErr := yearMonth(*settlementMonth)
				if monthErr != nil {
					return nil, fmt.Errorf("parse debt %s settlement: %w", debtID, monthErr)
				}
				builder.Settlement = &debtdomain.EarlySettlement{
					ID: *settlementID, ReferenceMonth: month,
					Amount: planningdomain.NewMoney(*settlementAmount),
				}
			}
			builders[debtID] = builder
			order = append(order, debtID)
		}
		interval, intervalErr := monthInterval(periodStart, &periodEnd)
		if intervalErr != nil {
			return nil, fmt.Errorf("parse debt period %s: %w", periodID, intervalErr)
		}
		builder.Periods = append(builder.Periods, debtdomain.InstallmentPeriod{
			ID: periodID, Interval: interval, Amount: planningdomain.NewMoney(amount),
			CashMonthOffset: cashOffset,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load invoice debt occurrences: %w", err)
	}

	result := make([]cardinvoice.DebtOccurrence, 0)
	for _, debtID := range order {
		builder := builders[debtID]
		debt, buildErr := debtdomain.NewDebt(debtdomain.NewDebtInput{
			ID: builder.ID, UserID: ownerID, CurrencyCode: currencyCode,
			Name: builder.Name, TotalInstallments: builder.TotalInstallments,
			FirstProjectedInstallment: builder.FirstProjectedInstallment,
			ScheduledStart:            builder.ScheduledStart, Periods: builder.Periods,
			Settlement: builder.Settlement, Status: planningdomain.FinancialItemStatusActive,
		})
		if buildErr != nil {
			return nil, fmt.Errorf("build invoice debt %s: %w", debtID, buildErr)
		}
		if debt.ScheduledEnd != builder.StoredEnd {
			return nil, fmt.Errorf("build invoice debt %s: stored end is inconsistent", debtID)
		}
		projection, projectErr := debtdomain.ProjectDebt(debtdomain.ProjectionInput{
			Debt: debt, From: debt.ScheduledStart, To: debt.ScheduledEnd,
			AsOf: debt.ScheduledStart,
		})
		if projectErr != nil {
			return nil, fmt.Errorf("project invoice debt %s: %w", debtID, projectErr)
		}
		for _, occurrence := range projection.Occurrences {
			result = append(result, cardinvoice.DebtOccurrence{
				DebtID: debt.ID, SourceID: occurrence.SourceID, Name: debt.Name,
				ReferenceMonth:    occurrence.ReferenceMonth,
				InstallmentNumber: occurrence.InstallmentNumber,
				InstallmentsTotal: occurrence.InstallmentsTotal,
				Amount:            occurrence.Amount, Kind: occurrence.Kind,
			})
		}
	}
	return result, nil
}

func loadPaymentPeriods(ctx context.Context, tx pgx.Tx, ownerID, currencyCode string, items []cardinvoice.FinancialItem, indexes map[string]int) error {
	rows, err := tx.Query(ctx, `
		select pp.id, pp.financial_item_id, pp.start_month, pp.end_month,
			pp.method, pp.credit_card_id
		from public.financial_item_payment_periods pp
		join public.financial_items i
			on i.id = pp.financial_item_id and i.user_id = pp.user_id
		where pp.user_id = $1 and i.currency_code = $2
		order by pp.financial_item_id, pp.start_month, pp.id
	`, ownerID, currencyCode)
	if err != nil {
		return fmt.Errorf("load invoice payment periods: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, itemID, methodValue string
		var start time.Time
		var end *time.Time
		var cardID *string
		if err := rows.Scan(&id, &itemID, &start, &end, &methodValue, &cardID); err != nil {
			return fmt.Errorf("scan invoice payment period: %w", err)
		}
		index, exists := indexes[itemID]
		if !exists {
			return fmt.Errorf("load invoice payment periods: item %s is not a projected expense", itemID)
		}
		method, err := cardinvoice.ParsePaymentMethod(methodValue)
		if err != nil {
			return fmt.Errorf("parse invoice payment period %s: %w", id, err)
		}
		interval, err := monthInterval(start, end)
		if err != nil {
			return fmt.Errorf("parse invoice payment period %s: %w", id, err)
		}
		linkedCard := ""
		if cardID != nil {
			linkedCard = *cardID
		}
		period, err := cardinvoice.NewPaymentMethodPeriod(id, interval, method, linkedCard)
		if err != nil {
			return fmt.Errorf("parse invoice payment period %s: %w", id, err)
		}
		items[index].PaymentPeriods = append(items[index].PaymentPeriods, period)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("load invoice payment periods: %w", err)
	}
	return nil
}

func loadMoves(ctx context.Context, tx pgx.Tx, ownerID, currencyCode string) ([]cardinvoice.OccurrenceMove, error) {
	rows, err := tx.Query(ctx, `
		select ae.id, ae.financial_item_id, ae.reference_month,
			ae.to_credit_card_id, ae.to_payment_month, ae.recorded_at
		from public.card_invoice_audit_events ae
		join public.financial_items i
			on i.id = ae.financial_item_id and i.user_id = ae.user_id
		where ae.user_id = $1 and i.currency_code = $2
			and ae.event_type = 'occurrence_moved'
		order by ae.recorded_at, ae.id
	`, ownerID, currencyCode)
	if err != nil {
		return nil, fmt.Errorf("load invoice occurrence moves: %w", err)
	}
	defer rows.Close()
	moves := []cardinvoice.OccurrenceMove{}
	for rows.Next() {
		var move cardinvoice.OccurrenceMove
		var referenceMonth, paymentMonth time.Time
		if err := rows.Scan(&move.ID, &move.ItemID, &referenceMonth, &move.ToCardID, &paymentMonth, &move.RecordedAt); err != nil {
			return nil, fmt.Errorf("scan invoice occurrence move: %w", err)
		}
		var err error
		move.ReferenceMonth, err = yearMonth(referenceMonth)
		if err != nil {
			return nil, fmt.Errorf("parse invoice occurrence move %s reference: %w", move.ID, err)
		}
		move.ToPaymentMonth, err = yearMonth(paymentMonth)
		if err != nil {
			return nil, fmt.Errorf("parse invoice occurrence move %s destination: %w", move.ID, err)
		}
		moves = append(moves, move)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load invoice occurrence moves: %w", err)
	}
	return moves, nil
}

func loadAdjustments(ctx context.Context, tx pgx.Tx, ownerID, currencyCode string) ([]cardinvoice.Adjustment, error) {
	rows, err := tx.Query(ctx, `
		select id, name, credit_card_id, payment_month, reference_month,
			amount_cents, status
		from public.card_invoice_adjustments
		where user_id = $1 and currency_code = $2
		order by created_at, id
	`, ownerID, currencyCode)
	if err != nil {
		return nil, fmt.Errorf("load invoice adjustments: %w", err)
	}
	defer rows.Close()
	adjustments := []cardinvoice.Adjustment{}
	for rows.Next() {
		var adjustment cardinvoice.Adjustment
		var paymentMonth time.Time
		var referenceMonth *time.Time
		var amount int64
		var statusValue string
		if err := rows.Scan(&adjustment.ID, &adjustment.Name, &adjustment.CardID, &paymentMonth, &referenceMonth, &amount, &statusValue); err != nil {
			return nil, fmt.Errorf("scan invoice adjustment: %w", err)
		}
		var err error
		adjustment.PaymentMonth, err = yearMonth(paymentMonth)
		if err != nil {
			return nil, fmt.Errorf("parse invoice adjustment %s payment month: %w", adjustment.ID, err)
		}
		if referenceMonth != nil {
			parsed, parseErr := yearMonth(*referenceMonth)
			if parseErr != nil {
				return nil, fmt.Errorf("parse invoice adjustment %s reference month: %w", adjustment.ID, parseErr)
			}
			adjustment.ReferenceMonth = &parsed
		}
		adjustment.Amount = planningdomain.NewMoney(amount)
		adjustment.Status, err = cardinvoice.ParseAdjustmentStatus(statusValue)
		if err != nil {
			return nil, fmt.Errorf("parse invoice adjustment %s status: %w", adjustment.ID, err)
		}
		adjustments = append(adjustments, adjustment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load invoice adjustments: %w", err)
	}
	return adjustments, nil
}

func monthInterval(start time.Time, end *time.Time) (planningdomain.MonthInterval, error) {
	startMonth, err := yearMonth(start)
	if err != nil {
		return planningdomain.MonthInterval{}, err
	}
	if end == nil {
		return planningdomain.NewOpenMonthInterval(startMonth)
	}
	endMonth, err := yearMonth(*end)
	if err != nil {
		return planningdomain.MonthInterval{}, err
	}
	return planningdomain.NewMonthInterval(startMonth, endMonth)
}

func yearMonth(value time.Time) (planningdomain.YearMonth, error) {
	return planningdomain.NewYearMonth(value.Year(), value.Month())
}
