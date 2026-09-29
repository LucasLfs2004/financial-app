"use client";

import { useRef, useState, type ChangeEvent, type FormEvent } from "react";
import { Alert, Button, Card, CardContent, Spinner } from "@lucaslfs2004/luke-ui";
import { FileUp, ReceiptText } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { browserApiFetch, BrowserApiError } from "@/lib/api/browser-api";
import type { CreditCard, InvoiceImportPreview, InvoiceImportResult, InvoiceImportRow } from "@/lib/api/types";

const MAX_FILE_BYTES = 2 * 1024 * 1024;
type ImportResponse<T> = T | { data: T };
type PreviewSelection = { cardId: string; month: string; file: File; data: InvoiceImportPreview };
type ImportOutcome = { cardId: string; month: string; data: InvoiceImportResult };

function unpack<T extends object>(response: ImportResponse<T>): T {
  return "data" in response ? response.data : response;
}

function money(cents: number) {
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(cents / 100);
}

function errorMessage(error: unknown) {
  if (error instanceof BrowserApiError) {
    const detail = error.message;
    switch (error.status) {
      case 400: return `Arquivo ou token inválido. ${detail}`;
      case 404: return "Cartão não encontrado. Atualize a página e selecione outro cartão.";
      case 409: return "Este cartão foi arquivado e não aceita importações.";
      case 422: return `Mês fora do horizonte ou sem configuração do cartão. ${detail}`;
      default: return detail;
    }
  }
  return error instanceof Error ? error.message : "Não foi possível importar a fatura.";
}

function statusLabel(row: InvoiceImportRow) {
  switch (row.status) {
    case "new": return "Nova";
    case "existing": return "Já existente";
    case "imported": return "Importada";
    case "skipped": return "Ignorada";
  }
}

function reasonLabel(reason: string | null | undefined) {
  if (reason === "non_expense") return "Pagamento, crédito ou valor zero — não entra nos gastos";
  return reason?.replaceAll("_", " ") ?? null;
}

function ImportRows({ rows }: { rows: InvoiceImportRow[] }) {
  return <div className="invoice-import-rows">
    {rows.map((row, index) => <div className="invoice-import-row" key={`${index}-${row.date}-${row.title}`}>
      <div><strong>{row.title}</strong><span>{row.date} · {money(row.amount_cents)}</span>{row.reason && <small>{reasonLabel(row.reason)}</small>}</div>
      <span className={`invoice-import-status status-${row.status}`}>{statusLabel(row)}</span>
    </div>)}
  </div>;
}

