package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lucas/financial-api/internal/cardinvoice"
	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

const (
	MaximumNameLength    = 120
	MaximumContextLength = 500
	MaximumReasonLength  = 500
)

var ErrValidation = errors.New("invalid credit card input")

type Status = institutiondomain.Status

const (
	StatusActive   = institutiondomain.StatusActive
	StatusArchived = institutiondomain.StatusArchived
)

type Configuration struct {
	ID                 string
	StartMonth         planningdomain.YearMonth
	EndMonth           *planningdomain.YearMonth
	NominalDueDay      int
	PaymentMonthOffset int
	Context            *string
	RecordedAt         time.Time
	CreatedAt          time.Time
}

type ConfigurationInput struct {
	EffectiveFrom      planningdomain.YearMonth
	EndMonth           *planningdomain.YearMonth
	NominalDueDay      int
	PaymentMonthOffset int
	Context            *string
}

type Card struct {
	ID             string
	UserID         string
	InstitutionID  string
	Institution    institutiondomain.Institution
	Name           string
	Status         Status
	ArchivedAt     *time.Time
	Configurations []Configuration
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NormalizeName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" || utf8.RuneCountInString(name) > MaximumNameLength {
		return "", fmt.Errorf("%w: name must contain between 1 and %d characters", ErrValidation, MaximumNameLength)
	}
	return name, nil
}

func ValidateConfiguration(input ConfigurationInput) error {
	if !input.EffectiveFrom.Valid() {
		return fmt.Errorf("%w: effective month is invalid", ErrValidation)
	}
	if input.EndMonth != nil && (!input.EndMonth.Valid() || input.EndMonth.Before(input.EffectiveFrom)) {
		return fmt.Errorf("%w: end month must not precede effective month", ErrValidation)
	}
	if _, err := cardinvoice.NewCardConfiguration(input.NominalDueDay, input.PaymentMonthOffset); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	if input.Context != nil && utf8.RuneCountInString(*input.Context) > MaximumContextLength {
		return fmt.Errorf("%w: context must contain at most %d characters", ErrValidation, MaximumContextLength)
	}
	return nil
}

func ValidateReason(reason *string) error {
	if reason != nil && utf8.RuneCountInString(*reason) > MaximumReasonLength {
		return fmt.Errorf("%w: reason must contain at most %d characters", ErrValidation, MaximumReasonLength)
	}
	return nil
}

func ValidID(value string) bool {
	return institutiondomain.ValidID(value)
}
