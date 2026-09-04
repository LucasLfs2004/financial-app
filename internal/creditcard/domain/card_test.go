package domain_test

import (
	"strings"
	"testing"

	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
	planningdomain "github.com/lucas/financial-api/internal/planning/domain"
)

func TestValidateConfigurationBoundaries(t *testing.T) {
	month, err := planningdomain.ParseYearMonth("2026-01")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		due       int
		offset    int
		context   *string
		wantError bool
	}{
		{name: "accepts day 31 and offset 12", due: 31, offset: 12},
		{name: "accepts day 1 and offset 0", due: 1, offset: 0},
		{name: "rejects day 32", due: 32, offset: 1, wantError: true},
		{name: "rejects negative offset", due: 10, offset: -1, wantError: true},
		{name: "rejects long context", due: 10, offset: 1, context: stringPointer(strings.Repeat("a", 501)), wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := carddomain.ValidateConfiguration(carddomain.ConfigurationInput{
				EffectiveFrom: month, NominalDueDay: test.due,
				PaymentMonthOffset: test.offset, Context: test.context,
			})
			if (err != nil) != test.wantError {
				t.Fatalf("error=%v wantError=%v", err, test.wantError)
			}
		})
	}
}

func TestValidateConfigurationRejectsEndBeforeStart(t *testing.T) {
	start, _ := planningdomain.ParseYearMonth("2026-05")
	end, _ := planningdomain.ParseYearMonth("2026-04")
	err := carddomain.ValidateConfiguration(carddomain.ConfigurationInput{
		EffectiveFrom: start, EndMonth: &end, NominalDueDay: 10, PaymentMonthOffset: 1,
	})
	if err == nil {
		t.Fatal("expected invalid interval")
	}
}

func stringPointer(value string) *string { return &value }
