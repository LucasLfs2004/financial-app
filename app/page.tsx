import { redirect } from "next/navigation";
import { getAuthenticatedClaims } from "@/lib/supabase/server";

export default async function Home() {
  const { data } = await getAuthenticatedClaims();

  redirect(data?.claims?.sub ? "/dashboard" : "/sign-in");
}
