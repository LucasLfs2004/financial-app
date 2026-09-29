"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { ApiEnvelope, CardInvoice, CardInvoiceSummary } from "@/lib/api/types";
import { browserApiFetch } from "@/lib/api/browser-api";
import { invoiceDueDateLabel, invoiceMonthLabel } from "@/lib/invoice-display";

const MONTHS_PER_PAGE = 6;
const shortMonth = new Intl.DateTimeFormat("pt-BR", { month: "short", timeZone: "UTC" });

function monthTick(month: string) {
  const [year, part] = month.split("-").map(Number);
  return `${shortMonth.format(new Date(Date.UTC(year, part - 1, 1))).replace(".", "")} ${String(year).slice(2)}`;
}

export function InvoiceSpendingChart({ cardId, invoices, currentPaymentMonth }: { cardId: string; invoices: CardInvoiceSummary[]; currentPaymentMonth: string }) {
  const initialIndex = invoices.findIndex((invoice) => invoice.payment_month === currentPaymentMonth);
  const [selectedMonth, setSelectedMonth] = useState(invoices[initialIndex >= 0 ? initialIndex : 0]?.payment_month ?? "");
  const [page, setPage] = useState(Math.floor(Math.max(initialIndex, 0) / MONTHS_PER_PAGE));
  const [detail, setDetail] = useState<CardInvoice | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const selected = invoices.find((invoice) => invoice.payment_month === selectedMonth);
  const visible = invoices.slice(page * MONTHS_PER_PAGE, (page + 1) * MONTHS_PER_PAGE);
  const maxCharges = Math.max(1, ...invoices.map((invoice) => Math.max(0, invoice.charges_total_cents)));
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: selected?.currency_code ?? "BRL" });

  useEffect(() => {
    if (!selectedMonth) return;
    const controller = new AbortController();
    setDetail(null);
    setLoading(true);
    setError("");
    browserApiFetch<ApiEnvelope<CardInvoice>>(`/credit-cards/${encodeURIComponent(cardId)}/invoices/${selectedMonth}`, { signal: controller.signal })
      .then((response) => setDetail(response.data))
      .catch((cause) => {
        if (!(cause instanceof DOMException && cause.name === "AbortError")) setError("Não foi possível carregar os lançamentos desta fatura.");
      })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [cardId, selectedMonth]);

  function changePage(nextPage: number) {
    const next = invoices[nextPage * MONTHS_PER_PAGE];
    if (!next) return;
    setPage(nextPage);
    setSelectedMonth(next.payment_month);
  }

  if (!selected) return null;
  const charges = detail?.payment_month === selectedMonth ? detail.components.filter((component) => component.source_type !== "invoice_payment") : [];

  return <section className="invoice-spending-panel" aria-label="Gastos por fatura">
    <div className="invoice-spending-heading"><div><span className="section-kicker">GASTOS DO CARTÃO</span><h2>Compras por fatura</h2><p>Selecione um mês para ver os lançamentos.</p></div><div className="invoice-chart-navigation"><button type="button" aria-label="Meses anteriores" onClick={() => changePage(page - 1)} disabled={page === 0}><ChevronLeft size={18} /></button><button type="button" aria-label="Próximos meses" onClick={() => changePage(page + 1)} disabled={(page + 1) * MONTHS_PER_PAGE >= invoices.length}><ChevronRight size={18} /></button></div></div>
    <div className="invoice-chart-bars" role="group" aria-label="Selecione uma fatura no gráfico">{visible.map((invoice) => {
      const amount = Math.max(0, invoice.charges_total_cents);
      return <button className={`invoice-chart-month${invoice.payment_month === selectedMonth ? " selected" : ""}`} type="button" key={invoice.payment_month} aria-pressed={invoice.payment_month === selectedMonth} aria-label={`Fatura de ${invoiceMonthLabel(invoice.payment_month)}: ${money.format(amount / 100)} em compras`} onClick={() => setSelectedMonth(invoice.payment_month)}>
        <span className="invoice-chart-track"><span className={`invoice-chart-bar${amount === 0 ? " empty" : ""}`} style={{ height: `${amount === 0 ? 4 : Math.max(5, amount / maxCharges * 100)}%` }} /></span><span className="invoice-chart-month-label">{monthTick(invoice.payment_month)}</span>
      </button>;
    })}</div>
    <div className="invoice-chart-selected" aria-live="polite"><div className="invoice-chart-selected-heading"><div><span className="section-kicker">FATURA SELECIONADA</span><h3>{invoiceMonthLabel(selected.payment_month)}</h3><p>Vencimento: {invoiceDueDateLabel(selected)}</p></div><strong>{money.format(selected.charges_total_cents / 100)}<small>em compras</small></strong></div>
      <div className="invoice-chart-totals"><span>Pagamentos <strong>{money.format(selected.payments_total_cents / 100)}</strong></span><span>Saldo da fatura <strong>{money.format(selected.projected_total_cents / 100)}</strong></span></div>
      {loading && <p className="invoice-chart-message" role="status">Carregando lançamentos...</p>}
      {error && <p className="invoice-chart-message" role="alert">{error}</p>}
      {!loading && !error && detail?.payment_month === selectedMonth && <div className="invoice-chart-expenses"><h4>Compras e encargos</h4>{charges.length === 0 ? <p className="invoice-chart-message">Nenhum gasto nesta fatura.</p> : <><ul>{charges.slice(0, 5).map((component) => <li key={component.source_id}><span>{component.name}</span><strong>{money.format(component.amount_cents / 100)}</strong></li>)}</ul>{charges.length > 5 && <p className="invoice-chart-message">Exibindo 5 de {charges.length} lançamentos.</p>}</>}</div>}
      <Link className="invoice-chart-detail-link" href={`/invoices/${cardId}/${selected.payment_month}`}>Ver fatura completa <ChevronRight size={16} /></Link>
    </div>
  </section>;
}
