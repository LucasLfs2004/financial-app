import type { Metadata } from "next";
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
    <header className="resource-page-header"><p>Organize entradas e saídas da sua projeção.</p></header>
    <div id="new-item" className="entries-form"><FinancialItemForm /></div>
    <FinancialItemsList items={items} cards={cards} />
  </main>;
}
