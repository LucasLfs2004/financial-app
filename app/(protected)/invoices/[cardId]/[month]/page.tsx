import type { Metadata } from "next";
import { ArrowRight, CalendarDays, Plus } from "lucide-react";
import Link from "next/link";
import { getCardInvoice, getCreditCards } from "@/lib/api/financial-api";
import { InvoiceAdjustmentForm } from "@/components/cards/invoice-adjustment-form";
import { InvoiceMoveForm } from "@/components/cards/invoice-move-form";
import { AppHeaderTitle } from "@/components/app-navigation";
import { formatDate } from "@/lib/format-date";
import { addInvoiceMonths, currentMonthInBrazil, invoiceMonthLabel } from "@/lib/invoice-display";

export const metadata: Metadata = { title: "Detalhe da fatura" };

export default async function InvoiceDetailPage({ params }: { params: Promise<{ cardId: string; month: string }> }) {
  const { cardId, month } = await params;
  const [invoice, cards] = await Promise.all([getCardInvoice(cardId, month).catch(() => null), getCreditCards().catch(() => [])]);
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: invoice?.currency_code ?? "BRL" });
  const charges = invoice ? (Number.isFinite(invoice.charges_total_cents) ? invoice.charges_total_cents : invoice.components.filter((component) => component.source_type !== "invoice_payment").reduce((total, component) => total + component.amount_cents, 0)) : 0;
  const payments = invoice ? (Number.isFinite(invoice.payments_total_cents) ? invoice.payments_total_cents : invoice.components.filter((component) => component.source_type === "invoice_payment").reduce((total, component) => total + component.amount_cents, 0)) : 0;
  const from = currentMonthInBrazil();
  const category = month === from ? "Fatura do mês passado" : month === addInvoiceMonths(from, 1) ? "Fatura atual" : month > from ? "Fatura futura" : "Fatura anterior";

  return <main className="resource-page-shell invoice-detail-page">
    <AppHeaderTitle title={`${invoice?.card_name ?? "Fatura"} · ${month.slice(5)}/${month.slice(0, 4)}`} />
    {!invoice ? <div className="empty-resource-inline">Não foi possível carregar esta fatura.</div> : <>
      <section className="invoice-detail-summary" aria-label="Resumo da fatura">
        <div className="invoice-detail-summary-main"><span className="section-kicker">{category.toUpperCase()}</span><h2>{invoiceMonthLabel(month)}</h2><span>Saldo projetado</span><strong>{money.format(invoice.projected_total_cents / 100)}</strong></div>
        <div className="invoice-detail-summary-meta"><span><CalendarDays size={17} /> Vencimento <strong>{invoice.nominal_due_date ? formatDate(invoice.nominal_due_date) : `Dia ${invoice.nominal_due_day}`}</strong></span><span>Compras <strong>{money.format(charges / 100)}</strong></span><span>Pagamentos <strong>{money.format(payments / 100)}</strong></span></div>
      </section>
      <section className="invoice-components-section" aria-labelledby="invoice-components-title">
        <div className="invoice-components-heading"><div><span className="section-kicker">COMPOSIÇÃO</span><h2 id="invoice-components-title">Lançamentos <small>({invoice.component_count})</small></h2></div><Link href={`/invoices/${cardId}/all`}>Todas as faturas <ArrowRight size={16} /></Link></div>
        {invoice.components.length === 0 ? <div className="empty-resource-inline">Nenhum lançamento nesta fatura.</div> : <ul className="invoice-components">
          {invoice.components.map((component) => <li className="invoice-component-row" key={component.source_id}>
            <div className="invoice-component-info"><strong>{component.name}</strong><span>{component.source_type === "invoice_payment" ? component.payment_date ? `Pagamento · ${formatDate(component.payment_date)}` : "Pagamento · data não informada" : component.reference_month ? `Competência · ${component.reference_month.slice(5)}/${component.reference_month.slice(0, 4)}` : "Competência não informada"}</span></div>
            <strong className="invoice-component-amount">{money.format(component.amount_cents / 100)}</strong>
            {component.item_id && component.reference_month && <div className="invoice-component-action"><InvoiceMoveForm itemId={component.item_id} referenceMonth={component.reference_month} cards={cards} /></div>}
          </li>)}
        </ul>}
        <details className="invoice-adjustment-details"><summary><Plus size={17} /> Adicionar cobrança</summary><InvoiceAdjustmentForm cardId={cardId} month={month} /></details>
      </section>
    </>}
  </main>;
}
