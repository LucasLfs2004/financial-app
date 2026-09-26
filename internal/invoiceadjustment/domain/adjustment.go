package domain

import (
	"errors"
	planning "github.com/lucas/financial-api/internal/planning/domain"
	"strings"
	"time"
)

var ErrValidation = errors.New("invalid invoice adjustment")

type Status string

const (
	Active   Status = "active"
	Archived Status = "archived"
)

type Adjustment struct {
	ID, UserID, CurrencyCode, CardID, Name string
	PaymentMonth                           planning.YearMonth
	ReferenceMonth                         *planning.YearMonth
	AmountCents                            int64
	Context                                *string
	Status                                 Status
	ArchivedAt                             *time.Time
	CreatedAt, UpdatedAt                   time.Time
}

func Validate(name string, amount int64, card string, payment planning.YearMonth, reference *planning.YearMonth) error {
	if strings.TrimSpace(name) == "" || len(name) > 120 || amount < 0 || strings.TrimSpace(card) == "" || !payment.Valid() {
		return ErrValidation
	}
	if reference != nil && !reference.Valid() {
		return ErrValidation
	}
	return nil
}
