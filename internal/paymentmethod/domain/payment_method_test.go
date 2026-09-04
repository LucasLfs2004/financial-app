package domain

import (
	"testing"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

func TestValidatePaymentMethodCompatibility(t *testing.T) {
	card := "card-1"
	if err := Validate(CreditCard, &card); err != nil {
		t.Fatal(err)
	}
	if err := Validate(Direct, nil); err != nil {
		t.Fatal(err)
	}
	if Validate(CreditCard, nil) == nil || Validate(Direct, &card) == nil {
		t.Fatal("expected incompatible methods to fail")
	}
}

func TestPeriodCarriesInclusiveMonths(t *testing.T) {
	start, _ := planning.ParseYearMonth("2026-08")
	end, _ := planning.ParseYearMonth("2026-12")
	p := Period{StartMonth: start, EndMonth: &end, Method: Direct}
	if p.StartMonth.String() != "2026-08" || p.EndMonth.String() != "2026-12" {
		t.Fatalf("unexpected period: %+v", p)
	}
}
