import Link from "next/link";
import { ArrowUpRight, LockKeyhole, ShieldCheck, Sparkles } from "lucide-react";

export default function AuthLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <main className="auth-shell">
      <section className="auth-brand-panel" aria-label="Apresentação do produto">
        <Link className="brand" href="/" aria-label="Projeção — início">
          <span className="brand-mark"><ArrowUpRight size={18} /></span>
          <span>projeção</span>
        </Link>

        <div className="auth-brand-copy">
          <span className="eyebrow"><Sparkles size={14} /> Clareza para decidir melhor</span>
          <h1>Seu dinheiro olhando para o futuro.</h1>
          <p>Planeje cenários, acompanhe seus compromissos e transforme números em decisões tranquilas.</p>
        </div>

        <div className="security-note">
          <ShieldCheck size={18} />
          <span><strong>Sessão protegida</strong>Autenticação segura pelo Supabase.</span>
        </div>
      </section>

      <section className="auth-form-panel">
        <div className="auth-mobile-header">
          <Link className="brand" href="/">
            <span className="brand-mark"><ArrowUpRight size={17} /></span>
            <span>projeção</span>
          </Link>
          <span className="secure-label"><LockKeyhole size={14} /> Ambiente seguro</span>
        </div>
        {children}
      </section>
    </main>
  );
}
