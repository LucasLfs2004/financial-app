"use client";

import { useState } from "react";
import { Alert, Badge, Button, Card, CardContent, Spinner } from "@lucaslfs2004/luke-ui";
import { Archive, ArrowRight, CalendarDays, Pencil } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { Debt } from "@/lib/api/types";
import { DebtForm } from "@/components/debts/debt-form";
import { formatMonth } from "@/lib/debt-month";

const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });

export function DebtCard({ debt }: { debt: Debt }) {
  const router = useRouter();
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState("");
  const [archiving, setArchiving] = useState(false);
  const statusLabel: Record<Debt["projection_status"], string> = { planned: "Planejada", active: "Em andamento", completed: "Concluída", settled_early: "Quitação projetada", archived: "Arquivada" };

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
        <div className="resource-card-heading"><div><h2>{debt.name}</h2><p>{debt.description || "Dívida parcelada"}</p></div><Badge color={debt.projection_status === "completed" || debt.projection_status === "settled_early" ? "forest" : "secondary"} variant="soft">{statusLabel[debt.projection_status]}</Badge></div>
        <div className="debt-metrics"><div><span>Restantes</span><strong>{debt.remaining_installments}</strong></div><div><span>Última parcela</span><strong>{formatMonth(debt.effective_end_month)}</strong></div><div><span>Libera por mês</span><strong>{money.format(debt.released_monthly_cents / 100)}</strong></div></div>
        <p className="debt-card-release">A partir de {formatMonth(debt.release_from_month)} · valor informativo</p>
        <div className="resource-card-footer"><span><CalendarDays size={15} /> Desde {formatMonth(debt.scheduled_start_month)}</span><div><Link className="debt-detail-link" href={`/debts/${debt.id}`}>Ver parcelas <ArrowRight size={14} /></Link><Button variant="ghost" size="sm" onClick={() => setEditing(true)} disabled={debt.status === "archived"}><Pencil size={15} /> Editar</Button>{debt.status === "active" && <Button variant="ghost" size="sm" onClick={archive} disabled={archiving}>{archiving ? <Spinner size="sm" /> : <Archive size={15} />} Arquivar</Button>}</div></div>
        {error && <Alert color="error">{error}</Alert>}
      </CardContent>
    </Card>
  );
}
