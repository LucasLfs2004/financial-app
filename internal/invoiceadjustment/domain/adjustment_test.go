package domain

import (
	planning "github.com/lucas/financial-api/internal/planning/domain"
	"testing"
)

func TestValidateAllowsZeroAndUnknownReference(t *testing.T) {
	m, _ := planning.ParseYearMonth("2026-12")
	if err := Validate("Taxa", 0, "card", m, nil); err != nil {
		t.Fatal(err)
	}
}
func TestValidateRejectsNegativeAmount(t *testing.T) {
	m, _ := planning.ParseYearMonth("2026-12")
	if Validate("Taxa", -1, "card", m, nil) == nil {
		t.Fatal("expected negative amount rejection")
	}
}
