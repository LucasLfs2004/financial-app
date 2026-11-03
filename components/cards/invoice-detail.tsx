import Link from "next/link";
import { Plus, ChevronRight } from "lucide-react";
import { MonthNavigation } from "@/components/month-navigation";
import { InvoiceCsvImport } from "./invoice-csv-import";
import { InvoiceAdjustmentForm } from "./invoice-adjustment-form";
import { InvoiceAdjustmentEditor } from "./invoice-adjustment-editor";
import { InvoiceMoveForm } from "./invoice-move-form";
import { featureFlags } from "@/lib/feature-flags";
import { formatDate } from "@/lib/format-date";
import { invoiceComponentLabel, invoiceComposition } from "@/lib/financial-overview";
import type { CardInvoice, CreditCard } from "@/lib/api/types";

export function InvoiceDetail({ invoice, cards, min, max, returnQuery = "", adjustment, moveItem, referenceMonth }: {
  invoice: CardInvoice; cards: CreditCard[]; min?: string; max?: string; returnQuery?: string;
  adjustment?: string; moveItem?: string; referenceMonth?: string;
}) {
  const cardId = invoice.card_id;
  const month = invoice.payment_month;
  const money = (cents: number) => new Intl.NumberFormat("pt-BR", { style: "currency", currency: invoice.currency_code }).format(cents / 100);
  const card = cards.find((item) => item.id === cardId);
  const canImport = card?.status === "active";
  const groups = invoiceComposition(invoice.components);
  return <div className="finance-invoice-detail">
    <MonthNavigation month={month} min={min} max={max} basePath={`/invoices/${encodeURIComponent(cardId)}`} query={returnQuery} monthInPath />
    <section className={`finance-hero${invoice.projected_total_cents < 0 ? " is-negative" : ""}`} aria-label="Resumo da fatura">
      <span className="section-kicker">SALDO PREVISTO DA FATURA</span><strong>{money(invoice.projected_total_cents)}</strong>
      <p>Vencimento · {invoice.nominal_due_date ? formatDate(invoice.nominal_due_date) : `dia nominal ${invoice.nominal_due_day}`}</p>
      <dl className="finance-invoice-totals"><div><dt>Compras e cobranças</dt><dd>{money(invoice.charges_total_cents)}</dd></div><div><dt>Pagamentos registrados</dt><dd>− {money(invoice.payments_total_cents)}</dd></div></dl>
    </section>
    <p className="finance-help">Inclui compromissos previstos e cobranças registradas. Registrar uma cobrança não confirma que ela foi paga.</p>
    {!invoice.nominal_due_date && <p className="finance-notice">O vencimento é nominal. Não há data válida informada para este mês.</p>}
    <section className="finance-composition" aria-labelledby="invoice-composition-heading">
      <h2 id="invoice-composition-heading">De onde vem esse valor?</h2>
      {groups.length === 0 ? <p className="finance-panel finance-help">Nenhum lançamento nesta fatura.</p> : <dl className="finance-arithmetic finance-panel">{groups.map((group) => <div key={group.id}><dt>{group.label} <small>({group.count})</small></dt><dd>{money(group.amount)}</dd></div>)}</dl>}
    </section>
    <section className="finance-invoice-entries" aria-labelledby="invoice-components-title"><div className="finance-section-heading"><h2 id="invoice-components-title">Lançamentos <small>({invoice.component_count})</small></h2></div>
      <ul className="finance-invoice-list">{invoice.components.map((component) => <li key={component.source_id}>
        <details className="finance-entry-details" open={adjustment === component.adjustment_id || moveItem === component.item_id && referenceMonth === component.reference_month ? true : undefined}>
          <summary><span><strong>{component.name}</strong><small>{invoiceComponentLabel(component)}{component.installment_number != null ? ` · ${component.installment_number}/${component.installments_total}` : ""}</small></span><strong className={component.amount_cents < 0 ? "finance-credit" : ""}>{money(component.amount_cents)}</strong><ChevronRight size={16} /></summary>
          <div className="finance-entry-expanded">
            <p className="finance-help">{component.source_type === "invoice_payment" ? component.payment_date ? `Pagamento em ${formatDate(component.payment_date)}` : "Data do pagamento não informada" : component.reference_month && component.reference_known !== false ? `Competência: ${component.reference_month.slice(5)}/${component.reference_month.slice(0, 4)}` : "Competência não informada"}</p>
            {component.debt_id && <Link className="finance-text-link" href={`/debts/${encodeURIComponent(component.debt_id)}${component.reference_month ? `?as_of=${component.reference_month}` : ""}`}>Consultar dívida →</Link>}
            {!component.debt_id && component.item_id && <Link className="finance-text-link" href={`/entries/${encodeURIComponent(component.item_id)}${component.reference_month ? `?month=${component.reference_month}` : ""}`}>Consultar registro →</Link>}
            {featureFlags.adjustmentMaintenance && component.adjustment_id && <InvoiceAdjustmentEditor id={component.adjustment_id} initialOpen={adjustment === component.adjustment_id} />}
            {component.item_id && component.reference_month && component.source_type !== "invoice_payment" && <InvoiceMoveForm itemId={component.item_id} referenceMonth={component.reference_month} cards={cards} initialOpen={moveItem === component.item_id && referenceMonth === component.reference_month} />}
          </div>
        </details>
      </li>)}</ul>
    </section>
    {canImport && <InvoiceCsvImport compact cards={cards} initialCardId={cardId} defaultMonth={month} currencyCode={invoice.currency_code} />}
    <details className="finance-disclosure"><summary><Plus size={18} /> Adicionar cobrança</summary><InvoiceAdjustmentForm cardId={cardId} month={month} /></details>
    <div className="finance-secondary-links"><Link href={`/invoices/${encodeURIComponent(cardId)}/all?${returnQuery || `from=${month}`}`}>Ver histórico e próximas faturas <ChevronRight size={16} /></Link><Link href={`/cards/${encodeURIComponent(cardId)}?month=${month}`}>Configurações do cartão <ChevronRight size={16} /></Link></div>
  </div>;
}
