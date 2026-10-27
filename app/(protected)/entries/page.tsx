import { currentMonth } from "@/lib/debt-month";
import type { Metadata } from "next";
import { FinancialItemForm } from "@/components/entries/financial-item-form";
import { FinancialItemsList } from "@/components/entries/financial-items-list";
import { getCreditCards, getFinancialItems, getProfile } from "@/lib/api/financial-api";

export const metadata: Metadata = { title: "Receitas e despesas" };

export default async function EntriesPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const query = await searchParams;
  const kinds = ["recurring_income", "one_time_income", "fixed_expense", "projected_variable_expense", "debt_installment"];
  const kind = typeof query.kind === "string" && kinds.includes(query.kind) ? query.kind : "all";
  const status = query.status === "active" || query.status === "archived" ? query.status : "all";
  const [items, cards, profile] = await Promise.all([
    getFinancialItems({ kind: kind === "all" ? undefined : kind, status: status === "all" ? undefined : status }),
    getCreditCards(),
    getProfile(),
  ]);

  return <main className="resource-page-shell entries-page">
    <header className="resource-page-header"><p>Organize entradas e saídas da sua projeção.</p></header>
    <div id="new-item" className="entries-form"><FinancialItemForm cards={cards} defaultMonth={currentMonth(profile.timezone)} currencyCode={profile.currency_code} /></div>
    <form className="resource-form resource-query-form" method="get"><div className="form-grid-two"><label className="native-field"><span>Tipo</span><select name="kind" defaultValue={kind}><option value="all">Todos</option><option value="recurring_income">Renda recorrente</option><option value="one_time_income">Renda pontual</option><option value="fixed_expense">Despesa fixa</option><option value="projected_variable_expense">Despesa variável</option><option value="debt_installment">Dívida</option></select></label><label className="native-field"><span>Status</span><select name="status" defaultValue={status}><option value="all">Todos</option><option value="active">Ativos</option><option value="archived">Arquivados</option></select></label></div><button type="submit">Filtrar</button></form>
    <FinancialItemsList items={items} cards={cards} month={currentMonth(profile.timezone)} returnQuery={new URLSearchParams({ kind, status }).toString()} />
  </main>;
}
