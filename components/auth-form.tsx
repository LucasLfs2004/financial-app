"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Alert, Button, Card, Input, Spinner } from "@lucaslfs2004/luke-ui";
import { ArrowRight, AtSign, LockKeyhole, UserRound } from "lucide-react";
import { createClient } from "@/lib/supabase/client";

type AuthFormProps = {
  mode: "sign-in" | "sign-up";
  initialError?: string;
  nextPath?: string;
};

function authErrorMessage(message: string) {
  const normalized = message.toLowerCase();
  if (normalized.includes("invalid login credentials")) return "E-mail ou senha incorretos.";
  if (normalized.includes("user already registered")) return "Já existe uma conta com este e-mail.";
  if (normalized.includes("password")) return "A senha precisa ter ao menos 8 caracteres, com letras e números.";
  if (normalized.includes("email")) return "Informe um endereço de e-mail válido.";
  return "Não foi possível concluir. Tente novamente em instantes.";
}

export function AuthForm({ mode, initialError, nextPath }: AuthFormProps) {
  const router = useRouter();
  const isSignUp = mode === "sign-up";
  const [error, setError] = useState(initialError ?? "");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSuccess("");

    const formData = new FormData(event.currentTarget);
    const email = String(formData.get("email") ?? "").trim();
    const password = String(formData.get("password") ?? "");
    const name = String(formData.get("name") ?? "").trim();
    const passwordConfirmation = String(formData.get("passwordConfirmation") ?? "");

    if (isSignUp && password !== passwordConfirmation) {
      setError("As senhas não coincidem.");
      return;
    }

    setLoading(true);
    const supabase = createClient();

    try {
      if (isSignUp) {
        const { data, error: signUpError } = await supabase.auth.signUp({
          email,
          password,
          options: {
            data: { name },
            emailRedirectTo: `${window.location.origin}/auth/callback`,
          },
        });

        if (signUpError) throw signUpError;
        if (!data.session) {
          setSuccess("Conta criada. Confira seu e-mail para confirmar o acesso.");
          return;
        }
      } else {
        const { error: signInError } = await supabase.auth.signInWithPassword({ email, password });
        if (signInError) throw signInError;
      }

      router.replace(nextPath?.startsWith("/") ? nextPath : "/dashboard");
      router.refresh();
    } catch (caughtError) {
      const message = caughtError instanceof Error ? caughtError.message : "";
      setError(authErrorMessage(message));
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="auth-card" padding="none" background="transparent" bordered={false}>
      <div className="auth-heading">
        <span className="auth-kicker">{isSignUp ? "COMECE AGORA" : "BEM-VINDO DE VOLTA"}</span>
        <h2>{isSignUp ? "Crie sua conta" : "Entre na sua conta"}</h2>
        <p>{isSignUp ? "Prepare hoje as decisões financeiras de amanhã." : "Continue de onde parou no seu planejamento."}</p>
      </div>

      {error && <Alert color="error" title="Não foi possível continuar">{error}</Alert>}
      {success && <Alert color="success" title="Tudo certo">{success}</Alert>}

      <form className="auth-form" onSubmit={handleSubmit}>
        {isSignUp && (
          <Input
            name="name"
            label="Como podemos chamar você?"
            placeholder="Seu nome"
            autoComplete="name"
            leftIcon={<UserRound />}
            size="lg"
            required
          />
        )}
        <Input
          name="email"
          type="email"
          label="E-mail"
          placeholder="voce@exemplo.com"
          autoComplete="email"
          inputMode="email"
          leftIcon={<AtSign />}
          size="lg"
          required
        />
        <Input
          name="password"
          type="password"
          label="Senha"
          placeholder="••••••••"
          autoComplete={isSignUp ? "new-password" : "current-password"}
          helperText={isSignUp ? "Mínimo de 8 caracteres, com letras e números." : undefined}
          minLength={8}
          leftIcon={<LockKeyhole />}
          size="lg"
          required
        />
        {isSignUp && (
          <Input
            name="passwordConfirmation"
            type="password"
            label="Confirme sua senha"
            placeholder="••••••••"
            autoComplete="new-password"
            minLength={8}
            leftIcon={<LockKeyhole />}
            size="lg"
            required
          />
        )}

        <Button className="auth-submit" type="submit" size="lg" disabled={loading}>
          {loading ? <><Spinner size="sm" /> Aguarde...</> : <>{isSignUp ? "Criar minha conta" : "Entrar"}<ArrowRight size={18} /></>}
        </Button>
      </form>

      <p className="auth-switch">
        {isSignUp ? "Já tem uma conta?" : "Ainda não tem uma conta?"}{" "}
        <Link href={isSignUp ? "/sign-in" : "/sign-up"}>
          {isSignUp ? "Sign in" : "Create account"}
        </Link>
      </p>
    </Card>
  );
}
