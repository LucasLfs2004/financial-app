package domain

import (
	"errors"
	"testing"
)

func TestMoneyArithmetic(t *testing.T) {
	tests := []struct {
		name string
		got  func() (Money, error)
		want int64
	}{
		{
			name: "add",
			got:  func() (Money, error) { return NewMoney(150).Add(NewMoney(25)) },
			want: 175,
		},
		{
			name: "subtract into negative result",
			got:  func() (Money, error) { return NewMoney(100).Subtract(NewMoney(125)) },
			want: -25,
		},
		{
			name: "negate",
			got:  func() (Money, error) { return NewMoney(75).Negate() },
			want: -75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.got()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Cents() != tt.want {
				t.Fatalf("got %d cents, want %d", got.Cents(), tt.want)
			}
		})
	}
}

func TestMoneyZero(t *testing.T) {
	zero := ZeroMoney()
	if !zero.IsZero() {
		t.Fatal("zero money must report IsZero")
	}
	if zero.IsNegative() {
		t.Fatal("zero money cannot be negative")
	}
}

func TestMoneyRejectsOverflow(t *testing.T) {
	tests := []struct {
		name string
		run  func() (Money, error)
	}{
		{"add above maximum", func() (Money, error) { return NewMoney(maxInt64).Add(NewMoney(1)) }},
		{"add below minimum", func() (Money, error) { return NewMoney(minInt64).Add(NewMoney(-1)) }},
		{"subtract below minimum", func() (Money, error) { return NewMoney(minInt64).Subtract(NewMoney(1)) }},
		{"subtract above maximum", func() (Money, error) { return NewMoney(maxInt64).Subtract(NewMoney(-1)) }},
		{"negate minimum", func() (Money, error) { return NewMoney(minInt64).Negate() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.run()
			if !errors.Is(err, ErrMoneyOverflow) {
				t.Fatalf("got error %v, want ErrMoneyOverflow", err)
			}
		})
	}
}
