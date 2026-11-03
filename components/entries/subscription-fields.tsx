"use client";

import { Input } from "@lucaslfs2004/luke-ui";

export function SubscriptionFields({ title, renewalDay }: { title?: string | null; renewalDay?: number | null }) {
  return <>
    <Input name="renewal_day" type="number" min="1" max="31" step="1" label="Dia de renovação · opcional" defaultValue={renewalDay ?? ""} placeholder="Ex.: 28" />
    <Input name="invoice_match_title" label="Descrição na fatura · opcional" defaultValue={title ?? ""} placeholder="Ex.: Totalpass" maxLength={120} />
    <p className="resource-hint">Copie a descrição completa da cobrança no CSV do Nubank. Ao importar a fatura do cartão vinculado, a cobrança substitui a previsão daquele mês sem duplicar o valor. O dia de renovação é uma previsão nominal.</p>
  </>;
}
