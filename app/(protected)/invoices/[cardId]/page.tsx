import type { Metadata } from "next";
import { ReceiptText } from "lucide-react";
import Link from "next/link";
import { getCardInvoices, getCreditCards } from "@/lib/api/financial-api";
import type { CardInvoiceSummary } from "@/lib/api/types";
import { addInvoiceMonths, currentMonthInBrazil, invoiceDueDateLabel, invoiceMonthLabel } from "@/lib/invoice-display";
import { InvoiceSpendingChart } from "@/components/cards/invoice-spending-chart";
import { AppHeaderTitle } from "@/components/app-navigation";

export const metadata: Metadata = { title: "Faturas" };

function InvoiceCard({ invoice, cardId, category }: { invoice: CardInvoiceSummary; cardId: string; category: "previous" | "current" | "future" }) {
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: invoice.currency_code });
  return <Link className={`invoice-card invoice-card-${category}`} href={`/invoices/${cardId}/${invoice.payment_month}`}>
    <div><span>Fatura de {invoiceMonthLabel(invoice.payment_month)}</span><strong>{money.format(invoice.projected_total_cents / 100)}</strong></div>
    <span className="invoice-card-due">Vencimento: {invoiceDueDateLabel(invoice)}</span>
    {Number.isFinite(invoice.charges_total_cents) && Number.isFinite(invoice.payments_total_cents) && <span className="invoice-card-totals">Compras {money.format(invoice.charges_total_cents / 100)} · Pagamentos {money.format(invoice.payments_total_cents / 100)}</span>}
    <small>{invoice.component_count} componente(s)</small>
  </Link>;
}

export default async function InvoicesPage({ params }: { params: Promise<{ cardId: string }> }) {
  const { cardId } = await params; const from = currentMonthInBrazil();
  const [cards, invoices] = await Promise.all([
    getCreditCards().catch(() => []),
    getCardInvoices(cardId, from, addInvoiceMonths(from, 11)).catch(() => []),
  ]);
  const card = cards.find((item) => item.id === cardId);
  const currentPaymentMonth = addInvoiceMonths(from, 1);
  const previous = invoices.filter((invoice) => invoice.payment_month === from);
  const current = invoices.filter((invoice) => invoice.payment_month === currentPaymentMonth);
  const future = invoices.filter((invoice) => invoice.payment_month > currentPaymentMonth);
  return <main className="resource-page-shell invoice-overview-page">
    <AppHeaderTitle title={card?.name ?? "Faturas do cartão"} />
    {invoices.length > 0 && <InvoiceSpendingChart cardId={cardId} invoices={invoices} currentPaymentMonth={currentPaymentMonth} />}
    <section className="resource-list-section"><div className="section-heading compact"><div><span className="section-kicker">PROJEÇÃO</span><h2>Suas faturas por período</h2></div><ReceiptText size={20} /></div>
      {invoices.length === 0 ? <div className="empty-resource-inline">Nenhuma fatura projetada para o intervalo.</div> : <div className="invoice-periods">
        {previous.length > 0 && <section className="invoice-period-section"><div className="invoice-period-heading"><h3>Fatura do mês passado</h3><p>Vence neste mês e reúne o ciclo anterior.</p></div><div className="invoice-grid">{previous.map((invoice) => <InvoiceCard key={invoice.payment_month} invoice={invoice} cardId={cardId} category="previous" />)}</div></section>}
        {current.length > 0 && <section className="invoice-period-section"><div className="invoice-period-heading"><h3>Fatura atual</h3><p>Vence no próximo mês.</p></div><div className="invoice-grid">{current.map((invoice) => <InvoiceCard key={invoice.payment_month} invoice={invoice} cardId={cardId} category="current" />)}</div></section>}
        {future.length > 0 && <section className="invoice-period-section"><div className="invoice-period-heading"><h3>Faturas futuras</h3><p>Vencimentos dos próximos ciclos.</p></div><div className="invoice-grid">{future.map((invoice) => <InvoiceCard key={invoice.payment_month} invoice={invoice} cardId={cardId} category="future" />)}</div></section>}
      </div>}
    </section>
  </main>;
}
