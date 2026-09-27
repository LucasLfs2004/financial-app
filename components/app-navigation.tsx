"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ArrowUpRight, ChartNoAxesCombined, CreditCard, House, Landmark, Plus, ReceiptText, WalletCards } from "lucide-react";

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

export function AppNavigation({ name, email }: { name: string; email: string }) {
  const pathname = usePathname();

  return <>
    <aside className="app-sidebar">
      <Link className="app-sidebar-brand" href="/dashboard"><span className="brand-mark"><ArrowUpRight size={18} /></span><strong>projeção</strong></Link>
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
