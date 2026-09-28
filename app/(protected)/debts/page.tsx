import type { Metadata } from "next";
import { ArrowLeft, Landmark, TrendingUp } from "lucide-react";
import Link from "next/link";
import { DebtCard } from "@/components/debts/debt-card";
import { DebtForm } from "@/components/debts/debt-form";
import { getCreditCards, getDebtReleases, getDebts } from "@/lib/api/financial-api";
import { addMonths, currentMonth, formatMonth } from "@/lib/debt-month";

export const metadata: Metadata = { title: "Debts" };

export default async function DebtsPage() {
  const month = currentMonth();
  const [debts, cards, releases] = await Promise.all([getDebts().catch(() => null), getCreditCards().catch(() => []), getDebtReleases(month, addMonths(month, 11)).catch(() => null)]);
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });

  return (
    <main className="resource-page-shell">
      <header className="resource-page-header">
        <Link className="back-link" href="/dashboard"><ArrowLeft size={16} /> Voltar ao dashboard</Link>
        <span className="eyebrow">DÍVIDAS</span>
        <h1>Suas dívidas</h1>
        <p>Cadastre parcelas, acompanhe o cronograma e veja quando cada compromisso termina.</p>
      </header>

      {releases && releases.releases.length > 0 && <section className="debt-releases-section"><div className="section-heading compact"><div><span className="section-kicker">PRÓXIMOS 12 MESES</span><h2>Valores liberados</h2></div><TrendingUp size={20} /></div><p>Quando uma dívida termina, este valor deixa de ser compromisso projetado. Ele não entra como renda no resumo.</p><div className="debt-release-months">{releases.monthly_totals.filter((total) => total.released_monthly_cents > 0).slice(0, 4).map((total) => <div key={total.month}><span>{formatMonth(total.month)}</span><strong>{money.format(total.released_monthly_cents / 100)}/mês</strong></div>)}</div><div className="debt-release-list">{releases.releases.map((release) => <Link href={`/debts/${release.debt_id}`} key={release.debt_id}><span><strong>{release.name}</strong><small>A partir de {formatMonth(release.release_from_month)}{release.reason === "early_settlement" ? " · quitação antecipada" : ""}</small></span><strong>{money.format(release.released_monthly_cents / 100)}/mês</strong></Link>)}</div></section>}

      <div className="resource-page-grid debts-grid">
        <div id="new-debt"><DebtForm cards={cards} /></div>
        <section className="resource-list-section">
          <div className="section-heading compact"><div><span className="section-kicker">ACOMPANHAMENTO</span><h2>Dívidas cadastradas</h2></div><Landmark size={20} /></div>
          {debts === null ? <div className="empty-resource-inline">Não foi possível carregar as dívidas. Tente atualizar a página.</div> : debts.length === 0 ? <div className="empty-resource-inline">Você ainda não cadastrou nenhuma dívida.</div> : <div className="resource-list">{debts.map((debt) => <DebtCard debt={debt} key={debt.id} />)}</div>}
        </section>
      </div>
    </main>
  );
}
