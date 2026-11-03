"use client";

import Link from "next/link";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { addMonths, isMonth } from "@/lib/debt-month";

export function MonthNavigation({ month, min, max, onChange, basePath, query = "", monthInPath = false }: {
  month: string; min?: string; max?: string; onChange?: (month: string) => void;
  basePath?: string; query?: string; monthInPath?: boolean;
}) {
  const previous = addMonths(month, -1);
  const next = addMonths(month, 1);
  const label = new Intl.DateTimeFormat("pt-BR", { month: "long", year: "numeric", timeZone: "UTC" }).format(new Date(`${month}-02T12:00:00Z`));
  function href(value: string) {
    const params = new URLSearchParams(query);
    if (!monthInPath) params.set("month", value);
    const search = params.toString();
    return `${basePath}${monthInPath ? `/${value}` : ""}${search ? `?${search}` : ""}`;
  }
  return <nav className="month-navigation" aria-label="Mês de consulta">
    {onChange ? <button type="button" aria-label="Mês anterior" disabled={Boolean(min && previous < min)} onClick={() => onChange(previous)}><ChevronLeft size={19} /></button>
      : min && previous < min ? <span /> : <Link aria-label="Mês anterior" href={href(previous)}><ChevronLeft size={19} /></Link>}
    {onChange ? <label className="month-navigation-picker"><span className="sr-only">Mês de consulta</span><input type="month" value={month} min={min} max={max} onChange={(event) => { if (isMonth(event.target.value)) onChange(event.target.value); }} /></label> : <span className="month-navigation-label">{label}</span>}
    {onChange ? <button type="button" aria-label="Próximo mês" disabled={Boolean(max && next > max)} onClick={() => onChange(next)}><ChevronRight size={19} /></button>
      : max && next > max ? <span /> : <Link aria-label="Próximo mês" href={href(next)}><ChevronRight size={19} /></Link>}
  </nav>;
}
