import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { AppHeaderTitle } from "@/components/app-navigation";
import { ItemActions } from "@/components/entries/item-detail";
import { PaymentHistory } from "@/components/payment-history";
import { PaymentChangeForm } from "@/components/payment-change-form";
import { FinancialApiError, getFinancialItem, getCreditCards, getPaymentMethodHistory, getProfile } from "@/lib/api/financial-api";
import { currentMonth, formatMonth, isMonth, selectPeriod } from "@/lib/debt-month";

export default async function ItemDetailPage({ params, searchParams }: { params: Promise<{ itemId: string }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const { itemId } = await params;
  const query = await searchParams;
  const [item, profile] = await Promise.all([getFinancialItem(itemId).catch((error: unknown) => { if (error instanceof FinancialApiError && error.status === 404) notFound(); throw error; }), getProfile()]);
  const month = isMonth(query.month) ? query.month : currentMonth(profile.timezone);
  if (item.kind === "debt_installment") redirect(`/debts/${encodeURIComponent(item.id)}?as_of=${month}`);
  const income = item.kind === "recurring_income" || item.kind === "one_time_income";
  const [cards, history] = await Promise.all([income ? [] : getCreditCards(), income ? null : getPaymentMethodHistory(item.id)]);
  const period = selectPeriod(item.periods, month);
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: item.currency_code });
  const tab = query.tab === "payments" && !income ? "payments" : "values";
  const returnQuery = new URLSearchParams();
  if (typeof query.kind === "string" && ["recurring_income", "one_time_income", "fixed_expense", "projected_variable_expense", "debt_installment"].includes(query.kind)) returnQuery.set("kind", query.kind);
  if (query.status === "active" || query.status === "archived") returnQuery.set("status", query.status);
  const queryFor = (nextTab: string) => new URLSearchParams({ ...Object.fromEntries(returnQuery), month, tab: nextTab });
  return <main className="resource-page-shell"><AppHeaderTitle title={item.name} /><Link className="back-link" href={`/entries?${returnQuery}`}>Voltar aos registros</Link><p>{item.description} · {item.status === "active" ? "Ativo" : "Arquivado"}</p><form className="resource-form resource-query-form" method="get">{[...returnQuery].map(([key, value]) => <input type="hidden" key={key} name={key} value={value} />)}<input type="hidden" name="tab" value={tab} /><label className="native-field"><span>Mês de consulta</span><input type="month" name="month" defaultValue={month} required /></label><button type="submit">Consultar</button></form><p>Valor em {formatMonth(month)}: <strong>{period ? money.format(period.amount_cents / 100) : "Sem vigência neste mês"}</strong></p><nav className="resource-card-footer" aria-label="Detalhes do registro"><Link aria-current={tab === "values" ? "page" : undefined} href={`?${queryFor("values")}`}>Valores</Link>{!income && <Link aria-current={tab === "payments" ? "page" : undefined} href={`?${queryFor("payments")}`}>Pagamentos</Link>}</nav>{tab === "values" ? <><section className="debt-action-panel"><h2>Histórico completo de valores</h2><div className="resource-history"><table><thead><tr><th>Vigência</th><th>Valor</th><th>Recorrência</th><th>Caixa</th><th>Contexto</th></tr></thead><tbody>{item.periods.map((value) => <tr key={value.id}><td>{formatMonth(value.start_month)} a {value.end_month ? formatMonth(value.end_month) : "sem término"}{value.start_month > month ? " · futura" : value === period ? " · vigente" : ""}</td><td>{money.format(value.amount_cents / 100)}</td><td>{value.recurrence === "once" ? "Pontual" : "Mensal"}</td><td>+{value.cash_month_offset} mês(es)</td><td>{value.context ?? "—"}</td></tr>)}</tbody></table></div></section><ItemActions key={item.updated_at + month} item={item} month={month} /></> : history && <><PaymentHistory history={history} cards={cards} /><PaymentChangeForm itemId={item.id} cards={cards} month={month} disabled={item.status !== "active"} /></>}</main>;
}
