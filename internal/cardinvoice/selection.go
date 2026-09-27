package cardinvoice

import (
	"fmt"
	"strings"
	"time"

	debtdomain "github.com/lucas/financial-api/internal/debt/domain"
	planning "github.com/lucas/financial-api/internal/planning/domain"
)

type Institution struct {
	ID         string
	Name       string
	Status     FinancialResourceStatus
	ArchivedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Card carries only the card data required by the pure invoice selector.
type Card struct {
	ID             string
	Name           string
	Institution    Institution
	Configurations []CardConfigurationPeriod
}

// FinancialItemPeriod is a persisted premise capable of producing one or more
// financial item occurrences within the plan horizon.
type FinancialItemPeriod struct {
	ID         string
	Interval   planning.MonthInterval
	Amount     planning.Money
	Recurrence planning.Recurrence
}

type FinancialItem struct {
	ID             string
	Name           string
	Kind           planning.FinancialItemKind
	Periods        []FinancialItemPeriod
	PaymentPeriods []PaymentMethodPeriod
}

// DebtOccurrence is the normalized projection produced by the debt module.
// Invoice selection only resolves payment allocation and never recalculates
// installment numbering or early-settlement cutoffs.
type DebtOccurrence struct {
	DebtID            string
	SourceID          string
	Name              string
	ReferenceMonth    planning.YearMonth
	DirectCashMonth   planning.YearMonth
	InstallmentNumber int
	InstallmentsTotal int
	Amount            planning.Money
	Kind              debtdomain.OccurrenceKind
}

// ProjectionData is the persistence-neutral source set used by application
// services to project one or more invoices consistently.
type ProjectionData struct {
	Cards           []Card
	Items           []FinancialItem
	DebtOccurrences []DebtOccurrence
	Adjustments     []Adjustment
	Moves           []OccurrenceMove
}

type SelectionInput struct {
	PlanStart       planning.YearMonth
	PlanEnd         planning.YearMonth
	CardID          string
	PaymentMonth    planning.YearMonth
	CurrencyCode    string
	Cards           []Card
	Items           []FinancialItem
	DebtOccurrences []DebtOccurrence
	Adjustments     []Adjustment
	Moves           []OccurrenceMove
}

type PaymentMonthComponentsInput struct {
	PlanStart       planning.YearMonth
	PlanEnd         planning.YearMonth
	PaymentMonth    planning.YearMonth
	Cards           []Card
	Items           []FinancialItem
	DebtOccurrences []DebtOccurrence
	Adjustments     []Adjustment
	Moves           []OccurrenceMove
}

type ReferenceMonthComponentsInput struct {
	PlanStart       planning.YearMonth
	PlanEnd         planning.YearMonth
	ReferenceMonth  planning.YearMonth
	Cards           []Card
	Items           []FinancialItem
	DebtOccurrences []DebtOccurrence
	Adjustments     []Adjustment
	Moves           []OccurrenceMove
}

// SelectInvoiceComponents expands financial premises into occurrences,
// resolves each occurrence's effective payment method and default invoice,
// and delegates final movement, adjustment, ordering and total rules to the
// projector.
func SelectInvoiceComponents(input SelectionInput) (Invoice, error) {
	if !input.PlanStart.Valid() || !input.PlanEnd.Valid() || input.PlanEnd.Before(input.PlanStart) ||
		!input.PaymentMonth.Valid() || strings.TrimSpace(input.CardID) == "" || !validCurrencyCode(input.CurrencyCode) {
		return Invoice{}, ErrInvalidProjectionInput
	}

	cards, targetCard, err := indexCards(input.Cards, input.CardID)
	if err != nil {
		return Invoice{}, err
	}
	dueConfiguration, err := SelectCardConfiguration(input.PaymentMonth, targetCard.Configurations)
	if err != nil {
		return Invoice{}, fmt.Errorf("%w: payment month configuration: %v", ErrInvalidProjectionInput, err)
	}

	occurrences, err := selectCardOccurrences(input.PlanStart, input.PlanEnd, input.Items, input.DebtOccurrences, cards)
	if err != nil {
		return Invoice{}, err
	}
	return ProjectInvoice(ProjectionInput{
		CardID:        targetCard.ID,
		CardName:      targetCard.Name,
		Institution:   targetCard.Institution,
		PaymentMonth:  input.PaymentMonth,
		CurrencyCode:  input.CurrencyCode,
		NominalDueDay: dueConfiguration.NominalDueDay(),
		Occurrences:   occurrences,
		Adjustments:   input.Adjustments,
		Moves:         input.Moves,
	})
}

// SelectPaymentMonthComponents returns the normalized components paid in one
// month across all cards. It intentionally does not resolve invoice due dates,
// which are presentation data and do not participate in monthly totals.
func SelectPaymentMonthComponents(input PaymentMonthComponentsInput) ([]Component, error) {
	if !input.PlanStart.Valid() || !input.PlanEnd.Valid() || input.PlanEnd.Before(input.PlanStart) || !input.PaymentMonth.Valid() {
		return nil, ErrInvalidProjectionInput
	}
	cards, err := indexAllCards(input.Cards)
	if err != nil {
		return nil, err
	}
	occurrences, err := selectCardOccurrences(input.PlanStart, input.PlanEnd, input.Items, input.DebtOccurrences, cards)
	if err != nil {
		return nil, err
	}
	components, err := projectComponents(occurrences, input.Adjustments, input.Moves, func(_ string, paymentMonth planning.YearMonth) bool {
		return paymentMonth == input.PaymentMonth
	})
	if err != nil {
		return nil, err
	}
	for _, component := range components {
		if _, exists := cards[component.CardID]; !exists {
			return nil, fmt.Errorf("%w: component references card %s", ErrInconsistentProjection, component.CardID)
		}
	}
	return components, nil
}

// SelectReferenceMonthComponents returns card-backed financial occurrences
// and explicitly referenced invoice adjustments for one competence month.
// Movement events are applied before filtering, so invoice metadata reflects
// the effective allocation without changing the component's competence.
func SelectReferenceMonthComponents(input ReferenceMonthComponentsInput) ([]Component, error) {
	if !input.PlanStart.Valid() || !input.PlanEnd.Valid() || input.PlanEnd.Before(input.PlanStart) || !input.ReferenceMonth.Valid() {
		return nil, ErrInvalidProjectionInput
	}
	cards, err := indexAllCards(input.Cards)
	if err != nil {
		return nil, err
	}
	occurrences, err := selectCardOccurrences(input.PlanStart, input.PlanEnd, input.Items, input.DebtOccurrences, cards)
	if err != nil {
		return nil, err
	}
	components, err := projectComponents(occurrences, input.Adjustments, input.Moves, func(_ string, _ planning.YearMonth) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	selected := make([]Component, 0, len(components))
	for _, component := range components {
		if component.ReferenceMonth == nil || *component.ReferenceMonth != input.ReferenceMonth {
			continue
		}
		if _, exists := cards[component.CardID]; !exists {
			return nil, fmt.Errorf("%w: component references card %s", ErrInconsistentProjection, component.CardID)
		}
		selected = append(selected, component)
	}
	return selected, nil
}

func indexCards(values []Card, targetID string) (map[string]Card, Card, error) {
	cards, err := indexAllCards(values)
	if err != nil {
		return nil, Card{}, err
	}
	var target Card
	for _, card := range cards {
		if card.ID == targetID {
			target = card
		}
	}
	if target.ID == "" {
		return nil, Card{}, ErrCardNotFound
	}
	return cards, target, nil
}

func indexAllCards(values []Card) (map[string]Card, error) {
	cards := make(map[string]Card, len(values))
	for _, card := range values {
		if strings.TrimSpace(card.ID) == "" || strings.TrimSpace(card.Name) == "" {
			return nil, ErrInvalidProjectionInput
		}
		if _, duplicate := cards[card.ID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate card %s", ErrInconsistentProjection, card.ID)
		}
		cards[card.ID] = card
	}
	return cards, nil
}

func selectCardOccurrences(planStart, planEnd planning.YearMonth, items []FinancialItem, debtOccurrences []DebtOccurrence, cards map[string]Card) ([]Occurrence, error) {
	occurrences := make([]Occurrence, 0)
	seenItems := make(map[string]struct{}, len(items))
	itemsByID := make(map[string]FinancialItem, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Name) == "" || !item.Kind.Valid() || !item.Kind.IsExpense() {
			return nil, ErrInvalidProjectionInput
		}
		if _, duplicate := seenItems[item.ID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate item %s", ErrInconsistentProjection, item.ID)
		}
		seenItems[item.ID] = struct{}{}
		itemsByID[item.ID] = item
		if item.Kind == planning.FinancialItemKindDebtInstallment {
			continue
		}
		for _, period := range item.Periods {
			if strings.TrimSpace(period.ID) == "" || !period.Interval.Valid() || period.Amount.IsNegative() || !period.Recurrence.Valid() {
				return nil, ErrInvalidProjectionInput
			}
			if err := planning.ValidateFinancialItemRecurrence(item.Kind, period.Recurrence); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrInvalidProjectionInput, err)
			}
			months, err := occurrenceMonths(period, planStart, planEnd)
			if err != nil {
				return nil, err
			}
			for _, referenceMonth := range months {
				payment, err := SelectPaymentMethod(referenceMonth, item.PaymentPeriods)
				if err != nil {
					return nil, err
				}
				if payment.Method == PaymentMethodDirect {
					continue
				}
				card, exists := cards[payment.CardID]
				if !exists {
					return nil, fmt.Errorf("%w: payment method references card %s", ErrInconsistentProjection, payment.CardID)
				}
				configuration, err := SelectCardConfiguration(referenceMonth, card.Configurations)
				if err != nil {
					return nil, fmt.Errorf("%w: reference month configuration for card %s: %v", ErrInconsistentProjection, card.ID, err)
				}
				paymentMonth, err := configuration.PaymentMonth(referenceMonth)
				if err != nil {
					return nil, fmt.Errorf("%w: derive payment month: %v", ErrInconsistentProjection, err)
				}
				occurrences = append(occurrences, Occurrence{
					SourceID: period.ID, ItemID: item.ID, Name: item.Name,
					ReferenceMonth: referenceMonth, DefaultCardID: card.ID,
					DefaultPaymentMonth: paymentMonth, Amount: period.Amount,
				})
			}
		}
	}
	for _, debtOccurrence := range debtOccurrences {
		if strings.TrimSpace(debtOccurrence.DebtID) == "" || strings.TrimSpace(debtOccurrence.SourceID) == "" ||
			strings.TrimSpace(debtOccurrence.Name) == "" || !debtOccurrence.ReferenceMonth.Valid() || !debtOccurrence.DirectCashMonth.Valid() ||
			debtOccurrence.InstallmentNumber < 1 || debtOccurrence.InstallmentsTotal < debtOccurrence.InstallmentNumber ||
			debtOccurrence.Amount.Cents() <= 0 || !debtOccurrence.Kind.Valid() {
			return nil, ErrInvalidProjectionInput
		}
		if debtOccurrence.ReferenceMonth.Before(planStart) || debtOccurrence.ReferenceMonth.After(planEnd) {
			continue
		}
		item, exists := itemsByID[debtOccurrence.DebtID]
		if !exists || item.Kind != planning.FinancialItemKindDebtInstallment {
			return nil, fmt.Errorf("%w: debt occurrence references item %s", ErrInconsistentProjection, debtOccurrence.DebtID)
		}
		payment, err := SelectPaymentMethod(debtOccurrence.ReferenceMonth, item.PaymentPeriods)
		if err != nil {
			return nil, err
		}
		if payment.Method == PaymentMethodDirect {
			continue
		}
		card, exists := cards[payment.CardID]
		if !exists {
			return nil, fmt.Errorf("%w: payment method references card %s", ErrInconsistentProjection, payment.CardID)
		}
		configuration, err := SelectCardConfiguration(debtOccurrence.ReferenceMonth, card.Configurations)
		if err != nil {
			return nil, fmt.Errorf("%w: reference month configuration for card %s: %v", ErrInconsistentProjection, card.ID, err)
		}
		paymentMonth, err := configuration.PaymentMonth(debtOccurrence.ReferenceMonth)
		if err != nil {
			return nil, fmt.Errorf("%w: derive payment month: %v", ErrInconsistentProjection, err)
		}
		debtID := debtOccurrence.DebtID
		installmentNumber := debtOccurrence.InstallmentNumber
		installmentsTotal := debtOccurrence.InstallmentsTotal
		occurrenceKind := debtOccurrence.Kind
		occurrences = append(occurrences, Occurrence{
			SourceID: debtOccurrence.SourceID, ItemID: debtOccurrence.DebtID, Name: debtOccurrence.Name,
			ReferenceMonth: debtOccurrence.ReferenceMonth, DefaultCardID: card.ID,
			DefaultPaymentMonth: paymentMonth, Amount: debtOccurrence.Amount,
			DebtID: &debtID, InstallmentNumber: &installmentNumber,
			InstallmentsTotal: &installmentsTotal, DebtOccurrenceKind: &occurrenceKind,
		})
	}
	return occurrences, nil
}

func occurrenceMonths(period FinancialItemPeriod, planStart, planEnd planning.YearMonth) ([]planning.YearMonth, error) {
	start := period.Interval.Start()
	if start.Before(planStart) {
		start = planStart
	}
	end := planEnd
	if periodEnd, hasEnd := period.Interval.End(); hasEnd && periodEnd.Before(end) {
		end = periodEnd
	}
	if end.Before(start) {
		return nil, nil
	}
	if period.Recurrence == planning.RecurrenceOnce {
		if period.Interval.Start().Before(planStart) || period.Interval.Start().After(planEnd) {
			return nil, nil
		}
		return []planning.YearMonth{period.Interval.Start()}, nil
	}
	months := make([]planning.YearMonth, 0)
	for month := start; !month.After(end); {
		months = append(months, month)
		next, err := month.AddMonths(1)
		if err != nil {
			if month == end {
				break
			}
			return nil, fmt.Errorf("%w: expand occurrence months: %v", ErrInconsistentProjection, err)
		}
		month = next
	}
	return months, nil
}
