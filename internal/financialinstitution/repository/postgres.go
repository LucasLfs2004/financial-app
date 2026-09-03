package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repository *PostgresRepository) Create(ctx context.Context, ownerID, name string) (institutiondomain.Institution, error) {
	institution, err := scanInstitution(repository.pool.QueryRow(ctx, `
		insert into public.financial_institutions (user_id, name)
		values ($1, $2)
		returning id, user_id, name, status, archived_at, created_at, updated_at
	`, ownerID, name))
	if err != nil {
		return institutiondomain.Institution{}, mapError("create financial institution", err)
	}
	return institution, nil
}

func (repository *PostgresRepository) List(ctx context.Context, ownerID string, status *institutiondomain.Status) ([]institutiondomain.Institution, error) {
	rows, err := repository.pool.Query(ctx, `
		select id, user_id, name, status, archived_at, created_at, updated_at
		from public.financial_institutions
		where user_id = $1
			and ($2::financial_resource_status is null or status = $2)
		order by created_at, id
	`, ownerID, statusValue(status))
	if err != nil {
		return nil, fmt.Errorf("list financial institutions: %w", err)
	}
	defer rows.Close()

	institutions := []institutiondomain.Institution{}
	for rows.Next() {
		institution, scanErr := scanInstitution(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan financial institution: %w", scanErr)
		}
		institutions = append(institutions, institution)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list financial institutions: %w", err)
	}
	return institutions, nil
}

func (repository *PostgresRepository) UpdateName(ctx context.Context, ownerID, institutionID, name string) (institutiondomain.Institution, error) {
	institution, err := scanInstitution(repository.pool.QueryRow(ctx, `
		update public.financial_institutions
		set name = $3
		where id = $1 and user_id = $2 and status = 'active'
		returning id, user_id, name, status, archived_at, created_at, updated_at
	`, institutionID, ownerID, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return institutiondomain.Institution{}, repository.missingOrArchived(ctx, ownerID, institutionID)
	}
	if err != nil {
		return institutiondomain.Institution{}, mapError("update financial institution", err)
	}
	return institution, nil
}

func (repository *PostgresRepository) Archive(ctx context.Context, ownerID, institutionID string) (institutiondomain.Institution, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return institutiondomain.Institution{}, fmt.Errorf("begin financial institution archive: %w", err)
	}
	defer tx.Rollback(ctx)

	institution, err := scanInstitution(tx.QueryRow(ctx, `
		select id, user_id, name, status, archived_at, created_at, updated_at
		from public.financial_institutions
		where id = $1 and user_id = $2
		for update
	`, institutionID, ownerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return institutiondomain.Institution{}, institutionapplication.ErrNotFound
	}
	if err != nil {
		return institutiondomain.Institution{}, fmt.Errorf("lock financial institution: %w", err)
	}
	if institution.Status == institutiondomain.StatusArchived {
		return institutiondomain.Institution{}, institutionapplication.ErrArchived
	}

	var hasActiveCards bool
	if err := tx.QueryRow(ctx, `
		select exists (
			select 1
			from public.credit_cards
			where institution_id = $1 and user_id = $2 and status = 'active'
		)
	`, institutionID, ownerID).Scan(&hasActiveCards); err != nil {
		return institutiondomain.Institution{}, fmt.Errorf("check active institution cards: %w", err)
	}
	if hasActiveCards {
		return institutiondomain.Institution{}, institutionapplication.ErrHasActiveCards
	}

	institution, err = scanInstitution(tx.QueryRow(ctx, `
		update public.financial_institutions
		set status = 'archived', archived_at = now()
		where id = $1 and user_id = $2 and status = 'active'
		returning id, user_id, name, status, archived_at, created_at, updated_at
	`, institutionID, ownerID))
	if err != nil {
		return institutiondomain.Institution{}, fmt.Errorf("archive financial institution: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return institutiondomain.Institution{}, fmt.Errorf("commit financial institution archive: %w", err)
	}
	return institution, nil
}

func (repository *PostgresRepository) missingOrArchived(ctx context.Context, ownerID, institutionID string) error {
	var status string
	err := repository.pool.QueryRow(ctx, `
		select status
		from public.financial_institutions
		where id = $1 and user_id = $2
	`, institutionID, ownerID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return institutionapplication.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("find financial institution state: %w", err)
	}
	if status == string(institutiondomain.StatusArchived) {
		return institutionapplication.ErrArchived
	}
	return institutionapplication.ErrNotFound
}

type rowScanner interface {
	Scan(...any) error
}

func scanInstitution(row rowScanner) (institutiondomain.Institution, error) {
	var institution institutiondomain.Institution
	var rawStatus string
	if err := row.Scan(
		&institution.ID,
		&institution.UserID,
		&institution.Name,
		&rawStatus,
		&institution.ArchivedAt,
		&institution.CreatedAt,
		&institution.UpdatedAt,
	); err != nil {
		return institutiondomain.Institution{}, err
	}
	status, err := institutiondomain.ParseStatus(rawStatus)
	if err != nil {
		return institutiondomain.Institution{}, fmt.Errorf("parse financial institution status: %w", err)
	}
	institution.Status = status
	return institution, nil
}

func mapError(operation string, err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return institutionapplication.ErrAlreadyExists
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func statusValue(status *institutiondomain.Status) any {
	if status == nil {
		return nil
	}
	return string(*status)
}
