import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { AppHeader, AppNavigation } from "@/components/app-navigation";

export default async function ProtectedLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const supabase = await createClient();
  const { data, error } = await supabase.auth.getUser();

  if (error || !data.user) redirect("/sign-in");

  const email = data.user.email ?? "";
  const name = String(data.user.user_metadata?.name ?? "").trim().split(" ")[0] || email.split("@")[0] || "Usuário";

  return <div className="app-shell"><AppNavigation name={name} email={email} /><div className="app-main"><AppHeader name={name} />{children}</div></div>;
}
