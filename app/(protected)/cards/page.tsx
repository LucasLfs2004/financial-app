import type { Metadata } from "next";
import { CardsHub } from "@/components/cards/cards-hub";
import { getCreditCards, getInstitutions } from "@/lib/api/financial-api";

export const metadata: Metadata = { title: "Cards" };

export default async function CardsPage() {
  const [institutions, cards] = await Promise.all([getInstitutions().catch(() => []), getCreditCards().catch(() => [])]);
  return <main className="resource-page-shell cards-page"><header className="resource-page-header"><p>Gerencie cartões, vencimentos e faturas.</p></header><CardsHub institutions={institutions} cards={cards} /></main>;
}
