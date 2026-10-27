import type { Metadata } from "next";
import { InstitutionList } from "@/components/cards/institution-list";
import { currentMonth } from "@/lib/debt-month";
import { CardsHub } from "@/components/cards/cards-hub";
import { getCreditCards, getInstitutions, getProfile } from "@/lib/api/financial-api";

export const metadata: Metadata = { title: "Cards" };

export default async function CardsPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const query = await searchParams;
  const status = query.status === "active" || query.status === "archived" ? query.status : "all";
  const institutionStatus = query.institutionStatus === "active" || query.institutionStatus === "archived" ? query.institutionStatus : "all";
  const [institutions, cards, profile] = await Promise.all([getInstitutions(), getCreditCards(), getProfile()]);
  return <main className="resource-page-shell cards-page"><header className="resource-page-header"><p>Gerencie cartões, vencimentos e faturas.</p></header><form className="resource-form resource-query-form" method="get"><div className="form-grid-two"><label className="native-field"><span>Cartões</span><select name="status" defaultValue={status}><option value="all">Todos</option><option value="active">Ativos</option><option value="archived">Arquivados</option></select></label><label className="native-field"><span>Instituições</span><select name="institutionStatus" defaultValue={institutionStatus}><option value="all">Todas</option><option value="active">Ativas</option><option value="archived">Arquivadas</option></select></label></div><button type="submit">Filtrar</button></form><CardsHub institutions={institutions} cards={cards.filter((card) => status === "all" || card.status === status)} defaultMonth={currentMonth(profile.timezone)} returnQuery={new URLSearchParams({ status, institutionStatus }).toString()} /><InstitutionList institutions={institutions.filter((institution) => institutionStatus === "all" || institution.status === institutionStatus)} /></main>;
}
