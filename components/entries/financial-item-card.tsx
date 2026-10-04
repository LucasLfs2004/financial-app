"use client";

import { useState } from "react";
import { Alert, Button, Card, CardContent, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { Archive, ArrowDownLeft, ArrowUpRight, CalendarClock, CalendarDays, Pencil, Save, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { CreditCard, FinancialItem } from "@/lib/api/types";

import { SubscriptionFields } from "./subscription-fields";

const labels: Record<string, string> = { recurring_income: "Renda recorrente", one_time_income: "Renda pontual", fixed_expense: "Despesa fixa", projected_variable_expense: "Despesa variável" };
const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
const monthLabel = new Intl.DateTimeFormat("pt-BR", { month: "long", year: "numeric", timeZone: "UTC" });

function formatMonth(month: string | undefined) {
  if (!month) return "Data não informada";
  return monthLabel.format(new Date(`${month}-02T12:00:00Z`));
}

export function FinancialItemCard({ item, cards }: { item: FinancialItem; cards: CreditCard[] }) {
  const router = useRouter();
  const [mode, setMode] = useState<"view" | "edit" | "change" | "payment">("view");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const period = item.periods[item.periods.length - 1];
  const income = item.kind === "recurring_income" || item.kind === "one_time_income";
  const changes = Math.max(0, item.periods.length - 1);

  async function saveMetadata(form: FormData) {
    return browserApiFetch(`/financial-items/${item.id}`, { method: "PATCH", body: JSON.stringify({ name: String(form.get("name")).trim(), description: String(form.get("description") || "").trim() || null, ...(item.kind === "fixed_expense" ? { invoice_match_title: String(form.get("invoice_match_title") || "").trim() || null, renewal_day: form.get("renewal_day") ? Number(form.get("renewal_day")) : null } : {}) }) });
  }

  async function archive() {
    setLoading(true); setError("");
    try { await browserApiFetch(`/financial-items/${item.id}/archive`, { method: "POST", body: JSON.stringify({}) }); router.refresh(); }
    catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : "Não foi possível arquivar o item."); }
    finally { setLoading(false); }
  }

  async function submitEdit(form: FormData) {
    setLoading(true); setError("");
    try { await saveMetadata(form); setMode("view"); router.refresh(); }
    catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : "Não foi possível atualizar o item."); }
    finally { setLoading(false); }
  }

  async function submitChange(form: FormData) {
    setLoading(true); setError("");
    try {
      await browserApiFetch(`/financial-items/${item.id}/changes`, { method: "POST", body: JSON.stringify({ effective_from: String(form.get("effective_from")), end_month: String(form.get("end_month") || "") || null, amount_cents: Math.round(Number(String(form.get("amount")).replace(",", ".")) * 100), cash_month_offset: Number(form.get("cash_month_offset") || 0), context: String(form.get("context") || "").trim() || null }) });
      setMode("view"); router.refresh();
    } catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : "Não foi possível registrar a mudança."); }
    finally { setLoading(false); }
  }

  async function submitPayment(form: FormData) {
    setLoading(true); setError("");
    try { await browserApiFetch(`/financial-items/${item.id}/payment-changes`, { method: "POST", body: JSON.stringify({ effective_from: String(form.get("effective_from")), end_month: String(form.get("end_month") || "") || null, method: String(form.get("method")), credit_card_id: String(form.get("credit_card_id") || "") || null, context: String(form.get("context") || "").trim() || null }) }); setMode("view"); router.refresh(); }
    catch (caughtError) { setError(caughtError instanceof Error ? caughtError.message : "Não foi possível alterar o meio de pagamento."); }
    finally { setLoading(false); }
  }

  if (mode === "edit") return <Card className="resource-card" padding="none"><CardContent className="resource-card-content"><form className="resource-form" action={submitEdit}><Input name="name" label="Nome" defaultValue={item.name} required /><Input name="description" label="Descrição" defaultValue={item.description ?? ""} />{item.kind === "fixed_expense" && <SubscriptionFields title={item.invoice_match_title} renewalDay={item.renewal_day} />}<div className="form-actions"><Button type="submit" disabled={loading}>{loading ? <Spinner size="sm" /> : <Save size={15} />} Salvar</Button><Button type="button" variant="ghost" onClick={() => setMode("view")}><X size={15} /> Cancelar</Button></div></form>{error && <Alert color="error">{error}</Alert>}</CardContent></Card>;
  if (mode === "change") return <Card className="resource-card" padding="none"><CardContent className="resource-card-content"><div className="section-heading"><h3>Alterar valor a partir de um mês</h3><Button variant="ghost" size="sm" onClick={() => setMode("view")}><X size={15} /></Button></div><form className="resource-form" action={submitChange}><div className="form-grid-two"><Input name="effective_from" type="month" label="A partir de" defaultValue={period?.start_month ?? "2026-01"} required /><Input name="end_month" type="month" label="Até" defaultValue={period?.end_month ?? ""} /></div><div className="form-grid-two"><Input name="amount" type="number" min="0" step="0.01" label="Novo valor (R$)" required /><Input name="cash_month_offset" type="number" min="0" max="12" label="Deslocamento" defaultValue={period?.cash_month_offset ?? 0} required /></div><Input name="context" label="Contexto" placeholder="Ex.: Reajuste" /><Button type="submit" disabled={loading}>{loading ? <Spinner size="sm" /> : <Save size={15} />} Registrar mudança</Button></form>{error && <Alert color="error">{error}</Alert>}</CardContent></Card>;
  if (mode === "payment") return <Card className="resource-card" padding="none"><CardContent className="resource-card-content"><div className="section-heading"><h3>Alterar forma de pagamento</h3><Button variant="ghost" size="sm" onClick={() => setMode("view")}><X size={15} /></Button></div><form className="resource-form" action={submitPayment}><div className="form-grid-two"><Input name="effective_from" type="month" label="A partir de" defaultValue="2026-01" required /><Input name="end_month" type="month" label="Até" /></div><label className="native-field"><span>Método</span><select name="method" defaultValue="direct"><option value="direct">Pagamento direto</option><option value="credit_card">Cartão de crédito</option></select></label><label className="native-field"><span>Cartão</span><select name="credit_card_id" defaultValue=""><option value="">Nenhum</option>{cards.filter((card) => card.status === "active").map((card) => <option value={card.id} key={card.id}>{card.institution.name} · {card.name}</option>)}</select></label><Input name="context" label="Contexto" placeholder="Opcional" /><Button type="submit" disabled={loading}>{loading ? <Spinner size="sm" /> : <Save size={15} />} Salvar método</Button></form>{error && <Alert color="error">{error}</Alert>}</CardContent></Card>;

  return <article className={`financial-record-card ${income ? "income" : "expense"} ${item.status === "archived" ? "archived" : ""}`}>
    <div className="financial-record-top">
      <span className="financial-record-icon">{income ? <ArrowDownLeft size={22} /> : <ArrowUpRight size={22} />}</span>
      <span className="financial-record-status">{item.status === "active" ? "Ativo" : "Arquivado"}</span>
    </div>
    <div className="financial-record-title"><div><span>{item.kind === "fixed_expense" && (period?.context === "Assinatura" || item.invoice_match_title) ? "Assinatura" : labels[item.kind] ?? item.kind}</span><h3>{item.name}</h3></div><small>{period?.recurrence === "once" ? "Pontual" : "Mensal"}</small></div>
    <div className="financial-record-value"><span>Valor {period?.recurrence === "once" ? "registrado" : "por mês"}</span><strong>{money.format((period?.amount_cents ?? 0) / 100)}</strong></div>
    <div className="financial-record-timing"><span><CalendarDays size={16} /> Desde {formatMonth(period?.start_month)}</span>{period?.end_month && period.end_month !== period.start_month && <span>Até {formatMonth(period.end_month)}</span>}</div>
    {item.renewal_day != null && <p className="resource-hint">Renovação prevista: dia {item.renewal_day}</p>}
    {item.invoice_match_title && <p className="resource-hint">Na fatura: {item.invoice_match_title}</p>}
    {item.description && <p className="financial-record-description">{item.description}</p>}
    <div className="financial-record-actions"><span>{changes > 0 ? `${changes} ${changes === 1 ? "alteração" : "alterações"} de valor` : ""}</span>{item.status === "active" && <div><Button variant="ghost" size="sm" onClick={() => setMode("edit")}><Pencil size={15} /> Editar</Button><Button variant="ghost" size="sm" onClick={() => setMode("change")}><CalendarClock size={15} /> Alterar valor</Button>{!income && <Button variant="ghost" size="sm" onClick={() => setMode("payment")}>Pagamento</Button>}<Button variant="ghost" size="sm" onClick={archive} disabled={loading}>{loading ? <Spinner size="sm" /> : <Archive size={15} />} Arquivar</Button></div>}</div>
    {error && <Alert color="error">{error}</Alert>}
  </article>;
}
