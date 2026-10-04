"use client";

import { useState, type FormEvent } from "react";
import { Alert, Button, Card, CardContent, CardTitle, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { CircleDollarSign, Save } from "lucide-react";
import { useRouter } from "next/navigation";
import type { CreditCard } from "@/lib/api/types";
import { SubscriptionFields } from "./subscription-fields";
import { useFinancialItemCreation } from "./use-financial-item-creation";

function toCents(value: FormDataEntryValue | null) {
  const parsed = Number(String(value ?? "0").replace(",", "."));
  return Number.isFinite(parsed) ? Math.round(parsed * 100) : 0;
}

export function FinancialItemForm({ cards = [] }: { cards?: CreditCard[] }) {
  const router = useRouter();
  const { saveItem, pendingItem } = useFinancialItemCreation();
  const activeCards = cards.filter((card) => card.status === "active");
  const [paymentMethod, setPaymentMethod] = useState("direct");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [selection, setSelection] = useState("recurring_income");
  const [recurrence, setRecurrence] = useState("monthly");
  const subscription = selection === "subscription";
  const once = recurrence === "once";
  const fixedExpense = subscription || selection === "fixed_expense";

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    setLoading(true);
    setError("");
    setSuccess("");
    const values = new FormData(form);
    const start = String(values.get("start_month"));
    const end = String(values.get("end_month") || "");
    const payload = {
      name: String(values.get("name")).trim(),
      kind: subscription ? "fixed_expense" : selection,
      description: String(values.get("description") || "").trim() || null,
      ...(fixedExpense ? { invoice_match_title: String(values.get("invoice_match_title") || "").trim() || null, renewal_day: values.get("renewal_day") ? Number(values.get("renewal_day")) : null } : {}),
      period: {
        start_month: start,
        end_month: once ? start : end || null,
        amount_cents: toCents(values.get("amount")),
        recurrence,
        cash_month_offset: Number(values.get("cash_month_offset") || 0),
        context: subscription ? "Assinatura" : String(values.get("context") || "").trim() || null,
      },
    };

    try {
      if (!once && end && end < start) throw new Error("O mês de término não pode vir antes do início.");
      const savedName = await saveItem(payload, fixedExpense && paymentMethod === "credit_card" ? {
        effective_from: start, end_month: end || null, method: "credit_card", credit_card_id: String(values.get("credit_card_id") || ""),
      } : null);
      form.reset();
      setSelection("recurring_income");
      setRecurrence("monthly");
      setPaymentMethod("direct");
      setSuccess(`${savedName} cadastrado com sucesso.`);
      router.refresh();
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : "Não foi possível cadastrar o item.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="form-card" padding="none">
      <CardContent className="form-card-content">
        <div className="section-heading"><div><span className="section-kicker">NOVO REGISTRO</span><CardTitle>Adicionar receita ou despesa</CardTitle></div><CircleDollarSign size={20} /></div>
        <p className="section-description">Informe quando começa e o valor. A data de término é opcional.</p>
        {error && <Alert color="error" title="Não foi possível salvar">{error}</Alert>}
        {success && <Alert color="success" title="Registro salvo">{success}</Alert>}
        {pendingItem && <Alert color="warning" title="Registro criado">{pendingItem.name} foi salvo. Falta concluir o vínculo com o cartão. Tente salvar novamente; o registro não será duplicado.</Alert>}
        <form className="resource-form" onSubmit={submit}>
          <fieldset className="resource-form subscription-form-fields" disabled={loading || Boolean(pendingItem)}>
          <Input name="name" label="Nome" placeholder="Ex.: Salário" required maxLength={120} />
          <label className="native-field"><span>Tipo</span><select name="kind" value={selection} onChange={(event) => { const value = event.target.value; setSelection(value); setPaymentMethod(value === "subscription" && activeCards.length > 0 ? "credit_card" : "direct"); setRecurrence(value === "one_time_income" ? "once" : "monthly"); setError(""); setSuccess(""); }}><option value="recurring_income">Renda recorrente</option><option value="one_time_income">Renda pontual</option><option value="fixed_expense">Despesa fixa</option><option value="projected_variable_expense">Despesa variável projetada</option><option value="subscription">Assinatura mensal</option></select></label>
          <Input name="description" label="Descrição" placeholder="Opcional" maxLength={1000} />
          <div className="form-grid-two"><Input name="amount" type="number" min="0.01" step="0.01" label={subscription ? "Valor por mês (R$)" : "Valor (R$)"} placeholder="0,00" required />{!subscription && <Input name="cash_month_offset" type="number" min="0" max="12" label="Deslocamento de caixa" defaultValue="0" required />}</div>
          <div className="form-grid-two"><Input name="start_month" type="month" label={subscription ? "Primeira cobrança" : "Começa em"} defaultValue={new Date().toISOString().slice(0, 7)} required />{!once && <Input name="end_month" type="month" label={subscription ? "Última cobrança · opcional" : "Termina em"} />}</div>
          {!subscription && selection !== "one_time_income" && <label className="native-field"><span>Recorrência</span><select name="recurrence" value={recurrence} onChange={(event) => setRecurrence(event.target.value)}><option value="monthly">Mensal</option><option value="once">Pontual</option></select></label>}
          {fixedExpense && <>
            <label className="native-field"><span>Forma de pagamento</span><select name="payment_method" value={paymentMethod} onChange={(event) => setPaymentMethod(event.target.value)}><option value="direct">Pagamento direto</option><option value="credit_card" disabled={activeCards.length === 0}>Cartão de crédito</option></select></label>
            {paymentMethod === "credit_card" && <label className="native-field"><span>Cartão de crédito</span><select name="credit_card_id" defaultValue="" required><option value="" disabled>Selecione um cartão</option>{activeCards.map((card) => <option value={card.id} key={card.id}>{card.institution.name} · {card.name}</option>)}</select></label>}
            <SubscriptionFields />
          </>}
          {subscription ? <p className="resource-hint">A assinatura entra como despesa fixa mensal na projeção.</p> : <Input name="context" label="Contexto" placeholder="Opcional" maxLength={500} />}
          </fieldset>
          <Button type="submit" size="lg" disabled={loading}>{loading ? <><Spinner size="sm" /> Salvando...</> : <><Save size={17} /> {subscription ? "Adicionar assinatura" : "Adicionar item"}</>}</Button>
        </form>
      </CardContent>
    </Card>
  );
}
