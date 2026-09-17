package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucas/financial-api/internal/cardinvoice"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// LoadProjectionData uses a repeatable-read snapshot so a projection never
// combines financial premises with card changes committed midway through the
// read.
func (repository *PostgresRepository) LoadProjectionData(ctx context.Context, ownerID, planID string) (cardinvoice.ProjectionData, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return cardinvoice.ProjectionData{}, fmt.Errorf("begin card invoice projection read: %w", err)
	}
	defer tx.Rollback(ctx)

	cards, err := loadCards(ctx, tx, ownerID)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	items, itemIndexes, err := loadItems(ctx, tx, ownerID, planID)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	if err := loadPaymentPeriods(ctx, tx, ownerID, planID, items, itemIndexes); err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	moves, err := loadMoves(ctx, tx, ownerID, planID)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	adjustments, err := loadAdjustments(ctx, tx, ownerID, planID)
	if err != nil {
		return cardinvoice.ProjectionData{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return cardinvoice.ProjectionData{}, fmt.Errorf("commit card invoice projection read: %w", err)
	}
	return cardinvoice.ProjectionData{Cards: cards, Items: items, Moves: moves, Adjustments: adjustments}, nil
}

func loadCards(ctx context.Context, tx pgx.Tx, ownerID string) ([]cardinvoice.Card, error) {
	rows, err := tx.Query(ctx, `
		select c.id, c.name, p.id, p.start_month, p.end_month,
			p.nominal_due_day, p.payment_month_offset
		from public.credit_cards c
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
		var periodID *string
		var start, end *time.Time
		var dueDay, offset *int
		if err := rows.Scan(&id, &name, &periodID, &start, &end, &dueDay, &offset); err != nil {
			return nil, fmt.Errorf("scan invoice card: %w", err)
		}
		index, exists := indexes[id]
		if !exists {
			index = len(cards)
			indexes[id] = index
			cards = append(cards, cardinvoice.Card{ID: id, Name: name, Configurations: []cardinvoice.CardConfigurationPeriod{}})
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

func loadItems(ctx context.Context, tx pgx.Tx, ownerID, planID string) ([]cardinvoice.FinancialItem, map[string]int, error) {
	rows, err := tx.Query(ctx, `
		select i.id, i.name, i.kind, p.id, p.start_month, p.end_month,
			p.amount_cents, p.recurrence
		from public.financial_items i
		join public.financial_item_periods p
			on p.financial_item_id = i.id and p.plan_id = i.plan_id and p.user_id = i.user_id
		where i.user_id = $1 and i.plan_id = $2
			and i.kind in ('fixed_expense', 'projected_variable_expense')
		order by i.created_at, i.id, p.start_month, p.id
	`, ownerID, planID)
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

func loadPaymentPeriods(ctx context.Context, tx pgx.Tx, ownerID, planID string, items []cardinvoice.FinancialItem, indexes map[string]int) error {
	rows, err := tx.Query(ctx, `
		select id, financial_item_id, start_month, end_month, method, credit_card_id
		from public.financial_item_payment_periods
		where user_id = $1 and plan_id = $2
		order by financial_item_id, start_month, id
	`, ownerID, planID)
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

func loadMoves(ctx context.Context, tx pgx.Tx, ownerID, planID string) ([]cardinvoice.OccurrenceMove, error) {
	rows, err := tx.Query(ctx, `
		select id, financial_item_id, reference_month, to_credit_card_id,
			to_payment_month, recorded_at
		from public.card_invoice_audit_events
		where user_id = $1 and plan_id = $2 and event_type = 'occurrence_moved'
		order by recorded_at, id
	`, ownerID, planID)
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

func loadAdjustments(ctx context.Context, tx pgx.Tx, ownerID, planID string) ([]cardinvoice.Adjustment, error) {
	rows, err := tx.Query(ctx, `
		select id, name, credit_card_id, payment_month, reference_month,
			amount_cents, status
		from public.card_invoice_adjustments
		where user_id = $1 and plan_id = $2
		order by created_at, id
	`, ownerID, planID)
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
