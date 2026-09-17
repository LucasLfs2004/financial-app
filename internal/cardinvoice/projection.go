package cardinvoice

import (
	"fmt"
	"sort"
	"strings"
	"time"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

type Occurrence struct {
	SourceID            string
	ItemID              string
	Name                string
	ReferenceMonth      planning.YearMonth
	DefaultCardID       string
	DefaultPaymentMonth planning.YearMonth
	Amount              planning.Money
}

type OccurrenceIdentity struct {
	ItemID         string
	ReferenceMonth planning.YearMonth
}

func NewOccurrenceIdentity(itemID string, referenceMonth planning.YearMonth) (OccurrenceIdentity, error) {
	if strings.TrimSpace(itemID) == "" || !referenceMonth.Valid() {
		return OccurrenceIdentity{}, ErrInvalidProjectionInput
	}
	return OccurrenceIdentity{ItemID: itemID, ReferenceMonth: referenceMonth}, nil
}

func (identity OccurrenceIdentity) Key() string {
	return occurrenceKey(identity.ItemID, identity.ReferenceMonth)
}

type Adjustment struct {
	ID             string
	Name           string
	CardID         string
	PaymentMonth   planning.YearMonth
	ReferenceMonth *planning.YearMonth
	Amount         planning.Money
	Status         AdjustmentStatus
}

type OccurrenceMove struct {
	ID             string
	ItemID         string
	ReferenceMonth planning.YearMonth
	ToCardID       string
	ToPaymentMonth planning.YearMonth
	RecordedAt     time.Time
}

type ProjectionInput struct {
	CardID        string
	CardName      string
	Institution   Institution
	PaymentMonth  planning.YearMonth
	CurrencyCode  string
	NominalDueDay int
	Occurrences   []Occurrence
	Adjustments   []Adjustment
	Moves         []OccurrenceMove
}

type Component struct {
	SourceID       string
	SourceType     ComponentType
	ItemID         *string
	AdjustmentID   *string
	Name           string
	ReferenceMonth *planning.YearMonth
	ReferenceKnown bool
	PaymentMonth   planning.YearMonth
	CardID         string
	Amount         planning.Money
	Allocation     AllocationOrigin
}

type Invoice struct {
	CardID         string
	CardName       string
	Institution    Institution
	PaymentMonth   planning.YearMonth
	CurrencyCode   string
	NominalDueDate NominalDueDate
	Total          planning.Money
	Components     []Component
}

func ProjectInvoice(input ProjectionInput) (Invoice, error) {
	if strings.TrimSpace(input.CardID) == "" || !input.PaymentMonth.Valid() || !validCurrencyCode(input.CurrencyCode) {
		return Invoice{}, ErrInvalidProjectionInput
	}
	dueDate, err := ResolveNominalDueDate(input.PaymentMonth, input.NominalDueDay)
	if err != nil {
		return Invoice{}, fmt.Errorf("%w: %v", ErrInvalidProjectionInput, err)
	}
	components, err := projectComponents(input.Occurrences, input.Adjustments, input.Moves, func(cardID string, paymentMonth planning.YearMonth) bool {
		return cardID == input.CardID && paymentMonth == input.PaymentMonth
	})
	if err != nil {
		return Invoice{}, err
	}
	total := planning.ZeroMoney()
	for _, component := range components {
		total, err = total.Add(component.Amount)
		if err != nil {
			return Invoice{}, fmt.Errorf("%w: total overflow", ErrInconsistentProjection)
		}
	}
	return Invoice{
		CardID: input.CardID, CardName: input.CardName, Institution: input.Institution, PaymentMonth: input.PaymentMonth,
		CurrencyCode: input.CurrencyCode, NominalDueDate: dueDate, Total: total,
		Components: components,
	}, nil
}

func projectComponents(occurrences []Occurrence, adjustments []Adjustment, moves []OccurrenceMove, include func(string, planning.YearMonth) bool) ([]Component, error) {
	latestMoves, err := selectLatestMoves(moves)
	if err != nil {
		return nil, err
	}
	components := make([]Component, 0, len(occurrences)+len(adjustments))
	seenOccurrences := make(map[string]struct{}, len(occurrences))
	for _, occurrence := range occurrences {
		key, validationError := validateOccurrence(occurrence)
		if validationError != nil {
			return nil, validationError
		}
		if _, duplicate := seenOccurrences[key]; duplicate {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateOccurrence, key)
		}
		seenOccurrences[key] = struct{}{}
		cardID := occurrence.DefaultCardID
		paymentMonth := occurrence.DefaultPaymentMonth
		allocation := AllocationCalculatedFromReference
		if move, moved := latestMoves[key]; moved {
			cardID = move.ToCardID
			paymentMonth = move.ToPaymentMonth
			allocation = AllocationMovedByUser
		}
		if !include(cardID, paymentMonth) {
			continue
		}
		itemID := occurrence.ItemID
		referenceMonth := occurrence.ReferenceMonth
		components = append(components, Component{
			SourceID: occurrence.SourceID, SourceType: ComponentTypeFinancialItemOccurrence,
			ItemID: &itemID, Name: occurrence.Name, ReferenceMonth: &referenceMonth,
			ReferenceKnown: true, PaymentMonth: paymentMonth, CardID: cardID,
			Amount: occurrence.Amount, Allocation: allocation,
		})
	}
	seenAdjustments := make(map[string]struct{}, len(adjustments))
	for _, adjustment := range adjustments {
		if err := validateAdjustment(adjustment); err != nil {
			return nil, err
		}
		if _, duplicate := seenAdjustments[adjustment.ID]; duplicate {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateAdjustment, adjustment.ID)
		}
		seenAdjustments[adjustment.ID] = struct{}{}
		if adjustment.Status == AdjustmentStatusArchived || !include(adjustment.CardID, adjustment.PaymentMonth) {
			continue
		}
		adjustmentID := adjustment.ID
		components = append(components, Component{
			SourceID: adjustment.ID, SourceType: ComponentTypeInvoiceAdjustment,
			AdjustmentID: &adjustmentID, Name: adjustment.Name,
			ReferenceMonth: adjustment.ReferenceMonth, ReferenceKnown: adjustment.ReferenceMonth != nil,
			PaymentMonth: adjustment.PaymentMonth, CardID: adjustment.CardID,
			Amount: adjustment.Amount, Allocation: AllocationSelectedInvoice,
		})
	}
	sortComponents(components)
	return components, nil
}

