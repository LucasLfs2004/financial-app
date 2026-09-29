import type { CardInvoiceSummary } from "@/lib/api/types";

const monthName = new Intl.DateTimeFormat("pt-BR", { month: "long", timeZone: "UTC" });

export function currentMonthInBrazil(now = new Date()): string {
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone: "America/Sao_Paulo", year: "numeric", month: "2-digit",
  }).formatToParts(now);
  const year = parts.find((part) => part.type === "year")?.value;
  const month = parts.find((part) => part.type === "month")?.value;
  return `${year}-${month}`;
}

export function addInvoiceMonths(month: string, amount: number): string {
  const [year, part] = month.split("-").map(Number);
  const date = new Date(Date.UTC(year, part - 1 + amount, 1));
  return `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, "0")}`;
}

export function invoiceMonthLabel(month: string): string {
  const [year, part] = month.split("-").map(Number);
  if (!year || !part || part < 1 || part > 12) return month;
  return `${monthName.format(new Date(Date.UTC(year, part - 1, 1)))} de ${year}`;
}

export function invoiceDueDateLabel(invoice: Pick<CardInvoiceSummary, "nominal_due_date" | "nominal_due_day">): string {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(invoice.nominal_due_date ?? "");
  if (!match) return `dia ${invoice.nominal_due_day} (data indisponível)`;
  const month = monthName.format(new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, 1)));
  return `${match[3]} de ${month} de ${match[1]}`;
}
