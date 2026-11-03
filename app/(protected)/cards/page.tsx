import type { Metadata } from "next";
import { InstitutionList } from "@/components/cards/institution-list";
import { MonthNavigation } from "@/components/month-navigation";
import { addMonths, clampMonth, currentMonth, isMonth } from "@/lib/debt-month";
import { CardsHub } from "@/components/cards/cards-hub";
import { getCardInvoice, getCreditCards, getCurrentPlan, getInstitutions, getProfile, absentOn404 } from "@/lib/api/financial-api";
import { unavailableRead } from "@/lib/api/optional-read";
import type { CardInvoice } from "@/lib/api/types";
import Link from "next/link";

export const metadata: Metadata = { title: "Cartões" };

export default async function CardsPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const query = await searchParams;
  const status = query.status === "active" || query.status === "archived" ? query.status : "all";
  const institutionStatus = query.institutionStatus === "active" || query.institutionStatus === "archived" ? query.institutionStatus : "all";
  const [institutions, allCards, profile, plan] = await Promise.all([getInstitutions(), getCreditCards(), getProfile(), getCurrentPlan().catch(absentOn404)]);
  const requested = isMonth(query.month) ? query.month : currentMonth(profile.timezone);
  const min = plan?.start_month;
  const max = plan ? addMonths(plan.end_month, 12) : undefined;
  const month = min && max ? clampMonth(requested, min, max) : requested;
  const cards = allCards.filter((card) => status === "all" || card.status === status);
  const entries = await Promise.all(cards.map(async (card) => [card.id, plan ? await getCardInvoice(card.id, month).catch(unavailableRead) : null] as const));
  const invoices: Record<string, CardInvoice | null> = Object.fromEntries(entries);
  const loaded = entries.map(([, invoice]) => invoice);
  const complete = cards.length > 0 && loaded.every((invoice) => invoice !== null && invoice.currency_code === profile.currency_code);
  const total = complete ? loaded.reduce((sum, invoice) => sum + invoice!.projected_total_cents, 0) : null;
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: profile.currency_code });
  const returnQuery = new URLSearchParams({ status, institutionStatus, month }).toString();
  return <main className="resource-page-shell cards-page finance-cards-page">
    <MonthNavigation month={month} min={min} max={max} basePath="/cards" query={new URLSearchParams({ status, institutionStatus }).toString()} />
    <section className="finance-hero" aria-label="Faturas do mês"><span className="section-kicker">FATURAS PREVISTAS NO MÊS</span><strong>{total === null ? "—" : money.format(total / 100)}</strong><p>{status === "archived" ? "Cartões arquivados no filtro." : "Valores líquidos de compras, cobranças e pagamentos."}</p><small>Não some este valor novamente aos compromissos do resumo.</small></section>
    {!plan && <p className="finance-notice">Crie um <Link href="/planning">planejamento</Link> para consultar os valores das faturas.</p>}
    {plan && !complete && cards.length > 0 && <p className="finance-notice" role="status">Não foi possível consolidar todas as faturas. Os valores indisponíveis não são tratados como zero. <Link href={`/cards?${returnQuery}`}>Tentar novamente</Link></p>}
    <details className="finance-filters"><summary>Filtrar cartões · {status === "all" ? "Todos" : status === "active" ? "Ativos" : "Arquivados"}</summary><form className="resource-form" method="get"><input type="hidden" name="month" value={month} /><label className="native-field"><span>Cartões</span><select name="status" defaultValue={status}><option value="all">Todos</option><option value="active">Ativos</option><option value="archived">Arquivados</option></select></label><input type="hidden" name="institutionStatus" value={institutionStatus} /><button type="submit">Aplicar filtro</button></form></details>
    <CardsHub institutions={institutions} cards={cards} invoices={invoices} defaultMonth={month} returnQuery={returnQuery} />
    <details className="finance-disclosure"><summary>Gerenciar instituições financeiras</summary><form className="resource-form" method="get"><input type="hidden" name="month" value={month} /><input type="hidden" name="status" value={status} /><label className="native-field"><span>Instituições</span><select name="institutionStatus" defaultValue={institutionStatus}><option value="all">Todas</option><option value="active">Ativas</option><option value="archived">Arquivadas</option></select></label><button type="submit">Aplicar filtro</button></form><InstitutionList institutions={institutions.filter((institution) => institutionStatus === "all" || institution.status === institutionStatus)} /></details>
  </main>;
}
