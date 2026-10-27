import { isMonth } from "../debt-month";

export function textField(form: FormData, key: string) { return String(form.get(key) ?? "").trim(); }
export function optionalText(form: FormData, key: string) { return textField(form, key) || null; }
export function centsField(form: FormData, key: string) {
  const value = Number(textField(form, key).replace(",", "."));
  const cents = Math.round(value * 100);
  if (!Number.isFinite(value) || value < 0 || !Number.isSafeInteger(cents)) throw new Error("Informe um valor válido.");
  return cents;
}
export function periodFields(form: FormData) {
  const effective_from = textField(form, "effective_from");
  const end_month = optionalText(form, "end_month");
  if (!isMonth(effective_from) || (end_month && (!isMonth(end_month) || end_month < effective_from))) throw new Error("Confira o início e o término da vigência.");
  return { effective_from, end_month, context: optionalText(form, "context") };
}
export function paymentChange(form: FormData) {
  const method = textField(form, "method");
  if (method !== "direct" && method !== "credit_card") throw new Error("Selecione o método de pagamento.");
  const credit_card_id = method === "credit_card" ? optionalText(form, "credit_card_id") : null;
  if (method === "credit_card" && !credit_card_id) throw new Error("Selecione o cartão.");
  return { ...periodFields(form), method, credit_card_id };
}
export function changedFields(original: Record<string, unknown>, values: Record<string, unknown>) {
  return Object.fromEntries(Object.entries(values).filter(([key, value]) => value !== original[key]));
}
