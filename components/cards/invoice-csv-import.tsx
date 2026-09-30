"use client";

import { useEffect, useRef, useState, type ChangeEvent, type FormEvent } from "react";
import { Alert, Button, Spinner } from "@lucaslfs2004/luke-ui";
import { ArrowUpRight, FileSpreadsheet, FileUp, ReceiptText, X } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { browserApiFetch, BrowserApiError } from "@/lib/api/browser-api";
import type { CreditCard, InvoiceImportPreview, InvoiceImportResult, InvoiceImportRow } from "@/lib/api/types";
import { formatDate } from "@/lib/format-date";

const MAX_FILE_BYTES = 2 * 1024 * 1024;
const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
type ImportResponse<T> = T | { data: T };
type Allocations = Record<string, string>;
type PreviewState = { data: InvoiceImportPreview; allocationsJson: string; cardId: string; month: string; file: File };
type ImportOutcome = { cardId: string; month: string; data: InvoiceImportResult };

function unpack<T extends object>(response: ImportResponse<T>): T {
  return "data" in response ? response.data : response;
}

function money(cents: number) {
  return currency.format(cents / 100);
}

function invoiceCount(data: InvoiceImportPreview) {
  return Number.isFinite(data.needs_invoice_count)
    ? data.needs_invoice_count!
    : data.rows.filter((row) => row.status === "needs_invoice").length;
}

function ignoredPayments(data: InvoiceImportPreview) {
  return data.rows.filter((row) => row.status === "skipped" && row.reason === "non_expense" && /^pagamento recebido/i.test(row.title.trim())).length;
}

function importTotals(data: InvoiceImportPreview, confirmed: boolean) {
  const rows = data.rows.filter((row) => row.status === (confirmed ? "imported" : "new"));
  const rowCharges = rows.reduce((total, row) => total + (row.kind === "payment" ? 0 : row.amount_cents), 0);
  const rowPayments = rows.reduce((total, row) => total + (row.kind === "payment" ? row.amount_cents : 0), 0);
  const net = Number.isFinite(data.new_total_cents) ? data.new_total_cents : rowCharges + rowPayments;
  const charges = Number.isFinite(data.new_charges_cents)
    ? data.new_charges_cents!
    : Number.isFinite(data.new_payments_cents) ? net - data.new_payments_cents! : rowCharges;
  const payments = Number.isFinite(data.new_payments_cents)
    ? data.new_payments_cents!
    : net - charges;
  return { charges, payments, net };
}

function errorMessage(error: unknown) {
  if (error instanceof BrowserApiError) {
    switch (error.status) {
      case 400: return "Arquivo, mês de pagamento ou token inválido. Confira as escolhas e gere uma nova prévia.";
      case 404: return "Cartão não encontrado. Atualize a página e selecione outro cartão.";
      case 409: return "Este cartão foi arquivado e não aceita importações.";
      case 422: return "O mês está fora do horizonte ou sem configuração do cartão.";
      default: return error.message;
    }
  }
  return error instanceof Error ? error.message : "Não foi possível importar a fatura.";
}

function statusLabel(status: InvoiceImportRow["status"]) {
  switch (status) {
    case "new": return "Novo";
    case "existing": return "Já existente";
    case "skipped": return "Ignorado";
    case "imported": return "Importado";
    case "needs_invoice": return "Escolha a fatura";
  }
}

function rowExplanation(row: InvoiceImportRow) {
  if (row.reason === "already_imported_to_other_invoice") return `Já importado na fatura ${row.invoice_payment_month ?? "original"}. Não será lançado novamente.`;
  if (row.status === "skipped" && row.reason === "non_expense" && /^pagamento recebido/i.test(row.title.trim())) return "Este pagamento foi ignorado pela API e não entrará na fatura.";
  if (row.reason === "non_expense") return "Crédito ou valor zero — não entra nos gastos.";
  return row.reason?.replaceAll("_", " ") ?? null;
}

