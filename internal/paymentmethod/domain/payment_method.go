package domain

import (
	"errors"
	"strings"
	"time"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

var (
	ErrValidation = errors.New("invalid payment method input")
	ErrOverlap    = errors.New("payment method period overlaps another period")
)

type Kind string

const (
	Direct     Kind = "direct"
	CreditCard Kind = "credit_card"
)

func (kind Kind) Valid() bool { return kind == Direct || kind == CreditCard }

type Period struct {
	ID           string
	StartMonth   planning.YearMonth
	EndMonth     *planning.YearMonth
	Method       Kind
	CreditCardID *string
	Context      *string
	RecordedAt   time.Time
	CreatedAt    time.Time
}

func Validate(method Kind, cardID *string) error {
	if !method.Valid() || (method == CreditCard && (cardID == nil || strings.TrimSpace(*cardID) == "")) || (method == Direct && cardID != nil) {
		return ErrValidation
	}
	return nil
}
