package cardinvoice

import (
	"strings"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

type PaymentMethodPeriod struct {
	ID       string
	Interval planning.MonthInterval
	Method   PaymentMethod
	CardID   string
}

func NewPaymentMethodPeriod(id string, interval planning.MonthInterval, method PaymentMethod, cardID string) (PaymentMethodPeriod, error) {
	if strings.TrimSpace(id) == "" || !interval.Valid() || !method.Valid() {
		return PaymentMethodPeriod{}, ErrInvalidPaymentMethod
	}
	if (method == PaymentMethodCreditCard && strings.TrimSpace(cardID) == "") || (method == PaymentMethodDirect && strings.TrimSpace(cardID) != "") {
		return PaymentMethodPeriod{}, ErrInvalidPaymentMethod
	}
	return PaymentMethodPeriod{ID: id, Interval: interval, Method: method, CardID: cardID}, nil
}

type PaymentMethodSelection struct {
	Method   PaymentMethod
	CardID   string
	Explicit bool
}

func SelectPaymentMethod(month planning.YearMonth, periods []PaymentMethodPeriod) (PaymentMethodSelection, error) {
	if !month.Valid() {
		return PaymentMethodSelection{}, ErrInvalidPaymentMethod
	}
	selection := PaymentMethodSelection{Method: PaymentMethodDirect}
	for _, period := range periods {
		if _, err := NewPaymentMethodPeriod(period.ID, period.Interval, period.Method, period.CardID); err != nil {
			return PaymentMethodSelection{}, err
		}
		if !period.Interval.Contains(month) {
			continue
		}
		if selection.Explicit {
			return PaymentMethodSelection{}, ErrPeriodOverlap
		}
		selection = PaymentMethodSelection{Method: period.Method, CardID: period.CardID, Explicit: true}
	}
	return selection, nil
}
