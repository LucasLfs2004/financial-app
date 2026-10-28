"use client";

import { useState } from "react";
import { Button, Input } from "@lucaslfs2004/luke-ui";
import type { FinancialInstitution } from "@/lib/api/types";
import { ArchiveAction, MutationForm } from "@/components/resource-mutation";
import { ResourceDialog } from "@/components/resource-dialog";
import { textField } from "@/lib/api/requests";

export function InstitutionList({ institutions }: { institutions: FinancialInstitution[] }) {
  const [editing, setEditing] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  return <section className="resource-list-section"><h2>Instituições financeiras</h2>{!institutions.length && <p>Nenhuma instituição corresponde ao filtro.</p>}<div className="resource-list">{institutions.map((institution) => <article key={institution.id + institution.updated_at} className="debt-action-panel"><h3>{institution.name}</h3><p>{institution.status === "active" ? "Ativa" : "Arquivada"}</p>{institution.status === "active" && <>{editing === institution.id ? <ResourceDialog title={`Editar ${institution.name}`} busy={busy} onClose={() => setEditing(null)}><MutationForm path={`/financial-institutions/${encodeURIComponent(institution.id)}`} method="PATCH" onBusyChange={setBusy} onSaved={() => setEditing(null)} payload={(form) => ({ name: textField(form, "name") })}><Input name="name" label="Nome da instituição" defaultValue={institution.name} required maxLength={120} /></MutationForm><Button variant="ghost" disabled={busy} onClick={() => setEditing(null)}>Cancelar edição</Button></ResourceDialog> : <Button variant="ghost" onClick={() => setEditing(institution.id)}>Editar nome</Button>}<ArchiveAction path={`/financial-institutions/${encodeURIComponent(institution.id)}/archive`} name={institution.name} effect="A instituição deixará de aceitar novos cartões. Os cartões existentes mantêm seu próprio status e histórico." /></>}</article>)}</div></section>;
}
