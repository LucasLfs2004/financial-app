import type { Metadata } from "next";
import { Landmark, TrendingUp } from "lucide-react";
import Link from "next/link";
import { DebtCard } from "@/components/debts/debt-card";
import { DebtForm } from "@/components/debts/debt-form";
import { getCreditCards, getDebtReleases, getDebts, getProfile } from "@/lib/api/financial-api";
import { addMonths, currentMonth, formatMonth, isMonth, inclusiveMonthCount } from "@/lib/debt-month";

export const metadata: Metadata = { title: "Debts" };

export default async function DebtsPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const query = await searchParams;
  const profile = await getProfile();
  const month = isMonth(query.as_of) ? query.as_of : currentMonth(profile.timezone);
  const status = query.status === "active" || query.status === "archived" ? query.status : "";
  const projections = ["planned", "active", "completed", "settled_early", "archived"];
  const projectionStatus = typeof query.projection_status === "string" && projections.includes(query.projection_status) ? query.projection_status : "";
  const releaseFrom = isMonth(query.releaseFrom) ? query.releaseFrom : month;
  const requestedTo = isMonth(query.releaseTo) ? query.releaseTo : addMonths(releaseFrom, 11);
  const validRange = inclusiveMonthCount(releaseFrom, requestedTo) >= 1 && inclusiveMonthCount(releaseFrom, requestedTo) <= 120;
  const releaseTo = validRange ? requestedTo : addMonths(releaseFrom, 11);
  const [debts, cards, releaseResponse] = await Promise.all([getDebts(month, { status, projection_status: projectionStatus }), getCreditCards(), getDebtReleases(releaseFrom, releaseTo)]);
  const releases = releaseResponse.data;
  const returnQuery = new URLSearchParams({ as_of: month, status: status || "all", projection_status: projectionStatus || "all", releaseFrom, releaseTo }).toString();
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: profile.currency_code });

  return (
    <main className="resource-page-shell">
      <header className="resource-page-header">
        <p>Acompanhe parcelas e o fim de cada compromisso.</p>
      </header>

      <form className="resource-form resource-query-form" method="get"><div className="form-grid-two"><label className="native-field"><span>Mês de análise</span><input type="month" name="as_of" defaultValue={month} required /></label><label className="native-field"><span>Status</span><select name="status" defaultValue={status || "all"}><option value="all">Todos</option><option value="active">Ativas</option><option value="archived">Arquivadas</option></select></label><label className="native-field"><span>Projeção</span><select name="projection_status" defaultValue={projectionStatus || "all"}><option value="all">Todas</option><option value="planned">Planejada</option><option value="active">Em andamento</option><option value="completed">Concluída</option><option value="settled_early">Quitação projetada</option><option value="archived">Arquivada</option></select></label><label className="native-field"><span>Liberações de</span><input type="month" name="releaseFrom" defaultValue={releaseFrom} required /></label><label className="native-field"><span>Liberações até (máximo 120 meses)</span><input type="month" name="releaseTo" defaultValue={releaseTo} required /></label></div><button type="submit">Consultar</button></form>
      {!validRange && <p role="alert">Intervalo de liberações inválido. Exibindo até 12 meses a partir do mês inicial.</p>}
      <p>Liberações: {formatMonth(releaseResponse.range.from)} a {formatMonth(releaseResponse.range.to)}.</p>
      {releases.releases.length === 0 && <p>Nenhuma liberação no período selecionado.</p>}
      {releases && releases.releases.length > 0 && <section className="debt-releases-section"><div className="section-heading compact"><div><span className="section-kicker">PERÍODO SELECIONADO</span><h2>Valores liberados</h2></div><TrendingUp size={20} /></div><p>Quando uma dívida termina, este valor deixa de ser compromisso projetado. Ele não entra como renda no resumo.</p><div className="debt-release-months">{releases.monthly_totals.filter((total) => total.released_monthly_cents > 0).slice(0, 4).map((total) => <div key={total.month}><span>{formatMonth(total.month)}</span><strong>{money.format(total.released_monthly_cents / 100)}/mês</strong></div>)}</div><div className="debt-release-list">{releases.releases.map((release) => <Link href={`/debts/${encodeURIComponent(release.debt_id)}?${new URLSearchParams({ as_of: month, returnTo: returnQuery })}`} key={release.debt_id}><span><strong>{release.name}</strong><small>A partir de {formatMonth(release.release_from_month)}{release.reason === "early_settlement" ? " · quitação antecipada" : ""}</small></span><strong>{money.format(release.released_monthly_cents / 100)}/mês</strong></Link>)}</div></section>}

      <div className="resource-page-grid debts-grid">
        <div id="new-debt"><DebtForm cards={cards} defaultMonth={currentMonth(profile.timezone)} currencyCode={profile.currency_code} /></div>
        <section className="resource-list-section">
          <div className="section-heading compact"><div><span className="section-kicker">ACOMPANHAMENTO</span><h2>Dívidas cadastradas</h2></div><Landmark size={20} /></div>
          {debts === null ? <div className="empty-resource-inline">Não foi possível carregar as dívidas. Tente atualizar a página.</div> : debts.length === 0 ? <div className="empty-resource-inline">Nenhuma dívida corresponde aos filtros.</div> : <div className="resource-list">{debts.map((debt) => <DebtCard debt={debt} returnQuery={returnQuery} key={debt.id} />)}</div>}
        </section>
      </div>
    </main>
  );
}
