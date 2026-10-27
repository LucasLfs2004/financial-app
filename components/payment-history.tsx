import Link from "next/link";
import type { CreditCard, PaymentMethodHistory } from "@/lib/api/types";
import { formatMonth } from "@/lib/debt-month";

export function PaymentHistory({ history, cards }: { history: PaymentMethodHistory; cards: CreditCard[] }) {
  return <section className="debt-action-panel"><h2>Histórico de pagamentos</h2>{history.periods.length === 0 ? <p>Sem vigências cadastradas: pagamento direto por padrão.</p> : <div className="resource-history"><table><thead><tr><th>Vigência</th><th>Método</th><th>Contexto</th><th>Registrado em</th></tr></thead><tbody>{history.periods.map((period) => <tr key={period.id}><td>{formatMonth(period.start_month)} a {period.end_month ? formatMonth(period.end_month) : "sem término"}</td><td>{period.method === "direct" ? "Direto" : period.credit_card_id ? <Link href={`/cards/${encodeURIComponent(period.credit_card_id)}`}>{cards.find((card) => card.id === period.credit_card_id)?.name ?? "Cartão histórico"}</Link> : "Cartão"}</td><td>{period.context ?? "—"}</td><td>{new Date(period.recorded_at).toLocaleString("pt-BR", { timeZone: "America/Sao_Paulo" })}</td></tr>)}</tbody></table></div>}</section>;
}
