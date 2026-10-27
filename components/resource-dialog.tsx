"use client";

import { useEffect, useId, useRef, type ReactNode } from "react";
import { Button } from "@lucaslfs2004/luke-ui";

export function ResourceDialog({ title, children, onClose, busy = false }: { title: string; children: ReactNode; onClose: () => void; busy?: boolean }) {
  const ref = useRef<HTMLDialogElement>(null);
  const id = useId();
  useEffect(() => { const dialog = ref.current; dialog?.showModal(); return () => dialog?.close(); }, []);
  return <dialog ref={ref} className="resource-dialog" aria-labelledby={id} onCancel={(event) => { event.preventDefault(); if (!busy) onClose(); }}><div className="resource-card-footer"><h2 id={id}>{title}</h2><Button variant="ghost" disabled={busy} onClick={onClose} aria-label="Fechar diálogo">Fechar</Button></div>{children}</dialog>;
}
