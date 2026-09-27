import type { Metadata } from "next";
import { Card, CardContent } from "@lucaslfs2004/luke-ui";
import { ArrowLeft, Camera, ChartNoAxesCombined, CircleDollarSign, ReceiptText } from "lucide-react";
import Link from "next/link";
import { PlanForm } from "@/components/planning/plan-form";
import { FinancialItemCard } from "@/components/planning/financial-item-card";
import { FinancialItemForm } from "@/components/planning/financial-item-form";
import { SavingsForm } from "@/components/planning/savings-form";
import { getCreditCards, getCurrentPlan, getFinancialItems, getOriginalPlan, getPlannedSavings } from "@/lib/api/financial-api";

export const metadata: Metadata = { title: "Planning" };

export default async function PlanningPage() {
  const [plan, items, savings, original, cards] = await Promise.all([
    getCurrentPlan().catch(() => null),
    getFinancialItems().catch(() => []),
    getPlannedSavings().catch(() => null),
    getOriginalPlan().catch(() => null),
    getCreditCards().catch(() => []),
  ]);

  return (
    <main className="resource-page-shell">
      <header className="resource-page-header">
        <Link className="back-link" href="/dashboard"><ArrowLeft size={16} /> Voltar ao dashboard</Link>
        <span className="eyebrow">PLANEJAMENTO</span>
        <h1>{plan ? "Seu planejamento" : "Crie seu planejamento"}</h1>
        <p>{plan ? "Revise o período e acompanhe os itens que formam a sua projeção." : "Defina o período base para começar a projetar seu dinheiro."}</p>
      </header>

      <div className="resource-page-grid planning-grid">
        <div className="planning-forms-column">
          <PlanForm plan={plan} />
          {plan && <SavingsForm savings={savings} />}
          {original && <Card className="snapshot-card" padding="none"><CardContent><Camera size={20} /><div><strong>Referência original criada</strong><span>Capturada em {new Date(original.captured_at).toLocaleDateString("pt-BR")}. Ela permanece imutável.</span></div></CardContent></Card>}
        </div>
        <section className="resource-list-section">
          <div className="section-heading compact"><div><span className="section-kicker">PROJEÇÃO</span><h2>Itens cadastrados</h2></div><ChartNoAxesCombined size={20} /></div>
          {plan && <FinancialItemForm />}
          {items.length === 0 ? (
            <Card className="empty-resource-card" padding="none"><CardContent><CircleDollarSign size={22} /><h3>Nenhum item ainda</h3><p>Adicione sua primeira receita ou despesa para preencher a projeção.</p></CardContent></Card>
          ) : (
            <div className="resource-list">{items.map((item) => <FinancialItemCard item={item} cards={cards} key={item.id} />)}</div>
          )}
          <div className="resource-hint"><ReceiptText size={16} /> As telas de receitas e despesas podem ser adicionadas sobre o mesmo contrato de itens financeiros.</div>
        </section>
      </div>
    </main>
  );
}
