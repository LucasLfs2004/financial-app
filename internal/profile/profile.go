package profile

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("profile not found")

type Profile struct {
	ID           string
	DisplayName  *string
	Timezone     string
	CurrencyCode string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (repository *Repository) FindByID(ctx context.Context, userID string) (Profile, error) {
	const query = `
		select id, display_name, timezone, currency_code, created_at, updated_at
		from public.profiles
		where id = $1
	`

	var result Profile
	err := repository.pool.QueryRow(ctx, query, userID).Scan(
		&result.ID,
		&result.DisplayName,
		&result.Timezone,
		&result.CurrencyCode,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	if err != nil {
		return Profile{}, fmt.Errorf("find profile: %w", err)
	}

	return result, nil
}
