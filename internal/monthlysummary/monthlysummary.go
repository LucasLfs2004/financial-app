package monthlysummary

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/lucas/financial-api/internal/cardinvoice"
	"github.com/lucas/financial-api/internal/financialitem"
	"github.com/lucas/financial-api/internal/planning"
	"github.com/lucas/financial-api/internal/planning/domain"
	"github.com/lucas/financial-api/internal/savings"
)

var (
	ErrValidation     = errors.New("invalid monthly summary input")
	ErrOutsideHorizon = errors.New("month outside plan horizon")
	ErrInconsistent   = errors.New("inconsistent monthly summary data")
)

const (
	PlannedSavingsKind        = "planned_savings"
	CardInvoiceAdjustmentKind = "card_invoice_adjustment"
)

type PlanReader interface {
	Current(context.Context, string) (planning.Plan, error)
}

type ItemReader interface {
	List(context.Context, string, string, financialitem.Filters) ([]financialitem.Item, error)
}

type SavingsReader interface {
	Get(context.Context, string, string) (savings.Configuration, error)
}

type InvoiceDataReader interface {
	LoadProjectionData(context.Context, string, string) (cardinvoice.ProjectionData, error)
}

type Service struct {
	plans    PlanReader
	items    ItemReader
	savings  SavingsReader
	invoices InvoiceDataReader
}

func NewService(plans PlanReader, items ItemReader, savingReader SavingsReader, invoices InvoiceDataReader) *Service {
	return &Service{plans: plans, items: items, savings: savingReader, invoices: invoices}
}

type Input struct {
	Month       domain.YearMonth
	Basis       domain.SummaryBasis
	Plan        planning.Plan
	Items       []financialitem.Item
	Savings     savings.Configuration
	InvoiceData cardinvoice.ProjectionData
}

type Breakdown struct {
	RecurringIncomeCents           int64
	OneTimeIncomeCents             int64
	FixedExpensesCents             int64
	ProjectedVariableExpensesCents int64
	CardInvoiceAdjustmentsCents    int64
}

type Source struct {
	SourceID            string
	ItemID              *string
	Name                string
	Kind                string
	Effect              domain.SummarySourceEffect
	ReferenceMonth      *domain.YearMonth
	CashMonth           domain.YearMonth
	AmountCents         int64
	SourceType          string
	PaymentMethod       *cardinvoice.PaymentMethod
	CreditCardID        *string
	InvoicePaymentMonth *domain.YearMonth
	InvoiceAllocation   *cardinvoice.AllocationOrigin
	ReferenceKnown      *bool
}

type Summary struct {
	Month               domain.YearMonth
	Basis               domain.SummaryBasis
	ResultKind          domain.SummaryResultKind
	CurrencyCode        string
	PlanStatus          domain.PlanStatus
	IncomeCents         int64
	CommitmentsCents    int64
	PlannedSavingsCents int64
	ResultCents         int64
	IsNegative          bool
	Completeness        domain.Completeness
	Breakdown           Breakdown
	Sources             []Source
}

func (service *Service) Get(ctx context.Context, ownerID string, month domain.YearMonth, basis domain.SummaryBasis) (Summary, error) {
	if strings.TrimSpace(ownerID) == "" || !month.Valid() || !basis.Valid() {
		return Summary{}, ErrValidation
	}
	plan, err := service.plans.Current(ctx, ownerID)
	if err != nil {
		return Summary{}, err
	}
	if month.Before(plan.StartMonth) || month.After(plan.EndMonth) {
		return Summary{}, ErrOutsideHorizon
	}
	items, err := service.items.List(ctx, ownerID, plan.ID, financialitem.Filters{})
	if err != nil {
		return Summary{}, err
	}
	configuration, err := service.savings.Get(ctx, ownerID, plan.ID)
	if err != nil {
		return Summary{}, err
	}
	invoiceData, err := service.invoices.LoadProjectionData(ctx, ownerID, plan.ID)
	if err != nil {
		return Summary{}, err
	}
	return CalculateMonthlySummary(Input{Month: month, Basis: basis, Plan: plan, Items: items, Savings: configuration, InvoiceData: invoiceData})
}

