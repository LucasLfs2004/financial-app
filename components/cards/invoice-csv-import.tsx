"use client";

import { useEffect, useRef, useState, type ChangeEvent, type FormEvent } from "react";
import { Alert, Button, Spinner } from "@lucaslfs2004/luke-ui";
import { ArrowLeft, ArrowRight, ArrowUpRight, Check, FileSpreadsheet, FileUp, X } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { browserApiFetch, BrowserApiError } from "@/lib/api/browser-api";
import type { CreditCard, InvoiceImportPreview, InvoiceImportProjectedInstallment, InvoiceImportResult, InvoiceImportRow } from "@/lib/api/types";
import { currentMonth } from "@/lib/debt-month";
import { invoiceCount, subscriptionMatchCount, subscriptionMatches, ignoredPayments, newProjectionCount, hasChanges, importTotals, csvTotals, duplicateCount, decisionRows, unsupportedCredits } from "@/lib/invoice-import";
import { formatDate } from "@/lib/format-date";

const MAX_FILE_BYTES = 2 * 1024 * 1024;
type ImportResponse<T> = T | { data: T };
type Allocations = Record<string, string>;
type PreviewState = { data: InvoiceImportPreview; allocationsJson: string; decisionsJson: string; cardId: string; month: string; file: File };
type ImportOutcome = { cardId: string; month: string; data: InvoiceImportResult };

function unpack<T extends object>(response: ImportResponse<T>): T {
  return "data" in response ? response.data : response;
}

function money(cents: number, currencyCode = "BRL") {
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency: currencyCode }).format(cents / 100);
}

