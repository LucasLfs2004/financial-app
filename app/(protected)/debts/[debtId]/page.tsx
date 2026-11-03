import { AppHeaderTitle } from "@/components/app-navigation";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { ArrowLeft, CalendarDays, CreditCard, Landmark } from "lucide-react";
import { FinancialApiError, getCreditCards, getDebt, getDebtSchedule, getPaymentMethodHistory } from "@/lib/api/financial-api";
import { DebtActions } from "@/components/debts/debt-actions";
import { queryWindow } from "@/lib/query-window";
import { addMonths, inclusiveMonthCount, isMonth, formatMonth } from "@/lib/debt-month";

export const metadata: Metadata = { title: "Detalhe da dívida" };

export default async function DebtDetailPage({ params, searchParams }: { params: Promise<{ debtId: string }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const { debtId } = await params;
  const query = await searchParams;
  const asOf = isMonth(query.as_of) ? query.as_of : undefined;
  const debt = await getDebt(debtId, asOf).catch((error: unknown) => {
    if (error instanceof FinancialApiError && error.status === 404) notFound();
    if (error instanceof FinancialApiError && error.status === 401) redirect("/sign-in");
    throw error;
  });
  const { from, to, adjusted } = queryWindow(query, debt.as_of_month, { from: debt.scheduled_start_month, to: debt.effective_end_month }, 120);
  const [schedule, cards, paymentHistory] = await Promise.all([
    getDebtSchedule(debt.id, from, to),
    getCreditCards(),
    getPaymentMethodHistory(debt.id),
  ]);
  const analysis = debt.as_of_month;
  const returnTo = new URLSearchParams(typeof query.returnTo === "string" ? query.returnTo : "");
  const listQuery = new URLSearchParams({ as_of: analysis });
  for (const key of ["status", "projection_status", "releaseFrom", "releaseTo"]) {
    const value = returnTo.get(key);
    if (value && (key.startsWith("release") ? isMonth(value) : ["all", "active", "archived", "planned", "completed", "settled_early"].includes(value))) listQuery.set(key, value);
  }
  const windowSize = inclusiveMonthCount(from, to);
  function windowLink(start: string) {
    const end = addMonths(start, Math.min(windowSize - 1, inclusiveMonthCount(start, debt.effective_end_month) - 1));
    return `/debts/${encodeURIComponent(debt.id)}?${new URLSearchParams({ as_of: analysis, from: start, to: end, returnTo: listQuery.toString() })}`;
  }
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: debt.currency_code });
  return <main className="resource-page-shell debt-detail-page"><AppHeaderTitle title={debt.name} />
    <Link className="back-link" href={`/debts?${listQuery}`}><ArrowLeft size={16} /> Voltar às dívidas</Link>
    <header className="resource-page-header"><p>{debt.description || "Acompanhe as parcelas, a forma de pagamento e o fim deste compromisso."}</p></header>
    <div className="debt-detail-stats"><div><span>Parcelas restantes</span><strong>{debt.remaining_installments}</strong><small>Em {formatMonth(debt.as_of_month)}</small></div><div><span>Última parcela</span><strong>{formatMonth(debt.effective_end_month)}</strong><small>{debt.settlement ? "Quitação antecipada projetada" : "Término previsto"}</small></div><div><span>Valor mensal liberado</span><strong>{money.format(debt.released_monthly_cents / 100)}</strong><small>A partir de {formatMonth(debt.release_from_month)} · informativo</small></div></div>
    <p className="debt-projection-note">Os valores abaixo são projeções. A quitação e as parcelas não confirmam pagamentos realizados.</p>
    <div className="debt-detail-grid"><section><div className="section-heading compact"><div><span className="section-kicker">PARCELAS</span><h2>Cronograma</h2></div><CalendarDays size={20} /></div>
      <form className="resource-form resource-query-form" method="get"><input type="hidden" name="as_of" value={analysis} /><input type="hidden" name="returnTo" value={listQuery.toString()} /><div className="form-grid-two"><label className="native-field"><span>De</span><input type="month" name="from" defaultValue={from} min={debt.scheduled_start_month} max={debt.effective_end_month} required /></label><label className="native-field"><span>Até (máximo 120 meses)</span><input type="month" name="to" defaultValue={to} min={from} max={debt.effective_end_month} required /></label></div><button type="submit">Consultar período</button></form>
      {adjusted && <p role="alert">Intervalo ajustado aos limites do cronograma.</p>}
      <nav className="resource-card-footer" aria-label="Janelas do cronograma">{from > debt.scheduled_start_month && <Link href={windowLink(addMonths(from, -Math.min(windowSize, inclusiveMonthCount(debt.scheduled_start_month, from) - 1)))}>Período anterior</Link>}<span>{formatMonth(schedule.range.from)} a {formatMonth(schedule.range.to)}</span>{to < debt.effective_end_month && <Link href={windowLink(addMonths(to, 1))}>Próximo período</Link>}</nav>
      {schedule.occurrences.length === 0 && <p>Sem parcelas neste intervalo.</p>}
      {!schedule ? <div className="empty-resource-inline">Não foi possível carregar o cronograma.</div> : <div className="debt-schedule-list">{schedule.occurrences.map((occurrence) => <div className="debt-schedule-row" key={`${occurrence.reference_month}-${occurrence.source_id}`}><span className="debt-schedule-index">{occurrence.installment_number}/{occurrence.installments_total}</span><div><strong>{formatMonth(occurrence.reference_month)}</strong><small>{occurrence.debt_occurrence_kind === "early_settlement" ? "Quitação antecipada projetada" : occurrence.payment_method === "credit_card" ? `No cartão · fatura ${occurrence.invoice_payment_month ? formatMonth(occurrence.invoice_payment_month) : "—"}` : `Pagamento direto · caixa ${formatMonth(occurrence.cash_month)}`}</small></div><strong>{money.format(occurrence.amount_cents / 100)}</strong>{occurrence.payment_method === "credit_card" && <CreditCard size={16} />}</div>)}</div>}
    </section><aside><div className="section-heading compact"><div><span className="section-kicker">AJUSTES</span><h2>Gerenciar dívida</h2></div><Landmark size={20} /></div><DebtActions debt={debt} cards={cards} paymentHistory={paymentHistory} /></aside></div>
  </main>;
}
