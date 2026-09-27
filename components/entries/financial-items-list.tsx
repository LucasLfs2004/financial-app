"use client";

import { useState } from "react";
import { ArrowDownLeft, ArrowUpRight, ListFilter } from "lucide-react";
import type { CreditCard, FinancialItem } from "@/lib/api/types";
import { FinancialItemCard } from "@/components/entries/financial-item-card";

type Filter = "all" | "income" | "expense";

function isIncome(item: FinancialItem) {
  return item.kind === "recurring_income" || item.kind === "one_time_income";
}

export function FinancialItemsList({ items, cards }: { items: FinancialItem[]; cards: CreditCard[] }) {
  const [filter, setFilter] = useState<Filter>("all");
  const visibleItems = items.filter((item) => filter === "all" || (filter === "income" ? isIncome(item) : !isIncome(item)));
  const incomeCount = items.filter(isIncome).length;
  const expenseCount = items.length - incomeCount;

  return <section className="financial-items-section" aria-label="Receitas e despesas cadastradas">
    <div className="financial-items-heading"><div><span className="section-kicker">SEUS REGISTROS</span><h2>Receitas e despesas</h2><p>Valores usados na projeção, organizados por tipo.</p></div><ListFilter size={19} /></div>
    <div className="financial-items-filters" role="group" aria-label="Filtrar registros">
      <button type="button" className={filter === "all" ? "active" : ""} aria-pressed={filter === "all"} onClick={() => setFilter("all")}>Todos <span>{items.length}</span></button>
      <button type="button" className={filter === "income" ? "active" : ""} aria-pressed={filter === "income"} onClick={() => setFilter("income")}><ArrowDownLeft size={15} /> Receitas <span>{incomeCount}</span></button>
      <button type="button" className={filter === "expense" ? "active" : ""} aria-pressed={filter === "expense"} onClick={() => setFilter("expense")}><ArrowUpRight size={15} /> Despesas <span>{expenseCount}</span></button>
    </div>
    {visibleItems.length ? <div className="financial-items-grid">{visibleItems.map((item) => <FinancialItemCard item={item} cards={cards} key={item.id} />)}</div> : <div className="empty-resource-inline">{items.length ? "Nenhum registro desse tipo." : "Nenhuma receita ou despesa cadastrada ainda."}</div>}
  </section>;
}
