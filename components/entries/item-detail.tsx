"use client";

import { Input } from "@lucaslfs2004/luke-ui";
import { ArchiveAction, MutationForm } from "@/components/resource-mutation";
import { SubscriptionFields } from "@/components/entries/subscription-fields";
import { changedFields, centsField, optionalText, periodFields, textField } from "@/lib/api/requests";
import type { FinancialItem } from "@/lib/api/types";

export function ItemActions({ item, month }: { item: FinancialItem; month: string }) {
  const path = `/financial-items/${encodeURIComponent(item.id)}`;
  const disabled = item.status !== "active";
  return <><div className="planning-settings-grid"><section className="debt-action-panel"><h2>Dados do registro</h2><MutationForm path={path} method="PATCH" disabled={disabled} payload={(form) => changedFields({ name: item.name, description: item.description, invoice_match_title: item.invoice_match_title ?? null, renewal_day: item.renewal_day ?? null }, { name: textField(form, "name"), description: optionalText(form, "description"), ...(item.kind === "fixed_expense" ? { invoice_match_title: optionalText(form, "invoice_match_title"), renewal_day: form.get("renewal_day") ? Number(form.get("renewal_day")) : null } : {}) })}><Input name="name" label="Nome" defaultValue={item.name} maxLength={120} required /><Input name="description" label="Descrição" defaultValue={item.description ?? ""} maxLength={1000} />{item.kind === "fixed_expense" && <SubscriptionFields title={item.invoice_match_title} renewalDay={item.renewal_day} />}</MutationForm></section><section className="debt-action-panel"><h2>Alterar valor por vigência</h2><MutationForm path={`${path}/changes`} disabled={disabled} payload={(form) => ({ ...periodFields(form), amount_cents: centsField(form, "amount"), cash_month_offset: Number(form.get("cash_month_offset")) })}><div className="form-grid-two"><Input name="effective_from" type="month" label="A partir de" defaultValue={month} required /><Input name="end_month" type="month" label="Até (opcional)" /></div><Input name="amount" type="number" min="0" step="0.01" label={`Valor (${item.currency_code})`} required /><Input name="cash_month_offset" type="number" min="0" max="12" step="1" label="Deslocamento para o caixa" defaultValue="0" required /><Input name="context" label="Contexto" maxLength={500} /></MutationForm></section></div>{!disabled && <ArchiveAction path={`${path}/archive`} name={item.name} effect="Sem competência informada, apenas o status é arquivado e o histórico é preservado. Informe um mês para interromper as ocorrências a partir dele." allowMonth />}</>;
}