function Totals({ data, confirmed = false }: { data: InvoiceImportPreview; confirmed?: boolean }) {
  const totals = importTotals(data, confirmed);
  return <div className="invoice-import-metrics">
    <span><strong>{money(totals.charges)}</strong>{confirmed ? "compras importadas" : "novas compras"}</span>
    <span><strong>{money(totals.payments)}</strong>{confirmed ? "pagamentos importados" : "novos pagamentos"}</span>
    <span><strong>{money(totals.net)}</strong>saldo líquido</span>
    <span><strong>{confirmed ? (data as InvoiceImportResult).imported_count ?? data.new_count : data.new_count}</strong>{confirmed ? "lançados" : "novos lançamentos"}</span>
    <span><strong>{data.existing_count}</strong>existentes</span>
    <span><strong>{data.skipped_count}</strong>ignorados</span>
  </div>;
}

export function InvoiceCsvImport({ cards, initialCardId = "" }: { cards: CreditCard[]; initialCardId?: string }) {
  const router = useRouter();
  const dialogRef = useRef<HTMLDialogElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const requestId = useRef(0);
  const previewAbort = useRef<AbortController | null>(null);
  const today = new Date();
  const defaultMonth = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}`;
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState<"file" | "review" | "done">("file");
  const [cardId, setCardId] = useState(initialCardId);
  const [month, setMonth] = useState(defaultMonth);
  const [file, setFile] = useState<File | null>(null);
  const [allocations, setAllocations] = useState<Allocations>({});
  const [preview, setPreview] = useState<PreviewState | null>(null);
  const [outcome, setOutcome] = useState<ImportOutcome | null>(null);
  const [pending, setPending] = useState<"preview" | "confirm" | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  }, [open]);

  function invalidate() {
    requestId.current += 1;
    previewAbort.current?.abort();
    setPreview(null);
    setOutcome(null);
    setAllocations({});
    setPending(null);
    setError("");
    setStep("file");
  }

  function start() {
    invalidate();
    setCardId(initialCardId);
    setMonth(defaultMonth);
    setFile(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
    setOpen(true);
  }

  function changeFile(event: ChangeEvent<HTMLInputElement>) {
    invalidate();
    const selected = event.target.files?.[0] ?? null;
    setFile(selected);
    if (selected && selected.size > MAX_FILE_BYTES) setError("O arquivo deve ter no máximo 2 MiB.");
  }

  async function loadPreview(nextAllocations: Allocations, selectedCard: string, selectedMonth: string, selectedFile: File) {
    const id = ++requestId.current;
    previewAbort.current?.abort();
    const controller = new AbortController();
    previewAbort.current = controller;
    const allocationsJson = JSON.stringify(nextAllocations);
    setPending("preview");
    setError("");
    const form = new FormData();
    form.append("file", selectedFile);
    form.append("payment_allocations", allocationsJson);
    try {
      const response = await browserApiFetch<ImportResponse<InvoiceImportPreview>>(`/credit-cards/${encodeURIComponent(selectedCard)}/invoices/${selectedMonth}/imports/preview`, { method: "POST", body: form, signal: controller.signal, successMessage: false });
      if (id === requestId.current) setPreview({ data: unpack(response), allocationsJson, cardId: selectedCard, month: selectedMonth, file: selectedFile });
    } catch (caught) {
      if (id === requestId.current && !(caught instanceof DOMException && caught.name === "AbortError")) setError(errorMessage(caught));
    } finally {
      if (id === requestId.current) setPending(null);
    }
  }

  function submitFile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!cardId || !/^\d{4}-(0[1-9]|1[0-2])$/.test(month) || !file) return;
    if (file.size > MAX_FILE_BYTES) { setError("O arquivo deve ter no máximo 2 MiB."); return; }
    setStep("review");
    void loadPreview({}, cardId, month, file);
  }

  function allocate(line: number, invoiceMonth: string) {
    if (!cardId || !file) return;
    const next = { ...allocations };
    if (invoiceMonth) next[String(line)] = invoiceMonth;
    else delete next[String(line)];
    setAllocations(next);
    void loadPreview(next, cardId, month, file);
  }

  async function confirm() {
    if (!preview || !file || pending || preview.file !== file || preview.cardId !== cardId || preview.month !== month || preview.allocationsJson !== JSON.stringify(allocations) || invoiceCount(preview.data) !== 0 || ignoredPayments(preview.data) > 0 || preview.data.new_count === 0) return;
    const selection = preview;
    const id = ++requestId.current;
    setPending("confirm");
    setError("");
    const form = new FormData();
    form.append("file", selection.file);
    form.append("payment_allocations", selection.allocationsJson);
    form.append("preview_token", selection.data.preview_token);
    try {
      const response = await browserApiFetch<ImportResponse<InvoiceImportResult>>(`/credit-cards/${encodeURIComponent(selection.cardId)}/invoices/${selection.month}/imports`, { method: "POST", body: form, successMessage: false });
      if (id === requestId.current) {
        setOutcome({ cardId: selection.cardId, month: selection.month, data: unpack(response) });
        setStep("done");
        router.refresh();
      }
    } catch (caught) {
      if (id === requestId.current) {
        setError(errorMessage(caught));
        if (caught instanceof BrowserApiError && caught.status === 400) setPreview(null);
      }
    } finally {
      if (id === requestId.current) setPending(null);
    }
  }

  const activeCards = cards.filter((card) => card.status === "active");
  const latestPreview = preview && preview.allocationsJson === JSON.stringify(allocations) && preview.cardId === cardId && preview.month === month && preview.file === file;
  const ready = latestPreview && !pending && invoiceCount(preview.data) === 0 && ignoredPayments(preview.data) === 0 && preview.data.new_count > 0;
  const affectedMonths = outcome ? [...new Set(outcome.data.rows.filter((row) => row.status === "imported").map((row) => row.kind === "payment" ? row.invoice_payment_month : outcome.month).filter((value): value is string => Boolean(value)))].sort() : [];

  return <>
    <button className="invoice-import-trigger" type="button" onClick={start}><span className="invoice-import-icon"><FileSpreadsheet size={26} /></span><span className="invoice-import-copy"><strong>Importar planilha da fatura</strong><small>Envie o CSV do Nubank e confira os lançamentos antes de salvar.</small></span><ArrowUpRight className="invoice-import-arrow" size={20} /></button>
    <dialog ref={dialogRef} className="invoice-import-dialog" onClose={() => setOpen(false)} onCancel={(event) => { if (pending === "confirm") event.preventDefault(); }} aria-labelledby="invoice-import-title">
      <div className="invoice-import-dialog-header"><div><span className="section-kicker">IMPORTAÇÃO NUBANK</span><h2 id="invoice-import-title">{step === "file" ? "Selecione o arquivo" : step === "review" ? "Revise a importação" : "Importação concluída"}</h2></div><button className="invoice-import-close" type="button" aria-label="Fechar" onClick={() => setOpen(false)} disabled={pending === "confirm"}><X size={19} /></button></div>
      <ol className="invoice-import-steps"><li className={step === "file" ? "active" : "done"}>1. Arquivo</li><li className={step === "review" ? "active" : step === "done" ? "done" : ""}>2. Revisão</li><li className={step === "done" ? "active" : ""}>3. Resultado</li></ol>
      <div className="invoice-import-dialog-body">
        {step === "file" && <><p className="section-description">O mês escolhido aqui recebe as compras. Você poderá escolher outra fatura para cada “Pagamento recebido” na próxima etapa. A prévia não grava dados.</p><form className="resource-form" onSubmit={submitFile}>
          <label className="native-field"><span>Cartão</span><select required value={cardId} onChange={(event) => { invalidate(); setCardId(event.target.value); }}><option value="" disabled>Selecione o cartão</option>{activeCards.map((card) => <option value={card.id} key={card.id}>{card.institution.name} · {card.name}</option>)}</select></label>
          <label className="native-field"><span>Fatura das compras</span><input type="month" required value={month} onChange={(event) => { invalidate(); setMonth(event.target.value); }} /></label>
          <label className="native-field"><span>Arquivo CSV</span><input ref={fileInputRef} type="file" accept=".csv,text/csv" required onChange={changeFile} /><small>Até 2 MiB e 5.000 linhas · colunas date,title,amount · data AAAA-MM-DD · valor com vírgula decimal</small></label>
          <Button type="submit" disabled={!cardId || !month || !file || file.size > MAX_FILE_BYTES}><ReceiptText size={16} /> Revisar CSV</Button>
        </form></>}
        {step === "review" && <><div className="invoice-import-review-heading"><p>Compras na fatura <strong>{month}</strong>. Escolha a fatura que cada pagamento vai abater.</p><Button type="button" variant="ghost" size="sm" onClick={() => { invalidate(); setFile(null); }} disabled={pending === "confirm"}>Trocar arquivo</Button></div>
          {preview && <><Totals data={preview.data} /><div className="invoice-import-counts">{invoiceCount(preview.data)} pagamento(s) sem fatura · {preview.data.existing_count} existente(s) · {preview.data.skipped_count} ignorado(s)</div><div className="invoice-import-rows">{preview.data.rows.map((row) => <div className="invoice-import-row" key={row.line}>
            <div className="invoice-import-row-main"><strong>{row.title}</strong><span>Linha {row.line} · {formatDate(row.date)} · {money(row.amount_cents)}</span>{row.kind === "payment" && <label className="native-field"><span>Fatura deste pagamento</span><input type="month" aria-label={`Fatura do pagamento na linha ${row.line}`} value={allocations[String(row.line)] ?? ""} onChange={(event) => allocate(row.line, event.target.value)} disabled={pending === "confirm"} /></label>}{rowExplanation(row) && <small>{rowExplanation(row)}</small>}</div>
            <span className={`invoice-import-status status-${row.status}`}>{statusLabel(row.status)}</span>
          </div>)}</div></>}
          {pending === "preview" && <p className="chart-message" role="status"><Spinner size="sm" /> Atualizando prévia e token...</p>}
          {!pending && error && file && <Button type="button" variant="ghost" onClick={() => void loadPreview(allocations, cardId, month, file)}>Tentar prévia novamente</Button>}
          {!pending && preview && invoiceCount(preview.data) > 0 && <p className="invoice-import-help">Escolha a fatura de todos os pagamentos para confirmar.</p>}
          {!pending && preview && ignoredPayments(preview.data) > 0 && <Alert color="error">{ignoredPayments(preview.data)} pagamento(s) recebido(s) foram ignorados pela API. A confirmação ficaria incompleta. Atualize a API e gere uma nova prévia antes de importar.</Alert>}
          {!pending && preview && invoiceCount(preview.data) === 0 && preview.data.new_count === 0 && <p className="invoice-import-help">Nenhum lançamento novo neste arquivo.</p>}
          <Button type="button" onClick={confirm} disabled={!ready}>{pending === "confirm" ? <Spinner size="sm" /> : <FileUp size={15} />} {pending === "confirm" ? "Importando..." : "Confirmar importação"}</Button>
        </>}
        {step === "done" && outcome && <><p className="section-description">Os lançamentos gravados estão marcados como “Importado”. As faturas afetadas e o planejamento serão carregados com os novos valores ao abrir essas telas.</p><Totals data={outcome.data} confirmed /><div className="invoice-import-counts">{outcome.data.existing_count} já existente(s) · {outcome.data.skipped_count} ignorado(s)</div><div className="invoice-import-rows">{outcome.data.rows.map((row) => <div className="invoice-import-row" key={row.line}><div className="invoice-import-row-main"><strong>{row.title}</strong><span>Linha {row.line} · {formatDate(row.date)} · {money(row.amount_cents)}</span>{row.kind === "payment" && row.invoice_payment_month && <small>Fatura {row.invoice_payment_month}</small>}{rowExplanation(row) && <small>{rowExplanation(row)}</small>}</div><span className={`invoice-import-status status-${row.status}`}>{statusLabel(row.status)}</span></div>)}</div><div className="invoice-import-links">{affectedMonths.map((value) => <Link key={value} className="small-link" href={`/invoices/${outcome.cardId}/${value}`} onClick={() => setOpen(false)}>Ver fatura {value}</Link>)}<Link className="small-link" href="/planning" onClick={() => setOpen(false)}>Ver planejamento atualizado</Link></div></>}
        {error && <Alert color="error">{error}</Alert>}
      </div>
    </dialog>
  </>;
}
