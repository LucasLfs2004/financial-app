"use client";

import { useRef, useState, type FormEvent, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { Alert, Button, Spinner } from "@lucaslfs2004/luke-ui";
import { browserApiFetch } from "@/lib/api/browser-api";
import { ResourceDialog } from "@/components/resource-dialog";

export function MutationForm({ path, method = "POST", payload, children, label = "Salvar", disabled = false, onSaved, onBusyChange }: { path: string; method?: string; payload: (form: FormData) => Record<string, unknown>; children: ReactNode; label?: string; disabled?: boolean; onSaved?: () => void; onBusyChange?: (value: boolean) => void }) {
  const router = useRouter();
  const lock = useRef(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (lock.current || disabled) return;
    const form = new FormData(event.currentTarget);
    lock.current = true; setBusy(true); onBusyChange?.(true); setError(""); setSaved(false);
    try {
      const body = payload(form);
      if (!Object.keys(body).length && method === "PATCH") { setSaved(true); return; }
      await browserApiFetch(path, { method, body: JSON.stringify(body) });
      setSaved(true); onSaved?.(); router.refresh();
    } catch (caught) { setError(caught instanceof Error ? caught.message : "Não foi possível salvar."); }
    finally { lock.current = false; setBusy(false); onBusyChange?.(false); }
  }
  return <form className="resource-form" onSubmit={submit}><fieldset className="resource-form subscription-form-fields" disabled={disabled || busy}>{children}</fieldset>{error && <Alert color="error">{error}</Alert>}{saved && <p role="status">Dados atualizados.</p>}<Button type="submit" disabled={disabled || busy}>{busy && <Spinner size="sm" />}{label}</Button></form>;
}

export function ArchiveAction({ path, name, effect, allowMonth = false }: { path: string; name: string; effect: string; allowMonth?: boolean }) {
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  return <div className="debt-action-panel"><Button variant="ghost" onClick={() => setOpen(true)} disabled={open}>Arquivar</Button>{open && <ResourceDialog title={`Arquivar ${name}?`} onClose={() => setOpen(false)} busy={busy}><p>{effect}</p><MutationForm path={path} label="Confirmar arquivamento" onBusyChange={setBusy} onSaved={() => setOpen(false)} payload={(form) => {
    const month = String(form.get("effective_from") || "");
    return { reason: String(form.get("reason") || "").trim() || null, ...(allowMonth && month ? { effective_from: month } : {}) };
  }}>{allowMonth && <label className="native-field"><span>Interromper ocorrências a partir de (opcional)</span><input type="month" name="effective_from" /></label>}<label className="native-field"><span>Motivo (opcional)</span><input name="reason" maxLength={500} /></label></MutationForm><Button variant="ghost" disabled={busy} onClick={() => setOpen(false)}>Cancelar</Button></ResourceDialog>}</div>;
}
