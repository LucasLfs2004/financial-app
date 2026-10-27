"use client";

import { VariableForecasts } from "./variable-forecasts";
import { useState } from "react";
import Link from "next/link";
import type { MonthlySummary } from "@/lib/api/types";
import { formatMonth } from "@/lib/debt-month";
import { summarySourceLink } from "@/lib/summary-source";

const sourceLabels: Record<string, string> = { recurring_income: "Renda recorrente", one_time_income: "Renda pontual", projected_variable_expense: "Despesa variável projetada", debt_installment: "Parcela de dívida", planned_savings: "Economia planejada", card_invoice_adjustment: "Ajuste de fatura", card_invoice_payment: "Pagamento de fatura", financial_item_occurrence: "Registro financeiro", invoice_payment_cash: "Pagamento de fatura em caixa", calculated_from_reference: "Calculada pela competência", selected_invoice: "Fatura selecionada", moved_by_user: "Movida pelo usuário", income: "Receita", fixed_expense: "Despesa fixa", variable_expense: "Despesa variável", debt: "Dívida", invoice_adjustment: "Ajuste de fatura", invoice_payment: "Pagamento de fatura", savings: "Economia", financial_item: "Registro", calculated: "Alocação calculada", moved: "Alocação movida", manual: "Alocação manual" };

type Category = "income" | "commitments" | "savings";
const labels = { income: "Receitas", commitments: "Compromissos", savings: "Economia" };
const buckets: Record<string, string> = { recurring_income_cents: "Renda recorrente", one_time_income_cents: "Renda pontual", fixed_expenses_cents: "Despesas fixas", projected_variable_expenses_cents: "Despesas variáveis", debt_installments_cents: "Parcelas de dívidas", card_invoice_adjustments_cents: "Ajustes de faturas", card_invoice_payments_cents: "Pagamentos de faturas" };
export function MonthlyComposition({ summary, initialCategory }: { summary: MonthlySummary; initialCategory?: Category }) {
  const [category, setCategory] = useState<Category | null>(initialCategory ?? null);
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: summary.currency_code });
  function choose(value: Category | null) {
    setCategory(value);
    const url = new URL(window.location.href);
    if (value) url.searchParams.set("composition", value); else url.searchParams.delete("composition");
    window.history.replaceState(null, "", url);
  }
  const total = category === "income" ? summary.income_cents : category === "commitments" ? summary.commitments_cents : summary.planned_savings_cents;
  const sources = summary.sources.filter((source) => source.effect === (category === "commitments" ? "commitment" : category));
  return <section className="monthly-composition"><VariableForecasts summary={summary} /><div className="form-actions" aria-label="Composição mensal">{(["income", "commitments", "savings"] as Category[]).map((value) => <button type="button" key={value} aria-pressed={category === value} onClick={() => choose(value)}>Ver composição: {labels[value]}</button>)}</div>{category && <div className="debt-action-panel"><div className="resource-card-footer"><h2>{labels[category]} em {formatMonth(summary.month)}</h2><button type="button" onClick={() => choose(null)}>Fechar composição</button></div><p>Total oficial: <strong>{money.format(total / 100)}</strong> · base {summary.basis === "cash" ? "Caixa" : "Competência"}</p><p>Categorias e fontes detalham o mesmo total.</p>{category !== "savings" && <dl className="snapshot-fields">{Object.entries(summary.breakdown).filter(([key]) => category === "income" ? key.endsWith("income_cents") : !key.endsWith("income_cents")).map(([key, amount]) => <div key={key}><dt>{buckets[key] ?? key}</dt><dd>{money.format(amount / 100)}</dd></div>)}</dl>}{sources.length === 0 ? <p>Nenhuma fonte nesta categoria.</p> : <div className="resource-history"><table><thead><tr><th>Origem</th><th>Competência</th><th>Caixa</th><th>Pagamento e fatura</th><th>Valor</th></tr></thead><tbody>{sources.map((source, index) => { const href = summarySourceLink(source); return <tr key={`${source.source_id}-${index}`}><td>{href ? <Link href={href}>{source.name}</Link> : source.name}{source.installment_number != null && <> · {source.installment_number}/{source.installments_total}</>}{source.debt_occurrence_kind === "early_settlement" && " · quitação antecipada"}{source.is_projected && " · projeção de parcela importada"}<small>{sourceLabels[source.kind] ?? sourceLabels[source.source_type ?? ""] ?? "Fonte financeira"}</small></td><td>{source.reference_month && source.reference_known !== false ? formatMonth(source.reference_month) : "Não informada"}</td><td>{formatMonth(source.cash_month)}</td><td>{source.payment_method === "credit_card" ? "Cartão" : source.payment_method === "direct" ? "Direto" : "—"}{source.invoice_payment_month && ` · ${formatMonth(source.invoice_payment_month)}`}{source.invoice_allocation && ` · ${sourceLabels[source.invoice_allocation] ?? "Alocação da fatura"}`}</td><td>{money.format(source.amount_cents / 100)}</td></tr>; })}</tbody></table></div>}</div>}</section>;
}
