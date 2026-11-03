"use client";

import { useState, type FormEvent } from "react";
import { Alert, Button, Card, CardContent, CardTitle, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { Check, Pencil, Save } from "lucide-react";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import { addMonths, currentMonth } from "@/lib/debt-month";
import type { Plan } from "@/lib/api/types";

type PlanFormProps = { plan?: Plan | null; defaultMonth?: string; currencyCode?: string };

export function PlanForm({ plan, defaultMonth = currentMonth(), currencyCode = "BRL" }: PlanFormProps) {
  const router = useRouter();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [activating, setActivating] = useState(false);
  const isEditing = Boolean(plan);
  const readonly = Boolean(plan && plan.status !== "draft");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (readonly || loading) return;
    setError("");
    setLoading(true);
    const values = Object.fromEntries(new FormData(event.currentTarget).entries());
    const payload = {
      name: String(values.name).trim(),
      start_month: String(values.start_month),
      end_month: String(values.end_month),
      currency_code: String(values.currency_code).toUpperCase(),
    };

    try {
      await browserApiFetch(isEditing ? "/plans/current" : "/plans", {
        method: isEditing ? "PATCH" : "POST",
        body: JSON.stringify(payload),
      });
      router.refresh();
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : "Não foi possível salvar o planejamento.");
    } finally {
      setLoading(false);
    }
  }

  async function activate() {
    setError("");
    setActivating(true);
    try {
      await browserApiFetch("/plans/current/activate", { method: "POST" });
      router.refresh();
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : "Não foi possível ativar o planejamento.");
    } finally {
      setActivating(false);
    }
  }

  return (
    <Card className="form-card" padding="none">
      <CardContent className="form-card-content">
        <div className="section-heading"><div><span className="section-kicker">{isEditing ? "CONFIGURAÇÃO" : "PRIMEIRO PASSO"}</span><CardTitle>{isEditing ? readonly ? "Planejamento ativo" : "Editar planejamento" : "Crie seu planejamento"}</CardTitle></div>{isEditing ? <Pencil size={19} /> : <Save size={19} />}</div>
        <p className="section-description">Defina o período que será usado para projetar suas receitas, despesas e dívidas.</p>
        {error && <Alert color="error" title="Não foi possível salvar">{error}</Alert>}
        <form className="resource-form" onSubmit={submit}>
          <Input name="name" disabled={readonly} label="Nome do planejamento" defaultValue={plan?.name ?? "Planejamento financeiro"} placeholder="Ex.: Planejamento 2026" required maxLength={120} />
          <div className="form-grid-two">
            <Input name="start_month" type="month" disabled={readonly} label="Começa em" defaultValue={plan?.start_month ?? defaultMonth} required />
            <Input name="end_month" type="month" disabled={readonly} label="Termina em" defaultValue={plan?.end_month ?? addMonths(defaultMonth, 11)} required />
          </div>
          <Input name="currency_code" disabled={readonly} label="Moeda" defaultValue={plan?.currency_code ?? currencyCode} maxLength={3} required />
          <Button type="submit" size="lg" disabled={loading || readonly}>{loading ? <><Spinner size="sm" /> Salvando...</> : <><Save size={17} /> {isEditing ? "Salvar alterações" : "Criar planejamento"}</>}</Button>
        </form>
        {plan?.status === "draft" && <Button className="secondary-action" variant="outline" size="lg" onClick={activate} disabled={activating}>{activating ? <><Spinner size="sm" /> Ativando...</> : <><Check size={17} /> Ativar planejamento</>}</Button>}
      </CardContent>
    </Card>
  );
}
