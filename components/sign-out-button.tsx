"use client";

import { Button } from "@lucaslfs2004/luke-ui";
import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { createClient } from "@/lib/supabase/client";

export function SignOutButton() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  async function signOut() {
    setLoading(true);
    const supabase = createClient();
    await supabase.auth.signOut();
    router.replace("/sign-in");
    router.refresh();
  }

  return (
    <Button variant="ghost" size="icon" aria-label="Sair da conta" onClick={signOut} disabled={loading}>
      <LogOut size={18} />
    </Button>
  );
}
