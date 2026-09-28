"use client";

import { useState, type FormEvent } from "react";
import { Alert, Button, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { CreditCard as CreditCardType, Debt, DebtOccurrence, PaymentMethodHistory } from "@/lib/api/types";
import { formatMonth } from "@/lib/debt-month";

function cents(value: FormDataEntryValue | null) { return Math.round(Number(String(value ?? "").replace(",", ".")) * 100); }

export function DebtActions({ debt, cards, paymentHistory, schedule }: { debt: Debt; cards: CreditCardType[]; paymentHistory: PaymentMethodHistory | null; schedule: DebtOccurrence[] }) {
  const router = useRouter();
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [method, setMethod] = useState("direct");
  const activeCards = cards.filter((card) => card.status === "active");
  const regularMonths = schedule.filter((occurrence) => occurrence.debt_occurrence_kind === "scheduled").map((occurrence) => occurrence.reference_month);
  const canChange = debt.status === "active";
  const canSettle = canChange && !debt.settlement && debt.projection_status !== "completed";

  async function submit(event: FormEvent<HTMLFormElement>, action: string, path: string, payload: Record<string, unknown>) {
    event.preventDefault(); setBusy(action); setError(""); setSuccess("");
    try { await browserApiFetch(path, { method: "POST", body: JSON.stringify(payload) }); setSuccess(action === "amount" ? "Valor projetado atualizado." : action === "payment" ? "Forma de pagamento atualizada." : "Quitação antecipada projetada registrada."); router.refresh(); }
    catch (caught) { setError(caught instanceof Error ? caught.message : "Não foi possível salvar a alteração."); }
    finally { setBusy(""); }
  }

  return <div className="debt-action-stack">
    {error && <Alert color="error" title="Não foi possível salvar">{error}</Alert>}
    {success && <Alert color="success">{success}</Alert>}
    <div className="debt-action-panel"><h3>Reajustar parcela</h3><p>Define um novo valor a partir de uma competência. As parcelas anteriores permanecem no histórico.</p>
      <form className="resource-form" onSubmit={(event) => { const data = new FormData(event.currentTarget); void submit(event, "amount", `/debts/${debt.id}/changes`, { effective_from: String(data.get("effective_from")), installment_amount_cents: cents(data.get("amount")), context: String(data.get("context") || "").trim() || null }); }}>
        <Input name="effective_from" type="month" label="A partir de" min={debt.scheduled_start_month} max={debt.effective_end_month} required disabled={!canChange} />
        <Input name="amount" type="number" min="0.01" step="0.01" label="Nova parcela (R$)" required disabled={!canChange} />
        <Input name="context" label="Motivo" maxLength={500} placeholder="Opcional" disabled={!canChange} />
        <Button type="submit" disabled={!canChange || Boolean(busy)}>{busy === "amount" && <Spinner size="sm" />} Salvar novo valor</Button>
      </form>
    </div>
    <div className="debt-action-panel"><h3>Forma de pagamento</h3><p>Vale a partir do mês escolhido. Use um cartão cadastrado para projetar a parcela na fatura.</p>
      {paymentHistory && paymentHistory.periods.length > 0 && <div className="debt-payment-history">{paymentHistory.periods.map((period) => <span key={period.id}>{formatMonth(period.start_month)}{period.end_month ? ` a ${formatMonth(period.end_month)}` : " em diante"} · {period.method === "credit_card" ? cards.find((card) => card.id === period.credit_card_id)?.name ?? "Cartão" : "Direto"}</span>)}</div>}
      <form className="resource-form" onSubmit={(event) => { const data = new FormData(event.currentTarget); void submit(event, "payment", `/financial-items/${debt.id}/payment-changes`, { effective_from: String(data.get("effective_from")), end_month: data.get("end_month") || null, method, credit_card_id: method === "credit_card" ? String(data.get("credit_card_id")) : null, context: String(data.get("context") || "").trim() || null }); }}>
        <div className="form-grid-two"><Input name="effective_from" type="month" label="A partir de" min={debt.scheduled_start_month} max={debt.effective_end_month} required disabled={!canChange} /><Input name="end_month" type="month" label="Até (opcional)" min={debt.scheduled_start_month} max={debt.effective_end_month} disabled={!canChange} /></div>
        <label className="native-field"><span>Método</span><select value={method} onChange={(event) => setMethod(event.target.value)} disabled={!canChange}><option value="direct">Pagamento direto</option><option value="credit_card" disabled={activeCards.length === 0}>Cartão de crédito</option></select></label>
        {method === "credit_card" && <label className="native-field"><span>Cartão</span><select name="credit_card_id" defaultValue="" required disabled={!canChange}><option value="" disabled>Selecione</option>{activeCards.map((card) => <option key={card.id} value={card.id}>{card.institution.name} · {card.name}</option>)}</select></label>}
        <Input name="context" label="Contexto" maxLength={500} placeholder="Opcional" disabled={!canChange} />
        <Button type="submit" disabled={!canChange || Boolean(busy)}>{busy === "payment" && <Spinner size="sm" />} Alterar pagamento</Button>
      </form>
    </div>
    <div className="debt-action-panel debt-settlement-panel"><h3>Quitação antecipada</h3>{debt.settlement ? <p>Projetada em {formatMonth(debt.settlement.reference_month)} por {new Intl.NumberFormat("pt-BR", { style: "currency", currency: debt.currency_code }).format(debt.settlement.amount_cents / 100)}. Este registro não confirma o pagamento.</p> : <><p>Substitui a parcela do mês escolhido e remove as projeções posteriores. O registro é único e não pode ser alterado.</p><form className="resource-form" onSubmit={(event) => { const data = new FormData(event.currentTarget); void submit(event, "settlement", `/debts/${debt.id}/early-settlement`, { reference_month: String(data.get("reference_month")), amount_cents: cents(data.get("amount")), reason: String(data.get("reason") || "").trim() || null }); }}>
        <label className="native-field"><span>Mês da quitação</span><select name="reference_month" defaultValue="" required disabled={!canSettle || regularMonths.length < 2}><option value="" disabled>Selecione um mês</option>{regularMonths.slice(0, -1).map((month) => <option key={month} value={month}>{formatMonth(month)}</option>)}</select></label>
        <Input name="amount" type="number" min="0.01" step="0.01" label="Valor projetado da quitação (R$)" required disabled={!canSettle} />
        <Input name="reason" label="Motivo" maxLength={500} placeholder="Opcional" disabled={!canSettle} />
        <Button type="submit" disabled={!canSettle || regularMonths.length < 2 || Boolean(busy)}>{busy === "settlement" && <Spinner size="sm" />} Registrar projeção de quitação</Button>
      </form></>}
    </div>
  </div>;
}