export function InvoiceCsvImport({ cards }: { cards: CreditCard[] }) {
  const router = useRouter();
  const today = new Date();
  const [cardId, setCardId] = useState("");
  const [month, setMonth] = useState(`${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}`);
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<PreviewSelection | null>(null);
  const [outcome, setOutcome] = useState<ImportOutcome | null>(null);
  const [pending, setPending] = useState<"preview" | "confirm" | null>(null);
  const [error, setError] = useState("");
  const selectionVersion = useRef(0);

  function invalidate() {
    selectionVersion.current += 1;
    setPreview(null);
    setOutcome(null);
    setError("");
  }

  function changeFile(event: ChangeEvent<HTMLInputElement>) {
    invalidate();
    const selected = event.target.files?.[0] ?? null;
    setFile(selected);
    if (selected && selected.size > MAX_FILE_BYTES) {
      setError("O arquivo deve ter no máximo 2 MiB.");
    }
  }

  async function requestPreview(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!cardId || !/^\d{4}-(0[1-9]|1[0-2])$/.test(month) || !file) return;
    if (file.size > MAX_FILE_BYTES) { setError("O arquivo deve ter no máximo 2 MiB."); return; }
    const version = ++selectionVersion.current;
    setPreview(null);
    setOutcome(null);
    setError("");
    setPending("preview");
    const form = new FormData();
    form.append("file", file);
    try {
      const response = await browserApiFetch<ImportResponse<InvoiceImportPreview>>(`/credit-cards/${encodeURIComponent(cardId)}/invoices/${month}/imports/preview`, { method: "POST", body: form, successMessage: false });
      if (version === selectionVersion.current) setPreview({ cardId, month, file, data: unpack(response) });
    } catch (caught) {
      if (version === selectionVersion.current) setError(errorMessage(caught));
    } finally {
      setPending(null);
    }
  }

  async function confirm() {
    if (!preview || preview.data.new_count === 0 || pending || preview.cardId !== cardId || preview.month !== month || preview.file !== file) return;
    const selection = preview;
    const version = selectionVersion.current;
    setError("");
    setPending("confirm");
    const form = new FormData();
    form.append("file", selection.file);
    form.append("preview_token", selection.data.preview_token);
    try {
      const response = await browserApiFetch<ImportResponse<InvoiceImportResult>>(`/credit-cards/${encodeURIComponent(selection.cardId)}/invoices/${selection.month}/imports`, { method: "POST", body: form });
      if (version === selectionVersion.current) {
        setOutcome({ cardId: selection.cardId, month: selection.month, data: unpack(response) });
        setPreview(null);
        router.refresh();
      }
    } catch (caught) {
      if (version === selectionVersion.current) {
        setError(errorMessage(caught));
        if (caught instanceof BrowserApiError && caught.status === 400) setPreview(null);
      }
    } finally {
      setPending(null);
    }
  }

  const activeCards = cards.filter((card) => card.status === "active");
  const resultRows = outcome?.data.rows ?? [];
  return <Card className="form-card invoice-import" padding="none"><CardContent className="form-card-content">
    <div className="section-heading"><div><span className="section-kicker">FATURA NUBANK</span><h2>Importar CSV</h2></div><FileUp size={20} /></div>
    <p className="section-description">Selecione o cartão, o mês de pagamento da fatura e o CSV. A prévia não grava compras.</p>
    <form className="resource-form" onSubmit={requestPreview}>
      <div className="form-grid-two">
        <label className="native-field"><span>Cartão</span><select required value={cardId} onChange={(event) => { invalidate(); setCardId(event.target.value); }} disabled={pending === "confirm"}><option value="" disabled>Selecione o cartão</option>{activeCards.map((card) => <option value={card.id} key={card.id}>{card.institution.name} · {card.name}</option>)}</select></label>
        <label className="native-field"><span>Mês da fatura</span><input type="month" required value={month} onChange={(event) => { invalidate(); setMonth(event.target.value); }} disabled={pending === "confirm"} /></label>
      </div>
      <label className="native-field"><span>Arquivo CSV</span><input type="file" accept=".csv,text/csv" required onChange={changeFile} disabled={pending === "confirm"} /><small>Até 2 MiB · colunas date,title,amount · data AAAA-MM-DD · valor em reais com vírgula decimal · até 5.000 linhas</small></label>
      <Button type="submit" disabled={!!pending || !cardId || !month || !file || file.size > MAX_FILE_BYTES}>{pending === "preview" ? <Spinner size="sm" /> : <ReceiptText size={15} />} {pending === "preview" ? "Gerando prévia..." : "Ver prévia"}</Button>
    </form>
    {error && <div className="invoice-import-feedback"><Alert color="error">{error}</Alert></div>}
    {preview && <section className="invoice-import-result" aria-label="Prévia da importação">
      <h3>Prévia</h3>
      <div className="invoice-import-metrics"><span><strong>{preview.data.new_count}</strong> novas</span><span><strong>{preview.data.existing_count}</strong> existentes</span><span><strong>{preview.data.skipped_count}</strong> ignoradas</span><span><strong>{money(preview.data.new_total_cents)}</strong> em novas compras</span></div>
      <ImportRows rows={preview.data.rows} />
      <Button type="button" onClick={confirm} disabled={!!pending || preview.data.new_count === 0}>{pending === "confirm" ? <Spinner size="sm" /> : <FileUp size={15} />} {pending === "confirm" ? "Importando..." : "Confirmar importação"}</Button>
    </section>}
    {outcome && <section className="invoice-import-result" aria-label="Resultado da importação" role="status">
      <h3>Importação concluída</h3>
      <div className="invoice-import-metrics"><span><strong>{outcome.data.imported_count ?? resultRows.filter((row) => row.status === "imported").length}</strong> importadas</span><span><strong>{outcome.data.existing_count ?? resultRows.filter((row) => row.status === "existing").length}</strong> existentes</span><span><strong>{outcome.data.skipped_count ?? resultRows.filter((row) => row.status === "skipped").length}</strong> ignoradas</span><span><strong>{money(outcome.data.imported_total_cents ?? resultRows.filter((row) => row.status === "imported").reduce((sum, row) => sum + row.amount_cents, 0))}</strong> importados</span></div>
      <ImportRows rows={resultRows} />
      <Link className="small-link" href={`/invoices/${outcome.cardId}/${outcome.month}`}>Ver fatura atualizada</Link>
    </section>}
  </CardContent></Card>;
}
