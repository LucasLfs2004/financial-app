package domain

import (
	"errors"
	planning "github.com/lucas/financial-api/internal/planning/domain"
	"strings"
	"time"
)

var ErrValidation = errors.New("invalid invoice move")

type Move struct {
	ID               string
	ItemID           string
	ReferenceMonth   planning.YearMonth
	FromCardID       string
	FromPaymentMonth planning.YearMonth
	ToCardID         string
	ToPaymentMonth   planning.YearMonth
	Reason           *string
	RecordedAt       time.Time
}

func Validate(itemID string, reference planning.YearMonth, targetCard string, targetMonth planning.YearMonth) error {
	if strings.TrimSpace(itemID) == "" || !reference.Valid() || strings.TrimSpace(targetCard) == "" || !targetMonth.Valid() {
		return ErrValidation
	}
	return nil
}
