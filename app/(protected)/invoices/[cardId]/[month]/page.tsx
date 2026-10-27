import { notFound } from "next/navigation";
import { isMonth, inclusiveMonthCount } from "@/lib/debt-month";
import { featureFlags } from "@/lib/feature-flags";
import { InvoiceAdjustmentEditor } from "@/components/cards/invoice-adjustment-editor";
import type { Metadata } from "next";
import { ArrowRight, CalendarDays, Plus } from "lucide-react";
import Link from "next/link";
import { getCardInvoice, getCreditCards, FinancialApiError } from "@/lib/api/financial-api";
import { InvoiceAdjustmentForm } from "@/components/cards/invoice-adjustment-form";
import { InvoiceMoveForm } from "@/components/cards/invoice-move-form";
import { AppHeaderTitle } from "@/components/app-navigation";
import { formatDate } from "@/lib/format-date";
import { invoiceMonthLabel } from "@/lib/invoice-display";

export const metadata: Metadata = { title: "Detalhe da fatura" };

export default async function InvoiceDetailPage({ params, searchParams }: { params: Promise<{ cardId: string; month: string }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const { cardId, month } = await params;
  const query = await searchParams;
  const returnQuery = isMonth(query.from) && isMonth(query.to) && inclusiveMonthCount(query.from, query.to) >= 1 && inclusiveMonthCount(query.from, query.to) <= 24 ? new URLSearchParams({ from: query.from, to: query.to }).toString() : "";
  if (!isMonth(month)) notFound();
  const [invoice, cards] = await Promise.all([getCardInvoice(cardId, month).catch((error: unknown) => { if (error instanceof FinancialApiError && error.status === 404) notFound(); throw error; }), getCreditCards()]);
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: invoice?.currency_code ?? "BRL" });
  const charges = invoice.charges_total_cents;
  const payments = invoice.payments_total_cents;
  const category = "Pagamento em " + invoiceMonthLabel(month);

  return <main className="resource-page-shell invoice-detail-page">
    <AppHeaderTitle title={`${invoice?.card_name ?? "Fatura"} · ${month.slice(5)}/${month.slice(0, 4)}`} />
    {!invoice ? <div className="empty-resource-inline">Não foi possível carregar esta fatura.</div> : <>
      <section className="invoice-detail-summary" aria-label="Resumo da fatura">
        <div className="invoice-detail-summary-main"><span className="section-kicker">{category.toUpperCase()}</span><h2>{invoiceMonthLabel(month)}</h2><span>Saldo projetado</span><strong>{money.format(invoice.projected_total_cents / 100)}</strong></div>
        <div className="invoice-detail-summary-meta"><span><CalendarDays size={17} /> Vencimento <strong>{invoice.nominal_due_date ? formatDate(invoice.nominal_due_date) : `Dia ${invoice.nominal_due_day}`}</strong></span><span>Compras <strong>{money.format(charges / 100)}</strong></span><span>Pagamentos <strong>{money.format(payments / 100)}</strong></span></div>
      </section>
      <section className="invoice-components-section" aria-labelledby="invoice-components-title">
        <div className="invoice-components-heading"><div><span className="section-kicker">COMPOSIÇÃO</span><h2 id="invoice-components-title">Lançamentos <small>({invoice.component_count})</small></h2></div><Link href={`/invoices/${encodeURIComponent(cardId)}/all?${returnQuery}`}>Todas as faturas <ArrowRight size={16} /></Link></div>
        {invoice.components.length === 0 ? <div className="empty-resource-inline">Nenhum lançamento nesta fatura.</div> : <ul className="invoice-components">
          {invoice.components.map((component) => <li className="invoice-component-row" key={component.source_id}>
            <div className="invoice-component-info"><strong>{component.name}{component.installment_number != null && ` · ${component.installment_number}/${component.installments_total}`}</strong>{component.is_projected && <small>Projeção de parcela importada</small>}{component.debt_occurrence_kind === "early_settlement" && <small>Quitação antecipada projetada</small>}{component.debt_id && <Link href={`/debts/${encodeURIComponent(component.debt_id)}${component.reference_month ? `?as_of=${component.reference_month}` : ""}`}>Consultar dívida</Link>}{!component.debt_id && component.item_id && <Link href={`/entries/${encodeURIComponent(component.item_id)}${component.reference_month ? `?month=${component.reference_month}` : ""}`}>Consultar registro</Link>}<span>{component.source_type === "invoice_payment" ? component.payment_date ? `Pagamento · ${formatDate(component.payment_date)}` : "Pagamento · data não informada" : component.reference_month ? `Competência · ${component.reference_month.slice(5)}/${component.reference_month.slice(0, 4)}` : "Competência não informada"}</span></div>
            {featureFlags.adjustmentMaintenance && component.adjustment_id && <InvoiceAdjustmentEditor id={component.adjustment_id} initialOpen={query.adjustment === component.adjustment_id} />}
            <strong className="invoice-component-amount">{money.format(component.amount_cents / 100)}</strong>
            {component.item_id && component.reference_month && component.source_type !== "invoice_payment" && <div className="invoice-component-action"><InvoiceMoveForm itemId={component.item_id} referenceMonth={component.reference_month} cards={cards} initialOpen={query.moveItem === component.item_id && query.referenceMonth === component.reference_month} /></div>}
          </li>)}
        </ul>}
        <details className="invoice-adjustment-details"><summary><Plus size={17} /> Adicionar cobrança</summary><InvoiceAdjustmentForm cardId={cardId} month={month} /></details>
      </section>
    </>}
  </main>;
}
