package domain

import "errors"

var (
	ErrMoneyOverflow          = errors.New("money operation overflows int64")
	ErrInvalidYearMonth       = errors.New("invalid year month")
	ErrInvalidMonthInterval   = errors.New("invalid month interval")
	ErrInvalidEnumValue       = errors.New("invalid enum value")
	ErrNegativeAmount         = errors.New("amount cannot be negative")
	ErrInvalidRecurrence      = errors.New("invalid recurrence for interval")
	ErrInvalidCashMonthOffset = errors.New("cash month offset must be between 0 and 12")
)
