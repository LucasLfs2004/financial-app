package cardinvoice

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	planning "github.com/lucas/financial-api/internal/planning/domain"
)

const (
	MinimumNominalDueDay = 1
	MaximumNominalDueDay = 31
)

var (
	ErrInvalidEnumValue         = errors.New("invalid card invoice enum value")
	ErrInvalidConfiguration     = errors.New("invalid credit card configuration")
	ErrCardConfigurationMissing = errors.New("credit card configuration missing")
	ErrInvalidPaymentMethod     = errors.New("invalid financial item payment method")
	ErrPeriodOverlap            = errors.New("card invoice period overlap")
	ErrInvalidProjectionInput   = errors.New("invalid card invoice projection input")
	ErrDuplicateOccurrence      = errors.New("duplicate financial item occurrence")
	ErrDuplicateAdjustment      = errors.New("duplicate invoice adjustment")
	ErrNegativeAmount           = errors.New("card invoice amount cannot be negative")
	ErrInconsistentProjection   = errors.New("inconsistent card invoice projection")
)

type FinancialResourceStatus string

const (
	FinancialResourceStatusActive   FinancialResourceStatus = "active"
	FinancialResourceStatusArchived FinancialResourceStatus = "archived"
)

func (status FinancialResourceStatus) Valid() bool {
	return status == FinancialResourceStatusActive || status == FinancialResourceStatusArchived
}

func ParseFinancialResourceStatus(value string) (FinancialResourceStatus, error) {
	status := FinancialResourceStatus(value)
	if !status.Valid() {
		return "", invalidEnum("financial resource status", value)
	}
	return status, nil
}

type PaymentMethod string

const (
	PaymentMethodDirect     PaymentMethod = "direct"
	PaymentMethodCreditCard PaymentMethod = "credit_card"
)

func (method PaymentMethod) Valid() bool {
	return method == PaymentMethodDirect || method == PaymentMethodCreditCard
}

func ParsePaymentMethod(value string) (PaymentMethod, error) {
	method := PaymentMethod(value)
	if !method.Valid() {
		return "", invalidEnum("payment method", value)
	}
	return method, nil
}

type ComponentType string

const (
	ComponentTypeFinancialItemOccurrence ComponentType = "financial_item_occurrence"
	ComponentTypeInvoiceAdjustment       ComponentType = "invoice_adjustment"
)

func (componentType ComponentType) Valid() bool {
	return componentType == ComponentTypeFinancialItemOccurrence || componentType == ComponentTypeInvoiceAdjustment
}

func ParseComponentType(value string) (ComponentType, error) {
	componentType := ComponentType(value)
	if !componentType.Valid() {
		return "", invalidEnum("component type", value)
	}
	return componentType, nil
}

type AllocationOrigin string

const (
	AllocationCalculatedFromReference AllocationOrigin = "calculated_from_reference"
	AllocationSelectedInvoice         AllocationOrigin = "selected_invoice"
	AllocationMovedByUser             AllocationOrigin = "moved_by_user"
)

func (origin AllocationOrigin) Valid() bool {
	return origin == AllocationCalculatedFromReference || origin == AllocationSelectedInvoice || origin == AllocationMovedByUser
}

func ParseAllocationOrigin(value string) (AllocationOrigin, error) {
	origin := AllocationOrigin(value)
	if !origin.Valid() {
		return "", invalidEnum("allocation origin", value)
	}
	return origin, nil
}

type DueDateResolution string

const (
	DueDateResolutionExact           DueDateResolution = "exact"
	DueDateResolutionInvalidForMonth DueDateResolution = "invalid_for_month"
)

func (resolution DueDateResolution) Valid() bool {
	return resolution == DueDateResolutionExact || resolution == DueDateResolutionInvalidForMonth
}

func ParseDueDateResolution(value string) (DueDateResolution, error) {
	resolution := DueDateResolution(value)
	if !resolution.Valid() {
		return "", invalidEnum("due date resolution", value)
	}
	return resolution, nil
}

type AdjustmentStatus string

const (
	AdjustmentStatusActive   AdjustmentStatus = "active"
	AdjustmentStatusArchived AdjustmentStatus = "archived"
)

func (status AdjustmentStatus) Valid() bool {
	return status == AdjustmentStatusActive || status == AdjustmentStatusArchived
}

func ParseAdjustmentStatus(value string) (AdjustmentStatus, error) {
	status := AdjustmentStatus(value)
	if !status.Valid() {
		return "", invalidEnum("adjustment status", value)
	}
	return status, nil
}

type AuditEventType string

