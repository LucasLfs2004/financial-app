"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Alert, Button, Input } from "@lucaslfs2004/luke-ui";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { ApiEnvelope, InvoiceAdjustmentDetail } from "@/lib/api/types";
import { adjustmentPatch, adjustmentPath } from "@/lib/api/adjustment";
import { MutationForm } from "@/components/resource-mutation";
import { formatMonth } from "@/lib/debt-month";

export function InvoiceAdjustmentEditor({ id, initialOpen }: { id: string; initialOpen: boolean }) {
  const [open, setOpen] = useState(initialOpen);
  const [detail, setDetail] = useState<InvoiceAdjustmentDetail | null>(null);
  const [error, setError] = useState("");
  const [revision, setRevision] = useState(0);
  const [confirmArchive, setConfirmArchive] = useState(false);
  function toggle(value: boolean) {
    setOpen(value); setConfirmArchive(false);
    const url = new URL(window.location.href);
    if (value) url.searchParams.set("adjustment", id); else url.searchParams.delete("adjustment");
    window.history.replaceState(null, "", url);
  }
  useEffect(() => {
    if (!open) return;
    const controller = new AbortController();
    setDetail(null); setError("");
    browserApiFetch<ApiEnvelope<InvoiceAdjustmentDetail>>(`/invoice-adjustments/${encodeURIComponent(id)}`, { signal: controller.signal }).then((response) => {
      if (response.data.id !== id || !response.data.capabilities || !Array.isArray(response.data.capabilities.editable_fields)) throw new Error("Dados de manutenção indisponíveis.");
      if (!controller.signal.aborted) setDetail(response.data);
    }).catch((caught) => { if (!controller.signal.aborted) setError(caught instanceof Error ? caught.message : "Não foi possível consultar o lançamento."); });
    return () => controller.abort();
  }, [open, id, revision]);
  if (!open) return <Button variant="ghost" size="sm" onClick={() => toggle(true)}>Consultar lançamento</Button>;
  const editable = detail && detail.status === "active" && detail.entry_kind !== "payment" && detail.capabilities.can_edit;
  const canArchive = detail && detail.status === "active" && detail.entry_kind !== "payment" && detail.capabilities.can_archive;
  return <section className="invoice-move-panel"><Button variant="ghost" onClick={() => toggle(false)}>Fechar lançamento</Button>{error ? <Alert color="error">{error}<Button variant="ghost" onClick={() => setRevision((value) => value + 1)}>Consultar novamente</Button></Alert> : !detail ? <p role="status">Consultando lançamento...</p> : <><h3>{detail.name}</h3><p>Pagamento em {formatMonth(detail.payment_month)} · {detail.entry_kind === "payment" ? "Pagamento recebido" : detail.is_projected ? "Projeção de parcela importada" : detail.import_source ? "Cobrança importada" : "Cobrança manual"} · {detail.status}</p>{detail.reconciled_item_id && <p>Vínculo persistido: <Link href={`/entries/${encodeURIComponent(detail.reconciled_item_id)}${detail.reconciled_reference_month ? `?month=${detail.reconciled_reference_month}` : ""}`}>Consultar assinatura{detail.reconciled_reference_month ? ` · ${formatMonth(detail.reconciled_reference_month)}` : ""}</Link></p>}{editable && <MutationForm key={detail.updated_at} path={adjustmentPath(detail)} method="PATCH" onSaved={() => setRevision((value) => value + 1)} payload={(form) => adjustmentPatch(detail, form)}>{detail.capabilities.editable_fields.includes("name") && <Input name="name" label="Nome" maxLength={120} defaultValue={detail.name} required />}{detail.capabilities.editable_fields.includes("amount_cents") && <Input name="amount" type="number" min="0" step="0.01" label={`Valor (${detail.currency_code})`} defaultValue={(detail.amount_cents / 100).toFixed(2)} required />}{detail.capabilities.editable_fields.includes("reference_month") && Object.hasOwn(detail, "reference_month") && <Input name="reference_month" type="month" label="Competência" defaultValue={detail.reference_month ?? ""} />}{detail.capabilities.editable_fields.includes("context") && Object.hasOwn(detail, "context") && <Input name="context" label="Contexto" maxLength={500} defaultValue={detail.context ?? ""} />}</MutationForm>}{canArchive && <><Button variant="ghost" onClick={() => setConfirmArchive(true)}>Arquivar lançamento</Button>{confirmArchive && <><p>{detail.capabilities.archive_effect === "restore_subscription_forecast" ? "A previsão da assinatura voltará a compor esse mês. A identidade da cobrança continuará reservada." : detail.capabilities.archive_effect === "remove_charge" ? "A cobrança será retirada da projeção, preservando seu histórico." : "O lançamento ficará arquivado conforme as regras da API."}</p><MutationForm path={`${adjustmentPath(detail)}/archive`} payload={() => ({})} onSaved={() => { setConfirmArchive(false); setRevision((value) => value + 1); }} label="Confirmar arquivamento"><p>Arquivar {detail.name}?</p></MutationForm><Button variant="ghost" onClick={() => setConfirmArchive(false)}>Cancelar</Button></>}</>}</>}</section>;
}
