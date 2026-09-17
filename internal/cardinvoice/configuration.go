package cardinvoice

import (
	"fmt"
	"strings"
	"time"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

const (
	MinimumNominalDueDay = 1
	MaximumNominalDueDay = 31
)

type CardConfiguration struct {
	nominalDueDay      int
	paymentMonthOffset int
}

func NewCardConfiguration(nominalDueDay, paymentMonthOffset int) (CardConfiguration, error) {
	if nominalDueDay < MinimumNominalDueDay || nominalDueDay > MaximumNominalDueDay {
		return CardConfiguration{}, fmt.Errorf("%w: nominal due day %d must be between %d and %d", ErrInvalidConfiguration, nominalDueDay, MinimumNominalDueDay, MaximumNominalDueDay)
	}
	if paymentMonthOffset < planning.MinimumCashMonthOffset || paymentMonthOffset > planning.MaximumCashMonthOffset {
		return CardConfiguration{}, fmt.Errorf("%w: payment month offset %d must be between %d and %d", ErrInvalidConfiguration, paymentMonthOffset, planning.MinimumCashMonthOffset, planning.MaximumCashMonthOffset)
	}
	return CardConfiguration{nominalDueDay: nominalDueDay, paymentMonthOffset: paymentMonthOffset}, nil
}

func (configuration CardConfiguration) NominalDueDay() int { return configuration.nominalDueDay }

func (configuration CardConfiguration) PaymentMonthOffset() int {
	return configuration.paymentMonthOffset
}

func (configuration CardConfiguration) PaymentMonth(referenceMonth planning.YearMonth) (planning.YearMonth, error) {
	if !referenceMonth.Valid() {
		return planning.YearMonth{}, fmt.Errorf("%w: invalid reference month", ErrInvalidConfiguration)
	}
	return referenceMonth.AddMonths(configuration.paymentMonthOffset)
}

type CardConfigurationPeriod struct {
	ID            string
	Interval      planning.MonthInterval
	Configuration CardConfiguration
}

func NewCardConfigurationPeriod(id string, interval planning.MonthInterval, configuration CardConfiguration) (CardConfigurationPeriod, error) {
	if strings.TrimSpace(id) == "" || !interval.Valid() || configuration.nominalDueDay < MinimumNominalDueDay || configuration.nominalDueDay > MaximumNominalDueDay ||
		configuration.paymentMonthOffset < planning.MinimumCashMonthOffset || configuration.paymentMonthOffset > planning.MaximumCashMonthOffset {
		return CardConfigurationPeriod{}, ErrInvalidConfiguration
	}
	return CardConfigurationPeriod{ID: id, Interval: interval, Configuration: configuration}, nil
}

func SelectCardConfiguration(month planning.YearMonth, periods []CardConfigurationPeriod) (CardConfiguration, error) {
	if !month.Valid() {
		return CardConfiguration{}, ErrInvalidConfiguration
	}
	var selected *CardConfiguration
	for _, period := range periods {
		if strings.TrimSpace(period.ID) == "" || !period.Interval.Valid() {
			return CardConfiguration{}, ErrInvalidConfiguration
		}
		if _, err := NewCardConfiguration(period.Configuration.nominalDueDay, period.Configuration.paymentMonthOffset); err != nil {
			return CardConfiguration{}, err
		}
		if !period.Interval.Contains(month) {
			continue
		}
		if selected != nil {
			return CardConfiguration{}, ErrPeriodOverlap
		}
		configuration := period.Configuration
		selected = &configuration
	}
	if selected == nil {
		return CardConfiguration{}, ErrCardConfigurationMissing
	}
	return *selected, nil
}

type NominalDueDate struct {
	Day        int
	Date       *time.Time
	Resolution DueDateResolution
}

func ResolveNominalDueDate(paymentMonth planning.YearMonth, nominalDueDay int) (NominalDueDate, error) {
	if !paymentMonth.Valid() || nominalDueDay < MinimumNominalDueDay || nominalDueDay > MaximumNominalDueDay {
		return NominalDueDate{}, ErrInvalidConfiguration
	}
	lastDay := time.Date(paymentMonth.Year(), paymentMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if nominalDueDay > lastDay {
		return NominalDueDate{Day: nominalDueDay, Resolution: DueDateResolutionInvalidForMonth}, nil
	}
	date := time.Date(paymentMonth.Year(), paymentMonth.Month(), nominalDueDay, 0, 0, 0, 0, time.UTC)
	return NominalDueDate{Day: nominalDueDay, Date: &date, Resolution: DueDateResolutionExact}, nil
}
