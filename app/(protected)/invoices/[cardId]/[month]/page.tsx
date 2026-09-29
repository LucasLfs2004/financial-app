import type { Metadata } from "next";
import { ArrowLeft, CalendarDays, CircleDollarSign } from "lucide-react";
import Link from "next/link";
import { Card, CardContent } from "@lucaslfs2004/luke-ui";
import { getCardInvoice, getCreditCards } from "@/lib/api/financial-api";
import { InvoiceAdjustmentForm } from "@/components/cards/invoice-adjustment-form";
import { InvoiceMoveForm } from "@/components/cards/invoice-move-form";

export const metadata: Metadata = { title: "Detalhe da fatura" };

export default async function InvoiceDetailPage({ params }: { params: Promise<{ cardId: string; month: string }> }) {
  const { cardId, month } = await params;
  const [invoice, cards] = await Promise.all([getCardInvoice(cardId, month).catch(() => null), getCreditCards().catch(() => [])]);
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: invoice?.currency_code ?? "BRL" });

  return <main className="resource-page-shell">
    <header className="resource-page-header"><Link className="back-link" href={`/invoices/${cardId}`}><ArrowLeft size={16} /> Voltar às faturas</Link><span className="eyebrow">DETALHE DA FATURA</span><h1>{invoice?.card_name ?? "Fatura"}</h1><p>{invoice?.institution.name ?? "Composição da fatura"} · {month}</p></header>
    {!invoice ? <div className="empty-resource-inline">Não foi possível carregar esta fatura.</div> : <div className="invoice-detail-grid">
      <Card className="invoice-total-card" padding="none"><CardContent>
        <CalendarDays size={20} /><span>Saldo projetado da fatura</span><strong>{money.format(invoice.projected_total_cents / 100)}</strong>
        <div className="invoice-total-breakdown"><span>Compras <strong>{money.format(invoice.charges_total_cents / 100)}</strong></span><span>Pagamentos <strong>{money.format(invoice.payments_total_cents / 100)}</strong></span></div>
        <small>Vencimento nominal: dia {invoice.nominal_due_day}{invoice.nominal_due_date ? ` · ${invoice.nominal_due_date}` : " · data inválida para este mês"}</small>
      </CardContent></Card>
      <section><div className="section-heading compact"><div><span className="section-kicker">COMPOSIÇÃO</span><h2>{invoice.component_count} componente(s)</h2></div><CircleDollarSign size={20} /></div>
        <div className="resource-list">{invoice.components.map((component) => <Card className="resource-card" padding="none" key={component.source_id}><CardContent className="resource-card-content">
          <div className="resource-card-heading"><div><h2>{component.name}</h2><p>{component.source_type === "invoice_payment" ? `Pagamento em ${component.payment_date ?? "data não informada"}` : component.reference_month ? `Competência ${component.reference_month}` : "Competência desconhecida"}</p></div><strong>{money.format(component.amount_cents / 100)}</strong></div>
          {component.debt_id && <Link className="component-origin" href={`/debts/${component.debt_id}`}>{component.debt_occurrence_kind === "early_settlement" ? "Quitação antecipada" : `Parcela ${component.installment_number ?? "?"}/${component.installments_total ?? "?"}`} · Ver dívida</Link>}
          {!component.debt_id && <span className="component-origin">{component.source_type === "invoice_payment" ? "Pagamento da fatura" : `Origem: ${component.allocation}`}</span>}
          {component.item_id && component.reference_month && <InvoiceMoveForm itemId={component.item_id} referenceMonth={component.reference_month} cards={cards} />}
        </CardContent></Card>)}</div>
        <InvoiceAdjustmentForm cardId={cardId} month={month} />
      </section>
    </div>}
  </main>;
}
