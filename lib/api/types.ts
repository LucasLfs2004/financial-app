export type PlanStatus = "draft" | "active" | "archived";
export type DebtProjectionStatus = "planned" | "active" | "completed" | "settled_early" | "archived";

export type Plan = {
  id: string;
  name: string;
  status: PlanStatus;
  start_month: string;
  end_month: string;
  currency_code: string;
  activated_at: string | null;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
};

export type FinancialPeriod = {
  id: string;
  start_month: string;
  end_month: string | null;
  amount_cents: number;
  recurrence: "monthly" | "once";
  cash_month_offset: number;
  context: string | null;
};

export type FinancialItem = {
  id: string;
  currency_code: string;
  name: string;
  kind: "recurring_income" | "one_time_income" | "fixed_expense" | "projected_variable_expense" | "debt_installment";
  description: string | null;
  status: "active" | "archived";
  archived_at: string | null;
  periods: FinancialPeriod[];
  created_at: string;
  updated_at: string;
};

export type FinancialItemKind = FinancialItem["kind"];

export type SavingPeriod = {
  id: string;
  start_month: string;
  end_month: string | null;
  amount_cents: number;
  context: string | null;
  created_at: string;
};

export type PlannedSavings = { configured: boolean; periods: SavingPeriod[] };

export type MonthlySummarySource = {
  source_id: string;
  item_id: string | null;
  name: string;
  kind: string;
  effect: "income" | "commitment" | "savings";
  reference_month: string | null;
  cash_month: string;
  amount_cents: number;
  debt_id?: string | null;
  installment_number?: number | null;
  installments_total?: number | null;
  debt_occurrence_kind?: "scheduled" | "early_settlement" | null;
};

export type MonthlySummary = {
  month: string;
  basis: "cash" | "reference";
  result_kind: string;
  currency_code: string;
  plan_status: PlanStatus;
  income_cents: number;
  commitments_cents: number;
  planned_savings_cents: number;
  result_cents: number;
  is_negative: boolean;
  completeness: string;
  breakdown: {
    recurring_income_cents: number;
    one_time_income_cents: number;
    fixed_expenses_cents: number;
    projected_variable_expenses_cents: number;
    debt_installments_cents: number;
    card_invoice_adjustments_cents: number;
    card_invoice_payments_cents: number;
  };
  sources: MonthlySummarySource[];
};

export type OriginalPlanSnapshot = {
  id: string;
  plan_id: string;
  kind: "original";
  schema_version: number;
  captured_at: string;
  plan: unknown;
};

export type FinancialInstitution = {
  id: string; name: string; status: "active" | "archived"; archived_at: string | null; created_at: string; updated_at: string;
};
export type CreditCardConfiguration = {
  id: string; start_month: string; end_month: string | null; nominal_due_day: number; payment_month_offset: number; context: string | null; recorded_at: string; created_at: string;
};
export type CreditCard = {
  id: string; institution: FinancialInstitution; name: string; status: "active" | "archived"; archived_at: string | null; configurations: CreditCardConfiguration[]; created_at: string; updated_at: string;
};
export type CardInvoiceComponent = {
  source_id: string; source_type: string; item_id: string | null; adjustment_id: string | null; name: string; reference_month: string | null; reference_known: boolean; payment_month: string; amount_cents: number; allocation: string;
  payment_date?: string | null;
  debt_id?: string | null; installment_number?: number | null; installments_total?: number | null; debt_occurrence_kind?: "scheduled" | "early_settlement" | null;
};
export type CardInvoiceSummary = {
  card_id: string; card_name: string; institution: FinancialInstitution; payment_month: string; nominal_due_day: number; nominal_due_date: string | null; nominal_due_date_resolution: string; currency_code: string; charges_total_cents: number; payments_total_cents: number; projected_total_cents: number; component_count: number;
};
export type CardInvoice = CardInvoiceSummary & { components: CardInvoiceComponent[] };
export type InvoiceImportRow = {
  line: number;
  date: string;
  title: string;
  amount_cents: number;
  kind: "expense" | "payment" | "unsupported";
  invoice_payment_month?: string;
  status: "new" | "existing" | "skipped" | "imported" | "needs_invoice";
  reason?: string | null;
};
export type InvoiceImportProjectedInstallment = {
  source_line: number;
  title: string;
  amount_cents: number;
  invoice_payment_month: string;
  installment_number: number;
  installments_total: number;
  status: "new" | "existing" | "imported";
};
export type InvoiceImportPreview = {
  preview_token: string;
  rows: InvoiceImportRow[];
  new_count: number;
  existing_count: number;
  skipped_count: number;
  needs_invoice_count?: number;
  new_total_cents: number;
  new_charges_cents?: number;
  new_payments_cents?: number;
  new_payment_count?: number;
  projected_installments?: InvoiceImportProjectedInstallment[];
  new_projected_count?: number;
};
export type InvoiceImportResult = InvoiceImportPreview & {
  imported_count?: number;
};
export type CardInvoiceAdjustment = {
  id: string; currency_code: string; credit_card_id: string; payment_month: string; reference_month: string | null; name: string; amount_cents: number; context: string | null; status: "active" | "archived"; archived_at: string | null; created_at: string; updated_at: string;
};

export type DebtEarlySettlement = { id: string; debt_id: string; reference_month: string; amount_cents: number; reason: string | null; recorded_by: string; recorded_at: string; created_at: string };
export type Debt = {
  id: string;
  currency_code: string;
  name: string;
  description: string | null;
  original_total_cents: number | null;
  total_installments: number;
  first_projected_installment: number;
  scheduled_start_month: string;
  scheduled_end_month: string;
  effective_end_month: string;
  release_from_month: string;
  released_monthly_cents: number;
  as_of_month: string;
  projection_status: DebtProjectionStatus;
  remaining_installments: number;
  status: "active" | "archived";
  settlement: DebtEarlySettlement | null;
  created_at: string;
  updated_at: string;
};

export type DebtOccurrence = {
  debt_id: string; source_id: string; reference_month: string; installment_number: number; installments_total: number;
  amount_cents: number; debt_occurrence_kind: "scheduled" | "early_settlement";
  payment_method: "direct" | "credit_card"; cash_month: string; credit_card_id: string | null;
  invoice_payment_month: string | null; completeness: "projected";
};
export type DebtSchedule = { debt: Debt; range: { from: string; to: string }; occurrences: DebtOccurrence[] };
export type DebtRelease = { debt_id: string; name: string; currency_code: string; scheduled_end_month: string; effective_end_month: string; release_from_month: string; released_monthly_cents: number; reason: "scheduled_completion" | "early_settlement" };
export type DebtReleaseList = { releases: DebtRelease[]; monthly_totals: { month: string; released_monthly_cents: number }[] };
export type PaymentMethodHistory = { financial_item_id: string; default_method: "direct"; periods: { id: string; start_month: string; end_month: string | null; method: "direct" | "credit_card"; credit_card_id: string | null; context: string | null }[] };

export type ApiEnvelope<T> = { data: T };
export type ApiList<T> = { data: T[] };
