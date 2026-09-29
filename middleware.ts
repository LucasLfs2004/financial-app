import type { NextRequest } from "next/server";
import { updateSession } from "@/lib/supabase/middleware";

export async function middleware(request: NextRequest) {
  return updateSession(request);
}

export const config = {
  matcher: ["/dashboard/:path*", "/planning/:path*", "/entries/:path*", "/debts/:path*", "/cards/:path*", "/invoices/:path*", "/sign-in", "/sign-up", "/entrar", "/criar-conta"],
};