func CalculateMonthlySummary(input Input) (Summary, error) {
	if !input.Month.Valid() || !input.Basis.Valid() || input.Month.Before(input.Plan.StartMonth) || input.Month.After(input.Plan.EndMonth) {
		return Summary{}, ErrValidation
	}
	sources, err := selectSummarySources(input)
	if err != nil {
		return Summary{}, err
	}
	result := Summary{Month: input.Month, Basis: input.Basis, CurrencyCode: input.Plan.CurrencyCode, PlanStatus: input.Plan.Status, Completeness: domain.CompletenessProjected, Sources: sources}
	if input.Basis == domain.SummaryBasisCash {
		result.ResultKind = domain.SummaryResultKindPlannedFree
	} else {
		result.ResultKind = domain.SummaryResultKindPlannedReferenceResult
	}
	recurringIncome := domain.ZeroMoney()
	oneTimeIncome := domain.ZeroMoney()
	fixedExpenses := domain.ZeroMoney()
	variableExpenses := domain.ZeroMoney()
	cardAdjustments := domain.ZeroMoney()
	plannedSavings := domain.ZeroMoney()
	for _, source := range sources {
		amount := domain.NewMoney(source.AmountCents)
		switch source.Kind {
		case string(domain.FinancialItemKindRecurringIncome):
			recurringIncome, err = recurringIncome.Add(amount)
		case string(domain.FinancialItemKindOneTimeIncome):
			oneTimeIncome, err = oneTimeIncome.Add(amount)
		case string(domain.FinancialItemKindFixedExpense):
			fixedExpenses, err = fixedExpenses.Add(amount)
		case string(domain.FinancialItemKindProjectedVariableExpense):
			variableExpenses, err = variableExpenses.Add(amount)
		case CardInvoiceAdjustmentKind:
			cardAdjustments, err = cardAdjustments.Add(amount)
		case PlannedSavingsKind:
			plannedSavings, err = plannedSavings.Add(amount)
		default:
			err = ErrInconsistent
		}
		if err != nil {
			return Summary{}, fmt.Errorf("%w: total overflow or unknown source", ErrInconsistent)
		}
	}
	result.Breakdown = Breakdown{
		RecurringIncomeCents: recurringIncome.Cents(), OneTimeIncomeCents: oneTimeIncome.Cents(),
		FixedExpensesCents: fixedExpenses.Cents(), ProjectedVariableExpensesCents: variableExpenses.Cents(),
		CardInvoiceAdjustmentsCents: cardAdjustments.Cents(),
	}
	result.PlannedSavingsCents = plannedSavings.Cents()
	income, err := recurringIncome.Add(oneTimeIncome)
	commitments, commitmentsErr := fixedExpenses.Add(variableExpenses)
	if commitmentsErr == nil {
		commitments, commitmentsErr = commitments.Add(cardAdjustments)
	}
	if err == nil {
		err = commitmentsErr
	}
	if err != nil {
		return Summary{}, fmt.Errorf("%w: total overflow", ErrInconsistent)
	}
	result.IncomeCents = income.Cents()
	result.CommitmentsCents = commitments.Cents()
	total, err := income.Subtract(commitments)
	if err == nil {
		total, err = total.Subtract(plannedSavings)
	}
	if err != nil {
		return Summary{}, fmt.Errorf("%w: result overflow", ErrInconsistent)
	}
	result.ResultCents = total.Cents()
	result.IsNegative = result.ResultCents < 0
	if err := validateConsistency(result); err != nil {
		return Summary{}, err
	}
	return result, nil
}

