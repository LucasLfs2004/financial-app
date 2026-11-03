import Link from "next/link";
import type { CreditCard, FinancialItem } from "@/lib/api/types";
import { formatMonth, selectPeriod } from "@/lib/debt-month";

const labels: Record<string, string> = { recurring_income: "Renda recorrente", one_time_income: "Renda pontual", fixed_expense: "Despesa fixa", projected_variable_expense: "Despesa variável", debt_installment: "Dívida" };
export function FinancialItemCard({ item, month, returnQuery }: { item: FinancialItem; cards: CreditCard[]; month: string; returnQuery: string }) {
  const period = selectPeriod(item.periods, month);
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: item.currency_code });
  const income = item.kind === "recurring_income" || item.kind === "one_time_income";
  const detail = item.kind === "debt_installment" ? `/debts/${encodeURIComponent(item.id)}?as_of=${month}` : `/entries/${encodeURIComponent(item.id)}?month=${month}&${returnQuery}`;
  return <article className={`financial-record-card ${income ? "income" : "expense"} ${item.status === "archived" ? "archived" : ""}`}><div className="financial-record-title"><div><span>{labels[item.kind]}</span><h3>{item.name}</h3></div><span>{item.status === "active" ? "Ativo" : "Arquivado"}</span></div><div className="financial-record-value"><span>Valor em {formatMonth(month)}</span><strong>{period ? money.format(period.amount_cents / 100) : "Sem vigência"}</strong></div>{item.description && <p>{item.description}</p>}{item.invoice_match_title && <p>Na fatura: {item.invoice_match_title}</p>}{item.renewal_day != null && <p>Renovação nominal: dia {item.renewal_day}</p>}<div className="financial-record-actions"><span>{item.periods.length} vigência(s)</span><Link href={detail}>Ver histórico e gerenciar</Link></div></article>;
}
