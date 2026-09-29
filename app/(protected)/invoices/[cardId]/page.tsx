import type { Metadata } from "next";
import { ArrowLeft, ReceiptText } from "lucide-react";
import Link from "next/link";
import { getCardInvoices, getCreditCards } from "@/lib/api/financial-api";

export const metadata: Metadata = { title: "Invoices" };

function currentMonth() { return new Date().toISOString().slice(0, 7); }
function addMonths(month: string, amount: number) { const [year, raw] = month.split("-").map(Number); const date = new Date(year, raw - 1 + amount, 1); return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`; }

export default async function InvoicesPage({ params }: { params: Promise<{ cardId: string }> }) {
  const { cardId } = await params; const from = currentMonth();
  const [cards, invoices] = await Promise.all([
    getCreditCards().catch(() => []),
    getCardInvoices(cardId, from, addMonths(from, 11)).catch(() => []),
  ]);
  const card = cards.find((item) => item.id === cardId);
  return <main className="resource-page-shell"><header className="resource-page-header"><Link className="back-link" href="/cards"><ArrowLeft size={16} /> Voltar aos cartões</Link><span className="eyebrow">FATURAS</span><h1>{card?.name ?? "Faturas do cartão"}</h1><p>{card?.institution.name ?? "Composição projetada"}</p></header><section className="resource-list-section"><div className="section-heading compact"><div><span className="section-kicker">PROJEÇÃO</span><h2>Próximas faturas</h2></div><ReceiptText size={20} /></div>{invoices.length === 0 ? <div className="empty-resource-inline">Nenhuma fatura projetada para o intervalo.</div> : <div className="invoice-grid">{invoices.map((invoice) => <Link className="invoice-card" href={`/invoices/${cardId}/${invoice.payment_month}`} key={invoice.payment_month}><div><span>{invoice.payment_month}</span><strong>{new Intl.NumberFormat("pt-BR", { style: "currency", currency: invoice.currency_code }).format(invoice.projected_total_cents / 100)}</strong></div><small>{invoice.component_count} componente(s) · vencimento dia {invoice.nominal_due_day}</small></Link>)}</div>}</section></main>;
}
