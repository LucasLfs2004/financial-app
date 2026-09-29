import type { Metadata } from "next";
import Link from "next/link";
import { ArrowLeft, CalendarDays, CreditCard, Landmark } from "lucide-react";
import { getCreditCards, getDebt, getDebtSchedule, getPaymentMethodHistory } from "@/lib/api/financial-api";
import { DebtActions } from "@/components/debts/debt-actions";
import { formatMonth } from "@/lib/debt-month";

export const metadata: Metadata = { title: "Detalhe da dívida" };

export default async function DebtDetailPage({ params }: { params: Promise<{ debtId: string }> }) {
  const { debtId } = await params;
  const cardsPromise = getCreditCards().catch(() => []);
  const debt = await getDebt(debtId).catch(() => null);
  if (!debt) return <main className="resource-page-shell"><Link className="back-link" href="/debts"><ArrowLeft size={16} /> Voltar às dívidas</Link><div className="empty-resource-inline">Dívida não encontrada ou indisponível.</div></main>;
  const [schedule, cards, paymentHistory] = await Promise.all([
    getDebtSchedule(debt.id, debt.scheduled_start_month, debt.effective_end_month).catch(() => null),
    cardsPromise,
    getPaymentMethodHistory(debt.id).catch(() => null),
  ]);
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: debt.currency_code });
  return <main className="resource-page-shell debt-detail-page">
    <header className="resource-page-header"><Link className="back-link" href="/debts"><ArrowLeft size={16} /> Voltar às dívidas</Link><span className="eyebrow">CRONOGRAMA PROJETADO</span><h1>{debt.name}</h1><p>{debt.description || "Acompanhe as parcelas, a forma de pagamento e o fim deste compromisso."}</p></header>
    <div className="debt-detail-stats"><div><span>Parcelas restantes</span><strong>{debt.remaining_installments}</strong><small>Em {formatMonth(debt.as_of_month)}</small></div><div><span>Última parcela</span><strong>{formatMonth(debt.effective_end_month)}</strong><small>{debt.settlement ? "Quitação antecipada projetada" : "Término previsto"}</small></div><div><span>Valor mensal liberado</span><strong>{money.format(debt.released_monthly_cents / 100)}</strong><small>A partir de {formatMonth(debt.release_from_month)} · informativo</small></div></div>
    <p className="debt-projection-note">Os valores abaixo são projeções. A quitação e as parcelas não confirmam pagamentos realizados.</p>
    <div className="debt-detail-grid"><section><div className="section-heading compact"><div><span className="section-kicker">PARCELAS</span><h2>Cronograma</h2></div><CalendarDays size={20} /></div>
      {!schedule ? <div className="empty-resource-inline">Não foi possível carregar o cronograma.</div> : <div className="debt-schedule-list">{schedule.occurrences.map((occurrence) => <div className="debt-schedule-row" key={`${occurrence.reference_month}-${occurrence.source_id}`}><span className="debt-schedule-index">{occurrence.installment_number}/{occurrence.installments_total}</span><div><strong>{formatMonth(occurrence.reference_month)}</strong><small>{occurrence.debt_occurrence_kind === "early_settlement" ? "Quitação antecipada projetada" : occurrence.payment_method === "credit_card" ? `No cartão · fatura ${occurrence.invoice_payment_month ? formatMonth(occurrence.invoice_payment_month) : "—"}` : `Pagamento direto · caixa ${formatMonth(occurrence.cash_month)}`}</small></div><strong>{money.format(occurrence.amount_cents / 100)}</strong>{occurrence.payment_method === "credit_card" && <CreditCard size={16} />}</div>)}</div>}
    </section><aside><div className="section-heading compact"><div><span className="section-kicker">AJUSTES</span><h2>Gerenciar dívida</h2></div><Landmark size={20} /></div><DebtActions debt={debt} cards={cards} paymentHistory={paymentHistory} schedule={schedule?.occurrences ?? []} /></aside></div>
  </main>;
}
