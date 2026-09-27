import type { Metadata } from "next";
import { Avatar, Badge, Card, CardContent } from "@lucaslfs2004/luke-ui";
import { ArrowDownLeft, ArrowUpRight, Bell, ChartNoAxesCombined, CircleDollarSign, CreditCard, House, PiggyBank, Plus, ReceiptText, WalletCards } from "lucide-react";
import { createClient } from "@/lib/supabase/server";
import { getMonthlySummary, getProfile } from "@/lib/api/financial-api";
import { SignOutButton } from "@/components/sign-out-button";
import Link from "next/link";
import { MonthlySummaryPanel } from "@/components/dashboard/monthly-summary";

export const metadata: Metadata = { title: "Visão geral" };

export default async function DashboardPage() {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  const profile = await getProfile().catch(() => null);
  const initialMonth = new Date().toISOString().slice(0, 7);
  const initialSummary = await getMonthlySummary(initialMonth).catch(() => null);
  const displayName = profile?.display_name ?? String(user?.user_metadata?.name ?? "").trim() ?? "";
  const firstName = displayName.split(" ")[0] || "por aqui";

  return (
    <main className="dashboard-shell">
      <header className="dashboard-header">
        <div className="dashboard-brand"><span className="brand-mark"><ArrowUpRight size={16} /></span><strong>projeção</strong></div>
        <div className="dashboard-actions">
          <button className="header-icon" aria-label="Notificações"><Bell size={18} /></button>
          <Avatar name={displayName || user?.email || "Usuário"} size="sm" />
          <SignOutButton />
        </div>
      </header>

      <div className="dashboard-content">
        <section className="dashboard-welcome">
          <div><span className="eyebrow">VISÃO GERAL</span><h1>Olá, {firstName}.</h1><p>Vamos construir uma visão mais tranquila do seu dinheiro.</p></div>
          <Link className="quick-action" href="/planning"><Plus size={19} /><span>Novo planejamento</span></Link>
        </section>

        {!profile && (
          <div className="api-status"><span className="status-dot" />Sua sessão está ativa. Inicie a API Go na porta 8080 para carregar os dados financeiros.</div>
        )}

        {initialSummary && <MonthlySummaryPanel initialMonth={initialMonth} initialSummary={initialSummary} />}

        <section className="summary-grid" aria-label="Resumo financeiro">
          <Card className="summary-card featured" padding="none"><CardContent className="summary-card-content"><div className="summary-top"><span>Saldo projetado</span><WalletCards size={19} /></div><strong>R$ —</strong><small>Crie seu primeiro planejamento</small></CardContent></Card>
          <Card className="summary-card" padding="none"><CardContent className="summary-card-content"><div className="summary-top"><span>Receitas</span><ArrowDownLeft size={19} /></div><strong>R$ —</strong><small className="positive">Entradas do mês</small></CardContent></Card>
          <Card className="summary-card" padding="none"><CardContent className="summary-card-content"><div className="summary-top"><span>Despesas</span><ArrowUpRight size={19} /></div><strong>R$ —</strong><small className="negative">Saídas do mês</small></CardContent></Card>
        </section>

        <section className="dashboard-grid">
          <Card className="empty-plan" padding="none">
            <CardContent className="empty-plan-content">
              <span className="empty-icon"><ChartNoAxesCombined size={25} /></span>
              <Badge color="forest" variant="soft">Primeiros passos</Badge>
              <h2>Seu planejamento começa aqui</h2>
              <p>Cadastre receitas e compromissos para enxergar os próximos meses com clareza.</p>
              <Link className="primary-action" href="/planning"><Plus size={18} /> Criar planejamento</Link>
            </CardContent>
          </Card>
          <Card className="next-steps" padding="none">
            <CardContent className="next-steps-content">
              <h2>Prepare sua projeção</h2>
              <div className="step-item"><span><CircleDollarSign size={18} /></span><div><strong>Adicione sua renda</strong><small>Salário e outras receitas recorrentes</small></div></div>
              <div className="step-item"><span><ReceiptText size={18} /></span><div><strong>Mapeie seus gastos</strong><small>Fixos, variáveis e cartões</small></div></div>
              <div className="step-item"><span><PiggyBank size={18} /></span><div><strong>Defina quanto guardar</strong><small>Transforme intenção em meta</small></div></div>
            </CardContent>
          </Card>
        </section>
      </div>

      <nav className="bottom-navigation" aria-label="Navegação principal">
        <a className="active" href="/dashboard"><House size={20} /><span>Início</span></a>
        <a href="/planning"><ChartNoAxesCombined size={20} /><span>Planejamento</span></a>
        <a href="/debts"><Plus size={22} /><span>Dívidas</span></a>
        <a href="#transacoes"><ReceiptText size={20} /><span>Transações</span></a>
        <a href="/cards"><CreditCard size={20} /><span>Cartões</span></a>
      </nav>
    </main>
  );
}
