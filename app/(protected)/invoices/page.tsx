import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, ReceiptText } from "lucide-react";
import { getCreditCards } from "@/lib/api/financial-api";
import { InvoiceCsvImport } from "@/components/cards/invoice-csv-import";

export const metadata: Metadata = { title: "Faturas" };

export default async function InvoicesPage() {
  const cards = await getCreditCards().catch(() => []);
  return <main className="resource-page-shell">
    <header className="resource-page-header"><span className="eyebrow">FATURAS</span><h1>Suas faturas</h1><p>Escolha um cartão para acompanhar as próximas faturas e seus componentes.</p></header>
    {cards.some((card) => card.status === "active") && <InvoiceCsvImport cards={cards} />}
    <div className="invoice-grid">{cards.length ? cards.map((card) => <Link className="invoice-card" href={`/invoices/${card.id}`} key={card.id}><div><span>{card.institution.name}</span><ReceiptText size={20} /></div><strong>{card.name}</strong><small>Ver próximas faturas <ArrowRight size={14} /></small></Link>) : <div className="empty-resource-inline">Nenhum cartão cadastrado. <Link className="small-link" href="/cards#new-card">Cadastrar cartão</Link></div>}</div>
  </main>;
}
