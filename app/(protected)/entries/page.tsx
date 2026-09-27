import type { Metadata } from "next";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { FinancialItemForm } from "@/components/entries/financial-item-form";
import { FinancialItemsList } from "@/components/entries/financial-items-list";
import { getCreditCards, getFinancialItems } from "@/lib/api/financial-api";

export const metadata: Metadata = { title: "Receitas e despesas" };

export default async function EntriesPage() {
  const [items, cards] = await Promise.all([
    getFinancialItems().catch(() => []),
    getCreditCards().catch(() => []),
  ]);

  return <main className="resource-page-shell entries-page">
    <header className="resource-page-header"><Link className="back-link" href="/dashboard"><ArrowLeft size={16} /> Voltar à visão geral</Link><span className="eyebrow">SEU DINHEIRO</span><h1>Receitas e despesas</h1><p>Cadastre cada entrada e saída uma vez. O planejamento usa esses registros para mostrar sua projeção.</p></header>
    <div id="new-item" className="entries-form"><FinancialItemForm /></div>
    <FinancialItemsList items={items} cards={cards} />
  </main>;
}
