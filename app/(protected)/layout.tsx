import { redirect } from "next/navigation";
import { getAuthenticatedClaims } from "@/lib/supabase/server";
import { AppHeader, AppNavigation } from "@/components/app-navigation";

export default async function ProtectedLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const { data, error } = await getAuthenticatedClaims();

  if (error || !data?.claims?.sub) redirect("/sign-in");

  const email = String(data.claims.email ?? "");
  const name = String(data.claims.user_metadata?.name ?? "").trim().split(" ")[0] || email.split("@")[0] || "Usuário";

  return <div className="app-shell"><AppNavigation name={name} email={email} /><div className="app-main"><AppHeader name={name} />{children}</div></div>;
}
