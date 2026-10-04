"use client";

import { createContext, useContext, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ArrowLeft, ChartNoAxesCombined, CreditCard, House, Landmark, Plus, ReceiptText, WalletCards } from "lucide-react";
import { SignOutButton } from "@/components/sign-out-button";
import type { MonthlySummary } from "@/lib/api/types";
import { API_NOTICE_EVENT, type ApiNotice } from "@/lib/api/notice-event";
import { BrandMark } from "@/components/brand-mark";

const items = [
  { href: "/dashboard", label: "Visão geral", mobileLabel: "Início", icon: House },
  { href: "/entries", label: "Receitas e despesas", mobileLabel: "Registros", icon: WalletCards },
  { href: "/planning", label: "Planejamento", mobileLabel: "Planos", icon: ChartNoAxesCombined },
  { href: "/debts", label: "Dívidas", mobileLabel: "Dívidas", icon: Landmark },
  { href: "/cards", label: "Cartões", mobileLabel: "Cartões", icon: CreditCard },
  { href: "/invoices", label: "Faturas", mobileLabel: "Faturas", icon: ReceiptText },
];

function isActive(pathname: string, href: string) {
  return pathname === href || pathname.startsWith(`${href}/`);
}

const HeaderTitleContext = createContext<{ title: { text: string; pathname: string } | null; setTitle: (title: { text: string; pathname: string } | null) => void } | null>(null);

export function AppHeaderProvider({ children }: { children: React.ReactNode }) {
  const [title, setTitle] = useState<{ text: string; pathname: string } | null>(null);
  return <HeaderTitleContext.Provider value={{ title, setTitle }}>{children}</HeaderTitleContext.Provider>;
}

export function AppHeaderTitle({ title }: { title: string }) {
  const pathname = usePathname();
  const setTitle = useContext(HeaderTitleContext)?.setTitle;
  useEffect(() => {
    setTitle?.({ text: title, pathname });
    return () => setTitle?.(null);
  }, [setTitle, title, pathname]);
  return null;
}

export function AppNavigation({ name, email }: { name: string; email: string }) {
  const pathname = usePathname();

  return <>
    <aside className="app-sidebar">
      <Link className="app-sidebar-brand" href="/dashboard"><BrandMark /><strong>projeção</strong></Link>
      <div className="app-sidebar-main">
        <span className="sidebar-label">MENU</span>
        <nav className="sidebar-links" aria-label="Navegação principal">
          {items.map(({ href, label, icon: Icon }) => <Link key={href} href={href} className={isActive(pathname, href) ? "active" : ""} aria-current={isActive(pathname, href) ? "page" : undefined}><Icon size={19} /><span>{label}</span></Link>)}
        </nav>
        <div className="sidebar-create"><span className="sidebar-label">CADASTRAR</span><Link href="/entries#new-item"><Plus size={16} /> Receita ou despesa</Link><Link href="/debts#new-debt"><Plus size={16} /> Nova dívida</Link><Link href="/cards#new-card"><Plus size={16} /> Novo cartão</Link></div>
      </div>
      <div className="sidebar-user"><span className="sidebar-avatar">{name.charAt(0).toUpperCase()}</span><span><strong>{name}</strong><small>{email}</small></span></div>
    </aside>
    <nav className="app-mobile-navigation" aria-label="Navegação principal">
      {items.map(({ href, label, mobileLabel, icon: Icon }) => <Link key={href} href={href} className={isActive(pathname, href) ? "active" : ""} aria-label={label} aria-current={isActive(pathname, href) ? "page" : undefined}><Icon size={19} /><span>{mobileLabel ?? label}</span></Link>)}
    </nav>
  </>;
}

export function AppHeader({ name }: { name: string }) {
  const pathname = usePathname();
  const section = items.find(({ href }) => isActive(pathname, href))?.label ?? "Visão geral";
  const customTitle = useContext(HeaderTitleContext)?.title;
  const pageTitle = customTitle?.pathname === pathname ? customTitle.text : section;
  const invoiceRoute = /^\/invoices\/([^/]+)(?:\/(.+))?$/.exec(pathname);
  const backHref = invoiceRoute ? invoiceRoute[2] ? `/invoices/${invoiceRoute[1]}` : "/invoices" : pathname.startsWith("/debts/") ? "/debts" : null;
  const [cash, setCash] = useState<MonthlySummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let controller: AbortController | undefined;
    async function refresh() {
      controller?.abort();
      const request = new AbortController();
      controller = request;
      setLoading(true);
      const parts = new Intl.DateTimeFormat("en-US", { timeZone: "America/Sao_Paulo", year: "numeric", month: "2-digit" }).formatToParts(new Date());
      const month = `${parts.find((part) => part.type === "year")?.value}-${parts.find((part) => part.type === "month")?.value}`;
      try {
        const response = await fetch(`/api/financial/plans/current/months/${month}/summary?basis=cash`, { cache: "no-store", signal: request.signal });
        const payload = response.ok ? await response.json() as { data: MonthlySummary } : null;
        if (!request.signal.aborted) setCash(payload?.data ?? null);
      } catch {
        if (!request.signal.aborted) setCash(null);
      } finally {
        if (!request.signal.aborted) setLoading(false);
      }
    }
    function onNotice(event: Event) {
      if ((event as CustomEvent<ApiNotice>).detail.type === "success") void refresh();
    }
    void refresh();
    window.addEventListener(API_NOTICE_EVENT, onNotice);
    window.addEventListener("focus", refresh);
    return () => {
      controller?.abort();
      window.removeEventListener(API_NOTICE_EVENT, onNotice);
      window.removeEventListener("focus", refresh);
    };
  }, [pathname]);

  const monthLabel = cash ? new Intl.DateTimeFormat("pt-BR", { month: "short", timeZone: "UTC" }).format(new Date(`${cash.month}-02T12:00:00Z`)).replace(".", "") : "mês atual";
  return <header className="app-header">
    <div className="app-header-leading">
      {backHref ? <Link className="app-header-back" href={backHref} aria-label="Voltar à lista anterior"><ArrowLeft size={19} /></Link> : <Link className="app-header-brand" href="/dashboard" aria-label="Projeção — início"><BrandMark /></Link>}
      <div className="app-header-location"><h1 title={pageTitle}>{pageTitle}</h1><span className={`app-header-cash${cash?.is_negative ? " is-negative" : ""}`} aria-busy={loading}>Caixa projetado · {monthLabel}<strong>{loading ? "…" : cash ? new Intl.NumberFormat("pt-BR", { style: "currency", currency: cash.currency_code }).format(cash.result_cents / 100) : "Indisponível"}</strong></span></div>
    </div>
    <div className="app-header-actions"><span className="app-header-greeting">Olá, {name}</span><SignOutButton /></div>
  </header>;
}
