"use client";

import { useState, type FormEvent } from "react";
import { Alert, Button, Card, CardContent, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { Plus, Save } from "lucide-react";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";

export function InvoiceAdjustmentForm({ cardId, month }: { cardId: string; month: string }) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const values = new FormData(form);
    const name = String(values.get("name") ?? "").trim();
    const payload = {
      name,
      amount_cents: Math.round(Number(String(values.get("amount")).replace(",", ".")) * 100),
      reference_month: String(values.get("reference_month") || "") || null,
      context: String(values.get("context") || "").trim() || null,
    };
    setLoading(true);
    setError("");
    setSuccess("");

    try {
      await browserApiFetch(`/credit-cards/${cardId}/invoices/${month}/adjustments`, { method: "POST", body: JSON.stringify(payload) });
      form.reset();
      setSuccess(`Cobrança ${name} adicionada com sucesso.`);
      router.refresh();
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : "Não foi possível adicionar o ajuste.");
    } finally {
      setLoading(false);
    }
  }

  return <Card className="form-card adjustment-form" padding="none"><CardContent className="form-card-content">
    <div className="section-heading"><div><span className="section-kicker">AJUSTE CONSOLIDADO</span><h2>Adicionar cobrança</h2></div><Plus size={20} /></div>
    <form className="resource-form" onSubmit={submit}>
      <Input name="name" label="Nome" placeholder="Ex.: Compra não identificada" required />
      <div className="form-grid-two"><Input name="amount" type="number" min="0" step="0.01" label="Valor (R$)" required /><Input name="reference_month" type="month" label="Competência" /></div>
      <Input name="context" label="Contexto" placeholder="Opcional" />
      <Button type="submit" disabled={loading}>{loading ? <Spinner size="sm" /> : <Save size={15} />} {loading ? "Adicionando..." : "Adicionar ajuste"}</Button>
    </form>
    <div aria-live="polite">{error && <Alert color="error">{error}</Alert>}{success && <Alert color="success">{success}</Alert>}</div>
  </CardContent></Card>;
}
