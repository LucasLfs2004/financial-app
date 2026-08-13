package domain

// Money represents a signed monetary value in cents. Input amounts that must
// be non-negative are validated by the domain object that consumes Money.
type Money struct {
	cents int64
}

func NewMoney(cents int64) Money {
	return Money{cents: cents}
}

func ZeroMoney() Money {
	return Money{}
}

func (m Money) Cents() int64 {
	return m.cents
}

func (m Money) IsZero() bool {
	return m.cents == 0
}

func (m Money) IsNegative() bool {
	return m.cents < 0
}

func (m Money) Add(other Money) (Money, error) {
	if other.cents > 0 && m.cents > maxInt64-other.cents {
		return Money{}, ErrMoneyOverflow
	}
	if other.cents < 0 && m.cents < minInt64-other.cents {
		return Money{}, ErrMoneyOverflow
	}

	return NewMoney(m.cents + other.cents), nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if other.cents > 0 && m.cents < minInt64+other.cents {
		return Money{}, ErrMoneyOverflow
	}
	if other.cents < 0 && m.cents > maxInt64+other.cents {
		return Money{}, ErrMoneyOverflow
	}

	return NewMoney(m.cents - other.cents), nil
}

func (m Money) Negate() (Money, error) {
	if m.cents == minInt64 {
		return Money{}, ErrMoneyOverflow
	}

	return NewMoney(-m.cents), nil
}

const (
	maxInt64 = int64(^uint64(0) >> 1)
	minInt64 = -maxInt64 - 1
)
