"use client";

import { useState } from "react";
import { Input } from "@lucaslfs2004/luke-ui";
import type { CreditCard } from "@/lib/api/types";
import { paymentChange } from "@/lib/api/requests";
import { MutationForm } from "@/components/resource-mutation";

export function PaymentChangeForm({ itemId, cards, month, disabled }: { itemId: string; cards: CreditCard[]; month: string; disabled: boolean }) {
  const [method, setMethod] = useState("direct");
  return <section className="debt-action-panel"><h2>Alterar pagamento por vigência</h2><MutationForm path={`/financial-items/${encodeURIComponent(itemId)}/payment-changes`} payload={paymentChange} disabled={disabled}><div className="form-grid-two"><Input name="effective_from" type="month" label="A partir de" defaultValue={month} required /><Input name="end_month" type="month" label="Até (opcional)" /></div><label className="native-field"><span>Método</span><select name="method" value={method} onChange={(event) => setMethod(event.target.value)}><option value="direct">Direto</option><option value="credit_card">Cartão de crédito</option></select></label>{method === "credit_card" && <label className="native-field"><span>Cartão</span><select name="credit_card_id" required defaultValue=""><option value="" disabled>Selecione</option>{cards.filter((card) => card.status === "active").map((card) => <option key={card.id} value={card.id}>{card.institution.name} · {card.name}</option>)}</select></label>}<Input name="context" label="Contexto" maxLength={500} /></MutationForm></section>;
}
