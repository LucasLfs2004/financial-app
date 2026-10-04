"use client";

import { useState, type FormEvent } from "react";
import { Alert, Button, Card, CardContent, CardTitle, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { ArrowLeft, CalendarClock, CreditCard as CreditCardIcon, Landmark, Receipt, Save } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { CreditCard, Debt } from "@/lib/api/types";
import { currentMonth } from "@/lib/debt-month";

type DebtFormProps = { debt?: Debt | null; cards?: CreditCard[]; onCancel?: () => void };
type EntryType = "installments" | "one_time" | "subscription";

const entryTypes = [
  { value: "installments", label: "Parcelada", description: "Compras e dívidas com parcelas", icon: CreditCardIcon },
  { value: "one_time", label: "À vista", description: "Um gasto em um único mês", icon: Receipt },
  { value: "subscription", label: "Assinatura", description: "Cobrança mensal recorrente", icon: CalendarClock },
] as const;

function toCents(value: FormDataEntryValue | null) {
  const amount = Number(String(value ?? "0").replace(",", "."));
  return Number.isFinite(amount) ? Math.round(amount * 100) : 0;
}

export function DebtForm({ debt, cards = [], onCancel }: DebtFormProps) {
  const router = useRouter();
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState("direct");
  const [cashMonthOffset, setCashMonthOffset] = useState(1);
  const [entryType, setEntryType] = useState<EntryType>("installments");
  const editing = Boolean(debt);
  const activeCards = cards.filter((card) => card.status === "active");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSuccess("");
    setLoading(true);
    const form = event.currentTarget;
    const values = new FormData(form);
    const name = String(values.get("name") ?? "").trim();
    const description = String(values.get("description") ?? "").trim() || null;

    try {
      if (editing) {
        await browserApiFetch(`/debts/${debt?.id}`, { method: "PATCH", body: JSON.stringify({ name, description }) });
      } else if (entryType === "installments") {
        const totalInstallments = Number(values.get("total_installments"));
        const firstProjectedInstallment = Number(values.get("first_projected_installment"));
        if (firstProjectedInstallment > totalInstallments) throw new Error("A parcela atual precisa estar dentro do total de parcelas.");
        await browserApiFetch("/debts", { method: "POST", body: JSON.stringify({
          name,
          description,
          original_total_cents: values.get("original_total") ? toCents(values.get("original_total")) : null,
          total_installments: totalInstallments,
          first_projected_installment: firstProjectedInstallment,
          scheduled_start_month: String(values.get("scheduled_start_month")),
          installment_amount_cents: toCents(values.get("installment_amount")),
          cash_month_offset: Number(values.get("cash_month_offset") ?? 1),
          context: String(values.get("context") ?? "").trim() || null,
          payment_method: String(values.get("payment_method")),
          credit_card_id: values.get("payment_method") === "credit_card" ? String(values.get("credit_card_id")) : null,
        }) });
      } else {
        const startMonth = String(values.get("start_month"));
        const endMonth = String(values.get("end_month") || "");
        if (entryType === "subscription" && endMonth && endMonth < startMonth) throw new Error("A última cobrança não pode vir antes da primeira.");
        await browserApiFetch("/financial-items", { method: "POST", body: JSON.stringify({
          name,
          description,
          kind: entryType === "subscription" ? "fixed_expense" : "projected_variable_expense",
          period: {
            start_month: startMonth,
            end_month: entryType === "subscription" ? endMonth || null : startMonth,
            amount_cents: toCents(values.get("amount")),
            recurrence: entryType === "subscription" ? "monthly" : "once",
            cash_month_offset: 0,
            context: entryType === "subscription" ? "Assinatura" : null,
          },
        }) });
      }
      setSuccess(editing ? "Dívida atualizada." : entryType === "installments" ? "Dívida cadastrada. O cronograma já está disponível na lista." : entryType === "subscription" ? "Assinatura cadastrada. Ela já entra na projeção mensal." : "Despesa à vista cadastrada no mês escolhido.");
      if (!editing) form.reset();
      onCancel?.();
      router.refresh();
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : "Não foi possível salvar o registro.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="form-card" padding="none">
      <CardContent className="form-card-content">
        <div className="section-heading"><div><span className="section-kicker">{editing ? "EDITAR DÍVIDA" : "NOVO COMPROMISSO"}</span><CardTitle>{editing ? "Atualizar dívida" : "O que você quer cadastrar?"}</CardTitle></div><Landmark size={20} /></div>
        <p className="section-description">{editing ? "Atualize o nome ou a descrição desta dívida." : "Escolha o tipo para ver apenas os campos necessários."}</p>
        {error && <Alert color="error" title="Não foi possível salvar">{error}</Alert>}
        {success && <Alert color="success">{success}{!editing && entryType !== "installments" && <> <Link href="/entries">Ver em receitas e despesas</Link></>}</Alert>}
        <form className="resource-form" onSubmit={submit}>
          {!editing && <div className="entry-type-picker" role="group" aria-label="Tipo de compromisso">
            {entryTypes.map(({ value, label, description: hint, icon: Icon }) => <button key={value} type="button" className={entryType === value ? "entry-type-option selected" : "entry-type-option"} aria-pressed={entryType === value} onClick={() => { setEntryType(value); setError(""); setSuccess(""); }}><Icon size={19} /><span><strong>{label}</strong><small>{hint}</small></span></button>)}
          </div>}
          <Input name="name" label={entryType === "subscription" && !editing ? "Nome da assinatura" : entryType === "one_time" && !editing ? "Nome da despesa" : "Nome da dívida"} defaultValue={debt?.name ?? ""} placeholder={entryType === "subscription" ? "Ex.: Streaming" : entryType === "one_time" ? "Ex.: Conserto do carro" : "Ex.: Financiamento do carro"} required maxLength={120} />
          <Input name="description" label="Descrição" defaultValue={debt?.description ?? ""} placeholder="Opcional" maxLength={1000} />
          {!editing && entryType === "installments" && <>
            <Input name="installment_amount" type="number" min="0.01" step="0.01" label="Valor de cada parcela (R$)" placeholder="0,00" required />
            <div className="form-grid-two"><Input name="total_installments" type="number" min="1" step="1" label="Total de parcelas" placeholder="12" required /><Input name="first_projected_installment" type="number" min="1" step="1" label="Próxima parcela" placeholder="1" defaultValue="1" required /></div>
            <Input name="scheduled_start_month" type="month" label="Mês da próxima parcela" defaultValue={currentMonth()} required />
            <label className="native-field"><span>Forma de pagamento</span><select name="payment_method" value={paymentMethod} onChange={(event) => setPaymentMethod(event.target.value)}><option value="direct">Pagamento direto</option><option value="credit_card" disabled={activeCards.length === 0}>Cartão de crédito</option></select></label>
            {paymentMethod === "credit_card" && <label className="native-field"><span>Cartão de crédito</span><select name="credit_card_id" defaultValue="" required><option value="" disabled>Selecione um cartão</option>{activeCards.map((card) => <option value={card.id} key={card.id}>{card.name} · {card.institution.name}</option>)}</select></label>}
            {activeCards.length === 0 && <p className="resource-hint">Quer pagar no cartão? <Link className="small-link" href="/cards#new-card">Cadastre um cartão primeiro</Link></p>}
            {paymentMethod === "direct" && cashMonthOffset === 1 && <p className="resource-hint">Exemplo: a última parcela de setembro entra no caixa de outubro, quando cai o salário no dia 1º. A partir do salário de novembro, esse valor fica livre.</p>}
            <details className="form-optional-details"><summary>Mais detalhes (opcional)</summary><div className="resource-form"><Input name="original_total" type="number" min="0" step="0.01" label="Valor original (R$)" placeholder="0,00" /><Input name="cash_month_offset" type="number" min="0" max="12" step="1" label="Deslocamento para o caixa (meses)" defaultValue="1" required onChange={(event) => setCashMonthOffset(Number(event.target.value))} /><Input name="context" label="Contexto da parcela" maxLength={500} /></div></details>
          </>}
          {!editing && entryType !== "installments" && <>
            <Input name="amount" type="number" min="0.01" step="0.01" label={entryType === "subscription" ? "Valor por mês (R$)" : "Valor pago (R$)"} placeholder="0,00" required />
            <div className="form-grid-two"><Input name="start_month" type="month" label={entryType === "subscription" ? "Primeira cobrança" : "Mês do gasto"} defaultValue={currentMonth()} required />{entryType === "subscription" && <Input name="end_month" type="month" label="Última cobrança · opcional" />}</div>
            {entryType === "subscription" && <p className="resource-hint">A assinatura entra como despesa fixa mensal na projeção.</p>}
          </>}
          <div className="form-actions"><Button type="submit" size="lg" disabled={loading}>{loading ? <><Spinner size="sm" /> Salvando...</> : <><Save size={17} /> {editing ? "Salvar alterações" : entryType === "installments" ? "Adicionar dívida" : entryType === "subscription" ? "Adicionar assinatura" : "Adicionar despesa"}</>}</Button>{onCancel && <Button type="button" variant="ghost" size="lg" onClick={onCancel}><ArrowLeft size={17} /> Cancelar</Button>}</div>
        </form>
      </CardContent>
    </Card>
  );
}
