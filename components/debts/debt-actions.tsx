"use client";

import { useRef, useState, type FormEvent } from "react";
import { Alert, Button, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { CreditCard as CreditCardType, Debt, PaymentMethodHistory } from "@/lib/api/types";
import { settlementMonthBounds, formatMonth } from "@/lib/debt-month";

function cents(value: FormDataEntryValue | null) { return Math.round(Number(String(value ?? "").replace(",", ".")) * 100); }

export function DebtActions({ debt, cards, paymentHistory }: { debt: Debt; cards: CreditCardType[]; paymentHistory: PaymentMethodHistory | null }) {
  const router = useRouter();
  const lock = useRef(false);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [method, setMethod] = useState("direct");
  const activeCards = cards.filter((card) => card.status === "active");
  const settlementBounds = settlementMonthBounds(debt);
  const lastSettlementMonth = settlementBounds.to;
  const canChange = debt.status === "active";
  const canSettle = canChange && !debt.settlement && debt.projection_status !== "completed" && settlementBounds.available;

  async function submit(event: FormEvent<HTMLFormElement>, action: string, path: string, payload: Record<string, unknown>) {
    event.preventDefault(); if (lock.current) return; lock.current = true; setBusy(action); setError(""); setSuccess("");
    try { await browserApiFetch(path, { method: "POST", body: JSON.stringify(payload) }); setSuccess(action === "amount" ? "Valor projetado atualizado." : action === "payment" ? "Forma de pagamento atualizada." : "Quitação antecipada projetada registrada."); router.refresh(); }
    catch (caught) { setError(caught instanceof Error ? caught.message : "Não foi possível salvar a alteração."); }
    finally { lock.current = false; setBusy(""); }
  }

  return <div className="debt-action-stack">
    {error && <Alert color="error" title="Não foi possível salvar">{error}</Alert>}
    {success && <Alert color="success">{success}</Alert>}
    <div className="debt-action-panel"><h3>Reajustar parcela</h3><p>Define um novo valor a partir de uma competência. As parcelas anteriores permanecem no histórico.</p>
      <form className="resource-form" onSubmit={(event) => { const data = new FormData(event.currentTarget); void submit(event, "amount", `/debts/${encodeURIComponent(debt.id)}/changes`, { effective_from: String(data.get("effective_from")), installment_amount_cents: cents(data.get("amount")), context: String(data.get("context") || "").trim() || null }); }}>
        <Input name="effective_from" type="month" label="A partir de" min={debt.scheduled_start_month} max={debt.effective_end_month} required disabled={!canChange} />
        <Input name="amount" type="number" min="0.01" step="0.01" label={`Nova parcela (${debt.currency_code})`} required disabled={!canChange} />
        <Input name="context" label="Motivo" maxLength={500} placeholder="Opcional" disabled={!canChange} />
        <Button type="submit" disabled={!canChange || Boolean(busy)}>{busy === "amount" && <Spinner size="sm" />} Salvar novo valor</Button>
      </form>
    </div>
    <div className="debt-action-panel"><h3>Forma de pagamento</h3><p>Vale a partir do mês escolhido. Use um cartão cadastrado para projetar a parcela na fatura.</p>
      {!paymentHistory && <p role="alert">Histórico indisponível. Atualize a página para consultar antes de alterar o pagamento.</p>}
      {paymentHistory?.periods.length === 0 && <p>Sem vigências cadastradas: pagamento direto por padrão.</p>}
      {paymentHistory && paymentHistory.periods.length > 0 && <div className="debt-payment-history">{paymentHistory.periods.map((period) => <span key={period.id}>{formatMonth(period.start_month)}{period.end_month ? ` a ${formatMonth(period.end_month)}` : " em diante"} · {period.method === "credit_card" ? cards.find((card) => card.id === period.credit_card_id)?.name ?? "Cartão" : "Direto"}{period.context ? ` · ${period.context}` : ""} · Registrado em {new Date(period.recorded_at).toLocaleString("pt-BR")}</span>)}</div>}
      <form className="resource-form" onSubmit={(event) => { const data = new FormData(event.currentTarget); void submit(event, "payment", `/financial-items/${encodeURIComponent(debt.id)}/payment-changes`, { effective_from: String(data.get("effective_from")), end_month: data.get("end_month") || null, method, credit_card_id: method === "credit_card" ? String(data.get("credit_card_id")) : null, context: String(data.get("context") || "").trim() || null }); }}>
        <div className="form-grid-two"><Input name="effective_from" type="month" label="A partir de" min={debt.scheduled_start_month} max={debt.effective_end_month} required disabled={!canChange} /><Input name="end_month" type="month" label="Até (opcional)" min={debt.scheduled_start_month} max={debt.effective_end_month} disabled={!canChange} /></div>
        <label className="native-field"><span>Método</span><select value={method} onChange={(event) => setMethod(event.target.value)} disabled={!canChange}><option value="direct">Pagamento direto</option><option value="credit_card" disabled={activeCards.length === 0}>Cartão de crédito</option></select></label>
        {method === "credit_card" && <label className="native-field"><span>Cartão</span><select name="credit_card_id" defaultValue="" required disabled={!canChange}><option value="" disabled>Selecione</option>{activeCards.map((card) => <option key={card.id} value={card.id}>{card.institution.name} · {card.name}</option>)}</select></label>}
        <Input name="context" label="Contexto" maxLength={500} placeholder="Opcional" disabled={!canChange} />
        <Button type="submit" disabled={!canChange || Boolean(busy)}>{busy === "payment" && <Spinner size="sm" />} Alterar pagamento</Button>
      </form>
    </div>
    <div className="debt-action-panel debt-settlement-panel"><h3>Quitação antecipada</h3>{debt.settlement ? <p>Projetada em {formatMonth(debt.settlement.reference_month)} por {new Intl.NumberFormat("pt-BR", { style: "currency", currency: debt.currency_code }).format(debt.settlement.amount_cents / 100)}. Este registro não confirma o pagamento. Motivo: {debt.settlement.reason || "Não informado"}. Registrado em {new Date(debt.settlement.recorded_at).toLocaleString("pt-BR")}.</p> : <><p>Substitui a parcela do mês escolhido e remove as projeções posteriores. O registro é único e não pode ser alterado.</p><form className="resource-form" onSubmit={(event) => { const data = new FormData(event.currentTarget); void submit(event, "settlement", `/debts/${encodeURIComponent(debt.id)}/early-settlement`, { reference_month: String(data.get("reference_month")), amount_cents: cents(data.get("amount")), reason: String(data.get("reason") || "").trim() || null }); }}>
        <Input name="reference_month" type="month" label="Mês da quitação" min={debt.scheduled_start_month} max={lastSettlementMonth} required disabled={!canSettle} />
        <Input name="amount" type="number" min="0.01" step="0.01" label={`Valor projetado da quitação (${debt.currency_code})`} required disabled={!canSettle} />
        <Input name="reason" label="Motivo" maxLength={500} placeholder="Opcional" disabled={!canSettle} />
        <Button type="submit" disabled={!canSettle || Boolean(busy)}>{busy === "settlement" && <Spinner size="sm" />} Registrar projeção de quitação</Button>
      </form></>}
    </div>
  </div>;
}
