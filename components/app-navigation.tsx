"use client";

import { createContext, useContext, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ArrowLeft, ChartNoAxesCombined, CreditCard, House, Landmark, Plus, ReceiptText, WalletCards } from "lucide-react";
import { SignOutButton } from "@/components/sign-out-button";
import { BrandMark } from "@/components/brand-mark";

const items = [
  { href: "/dashboard", label: "Visão geral", mobileLabel: "Início", icon: House },
  { href: "/entries", label: "Receitas e despesas", mobileLabel: "Registros", icon: WalletCards },
  { href: "/planning", label: "Planejamento", mobileLabel: "Planos", icon: ChartNoAxesCombined },
  { href: "/debts", label: "Dívidas", mobileLabel: "Dívidas", icon: Landmark },
  { href: "/cards", label: "Cartões", mobileLabel: "Cartões", icon: CreditCard },
  { href: "/invoices", label: "Faturas", mobileLabel: "Faturas", icon: ReceiptText },
];

const mobileItems = [
  { href: "/dashboard", label: "Início", icon: House, active: (path: string) => path.startsWith("/dashboard") },
  { href: "/entries", label: "Lançamentos", icon: WalletCards, active: (path: string) => path.startsWith("/entries") || path.startsWith("/debts") },
  { href: "/cards", label: "Cartões", icon: CreditCard, active: (path: string) => path.startsWith("/cards") || path.startsWith("/invoices") },
  { href: "/planning", label: "Planejar", icon: ChartNoAxesCombined, active: (path: string) => path.startsWith("/planning") },
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
      {mobileItems.map(({ href, label, icon: Icon, active }) => <Link key={href} href={href} className={active(pathname) ? "active" : ""} aria-current={active(pathname) ? "page" : undefined}><Icon size={21} /><span>{label}</span></Link>)}
    </nav>
  </>;
}

export function AppHeader({ name }: { name: string }) {
  const pathname = usePathname();
  const section = items.find(({ href }) => isActive(pathname, href))?.label ?? "Visão geral";
  const customTitle = useContext(HeaderTitleContext)?.title;
  const pageTitle = customTitle?.pathname === pathname ? customTitle.text : section;
  const invoiceRoute = /^\/invoices\/([^/]+)(?:\/(.+))?$/.exec(pathname);
  const backHref = invoiceRoute ? "/cards" : pathname.startsWith("/debts/") ? "/debts" : pathname.startsWith("/cards/") ? "/cards" : null;
  return <header className="app-header">
    <div className="app-header-leading">
      {backHref ? <Link className="app-header-back" href={backHref} aria-label="Voltar à lista anterior"><ArrowLeft size={19} /></Link> : <Link className="app-header-brand" href="/dashboard" aria-label="Projeção — início"><BrandMark /></Link>}
      <div className="app-header-location"><h1 title={pageTitle}>{pageTitle}</h1></div>
    </div>
    <div className="app-header-actions"><span className="app-header-greeting">Olá, {name}</span><SignOutButton /></div>
  </header>;
}
