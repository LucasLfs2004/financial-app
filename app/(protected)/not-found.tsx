import Link from "next/link";

export default function ResourceNotFound() {
  return <main className="resource-page-shell"><div className="empty-resource-inline"><h2>Recurso não encontrado</h2><p>Confira o endereço ou volte à lista.</p><Link href="/dashboard">Voltar à visão geral</Link></div></main>;
}
