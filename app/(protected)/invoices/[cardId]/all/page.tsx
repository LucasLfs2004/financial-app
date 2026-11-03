import { AppHeaderTitle } from "@/components/app-navigation";
import { notFound } from "next/navigation";
import { addMonths, currentMonth, inclusiveMonthCount } from "@/lib/debt-month";
import { queryWindow } from "@/lib/query-window";
import type { Metadata } from "next";
import { ReceiptText } from "lucide-react";
import Link from "next/link";
import { getCardInvoices, getCreditCards, getCurrentPlan, getProfile, absentOn404 } from "@/lib/api/financial-api";
import type { CardInvoiceSummary } from "@/lib/api/types";
import { invoiceDueDateLabel, invoiceMonthLabel } from "@/lib/invoice-display";
import { InvoiceSpendingChart } from "@/components/cards/invoice-spending-chart";

export const metadata: Metadata = { title: "Todas as faturas" };

function InvoiceCard({ invoice, cardId, category, query }: { invoice: CardInvoiceSummary; cardId: string; query: string; category: "previous" | "current" | "future" }) {
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: invoice.currency_code });
  return <Link className={`invoice-card invoice-card-${category}`} href={`/invoices/${encodeURIComponent(cardId)}/${invoice.payment_month}?${query}`}>
    <div><span>Pagamento em {invoiceMonthLabel(invoice.payment_month)}</span><strong>{money.format(invoice.projected_total_cents / 100)}</strong></div>
    <span className="invoice-card-due">Vencimento: {invoiceDueDateLabel(invoice)}</span>
    {Number.isFinite(invoice.charges_total_cents) && Number.isFinite(invoice.payments_total_cents) && <span className="invoice-card-totals">Compras {money.format(invoice.charges_total_cents / 100)} · Pagamentos {money.format(invoice.payments_total_cents / 100)}</span>}
    <small>{invoice.component_count} componente(s)</small>
  </Link>;
}

export default async function AllInvoicesPage({ params, searchParams }: { params: Promise<{ cardId: string }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const { cardId } = await params;
  const query = await searchParams;
  const [cards, plan, profile] = await Promise.all([getCreditCards(), getCurrentPlan().catch(absentOn404), getProfile()]);
  const card = cards.find((item) => item.id === cardId);
  if (!card) notFound();
  if (!plan) return <main className="resource-page-shell"><p>Crie um planejamento para consultar as faturas.</p><Link href="/planning">Criar planejamento</Link></main>;
  const bounds = { from: plan.start_month, to: addMonths(plan.end_month, 12) };
  const today = currentMonth(profile.timezone);
  const { from, to, adjusted } = queryWindow(query, today, bounds, 24);
  const invoices = await getCardInvoices(cardId, from, to);
  const size = inclusiveMonthCount(from, to);
  const selectedQuery = new URLSearchParams({ from, to }).toString();
  function windowLink(start: string) {
    return `/invoices/${encodeURIComponent(cardId)}/all?${new URLSearchParams({ from: start, to: addMonths(start, Math.min(size - 1, inclusiveMonthCount(start, bounds.to) - 1)) })}`;
  }

  return <main className="resource-page-shell"><AppHeaderTitle title={`Todas as faturas · ${card?.name ?? "Cartão"}`} />
    <header className="resource-page-header"><p>Consulte os valores e abra cada fatura mês a mês.</p></header>
    <form className="resource-form resource-query-form" method="get"><div className="form-grid-two"><label className="native-field"><span>Pagamento de</span><input type="month" name="from" defaultValue={from} min={bounds.from} max={bounds.to} required /></label><label className="native-field"><span>Pagamento até (máximo 24 meses)</span><input type="month" name="to" defaultValue={to} min={from} max={bounds.to} required /></label></div><button type="submit">Consultar faturas</button></form>
    {adjusted && <p role="status">Intervalo ajustado aos limites de consulta do planejamento.</p>}
    <nav className="resource-card-footer" aria-label="Períodos de faturas">{from > bounds.from && <Link href={windowLink(addMonths(from, -Math.min(size, inclusiveMonthCount(bounds.from, from) - 1)))}>Período anterior</Link>}<span>{invoiceMonthLabel(from)} a {invoiceMonthLabel(to)}</span>{to < bounds.to && <Link href={windowLink(addMonths(to, 1))}>Próximo período</Link>}</nav>
    {invoices.length > 0 && <InvoiceSpendingChart key={selectedQuery} cardId={cardId} invoices={invoices} currentPaymentMonth={today} returnQuery={selectedQuery} />}
    <section className="resource-list-section"><div className="section-heading compact"><div><span className="section-kicker">PROJEÇÃO</span><h2>Faturas por mês de pagamento</h2></div><ReceiptText size={20} /></div>
      {invoices.length === 0 ? <div className="empty-resource-inline">Nenhuma fatura projetada para o intervalo.</div> : <div className="invoice-grid">{invoices.map((invoice) => <InvoiceCard key={invoice.payment_month} invoice={invoice} cardId={cardId} query={selectedQuery} category={invoice.payment_month < today ? "previous" : invoice.payment_month === today ? "current" : "future"} />)}</div>}
    </section>
  </main>;
}
