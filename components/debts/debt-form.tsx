"use client";

import { useState, type FormEvent } from "react";
import { Alert, Button, Card, CardContent, CardTitle, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { ArrowLeft, Landmark, Save } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { CreditCard, Debt } from "@/lib/api/types";

type DebtFormProps = { debt?: Debt | null; cards?: CreditCard[]; onCancel?: () => void };

function toCents(value: FormDataEntryValue | null) {
  const amount = Number(String(value ?? "0").replace(",", "."));
  return Number.isFinite(amount) ? Math.round(amount * 100) : 0;
}

export function DebtForm({ debt, cards = [], onCancel }: DebtFormProps) {
  const router = useRouter();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState("direct");
  const editing = Boolean(debt);
  const activeCards = cards.filter((card) => card.status === "active");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setLoading(true);
    const values = new FormData(event.currentTarget);
    const payload = editing
      ? { name: String(values.get("name")).trim(), description: String(values.get("description") ?? "").trim() || null }
      : {
          name: String(values.get("name")).trim(),
          description: String(values.get("description") ?? "").trim() || null,
          original_total_cents: toCents(values.get("original_total")),
          total_installments: Number(values.get("total_installments")),
          first_projected_installment: Number(values.get("first_projected_installment")),
          scheduled_start_month: String(values.get("scheduled_start_month")),
          installment_amount_cents: toCents(values.get("installment_amount")),
          cash_month_offset: Number(values.get("cash_month_offset") || 0),
          payment_method: String(values.get("payment_method")),
          credit_card_id: values.get("payment_method") === "credit_card" ? String(values.get("credit_card_id")) : null,
        };

    try {
      await browserApiFetch(editing ? `/debts/${debt?.id}` : "/debts", { method: editing ? "PATCH" : "POST", body: JSON.stringify(payload) });
      onCancel?.();
      router.refresh();
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : "Não foi possível salvar a dívida.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="form-card" padding="none">
      <CardContent className="form-card-content">
        <div className="section-heading"><div><span className="section-kicker">{editing ? "EDITAR DÍVIDA" : "NOVA DÍVIDA"}</span><CardTitle>{editing ? "Atualizar dívida" : "Adicione uma dívida"}</CardTitle></div><Landmark size={20} /></div>
        <p className="section-description">O cronograma será calculado automaticamente a partir da primeira parcela projetada.</p>
        {error && <Alert color="error" title="Não foi possível salvar">{error}</Alert>}
        <form className="resource-form" onSubmit={submit}>
          <Input name="name" label="Nome da dívida" defaultValue={debt?.name ?? ""} placeholder="Ex.: Financiamento do carro" required maxLength={120} />
          <Input name="description" label="Descrição" defaultValue={debt?.description ?? ""} placeholder="Opcional" maxLength={1000} />
          {!editing && <>
            <div className="form-grid-two"><Input name="original_total" type="number" min="0" step="0.01" label="Valor original (R$)" placeholder="0,00" required /><Input name="installment_amount" type="number" min="0.01" step="0.01" label="Valor da parcela (R$)" placeholder="0,00" required /></div>
            <div className="form-grid-two"><Input name="total_installments" type="number" min="1" label="Total de parcelas" placeholder="12" required /><Input name="first_projected_installment" type="number" min="1" label="Parcela atual" placeholder="1" required /></div>
            <div className="form-grid-two"><Input name="scheduled_start_month" type="month" label="Início projetado" defaultValue={new Date().toISOString().slice(0, 7)} required /><Input name="cash_month_offset" type="number" min="0" max="12" label="Atraso de caixa (meses)" defaultValue="0" required /></div>
            <label className="native-field"><span>Forma de pagamento</span><select name="payment_method" value={paymentMethod} onChange={(event) => setPaymentMethod(event.target.value)}><option value="direct">Pagamento direto</option><option value="credit_card" disabled={activeCards.length === 0}>Cartão de crédito</option></select></label>
            {paymentMethod === "credit_card" && <label className="native-field"><span>Cartão de crédito</span><select name="credit_card_id" defaultValue="" required><option value="" disabled>Selecione um cartão</option>{activeCards.map((card) => <option value={card.id} key={card.id}>{card.name} · {card.institution.name}</option>)}</select></label>}
            {activeCards.length === 0 && <p className="resource-hint">Quer pagar no cartão? <Link className="small-link" href="/cards#new-card">Cadastre um cartão primeiro</Link></p>}
          </>}
          <div className="form-actions"><Button type="submit" size="lg" disabled={loading}>{loading ? <><Spinner size="sm" /> Salvando...</> : <><Save size={17} /> {editing ? "Salvar alterações" : "Adicionar dívida"}</>}</Button>{onCancel && <Button type="button" variant="ghost" size="lg" onClick={onCancel}><ArrowLeft size={17} /> Cancelar</Button>}</div>
        </form>
      </CardContent>
    </Card>
  );
}