func validateOccurrence(occurrence Occurrence) (string, error) {
	if strings.TrimSpace(occurrence.SourceID) == "" || strings.TrimSpace(occurrence.ItemID) == "" ||
		strings.TrimSpace(occurrence.Name) == "" || strings.TrimSpace(occurrence.DefaultCardID) == "" ||
		!occurrence.ReferenceMonth.Valid() || !occurrence.DefaultPaymentMonth.Valid() {
		return "", ErrInvalidProjectionInput
	}
	if occurrence.Amount.IsNegative() {
		return "", ErrNegativeAmount
	}
	identity, err := NewOccurrenceIdentity(occurrence.ItemID, occurrence.ReferenceMonth)
	if err != nil {
		return "", err
	}
	return identity.Key(), nil
}

func validateAdjustment(adjustment Adjustment) error {
	if strings.TrimSpace(adjustment.ID) == "" || strings.TrimSpace(adjustment.Name) == "" ||
		strings.TrimSpace(adjustment.CardID) == "" || !adjustment.PaymentMonth.Valid() || !adjustment.Status.Valid() {
		return ErrInvalidProjectionInput
	}
	if adjustment.ReferenceMonth != nil && !adjustment.ReferenceMonth.Valid() {
		return ErrInvalidProjectionInput
	}
	if adjustment.Amount.IsNegative() {
		return ErrNegativeAmount
	}
	return nil
}

func selectLatestMoves(moves []OccurrenceMove) (map[string]OccurrenceMove, error) {
	latest := make(map[string]OccurrenceMove, len(moves))
	for _, move := range moves {
		if strings.TrimSpace(move.ID) == "" || strings.TrimSpace(move.ItemID) == "" ||
			strings.TrimSpace(move.ToCardID) == "" || !move.ReferenceMonth.Valid() ||
			!move.ToPaymentMonth.Valid() || move.RecordedAt.IsZero() {
			return nil, ErrInvalidProjectionInput
		}
		key := occurrenceKey(move.ItemID, move.ReferenceMonth)
		current, exists := latest[key]
		if !exists || move.RecordedAt.After(current.RecordedAt) || (move.RecordedAt.Equal(current.RecordedAt) && move.ID > current.ID) {
			latest[key] = move
		}
	}
	return latest, nil
}

func occurrenceKey(itemID string, referenceMonth planning.YearMonth) string {
	return itemID + ":" + referenceMonth.String()
}

func sortComponents(components []Component) {
	sort.Slice(components, func(left, right int) bool {
		if componentRank(components[left].SourceType) != componentRank(components[right].SourceType) {
			return componentRank(components[left].SourceType) < componentRank(components[right].SourceType)
		}
		if components[left].Name != components[right].Name {
			return components[left].Name < components[right].Name
		}
		return components[left].SourceID < components[right].SourceID
	})
}

func componentRank(componentType ComponentType) int {
	if componentType == ComponentTypeFinancialItemOccurrence {
		return 0
	}
	return 1
}

func validCurrencyCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, character := range value {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}
