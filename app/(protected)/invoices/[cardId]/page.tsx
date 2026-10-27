import { notFound } from "next/navigation";
import { addMonths, currentMonth } from "@/lib/debt-month";
import { queryWindow } from "@/lib/query-window";
import type { Metadata } from "next";
import { ArrowRight, CalendarDays, ReceiptText } from "lucide-react";
import Link from "next/link";
import { getCardInvoices, getCreditCards, getCurrentPlan, getProfile, absentOn404 } from "@/lib/api/financial-api";
import { invoiceDueDateLabel, invoiceMonthLabel } from "@/lib/invoice-display";
import { InvoiceCsvImport } from "@/components/cards/invoice-csv-import";
import { AppHeaderTitle } from "@/components/app-navigation";

export const metadata: Metadata = { title: "Faturas" };

export default async function InvoicesPage({ params, searchParams }: { params: Promise<{ cardId: string }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const { cardId } = await params;
  const query = await searchParams;
  const [cards, plan, profile] = await Promise.all([getCreditCards(), getCurrentPlan().catch(absentOn404), getProfile()]);
  const card = cards.find((item) => item.id === cardId);
  if (!card) notFound();
  if (!plan) return <main className="resource-page-shell"><p>Crie um planejamento para consultar as faturas.</p><Link href="/planning">Criar planejamento</Link></main>;
  const { from, to } = queryWindow(query, currentMonth(profile.timezone), { from: plan.start_month, to: addMonths(plan.end_month, 12) }, 24);
  const invoices = await getCardInvoices(cardId, from, to);
  const featured = invoices[0];
  const selectedQuery = new URLSearchParams({ from, to }).toString();
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: featured?.currency_code ?? "BRL" });

  return <main className="resource-page-shell invoice-overview-page">
    <AppHeaderTitle title={card?.name ?? "Faturas do cartão"} />
    <div className="invoice-overview-content">
      {card?.status === "active" && <InvoiceCsvImport cards={cards} initialCardId={cardId} defaultMonth={currentMonth(profile.timezone)} currencyCode={profile.currency_code} />}
      {featured ? <section className="invoice-featured" aria-labelledby="invoice-featured-title">
        <div className="invoice-featured-heading"><span className="section-kicker">FATURA EM DESTAQUE</span><CalendarDays size={19} /></div>
        <h2 id="invoice-featured-title">{invoiceMonthLabel(featured.payment_month)}</h2>
        <strong>{money.format(featured.projected_total_cents / 100)}</strong>
        <p>Vencimento: {invoiceDueDateLabel(featured)}</p>
        <Link href={`/invoices/${encodeURIComponent(cardId)}/${featured.payment_month}?${selectedQuery}`}>Abrir esta fatura <ArrowRight size={17} /></Link>
      </section> : <div className="empty-resource-inline">Nenhuma fatura projetada para o intervalo.</div>}
      <Link className="invoice-all-link" href={`/invoices/${encodeURIComponent(cardId)}/all?${selectedQuery}`}><ReceiptText size={21} /><span><strong>Ver todas as faturas</strong><small>Consulte a lista mês a mês e o gráfico de gastos.</small></span><ArrowRight size={19} /></Link>
    </div>
  </main>;
}