func selectSummarySources(input Input) ([]Source, error) {
	base, err := SelectOccurrences(input.Month, input.Basis, input.Items, input.Savings)
	if err != nil {
		return nil, err
	}
	projectionItems := make(map[string]cardinvoice.FinancialItem, len(input.InvoiceData.Items))
	for _, item := range input.InvoiceData.Items {
		projectionItems[item.ID] = item
	}
	if input.Basis == domain.SummaryBasisCash {
		filtered := make([]Source, 0, len(base))
		for _, source := range base {
			if source.ItemID != nil && source.ReferenceMonth != nil {
				if item, exists := projectionItems[*source.ItemID]; exists && item.Kind.IsExpense() {
					payment, selectionErr := cardinvoice.SelectPaymentMethod(*source.ReferenceMonth, item.PaymentPeriods)
					if selectionErr != nil {
						return nil, fmt.Errorf("%w: select payment method: %v", ErrInconsistent, selectionErr)
					}
					if payment.Method == cardinvoice.PaymentMethodCreditCard {
						continue
					}
				}
			}
			filtered = append(filtered, source)
		}
		components, selectionErr := cardinvoice.SelectPaymentMonthComponents(cardinvoice.PaymentMonthComponentsInput{
			PlanStart: input.Plan.StartMonth, PlanEnd: input.Plan.EndMonth, PaymentMonth: input.Month,
			Cards: input.InvoiceData.Cards, Items: input.InvoiceData.Items,
			Adjustments: input.InvoiceData.Adjustments, Moves: input.InvoiceData.Moves,
		})
		if selectionErr != nil {
			return nil, fmt.Errorf("%w: select invoice components: %v", ErrInconsistent, selectionErr)
		}
		base = filtered
		for _, component := range components {
			source, conversionErr := componentSource(component, projectionItems)
			if conversionErr != nil {
				return nil, conversionErr
			}
			base = append(base, source)
		}
	} else {
		components, selectionErr := cardinvoice.SelectReferenceMonthComponents(cardinvoice.ReferenceMonthComponentsInput{
			PlanStart: input.Plan.StartMonth, PlanEnd: input.Plan.EndMonth, ReferenceMonth: input.Month,
			Cards: input.InvoiceData.Cards, Items: input.InvoiceData.Items,
			Adjustments: input.InvoiceData.Adjustments, Moves: input.InvoiceData.Moves,
		})
		if selectionErr != nil {
			return nil, fmt.Errorf("%w: select invoice components: %v", ErrInconsistent, selectionErr)
		}
		cardOccurrences := make(map[string]cardinvoice.Component)
		for _, component := range components {
			if component.SourceType == cardinvoice.ComponentTypeInvoiceAdjustment {
				source, conversionErr := componentSource(component, projectionItems)
				if conversionErr != nil {
					return nil, conversionErr
				}
				base = append(base, source)
				continue
			}
			if component.ItemID != nil && component.ReferenceMonth != nil {
				cardOccurrences[*component.ItemID+":"+component.ReferenceMonth.String()] = component
			}
		}
		for index := range base {
			if base[index].ItemID == nil || base[index].ReferenceMonth == nil {
				continue
			}
			component, exists := cardOccurrences[*base[index].ItemID+":"+base[index].ReferenceMonth.String()]
			if exists {
				addCardMetadata(&base[index], component)
			}
		}
	}
	sortSources(base)
	return base, nil
}

func componentSource(component cardinvoice.Component, items map[string]cardinvoice.FinancialItem) (Source, error) {
	kind := CardInvoiceAdjustmentKind
	itemID := component.ItemID
	if component.SourceType == cardinvoice.ComponentTypeFinancialItemOccurrence {
		if itemID == nil {
			return Source{}, fmt.Errorf("%w: invoice occurrence without item", ErrInconsistent)
		}
		item, exists := items[*itemID]
		if !exists || !item.Kind.IsExpense() {
			return Source{}, fmt.Errorf("%w: invoice occurrence references unknown item", ErrInconsistent)
		}
		kind = string(item.Kind)
	}
	source := Source{
		SourceID: component.SourceID, ItemID: itemID, Name: component.Name, Kind: kind,
		Effect: domain.SummarySourceEffectCommitment, ReferenceMonth: component.ReferenceMonth,
		CashMonth: component.PaymentMonth, AmountCents: component.Amount.Cents(), SourceType: string(component.SourceType),
	}
	addCardMetadata(&source, component)
	return source, nil
}

func addCardMetadata(source *Source, component cardinvoice.Component) {
	paymentMethod := cardinvoice.PaymentMethodCreditCard
	cardID := component.CardID
	paymentMonth := component.PaymentMonth
	allocation := component.Allocation
	referenceKnown := component.ReferenceKnown
	source.PaymentMethod = &paymentMethod
	source.CreditCardID = &cardID
	source.InvoicePaymentMonth = &paymentMonth
	source.InvoiceAllocation = &allocation
	source.ReferenceKnown = &referenceKnown
}

