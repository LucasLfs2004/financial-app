"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import { Button, Input, Alert, Spinner } from "@lucaslfs2004/luke-ui";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { CreditCard, InvoiceMoveHistory } from "@/lib/api/types";
import { formatMonth } from "@/lib/debt-month";

export function InvoiceMoveForm({ itemId, referenceMonth, cards, initialOpen = false }: { itemId: string; referenceMonth: string; cards: CreditCard[]; initialOpen?: boolean }) {
  const router = useRouter();
  const [open, setOpen] = useState(initialOpen);
  const [history, setHistory] = useState<InvoiceMoveHistory | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [revision, setRevision] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const lock = useRef(false);
  const path = `/financial-items/${encodeURIComponent(itemId)}/occurrences/${encodeURIComponent(referenceMonth)}/invoice-moves`;
  function toggle(value: boolean) {
    setOpen(value);
    const url = new URL(window.location.href);
    if (value) { url.searchParams.set("moveItem", itemId); url.searchParams.set("referenceMonth", referenceMonth); }
    else { url.searchParams.delete("moveItem"); url.searchParams.delete("referenceMonth"); }
    window.history.replaceState(null, "", url);
  }
  useEffect(() => {
    if (!open) return;
    const controller = new AbortController();
    setLoading(true); setHistory(null); setError("");
    browserApiFetch<InvoiceMoveHistory>(path, { signal: controller.signal }).then((response) => {
      if (!Array.isArray(response.data)) throw new Error("Histórico de movimentações inválido.");
      if (!controller.signal.aborted) setHistory(response);
    }).catch((caught) => { if (!controller.signal.aborted) setError(caught instanceof Error ? caught.message : "Histórico indisponível."); }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [open, path, revision]);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (lock.current) return;
    const values = new FormData(event.currentTarget);
    lock.current = true; setSubmitting(true); setError("");
    try {
      await browserApiFetch(path, { method: "POST", body: JSON.stringify({ target_credit_card_id: String(values.get("target_credit_card_id")), target_payment_month: String(values.get("target_payment_month")), reason: String(values.get("reason") || "").trim() || null }) });
      setRevision((value) => value + 1); router.refresh();
    } catch (caught) { setError(caught instanceof Error ? caught.message : "Não foi possível mover a ocorrência."); }
    finally { lock.current = false; setSubmitting(false); }
  }
  const name = (id: string) => cards.find((card) => card.id === id)?.name ?? "Cartão histórico";
  if (!open) return <Button variant="ghost" size="sm" onClick={() => toggle(true)}>Movimentações</Button>;
  return <section className="invoice-move-panel"><div className="resource-card-footer"><h3>Movimentações · {formatMonth(referenceMonth)}</h3><Button variant="ghost" onClick={() => toggle(false)} disabled={submitting}>Fechar</Button></div>{loading && <p role="status">Carregando histórico...</p>}{error && <Alert color="error">{error}<Button variant="ghost" onClick={() => setRevision((value) => value + 1)}>Consultar novamente</Button></Alert>}{history && <>{history.current_allocation ? <p>Alocação atual: <Link href={`/invoices/${encodeURIComponent(history.current_allocation.credit_card_id)}/${history.current_allocation.payment_month}`}>{name(history.current_allocation.credit_card_id)} · {formatMonth(history.current_allocation.payment_month)}</Link> · {({ calculated_from_reference: "Calculada pela competência", selected_invoice: "Fatura selecionada", moved_by_user: "Movida pelo usuário" } as Record<string, string>)[history.current_allocation.origin] ?? "Alocação registrada"}</p> : <p>{history.current_allocation === null ? "Sem alocação atual de cartão." : "Alocação atual indisponível neste histórico."}</p>}{!history.data.length ? <p>Nenhuma movimentação registrada.</p> : <ol>{[...history.data].sort((a, b) => a.recorded_at.localeCompare(b.recorded_at) || a.id.localeCompare(b.id)).map((move) => <li key={move.id}><Link href={`/invoices/${encodeURIComponent(move.from_credit_card_id)}/${move.from_payment_month}`}>{name(move.from_credit_card_id)} · {formatMonth(move.from_payment_month)}</Link> → <Link href={`/invoices/${encodeURIComponent(move.to_credit_card_id)}/${move.to_payment_month}`}>{name(move.to_credit_card_id)} · {formatMonth(move.to_payment_month)}</Link><p>{move.reason ?? "Sem motivo informado"} · {new Date(move.recorded_at).toLocaleString("pt-BR")}</p></li>)}</ol>}</>}<p>Uma nova movimentação preserva os eventos anteriores, inclusive para voltar à fatura de origem.</p><form className="resource-form" onSubmit={submit}><fieldset className="resource-form subscription-form-fields" disabled={submitting || loading || !history}><label className="native-field"><span>Cartão de destino</span><select name="target_credit_card_id" required defaultValue=""><option value="" disabled>Selecione</option>{cards.filter((card) => card.status === "active").map((card) => <option key={card.id} value={card.id}>{card.institution.name} · {card.name}</option>)}</select></label><Input name="target_payment_month" type="month" label="Mês de pagamento" required /><Input name="reason" label="Motivo" maxLength={500} /></fieldset><Button type="submit" disabled={submitting || loading || !history}>{submitting && <Spinner size="sm" />}Registrar movimentação</Button></form></section>;
}