function errorMessage(error: unknown) {
  if (error instanceof BrowserApiError) {
    switch (error.status) {
      case 400: return "Arquivo, mês de pagamento ou token inválido. Confira as escolhas e gere uma nova prévia.";
      case 404: return "Cartão não encontrado. Atualize a página e selecione outro cartão.";
      case 409: return error.message;
      case 422: return error.message;
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
    case "needs_subscription_match": return "Revisar assinatura";
    case "needs_duplicate_decision": return "Possível duplicata";
  }
}

function rowExplanation(row: InvoiceImportRow, currencyCode: string) {
  if (row.reason === "duplicate_skipped") return "Você escolheu não importar esta linha. O registro existente será preservado.";
  if (row.status === "needs_duplicate_decision") return "Há um lançamento semelhante nesta fatura. Revise na etapa Decisões.";
  if (row.kind === "credit") return "Crédito ou estorno: reduz as cobranças da fatura e não é um pagamento.";
  if (row.reason === "ambiguous_subscription") return "Mais de uma cobrança ou assinatura corresponde a esta descrição. Revise o cadastro em Receitas e despesas e gere uma nova prévia.";
  if (row.reason === "subscription_match") return row.planned_amount_cents != null
    ? `Substitui a previsão de ${money(row.planned_amount_cents, currencyCode)} da assinatura em ${row.matched_reference_month}. A cobrança de ${money(row.amount_cents, currencyCode)} será contada uma única vez.`
    : `Cobrança vinculada à assinatura em ${row.matched_reference_month}. A previsão já foi substituída, sem duplicação.`;
  if (row.reason === "projected_installment") return "Esta cobrança substitui uma parcela já projetada, sem duplicar o valor.";
  if (row.reason === "already_imported_to_other_invoice") return `Já importado na fatura ${row.invoice_payment_month ?? "original"}. Não será lançado novamente.`;
  if (row.status === "skipped" && row.reason === "non_expense" && /^pagamento recebido/i.test(row.title.trim())) return "Este pagamento foi ignorado pela API e não entrará na fatura.";
  if (row.reason === "non_expense") return row.amount_cents < 0 ? "Crédito não suportado por esta versão da API. A importação será bloqueada." : "Valor zero — não altera a fatura.";
  return row.reason?.replaceAll("_", " ") ?? null;
}

function Totals({ data, confirmed = false, currencyCode }: { data: InvoiceImportPreview; confirmed?: boolean; currencyCode: string }) {
  const totals = importTotals(data, confirmed);
  return <div className="invoice-import-metrics">
    <span><strong>{money(totals.charges, currencyCode)}</strong>{confirmed ? "compras importadas" : "novas compras"}</span>
    <span><strong>{money(totals.credits, currencyCode)}</strong>{confirmed ? "créditos importados" : "novos créditos"}</span>
    <span><strong>{money(totals.payments, currencyCode)}</strong>{confirmed ? "pagamentos importados" : "novos pagamentos"}</span>
    <span><strong>{money(totals.net, currencyCode)}</strong>{confirmed ? "saldo dos lançamentos importados" : "saldo dos novos lançamentos"}</span>
    {subscriptionMatches(data).length > 0 && <span><strong>{money(subscriptionMatches(data).reduce((total, row) => total + (row.planned_amount_cents ?? 0), 0), currencyCode)}</strong>previsões substituídas</span>}
    <span><strong>{confirmed ? (data as InvoiceImportResult).imported_count ?? data.new_count : data.new_count}</strong>{confirmed ? "lançados" : "novos lançamentos"}</span>
    <span><strong>{newProjectionCount(data)}</strong>{confirmed ? "parcelas futuras projetadas" : "novas parcelas futuras"}</span>
    <span><strong>{data.existing_count}</strong>existentes</span>
    <span><strong>{data.skipped_count}</strong>ignorados</span>
  </div>;
}

function ProjectedRows({ rows, confirmed = false, currencyCode }: { rows: InvoiceImportProjectedInstallment[]; confirmed?: boolean; currencyCode: string }) {
  if (rows.length === 0) return null;
  return <><p className="invoice-import-help">Parcelas futuras por fatura. As já existentes não serão duplicadas.</p><div className="invoice-import-rows">{rows.map((row) => <div className="invoice-import-row" key={`${row.source_line}-${row.invoice_payment_month}-${row.installment_number}`}>
    <div className="invoice-import-row-main"><strong>{row.title} · {row.installment_number}/{row.installments_total}</strong><span>Projeção de parcela importada · Fatura {row.invoice_payment_month} · {money(row.amount_cents, currencyCode)} · origem na linha {row.source_line}</span></div>
    <span className={`invoice-import-status status-${row.status}`}>{row.status === "existing" ? "Já existente" : confirmed ? "Projetada" : "Nova projeção"}</span>
  </div>)}</div></>;
}

export function InvoiceCsvImport({ cards, initialCardId = "", defaultMonth = currentMonth(), currencyCode = "BRL", compact = false }: { cards: CreditCard[]; initialCardId?: string; defaultMonth?: string; currencyCode?: string; compact?: boolean }) {
  const router = useRouter();
  const dialogRef = useRef<HTMLDialogElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const requestId = useRef(0);
  const previewAbort = useRef<AbortController | null>(null);
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState<"file" | "overview" | "rows" | "decisions" | "done">("file");
  const bodyRef = useRef<HTMLDivElement>(null);
  const [cardId, setCardId] = useState(initialCardId);
  const [month, setMonth] = useState(defaultMonth);
  const [file, setFile] = useState<File | null>(null);
  const [allocations, setAllocations] = useState<Allocations>({});
  const [decisions, setDecisions] = useState<Allocations>({});
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

  useEffect(() => () => previewAbort.current?.abort(), []);

  useEffect(() => {
    bodyRef.current?.scrollTo({ top: 0 });
    bodyRef.current?.focus();
  }, [step]);

  useEffect(() => {
    if (!open) return;
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => { document.body.style.overflow = previous; };
  }, [open]);

  function invalidate() {
    requestId.current += 1;
    previewAbort.current?.abort();
    setPreview(null);
    setOutcome(null);
    setAllocations({});
    setDecisions({});
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

  async function loadPreview(nextAllocations: Allocations, selectedCard: string, selectedMonth: string, selectedFile: File, nextDecisions: Allocations = decisions) {
    const id = ++requestId.current;
    previewAbort.current?.abort();
    const controller = new AbortController();
    previewAbort.current = controller;
    const allocationsJson = JSON.stringify(nextAllocations);
    const decisionsJson = JSON.stringify(nextDecisions);
    setPending("preview");
    setError("");
    const form = new FormData();
    form.append("file", selectedFile);
    form.append("payment_allocations", allocationsJson);
    form.append("duplicate_decisions", decisionsJson);
    try {
      const response = await browserApiFetch<ImportResponse<InvoiceImportPreview>>(`/credit-cards/${encodeURIComponent(selectedCard)}/invoices/${selectedMonth}/imports/preview`, { method: "POST", body: form, signal: controller.signal, successMessage: false });
      if (id === requestId.current) {
        setPreview({ data: unpack(response), allocationsJson, decisionsJson, cardId: selectedCard, month: selectedMonth, file: selectedFile });
        if (step === "file") setStep("overview");
      }
    } catch (caught) {
      if (id === requestId.current && !(caught instanceof DOMException && caught.name === "AbortError")) {
        setError(errorMessage(caught));
        if (caught instanceof BrowserApiError && caught.status === 400) { setPreview(null); setDecisions({}); }
      }
    } finally {
      if (id === requestId.current) setPending(null);
    }
  }

  function submitFile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!cardId || !/^\d{4}-(0[1-9]|1[0-2])$/.test(month) || !file) return;
    if (file.size > MAX_FILE_BYTES) { setError("O arquivo deve ter no máximo 2 MiB."); return; }
    if (latestPreview && !error) { setStep("overview"); return; }
    void loadPreview(allocations, cardId, month, file, decisions);
  }

  function allocate(line: number, invoiceMonth: string) {
    if (!cardId || !file) return;
    const next = { ...allocations };
    if (invoiceMonth) next[String(line)] = invoiceMonth;
    else delete next[String(line)];
    setAllocations(next);
    void loadPreview(next, cardId, month, file);
  }

  function decide(line: number, value: string) {
    if (!file) return;
    const next = { ...decisions };
    if (value) next[String(line)] = value;
    else delete next[String(line)];
    setDecisions(next);
    void loadPreview(allocations, cardId, month, file, next);
  }

  async function confirm() {
    if (!preview || !file || !ready) return;
    const selection = preview;
    const id = ++requestId.current;
    setPending("confirm");
    setError("");
    const form = new FormData();
    form.append("file", selection.file);
    form.append("payment_allocations", selection.allocationsJson);
    form.append("duplicate_decisions", selection.decisionsJson);
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
        setPreview(null);
      }
    } finally {
      if (id === requestId.current) setPending(null);
    }
  }

  const activeCards = cards.filter((card) => card.status === "active");
  const latestPreview = preview && preview.allocationsJson === JSON.stringify(allocations) && preview.decisionsJson === JSON.stringify(decisions) && preview.cardId === cardId && preview.month === month && preview.file === file;
  const supported = preview && (preview.data.review_version ?? 0) >= 2;
  const ready = latestPreview && supported && !pending && !error && invoiceCount(preview.data) === 0 && duplicateCount(preview.data) === 0 && subscriptionMatchCount(preview.data) === 0 && ignoredPayments(preview.data) === 0 && unsupportedCredits(preview.data) === 0 && hasChanges(preview.data);
  const affectedMonths = outcome ? [...new Set([
    ...outcome.data.rows.filter((row) => row.status === "imported" || (row.matched_item_id && row.planned_amount_cents != null)).map((row) => row.kind === "payment" ? row.invoice_payment_month : outcome.month),
    ...(outcome.data.projected_installments ?? []).filter((row) => row.status === "imported").map((row) => row.invoice_payment_month),
  ].filter((value): value is string => Boolean(value)))].sort() : [];

  const stages = ["file", "overview", "rows", "decisions", "done"] as const;
  const stageNames = ["Arquivo", "Visão geral", "Lançamentos", "Decisões", "Resultado"];
  const stageIndex = stages.indexOf(step);
  const data = outcome?.data ?? preview?.data;
  const totals = data ? csvTotals(data) : null;
  const choices = preview ? decisionRows(preview.data) : [];
  const unresolved = preview ? invoiceCount(preview.data) + duplicateCount(preview.data) + subscriptionMatchCount(preview.data) : 0;
  const canReview = Boolean(latestPreview && !pending && !error);

  function renderRow(row: InvoiceImportRow, editable = false) {
    return <div className="invoice-import-row" key={row.line}>
      <div className="invoice-import-row-main">
        <strong>{row.title}</strong>
        <span>Linha {row.line} · {formatDate(row.date)} · {money(row.amount_cents, currencyCode)}</span>
        <span>{row.kind === "payment" ? "Pagamento" : row.kind === "credit" ? "Crédito / estorno" : "Compra"}</span>
        {rowExplanation(row, currencyCode) && <small>{rowExplanation(row, currencyCode)}</small>}
        {editable && row.kind === "payment" && <label className="native-field">
          <span>Fatura que este pagamento abate</span>
          <input type="month" aria-label={`Fatura do pagamento na linha ${row.line}`} value={allocations[String(row.line)] ?? ""} onChange={(event) => allocate(row.line, event.target.value)} disabled={Boolean(pending)} />
          {row.invoice_payment_month && row.reason === "already_imported_to_other_invoice" && <small>Já registrado em {row.invoice_payment_month}; não será movido.</small>}
        </label>}
        {editable && (row.duplicate_candidates?.length ?? 0) > 0 && <fieldset className="invoice-import-decision">
          <legend>Este lançamento já existe?</legend>
          <p>O nome e o valor coincidem com registros desta fatura. A data pode diferir em até um dia.</p>
          {row.duplicate_candidates!.map((candidate) => <p key={candidate.id}><strong>{candidate.title}</strong> · {formatDate(candidate.date)} · {money(candidate.amount_cents, currencyCode)}</p>)}
          <label><input type="radio" name={`duplicate-${row.line}`} checked={decisions[String(row.line)] === "skip"} disabled={Boolean(pending)} onChange={() => decide(row.line, "skip")} /> Não importar esta linha; preservar o existente</label>
          <label><input type="radio" name={`duplicate-${row.line}`} checked={decisions[String(row.line)] === "import"} disabled={Boolean(pending)} onChange={() => decide(row.line, "import")} /> Importar como outra compra ou crédito</label>
        </fieldset>}
        {editable && row.status === "needs_subscription_match" && <small>Revise a descrição em <Link href="/entries" target="_blank" rel="noopener noreferrer">Receitas e despesas (abre em outra aba)</Link> e atualize a prévia.</small>}
      </div>
      <span className={`invoice-import-status status-${row.status}`}>{statusLabel(row.status)}</span>
    </div>;
  }

  return <>
    <button className={`invoice-import-trigger${compact ? " finance-import-compact" : ""}`} type="button" onClick={start}><span className="invoice-import-icon"><FileSpreadsheet size={26} /></span><span className="invoice-import-copy"><strong>{compact ? "Importar fatura do Nubank" : "Importar planilha da fatura"}</strong><small>Envie o CSV do Nubank e confira os lançamentos antes de salvar.</small></span><ArrowUpRight className="invoice-import-arrow" size={20} /></button>
    <dialog ref={dialogRef} className="invoice-import-dialog" onClose={() => { previewAbort.current?.abort(); requestId.current += 1; setOpen(false); }} onCancel={(event) => { if (pending === "confirm") event.preventDefault(); }} aria-labelledby="invoice-import-title">
      <div className="invoice-import-dialog-header"><div><span className="section-kicker">IMPORTAÇÃO NUBANK</span><h2 id="invoice-import-title">{step === "done" ? "Importação concluída" : stageNames[stageIndex]}</h2></div><button className="invoice-import-close" type="button" aria-label="Fechar importação" onClick={() => setOpen(false)} disabled={pending === "confirm"}><X size={19} /></button></div>
      <ol className="invoice-import-steps" aria-label="Etapas da importação">{stages.map((stage, index) => <li key={stage} className={index === stageIndex ? "active" : index < stageIndex ? "done" : ""} aria-current={stage === step ? "step" : undefined}><span>{index + 1}.</span> {stageNames[index]}</li>)}</ol>
      <div ref={bodyRef} className="invoice-import-dialog-body" tabIndex={-1} aria-label={stageNames[stageIndex]} aria-busy={Boolean(pending)}>
        {step === "file" && <>
          <p className="section-description">Selecione o cartão, a fatura das compras e o arquivo CSV. Na próxima etapa, você verá o resumo dos valores antes de revisar os lançamentos. A prévia não salva dados.</p>
          <form id="invoice-import-file-form" className="resource-form" onSubmit={submitFile}>
            <label className="native-field"><span>Cartão</span><select required value={cardId} disabled={Boolean(pending)} onChange={(event) => { invalidate(); setCardId(event.target.value); }}><option value="" disabled>Selecione o cartão</option>{activeCards.map((card) => <option value={card.id} key={card.id}>{card.institution.name} · {card.name}</option>)}</select></label>
            <label className="native-field"><span>Fatura das compras</span><input type="month" required value={month} disabled={Boolean(pending)} onChange={(event) => { invalidate(); setMonth(event.target.value); }} /></label>
            <label className="native-field"><span>Arquivo CSV</span><input ref={fileInputRef} type="file" accept=".csv,text/csv" required={!file} disabled={Boolean(pending)} onChange={changeFile} />{file && <small>Arquivo selecionado: {file.name}</small>}<small>Até 2 MiB e 5.000 linhas · colunas date,title,amount</small></label>
          </form>
        </>}
        {step === "overview" && preview && totals && <>
            <p className="section-description">Confira os valores de <strong>{file?.name}</strong>. As compras entram na fatura <strong>{month}</strong>; pagamentos e possíveis duplicatas serão revisados na etapa Decisões.</p>
            <h3 className="invoice-import-section-title">Dados do arquivo · {preview.data.rows.length} lançamentos</h3>
            <div className="invoice-import-metrics">
              <span><strong>{money(totals.charges, currencyCode)}</strong>compras e parcelas no CSV</span>
              <span><strong>{money(totals.credits, currencyCode)}</strong>créditos e estornos no CSV</span>
              <span><strong>{money(totals.charges - totals.credits, currencyCode)}</strong>compras menos créditos</span>
              <span><strong>{money(totals.payments, currencyCode)}</strong>pagamentos para conferir</span>
              <span><strong>{preview.data.existing_count}</strong>já existentes</span>
              <span><strong>{unresolved}</strong>decisões pendentes</span>
            </div>
            <p className="invoice-import-help">Os totais acima incluem todas as linhas do arquivo. O valor a importar será calculado após suas decisões; os pagamentos podem pertencer a outras faturas.</p>
        </>}
        {step === "rows" && preview && <>
          <p className="section-description">Confira todas as linhas e os motivos de cada classificação. As escolhas ficam na próxima etapa.</p>
          <div className="invoice-import-list-scroll" tabIndex={0} role="region" aria-label="Lançamentos e parcelas futuras">
            <div className="invoice-import-rows">{preview.data.rows.map((row) => renderRow(row))}</div>
            <ProjectedRows rows={preview.data.projected_installments ?? []} currencyCode={currencyCode} />
          </div>
        </>}
        {step === "decisions" && preview && <>
          <p className="section-description">{unresolved > 0 ? `${unresolved} decisão(ões) pendente(s). Confira os pagamentos e as possíveis duplicatas antes de salvar.` : "Todas as decisões foram resolvidas. Confira o resumo antes de salvar."}</p>
          <details className="invoice-import-impact"><summary>Valores a importar</summary><Totals data={preview.data} currencyCode={currencyCode} /></details>
          {choices.length > 0 ? <div className="invoice-import-list-scroll" tabIndex={0} role="region" aria-label="Decisões da importação"><div className="invoice-import-rows">{choices.map((row) => renderRow(row, true))}</div></div> : <p className="invoice-import-help">Nenhum lançamento precisa de uma escolha adicional.</p>}
          {subscriptionMatchCount(preview.data) > 0 && <Alert color="warning">Há assinaturas ambíguas. Ajuste o cadastro e atualize a prévia para continuar.</Alert>}
          {!hasChanges(preview.data) && unresolved === 0 && <p className="invoice-import-help">Nenhum lançamento, vínculo de assinatura ou parcela futura nova neste arquivo.</p>}
        </>}
        {step === "done" && outcome && <>
          <p className="section-description">Importação salva. As linhas que você decidiu não importar foram preservadas no resumo; os registros existentes não foram removidos.</p>
          <Totals data={outcome.data} confirmed currencyCode={currencyCode} />
          <div className="invoice-import-list-scroll" tabIndex={0} role="region" aria-label="Resultado dos lançamentos"><div className="invoice-import-rows">{outcome.data.rows.map((row) => renderRow(row))}</div><ProjectedRows rows={outcome.data.projected_installments ?? []} confirmed currencyCode={currencyCode} /><div className="invoice-import-links">{affectedMonths.map((value) => <Link key={value} className="small-link" href={`/invoices/${encodeURIComponent(outcome.cardId)}/${value}`} onClick={() => setOpen(false)}>Ver fatura {value}</Link>)}<Link className="small-link" href="/planning" onClick={() => setOpen(false)}>Ver planejamento atualizado</Link></div></div>
        </>}
        {preview && step !== "done" && (!supported || unsupportedCredits(preview.data) > 0 || ignoredPayments(preview.data) > 0) && <Alert color="error">A API precisa ser atualizada para revisar duplicatas e importar créditos e pagamentos com segurança. Nenhum dado será salvo por este fluxo até a atualização.</Alert>}
        {error && <Alert color="error">{error}</Alert>}
      </div>
      <footer className="invoice-import-footer">
        <span className="invoice-import-progress" role="status">{pending === "preview" ? "Atualizando prévia…" : pending === "confirm" ? "Salvando importação…" : `Etapa ${stageIndex + 1} de ${stages.length}`}</span>
        <div className="invoice-import-navigation">
          {step !== "file" && step !== "done" && <Button type="button" variant="ghost" size="sm" disabled={Boolean(pending)} onClick={() => setStep(stages[stageIndex - 1])}><ArrowLeft size={16} aria-hidden="true" /> Voltar</Button>}
          {step === "file" && <Button form="invoice-import-file-form" type="submit" size="sm" disabled={Boolean(pending) || !cardId || !month || !file || file.size > MAX_FILE_BYTES}>{pending ? <Spinner size="sm" /> : <ArrowRight size={16} />} Avançar</Button>}
          {step === "overview" && preview && <Button type="button" size="sm" disabled={!canReview} onClick={() => setStep("rows")}>Avançar <ArrowRight size={16} aria-hidden="true" /></Button>}
          {step === "rows" && <Button type="button" size="sm" disabled={!canReview} onClick={() => setStep("decisions")}>Avançar <ArrowRight size={16} aria-hidden="true" /></Button>}
          {step === "decisions" && <Button type="button" size="sm" onClick={confirm} disabled={!ready}>{pending === "confirm" ? <Spinner size="sm" /> : <FileUp size={16} />} Importar</Button>}
          {step === "done" && <Button type="button" size="sm" onClick={() => setOpen(false)}><Check size={16} aria-hidden="true" /> Concluir</Button>}
          {error && step !== "done" && file && <Button type="button" size="sm" variant="ghost" disabled={Boolean(pending)} onClick={() => void loadPreview(allocations, cardId, month, file)}>Tentar novamente</Button>}
          {step === "decisions" && preview && subscriptionMatchCount(preview.data) > 0 && <Button type="button" size="sm" variant="ghost" disabled={Boolean(pending)} onClick={() => file && void loadPreview(allocations, cardId, month, file)}>Atualizar prévia</Button>}
        </div>
      </footer>
    </dialog>
  </>;
}
