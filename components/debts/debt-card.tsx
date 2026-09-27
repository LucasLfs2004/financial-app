"use client";

import { useState } from "react";
import { Alert, Badge, Button, Card, CardContent } from "@lucaslfs2004/luke-ui";
import { Archive, CalendarDays, Pencil } from "lucide-react";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { Debt } from "@/lib/api/types";
import { DebtForm } from "@/components/debts/debt-form";

const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });

export function DebtCard({ debt }: { debt: Debt }) {
  const router = useRouter();
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState("");
  const [archiving, setArchiving] = useState(false);

  async function archive() {
    setError("");
    setArchiving(true);
    try {
      await browserApiFetch(`/debts/${debt.id}/archive`, { method: "POST" });
      router.refresh();
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : "Não foi possível arquivar a dívida.");
    } finally {
      setArchiving(false);
    }
  }

  if (editing) return <DebtForm debt={debt} onCancel={() => setEditing(false)} />;

  return (
    <Card className="resource-card" padding="none">
      <CardContent className="resource-card-content">
        <div className="resource-card-heading"><div><h2>{debt.name}</h2><p>{debt.description || "Dívida parcelada"}</p></div><Badge color={debt.projection_status === "completed" ? "forest" : "secondary"} variant="soft">{debt.projection_status === "completed" ? "Concluída" : "Em andamento"}</Badge></div>
        <div className="debt-metrics"><div><span>Parcela atual</span><strong>{money.format(debt.released_monthly_cents / 100)}</strong></div><div><span>Restantes</span><strong>{debt.remaining_installments}</strong></div><div><span>Termina em</span><strong>{debt.effective_end_month}</strong></div></div>
        <div className="resource-card-footer"><span><CalendarDays size={15} /> Desde {debt.scheduled_start_month}</span><div><Button variant="ghost" size="sm" onClick={() => setEditing(true)}><Pencil size={15} /> Editar</Button><Button variant="ghost" size="sm" onClick={archive} disabled={archiving}><Archive size={15} /> {archiving ? "..." : "Arquivar"}</Button></div></div>
        {error && <Alert color="error">{error}</Alert>}
      </CardContent>
    </Card>
  );
}
