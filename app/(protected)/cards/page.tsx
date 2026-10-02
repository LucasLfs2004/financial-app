import type { Metadata } from "next";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { CardsHub } from "@/components/cards/cards-hub";
import { getCreditCards, getInstitutions } from "@/lib/api/financial-api";

export const metadata: Metadata = { title: "Cards" };

export default async function CardsPage() {
  const [institutions, cards] = await Promise.all([getInstitutions().catch(() => []), getCreditCards().catch(() => [])]);
  return <main className="resource-page-shell cards-page"><header className="resource-page-header"><Link className="back-link" href="/dashboard"><ArrowLeft size={16} /> Voltar ao dashboard</Link><span className="eyebrow">CARTÕES</span><h1>Instituições e cartões</h1><p>Cadastre seus cartões para explicar de onde vêm as faturas e quando elas entram no caixa.</p></header><CardsHub institutions={institutions} cards={cards} /></main>;
}
