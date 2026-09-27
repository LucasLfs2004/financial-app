import type { Metadata } from "next";
import { ArrowLeft, Landmark } from "lucide-react";
import Link from "next/link";
import { DebtCard } from "@/components/debts/debt-card";
import { DebtForm } from "@/components/debts/debt-form";
import { getCreditCards, getDebts } from "@/lib/api/financial-api";

export const metadata: Metadata = { title: "Debts" };

export default async function DebtsPage() {
  const [debts, cards] = await Promise.all([getDebts().catch(() => []), getCreditCards().catch(() => [])]);

  return (
    <main className="resource-page-shell">
      <header className="resource-page-header">
        <Link className="back-link" href="/dashboard"><ArrowLeft size={16} /> Voltar ao dashboard</Link>
        <span className="eyebrow">DÍVIDAS</span>
        <h1>Suas dívidas</h1>
        <p>Cadastre parcelas e acompanhe o impacto de cada compromisso na sua projeção.</p>
      </header>

      <div className="resource-page-grid debts-grid">
        <div id="new-debt"><DebtForm cards={cards} /></div>
        <section className="resource-list-section">
          <div className="section-heading compact"><div><span className="section-kicker">ACOMPANHAMENTO</span><h2>Dívidas cadastradas</h2></div><Landmark size={20} /></div>
          {debts.length === 0 ? <div className="empty-resource-inline">Você ainda não cadastrou nenhuma dívida.</div> : <div className="resource-list">{debts.map((debt) => <DebtCard debt={debt} key={debt.id} />)}</div>}
        </section>
      </div>
    </main>
  );
}
