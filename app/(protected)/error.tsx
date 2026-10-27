"use client";

import Link from "next/link";
import { Button } from "@lucaslfs2004/luke-ui";

export default function ResourceError({ reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return <main className="resource-page-shell"><div className="empty-resource-inline" role="alert"><h2>Não foi possível carregar os dados</h2><p>O serviço pode estar indisponível ou sua sessão pode ter expirado.</p><Button onClick={reset}>Tentar novamente</Button> <Link href="/sign-in">Entrar novamente</Link></div></main>;
}
