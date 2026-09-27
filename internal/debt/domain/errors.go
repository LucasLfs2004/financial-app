package domain

import "errors"

var (
	ErrInvalidDebt            = errors.New("invalid debt")
	ErrDebtScheduleTooLong    = errors.New("debt schedule exceeds 120 months")
	ErrDebtPeriodGap          = errors.New("debt periods contain a gap")
	ErrDebtPeriodOverlap      = errors.New("debt periods overlap")
	ErrInvalidSettlement      = errors.New("invalid early settlement")
	ErrInvalidProjectionRange = errors.New("invalid projection range")
)
