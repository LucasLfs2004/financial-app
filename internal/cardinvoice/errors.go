package cardinvoice

import "errors"

var (
	ErrInvalidEnumValue         = errors.New("invalid card invoice enum value")
	ErrInvalidConfiguration     = errors.New("invalid credit card configuration")
	ErrCardConfigurationMissing = errors.New("credit card configuration missing")
	ErrInvalidPaymentMethod     = errors.New("invalid financial item payment method")
	ErrPeriodOverlap            = errors.New("card invoice period overlap")
	ErrInvalidProjectionInput   = errors.New("invalid card invoice projection input")
	ErrCardNotFound             = errors.New("credit card not found")
	ErrDuplicateOccurrence      = errors.New("duplicate financial item occurrence")
	ErrDuplicateAdjustment      = errors.New("duplicate invoice adjustment")
	ErrNegativeAmount           = errors.New("card invoice amount cannot be negative")
	ErrInconsistentProjection   = errors.New("inconsistent card invoice projection")
)
