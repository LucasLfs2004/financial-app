package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	name, err := NormalizeName("  Banco Alfa  ")
	if err != nil || name != "Banco Alfa" {
		t.Fatalf("name=%q error=%v", name, err)
	}
	if _, err := NormalizeName(strings.Repeat("á", MaximumNameLength)); err != nil {
		t.Fatalf("multibyte name error=%v", err)
	}

	for _, value := range []string{"", "   ", strings.Repeat("a", MaximumNameLength+1)} {
		if _, err := NormalizeName(value); !errors.Is(err, ErrValidation) {
			t.Fatalf("value=%q error=%v", value, err)
		}
	}
}

func TestStatus(t *testing.T) {
	for _, value := range []string{"active", "archived"} {
		status, err := ParseStatus(value)
		if err != nil || !status.Valid() {
			t.Fatalf("value=%q status=%q error=%v", value, status, err)
		}
	}
	if _, err := ParseStatus("disabled"); !errors.Is(err, ErrValidation) {
		t.Fatalf("error=%v", err)
	}
}

func TestValidID(t *testing.T) {
	if !ValidID("a1100000-0000-0000-0000-000000000001") {
		t.Fatal("expected valid UUID")
	}
	for _, value := range []string{"", "institution", "a1100000-0000-0000-0000-00000000000z"} {
		if ValidID(value) {
			t.Fatalf("expected invalid UUID %q", value)
		}
	}
}

func TestValidateReason(t *testing.T) {
	reason := strings.Repeat("a", MaximumReasonLength+1)
	if err := ValidateReason(&reason); !errors.Is(err, ErrValidation) {
		t.Fatalf("error=%v", err)
	}
	if err := ValidateReason(nil); err != nil {
		t.Fatalf("error=%v", err)
	}
}
