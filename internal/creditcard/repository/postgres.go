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
	cardapplication "github.com/lucas/financial-api/internal/creditcard/application"
	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repository *PostgresRepository) Create(ctx context.Context, ownerID, institutionID, name string, configuration carddomain.ConfigurationInput) (carddomain.Card, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return carddomain.Card{}, fmt.Errorf("begin credit card creation: %w", err)
	}
	defer tx.Rollback(ctx)

	var institutionStatus string
	err = tx.QueryRow(ctx, `
		select status
		from public.financial_institutions
		where id = $1 and user_id = $2
		for share
	`, institutionID, ownerID).Scan(&institutionStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return carddomain.Card{}, cardapplication.ErrInstitutionNotFound
	}
	if err != nil {
		return carddomain.Card{}, fmt.Errorf("lock credit card institution: %w", err)
	}
	if institutionStatus == string(institutiondomain.StatusArchived) {
		return carddomain.Card{}, cardapplication.ErrInstitutionArchived
	}

	var cardID string
	err = tx.QueryRow(ctx, `
		insert into public.credit_cards (user_id, institution_id, name)
		values ($1, $2, $3)
		returning id
	`, ownerID, institutionID, name).Scan(&cardID)
	if err != nil {
		return carddomain.Card{}, mapError("create credit card", err)
	}

	_, err = tx.Exec(ctx, `
		insert into public.credit_card_periods (
			credit_card_id, user_id, start_month, end_month,
			nominal_due_day, payment_month_offset, context
		) values ($1, $2, $3, $4, $5, $6, $7)
	`, cardID, ownerID, configuration.EffectiveFrom.Time(), monthTime(configuration.EndMonth), configuration.NominalDueDay, configuration.PaymentMonthOffset, configuration.Context)
	if err != nil {
		return carddomain.Card{}, mapError("create initial credit card period", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return carddomain.Card{}, fmt.Errorf("commit credit card creation: %w", err)
	}
	return repository.Find(ctx, ownerID, cardID)
}

func (repository *PostgresRepository) List(ctx context.Context, ownerID string, status *carddomain.Status) ([]carddomain.Card, error) {
	rows, err := repository.pool.Query(ctx, cardSelect+`
		where c.user_id = $1
			and ($2::financial_resource_status is null or c.status = $2)
		order by c.created_at, c.id
	`, ownerID, statusValue(status))
	if err != nil {
		return nil, fmt.Errorf("list credit cards: %w", err)
	}
	defer rows.Close()

	cards := []carddomain.Card{}
	for rows.Next() {
		card, scanErr := scanCard(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan credit card: %w", scanErr)
		}
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list credit cards: %w", err)
	}
	return cards, nil
}

func (repository *PostgresRepository) Find(ctx context.Context, ownerID, cardID string) (carddomain.Card, error) {
	card, err := scanCard(repository.pool.QueryRow(ctx, cardSelect+`
		where c.user_id = $1 and c.id = $2
	`, ownerID, cardID))
	if errors.Is(err, pgx.ErrNoRows) {
		return carddomain.Card{}, cardapplication.ErrNotFound
	}
	if err != nil {
		return carddomain.Card{}, fmt.Errorf("find credit card: %w", err)
	}
	return card, nil
}

func (repository *PostgresRepository) UpdateName(ctx context.Context, ownerID, cardID, name string) (carddomain.Card, error) {
	result, err := repository.pool.Exec(ctx, `
		update public.credit_cards
		set name = $3
		where id = $1 and user_id = $2 and status = 'active'
	`, cardID, ownerID, name)
	if err != nil {
		return carddomain.Card{}, mapError("update credit card", err)
	}
	if result.RowsAffected() == 0 {
		return carddomain.Card{}, repository.missingOrArchived(ctx, ownerID, cardID)
	}
	return repository.Find(ctx, ownerID, cardID)
}

