import type { Metadata } from "next";
import { ArrowRight, CalendarDays, ReceiptText } from "lucide-react";
import Link from "next/link";
import { getCardInvoices, getCreditCards } from "@/lib/api/financial-api";
import { addInvoiceMonths, currentMonthInBrazil, invoiceDueDateLabel, invoiceMonthLabel } from "@/lib/invoice-display";
import { InvoiceCsvImport } from "@/components/cards/invoice-csv-import";
import { AppHeaderTitle } from "@/components/app-navigation";

export const metadata: Metadata = { title: "Faturas" };

export default async function InvoicesPage({ params }: { params: Promise<{ cardId: string }> }) {
  const { cardId } = await params;
  const from = currentMonthInBrazil();
  const [cards, invoices] = await Promise.all([
    getCreditCards().catch(() => []),
    getCardInvoices(cardId, from, addInvoiceMonths(from, 11)).catch(() => []),
  ]);
  const card = cards.find((item) => item.id === cardId);
  const featured = invoices.find((invoice) => invoice.payment_month === addInvoiceMonths(from, 1)) ?? invoices[0];
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: featured?.currency_code ?? "BRL" });

  return <main className="resource-page-shell invoice-overview-page">
    <AppHeaderTitle title={card?.name ?? "Faturas do cartão"} />
    <div className="invoice-overview-content">
      {card?.status === "active" && <InvoiceCsvImport cards={cards} initialCardId={cardId} />}
      {featured ? <section className="invoice-featured" aria-labelledby="invoice-featured-title">
        <div className="invoice-featured-heading"><span className="section-kicker">FATURA EM DESTAQUE</span><CalendarDays size={19} /></div>
        <h2 id="invoice-featured-title">{invoiceMonthLabel(featured.payment_month)}</h2>
        <strong>{money.format(featured.projected_total_cents / 100)}</strong>
        <p>Vencimento: {invoiceDueDateLabel(featured)}</p>
        <Link href={`/invoices/${cardId}/${featured.payment_month}`}>Abrir esta fatura <ArrowRight size={17} /></Link>
      </section> : <div className="empty-resource-inline">Nenhuma fatura projetada para o intervalo.</div>}
      <Link className="invoice-all-link" href={`/invoices/${cardId}/all`}><ReceiptText size={21} /><span><strong>Ver todas as faturas</strong><small>Consulte a lista mês a mês e o gráfico de gastos.</small></span><ArrowRight size={19} /></Link>
    </div>
  </main>;
}
