"use client";

import { useState, type FormEvent } from "react";
import { Alert, Button, Card, CardContent, CardTitle, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { PiggyBank, Save } from "lucide-react";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { PlannedSavings } from "@/lib/api/types";

export function SavingsForm({ savings }: { savings: PlannedSavings | null }) {
  const router = useRouter(); const [loading, setLoading] = useState(false); const [error, setError] = useState(""); const period = savings?.periods[savings.periods.length - 1];
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setLoading(true); setError(""); const values = new FormData(event.currentTarget);
    try { await browserApiFetch("/plans/current/savings", { method: "PUT", body: JSON.stringify({ effective_from: String(values.get("effective_from")), end_month: String(values.get("end_month") || "") || null, amount_cents: Math.round(Number(String(values.get("amount")).replace(",", ".")) * 100), context: String(values.get("context") || "").trim() || null }) }); router.refresh(); }
    catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : "Não foi possível salvar o valor planejado."); }
    finally { setLoading(false); }
  }
  return <Card className="form-card" padding="none"><CardContent className="form-card-content"><div className="section-heading"><div><span className="section-kicker">ECONOMIA</span><CardTitle>Quanto guardar</CardTitle></div><PiggyBank size={20} /></div><p className="section-description">Um valor zero também é uma configuração válida e permite ativar o planejamento.</p>{error && <Alert color="error">{error}</Alert>}<form className="resource-form" onSubmit={submit}><div className="form-grid-two"><Input name="effective_from" type="month" label="A partir de" defaultValue={period?.start_month ?? "2026-01"} required /><Input name="end_month" type="month" label="Até" defaultValue={period?.end_month ?? "2026-12"} /></div><Input name="amount" type="number" min="0" step="0.01" label="Valor mensal (R$)" defaultValue={period ? (period.amount_cents / 100).toFixed(2) : "0.00"} required /><Input name="context" label="Contexto" defaultValue={period?.context ?? ""} placeholder="Ex.: Reserva mensal" /><Button type="submit" size="lg" disabled={loading}>{loading ? <><Spinner size="sm" /> Salvando...</> : <><Save size={17} /> Salvar economia</>}</Button></form></CardContent></Card>;
}