const (
	AuditEventOccurrenceMoved    AuditEventType = "occurrence_moved"
	AuditEventAdjustmentChanged  AuditEventType = "adjustment_changed"
	AuditEventAdjustmentArchived AuditEventType = "adjustment_archived"
)

func (eventType AuditEventType) Valid() bool {
	return eventType == AuditEventOccurrenceMoved || eventType == AuditEventAdjustmentChanged || eventType == AuditEventAdjustmentArchived
}

func ParseAuditEventType(value string) (AuditEventType, error) {
	eventType := AuditEventType(value)
	if !eventType.Valid() {
		return "", invalidEnum("audit event type", value)
	}
	return eventType, nil
}

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

func (configuration CardConfiguration) NominalDueDay() int {
	return configuration.nominalDueDay
}

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

type PaymentMethodPeriod struct {
	ID       string
	Interval planning.MonthInterval
	Method   PaymentMethod
	CardID   string
}

func NewPaymentMethodPeriod(id string, interval planning.MonthInterval, method PaymentMethod, cardID string) (PaymentMethodPeriod, error) {
	if strings.TrimSpace(id) == "" || !interval.Valid() || !method.Valid() {
		return PaymentMethodPeriod{}, ErrInvalidPaymentMethod
	}
	if (method == PaymentMethodCreditCard && strings.TrimSpace(cardID) == "") || (method == PaymentMethodDirect && strings.TrimSpace(cardID) != "") {
		return PaymentMethodPeriod{}, ErrInvalidPaymentMethod
	}
	return PaymentMethodPeriod{ID: id, Interval: interval, Method: method, CardID: cardID}, nil
}

type PaymentMethodSelection struct {
	Method   PaymentMethod
	CardID   string
	Explicit bool
}

func SelectPaymentMethod(month planning.YearMonth, periods []PaymentMethodPeriod) (PaymentMethodSelection, error) {
	if !month.Valid() {
		return PaymentMethodSelection{}, ErrInvalidPaymentMethod
	}
	selection := PaymentMethodSelection{Method: PaymentMethodDirect}
	for _, period := range periods {
		if _, err := NewPaymentMethodPeriod(period.ID, period.Interval, period.Method, period.CardID); err != nil {
			return PaymentMethodSelection{}, err
		}
		if !period.Interval.Contains(month) {
			continue
		}
		if selection.Explicit {
			return PaymentMethodSelection{}, ErrPeriodOverlap
		}
		selection = PaymentMethodSelection{Method: period.Method, CardID: period.CardID, Explicit: true}
	}
	return selection, nil
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

	latestMoves, err := selectLatestMoves(input.Moves)
	if err != nil {
		return Invoice{}, err
	}

	components := make([]Component, 0, len(input.Occurrences)+len(input.Adjustments))
	seenOccurrences := make(map[string]struct{}, len(input.Occurrences))
	for _, occurrence := range input.Occurrences {
		key, validationError := validateOccurrence(occurrence)
		if validationError != nil {
			return Invoice{}, validationError
		}
		if _, duplicate := seenOccurrences[key]; duplicate {
			return Invoice{}, fmt.Errorf("%w: %s", ErrDuplicateOccurrence, key)
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
		if cardID != input.CardID || paymentMonth != input.PaymentMonth {
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

	seenAdjustments := make(map[string]struct{}, len(input.Adjustments))
	for _, adjustment := range input.Adjustments {
		if err := validateAdjustment(adjustment); err != nil {
			return Invoice{}, err
		}
		if _, duplicate := seenAdjustments[adjustment.ID]; duplicate {
			return Invoice{}, fmt.Errorf("%w: %s", ErrDuplicateAdjustment, adjustment.ID)
		}
		seenAdjustments[adjustment.ID] = struct{}{}
		if adjustment.Status == AdjustmentStatusArchived || adjustment.CardID != input.CardID || adjustment.PaymentMonth != input.PaymentMonth {
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
	total := planning.ZeroMoney()
	for _, component := range components {
		total, err = total.Add(component.Amount)
		if err != nil {
			return Invoice{}, fmt.Errorf("%w: total overflow", ErrInconsistentProjection)
		}
	}

	return Invoice{
		CardID: input.CardID, CardName: input.CardName, PaymentMonth: input.PaymentMonth,
		CurrencyCode: input.CurrencyCode, NominalDueDate: dueDate, Total: total,
		Components: components,
	}, nil
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

func invalidEnum(name, value string) error {
	return fmt.Errorf("%w: %s %q", ErrInvalidEnumValue, name, value)
}
