import Link from "next/link";
import { AppHeaderTitle } from "@/components/app-navigation";
import { OriginalSnapshot } from "@/components/planning/original-snapshot";
import { getOriginalPlan, absentOn404 } from "@/lib/api/financial-api";

export default async function OriginalPage() {
  const snapshot = await getOriginalPlan().catch(absentOn404);
  return <main className="resource-page-shell"><AppHeaderTitle title="Referência original" /><Link className="back-link" href="/planning">Voltar ao planejamento</Link>{snapshot ? <OriginalSnapshot snapshot={snapshot} /> : <p>A referência original estará disponível após a ativação do planejamento.</p>}</main>;
}
