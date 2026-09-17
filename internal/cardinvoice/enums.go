package cardinvoice

import "fmt"

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

func invalidEnum(name, value string) error {
	return fmt.Errorf("%w: %s %q", ErrInvalidEnumValue, name, value)
}
