"use client";

import { useState, type FormEvent } from "react";
import { Alert, Button, Card, CardContent, CardTitle, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { CircleDollarSign, Save } from "lucide-react";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";

function toCents(value: FormDataEntryValue | null) {
  const parsed = Number(String(value ?? "0").replace(",", "."));
  return Number.isFinite(parsed) ? Math.round(parsed * 100) : 0;
}

export function FinancialItemForm() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    setLoading(true);
    setError("");
    setSuccess("");
    const values = new FormData(form);
    const start = String(values.get("start_month"));
    const recurrence = String(values.get("recurrence"));
    const end = String(values.get("end_month") || "");
    const payload = {
      name: String(values.get("name")).trim(),
      kind: String(values.get("kind")),
      description: String(values.get("description") || "").trim() || null,
      period: {
        start_month: start,
        end_month: recurrence === "once" ? start : end || null,
        amount_cents: toCents(values.get("amount")),
        recurrence,
        cash_month_offset: Number(values.get("cash_month_offset") || 0),
        context: String(values.get("context") || "").trim() || null,
      },
    };

    try {
      await browserApiFetch("/financial-items", { method: "POST", body: JSON.stringify(payload) });
      form.reset();
      setSuccess(`${payload.name} cadastrado com sucesso.`);
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
        <form className="resource-form" onSubmit={submit}>
          <Input name="name" label="Nome" placeholder="Ex.: Salário" required maxLength={120} />
          <label className="native-field"><span>Tipo</span><select name="kind" defaultValue="recurring_income"><option value="recurring_income">Renda recorrente</option><option value="one_time_income">Renda pontual</option><option value="fixed_expense">Despesa fixa</option><option value="projected_variable_expense">Despesa variável projetada</option></select></label>
          <Input name="description" label="Descrição" placeholder="Opcional" maxLength={1000} />
          <div className="form-grid-two"><Input name="amount" type="number" min="0" step="0.01" label="Valor (R$)" placeholder="0,00" required /><Input name="cash_month_offset" type="number" min="0" max="12" label="Deslocamento de caixa" defaultValue="0" required /></div>
          <div className="form-grid-two"><Input name="start_month" type="month" label="Começa em" defaultValue={new Date().toISOString().slice(0, 7)} required /><Input name="end_month" type="month" label="Termina em" /></div>
          <label className="native-field"><span>Recorrência</span><select name="recurrence" defaultValue="monthly"><option value="monthly">Mensal</option><option value="once">Pontual</option></select></label>
          <Input name="context" label="Contexto" placeholder="Opcional" maxLength={500} />
          <Button type="submit" size="lg" disabled={loading}>{loading ? <><Spinner size="sm" /> Salvando...</> : <><Save size={17} /> Adicionar item</>}</Button>
        </form>
      </CardContent>
    </Card>
  );
}