func (repository *PostgresRepository) ChangeConfiguration(ctx context.Context, ownerID, cardID string, configuration carddomain.ConfigurationInput) (carddomain.Card, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return carddomain.Card{}, fmt.Errorf("begin credit card configuration change: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		select status from public.credit_cards
		where id = $1 and user_id = $2
		for update
	`, cardID, ownerID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return carddomain.Card{}, cardapplication.ErrNotFound
	}
	if err != nil {
		return carddomain.Card{}, fmt.Errorf("lock credit card: %w", err)
	}
	if status == string(carddomain.StatusArchived) {
		return carddomain.Card{}, cardapplication.ErrArchived
	}

	var periodID string
	var startMonth time.Time
	var previousEnd *time.Time
	err = tx.QueryRow(ctx, `
		select id, start_month, end_month
		from public.credit_card_periods
		where credit_card_id = $1 and user_id = $2
			and start_month <= $3
			and (end_month is null or end_month >= $3)
		for update
	`, cardID, ownerID, configuration.EffectiveFrom.Time()).Scan(&periodID, &startMonth, &previousEnd)
	if errors.Is(err, pgx.ErrNoRows) {
		return carddomain.Card{}, cardapplication.ErrConfigurationMissing
	}
	if err != nil {
		return carddomain.Card{}, fmt.Errorf("lock applicable credit card period: %w", err)
	}

	if sameMonth(startMonth, configuration.EffectiveFrom) {
		if _, err = tx.Exec(ctx, `delete from public.credit_card_periods where id = $1`, periodID); err != nil {
			return carddomain.Card{}, fmt.Errorf("replace credit card period: %w", err)
		}
	} else {
		previousMonth, previousErr := configuration.EffectiveFrom.AddMonths(-1)
		if previousErr != nil {
			return carddomain.Card{}, fmt.Errorf("calculate previous credit card month: %w", previousErr)
		}
		if _, err = tx.Exec(ctx, `update public.credit_card_periods set end_month = $2 where id = $1`, periodID, previousMonth.Time()); err != nil {
			return carddomain.Card{}, fmt.Errorf("close credit card period: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		insert into public.credit_card_periods (
			credit_card_id, user_id, start_month, end_month,
			nominal_due_day, payment_month_offset, context
		) values ($1, $2, $3, $4, $5, $6, $7)
	`, cardID, ownerID, configuration.EffectiveFrom.Time(), monthTime(configuration.EndMonth), configuration.NominalDueDay, configuration.PaymentMonthOffset, configuration.Context)
	if err != nil {
		return carddomain.Card{}, mapError("create credit card configuration change", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return carddomain.Card{}, fmt.Errorf("commit credit card configuration change: %w", err)
	}
	return repository.Find(ctx, ownerID, cardID)
}

func (repository *PostgresRepository) Archive(ctx context.Context, ownerID, cardID string) (carddomain.Card, error) {
	result, err := repository.pool.Exec(ctx, `
		update public.credit_cards
		set status = 'archived', archived_at = now()
		where id = $1 and user_id = $2 and status = 'active'
	`, cardID, ownerID)
	if err != nil {
		return carddomain.Card{}, fmt.Errorf("archive credit card: %w", err)
	}
	if result.RowsAffected() == 0 {
		return carddomain.Card{}, repository.missingOrArchived(ctx, ownerID, cardID)
	}
	return repository.Find(ctx, ownerID, cardID)
}

func (repository *PostgresRepository) missingOrArchived(ctx context.Context, ownerID, cardID string) error {
	var status string
	err := repository.pool.QueryRow(ctx, `select status from public.credit_cards where id = $1 and user_id = $2`, cardID, ownerID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return cardapplication.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("find credit card state: %w", err)
	}
	if status == string(carddomain.StatusArchived) {
		return cardapplication.ErrArchived
	}
	return cardapplication.ErrNotFound
}

const cardSelect = `
	select
		c.id, c.user_id, c.institution_id, c.name, c.status, c.archived_at, c.created_at, c.updated_at,
		i.id, i.user_id, i.name, i.status, i.archived_at, i.created_at, i.updated_at,
		coalesce((
			select jsonb_agg(jsonb_build_object(
				'id', p.id,
				'start_month', to_char(p.start_month, 'YYYY-MM'),
				'end_month', case when p.end_month is null then null else to_char(p.end_month, 'YYYY-MM') end,
				'nominal_due_day', p.nominal_due_day,
				'payment_month_offset', p.payment_month_offset,
				'context', p.context,
				'recorded_at', p.recorded_at,
				'created_at', p.created_at
			) order by p.start_month, p.id)
			from public.credit_card_periods p
			where p.credit_card_id = c.id and p.user_id = c.user_id
		), '[]'::jsonb)
	from public.credit_cards c
	join public.financial_institutions i on i.id = c.institution_id and i.user_id = c.user_id
`

type configurationJSON struct {
	ID                 string    `json:"id"`
	StartMonth         string    `json:"start_month"`
	EndMonth           *string   `json:"end_month"`
	NominalDueDay      int       `json:"nominal_due_day"`
	PaymentMonthOffset int       `json:"payment_month_offset"`
	Context            *string   `json:"context"`
	RecordedAt         time.Time `json:"recorded_at"`
	CreatedAt          time.Time `json:"created_at"`
}

type rowScanner interface{ Scan(...any) error }

func scanCard(row rowScanner) (carddomain.Card, error) {
	var card carddomain.Card
	var cardStatus, institutionStatus string
	var rawConfigurations []byte
	err := row.Scan(
		&card.ID, &card.UserID, &card.InstitutionID, &card.Name, &cardStatus, &card.ArchivedAt, &card.CreatedAt, &card.UpdatedAt,
		&card.Institution.ID, &card.Institution.UserID, &card.Institution.Name, &institutionStatus, &card.Institution.ArchivedAt, &card.Institution.CreatedAt, &card.Institution.UpdatedAt,
		&rawConfigurations,
	)
	if err != nil {
		return carddomain.Card{}, err
	}
	parsedCardStatus, err := institutiondomain.ParseStatus(cardStatus)
	if err != nil {
		return carddomain.Card{}, err
	}
	parsedInstitutionStatus, err := institutiondomain.ParseStatus(institutionStatus)
	if err != nil {
		return carddomain.Card{}, err
	}
	card.Status = parsedCardStatus
	card.Institution.Status = parsedInstitutionStatus

	var raw []configurationJSON
	if err := json.Unmarshal(rawConfigurations, &raw); err != nil {
		return carddomain.Card{}, fmt.Errorf("decode credit card periods: %w", err)
	}
	card.Configurations = make([]carddomain.Configuration, 0, len(raw))
	for _, item := range raw {
		startMonth, err := planningdomain.ParseYearMonth(item.StartMonth)
		if err != nil {
			return carddomain.Card{}, fmt.Errorf("parse credit card start month: %w", err)
		}
		var endMonth *planningdomain.YearMonth
		if item.EndMonth != nil {
			parsed, parseErr := planningdomain.ParseYearMonth(*item.EndMonth)
			if parseErr != nil {
				return carddomain.Card{}, fmt.Errorf("parse credit card end month: %w", parseErr)
			}
			endMonth = &parsed
		}
		card.Configurations = append(card.Configurations, carddomain.Configuration{
			ID: item.ID, StartMonth: startMonth, EndMonth: endMonth,
			NominalDueDay: item.NominalDueDay, PaymentMonthOffset: item.PaymentMonthOffset,
			Context: item.Context, RecordedAt: item.RecordedAt, CreatedAt: item.CreatedAt,
		})
	}
	return card, nil
}

func mapError(operation string, err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return cardapplication.ErrAlreadyExists
		case "23P01":
			return cardapplication.ErrPeriodOverlap
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func monthTime(month *planningdomain.YearMonth) any {
	if month == nil {
		return nil
	}
	return month.Time()
}

func sameMonth(value time.Time, month planningdomain.YearMonth) bool {
	return value.Year() == month.Year() && value.Month() == month.Month()
}

func statusValue(status *carddomain.Status) any {
	if status == nil {
		return nil
	}
	return string(*status)
}
