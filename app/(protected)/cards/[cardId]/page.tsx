import Link from "next/link";
import { notFound } from "next/navigation";
import { AppHeaderTitle } from "@/components/app-navigation";
import { CardDetail } from "@/components/cards/card-detail";
import { FinancialApiError, getCreditCard, getProfile } from "@/lib/api/financial-api";
import { currentMonth, isMonth } from "@/lib/debt-month";

export default async function CardDetailPage({ params, searchParams }: { params: Promise<{ cardId: string }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const { cardId } = await params;
  const query = await searchParams;
  const [card, profile] = await Promise.all([getCreditCard(cardId).catch((error: unknown) => { if (error instanceof FinancialApiError && error.status === 404) notFound(); throw error; }), getProfile()]);
  const month = isMonth(query.month) ? query.month : currentMonth(profile.timezone);
  const status = query.status === "active" || query.status === "archived" ? query.status : "all";
  const institutionStatus = query.institutionStatus === "active" || query.institutionStatus === "archived" ? query.institutionStatus : "all";
  return <main className="resource-page-shell"><AppHeaderTitle title={card.name} /><Link className="back-link" href={`/cards?${new URLSearchParams({ status, institutionStatus })}`}>Voltar aos cartões</Link><CardDetail key={card.updated_at + month} card={card} month={month} returnQuery={new URLSearchParams({ status, institutionStatus }).toString()} /></main>;
}
