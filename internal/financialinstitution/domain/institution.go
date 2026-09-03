package domain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaximumNameLength   = 120
	MaximumReasonLength = 500
)

var ErrValidation = errors.New("invalid financial institution input")

type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

func (status Status) Valid() bool {
	return status == StatusActive || status == StatusArchived
}

func ParseStatus(value string) (Status, error) {
	status := Status(value)
	if !status.Valid() {
		return "", fmt.Errorf("%w: invalid status %q", ErrValidation, value)
	}
	return status, nil
}

type Institution struct {
	ID         string
	UserID     string
	Name       string
	Status     Status
	ArchivedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NormalizeName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" || utf8.RuneCountInString(name) > MaximumNameLength {
		return "", fmt.Errorf("%w: name must contain between 1 and %d characters", ErrValidation, MaximumNameLength)
	}
	return name, nil
}

func ValidateReason(reason *string) error {
	if reason != nil && utf8.RuneCountInString(*reason) > MaximumReasonLength {
		return fmt.Errorf("%w: reason must contain at most %d characters", ErrValidation, MaximumReasonLength)
	}
	return nil
}

func ValidID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	compact := strings.ReplaceAll(value, "-", "")
	decoded, err := hex.DecodeString(compact)
	return err == nil && len(decoded) == 16
}