func SelectOccurrences(month domain.YearMonth, basis domain.SummaryBasis, items []financialitem.Item, configuration savings.Configuration) ([]Source, error) {
	if !month.Valid() || !basis.Valid() {
		return nil, ErrValidation
	}
	sources := make([]Source, 0)
	seenPeriods := make(map[string]struct{})
	selectedItems := make(map[string]struct{})
	for _, item := range items {
		for _, period := range item.Periods {
			if item.ID == "" || period.ID == "" || period.AmountCents < 0 || !item.Kind.Valid() || !period.Recurrence.Valid() || period.CashMonthOffset < domain.MinimumCashMonthOffset || period.CashMonthOffset > domain.MaximumCashMonthOffset {
				return nil, ErrInconsistent
			}
			referenceMonth := month
			if basis == domain.SummaryBasisCash {
				var err error
				referenceMonth, err = month.AddMonths(-period.CashMonthOffset)
				if err != nil {
					continue
				}
			}
			if !periodApplies(period, referenceMonth) {
				continue
			}
			cashMonth, err := referenceMonth.AddMonths(period.CashMonthOffset)
			if err != nil {
				return nil, ErrInconsistent
			}
			periodKey := item.ID + ":" + period.ID
			if _, duplicate := seenPeriods[periodKey]; duplicate {
				continue
			}
			seenPeriods[periodKey] = struct{}{}
			if _, duplicate := selectedItems[item.ID]; duplicate {
				return nil, fmt.Errorf("%w: multiple periods apply to item %s", ErrInconsistent, item.ID)
			}
			selectedItems[item.ID] = struct{}{}
			itemID := item.ID
			effect := domain.SummarySourceEffectCommitment
			if item.Kind.IsIncome() {
				effect = domain.SummarySourceEffectIncome
			}
			reference := referenceMonth
			sources = append(sources, Source{SourceID: period.ID, ItemID: &itemID, Name: item.Name, Kind: string(item.Kind), Effect: effect, ReferenceMonth: &reference, CashMonth: cashMonth, AmountCents: period.AmountCents, SourceType: string(cardinvoice.ComponentTypeFinancialItemOccurrence)})
		}
	}
	savingSelected := false
	for _, period := range configuration.Periods {
		if period.ID == "" || period.AmountCents < 0 {
			return nil, ErrInconsistent
		}
		if month.Before(period.StartMonth) || (period.EndMonth != nil && month.After(*period.EndMonth)) {
			continue
		}
		periodKey := "saving:" + period.ID
		if _, duplicate := seenPeriods[periodKey]; duplicate {
			continue
		}
		seenPeriods[periodKey] = struct{}{}
		if savingSelected {
			return nil, fmt.Errorf("%w: multiple saving periods apply", ErrInconsistent)
		}
		savingSelected = true
		reference := month
		sources = append(sources, Source{SourceID: period.ID, Name: "Valor planejado para guardar", Kind: PlannedSavingsKind, Effect: domain.SummarySourceEffectSavings, ReferenceMonth: &reference, CashMonth: month, AmountCents: period.AmountCents, SourceType: PlannedSavingsKind})
	}
	sortSources(sources)
	return sources, nil
}

func sortSources(sources []Source) {
	sort.Slice(sources, func(i, j int) bool {
		left, right := sources[i], sources[j]
		if effectRank(left.Effect) != effectRank(right.Effect) {
			return effectRank(left.Effect) < effectRank(right.Effect)
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		return left.SourceID < right.SourceID
	})
}

func periodApplies(period financialitem.Period, month domain.YearMonth) bool {
	if month.Before(period.StartMonth) || (period.EndMonth != nil && month.After(*period.EndMonth)) {
		return false
	}
	return period.Recurrence == domain.RecurrenceMonthly || (period.Recurrence == domain.RecurrenceOnce && month == period.StartMonth)
}

func effectRank(effect domain.SummarySourceEffect) int {
	switch effect {
	case domain.SummarySourceEffectIncome:
		return 0
	case domain.SummarySourceEffectCommitment:
		return 1
	default:
		return 2
	}
}

func validateConsistency(summary Summary) error {
	income := domain.ZeroMoney()
	commitments := domain.ZeroMoney()
	saving := domain.ZeroMoney()
	var err error
	for _, source := range summary.Sources {
		amount := domain.NewMoney(source.AmountCents)
		switch source.Effect {
		case domain.SummarySourceEffectIncome:
			income, err = income.Add(amount)
		case domain.SummarySourceEffectCommitment:
			commitments, err = commitments.Add(amount)
		case domain.SummarySourceEffectSavings:
			saving, err = saving.Add(amount)
		default:
			err = ErrInconsistent
		}
		if err != nil {
			return fmt.Errorf("%w: source totals", ErrInconsistent)
		}
	}
	breakdownIncome, err := domain.NewMoney(summary.Breakdown.RecurringIncomeCents).Add(domain.NewMoney(summary.Breakdown.OneTimeIncomeCents))
	if err != nil {
		return ErrInconsistent
	}
	breakdownCommitments, err := domain.NewMoney(summary.Breakdown.FixedExpensesCents).Add(domain.NewMoney(summary.Breakdown.ProjectedVariableExpensesCents))
	if err == nil {
		breakdownCommitments, err = breakdownCommitments.Add(domain.NewMoney(summary.Breakdown.CardInvoiceAdjustmentsCents))
	}
	if err != nil {
		return ErrInconsistent
	}
	expectedResult, err := income.Subtract(commitments)
	if err == nil {
		expectedResult, err = expectedResult.Subtract(saving)
	}
	if err != nil || income.Cents() != summary.IncomeCents || income.Cents() != breakdownIncome.Cents() || commitments.Cents() != summary.CommitmentsCents || commitments.Cents() != breakdownCommitments.Cents() || saving.Cents() != summary.PlannedSavingsCents || expectedResult.Cents() != summary.ResultCents {
		return fmt.Errorf("%w: totals do not match components", ErrInconsistent)
	}
	return nil
}
