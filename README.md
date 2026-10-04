# Projeção Financeira

Aplicação Next.js para acompanhar planejamento, cartões, faturas e dívidas.

## Visualização por competência

As projeções e o resumo mensal usam a base **Caixa** por padrão. A opção **Competência** fica oculta enquanto `NEXT_PUBLIC_ENABLE_REFERENCE_BASIS` não estiver definida como `true`.

Para reativá-la, configure a variável no ambiente do app e faça um novo build:

```env
NEXT_PUBLIC_ENABLE_REFERENCE_BASIS=true
```

A flag controla a opção nas telas; os dados e a API de competência continuam disponíveis para uma reativação futura.

## Mês de pagamento das dívidas

No cadastro de uma dívida parcelada com pagamento direto, o campo **Deslocamento para o caixa (meses)** começa em `1` e pode ser alterado em **Mais detalhes**. Ele corresponde a `cash_month_offset` enviado à API. Com `1`, uma parcela referente a setembro é projetada no caixa de outubro: se o salário cai em 1º de outubro, é esse salário que cobre a parcela. O salário de novembro já não inclui esse pagamento.

Para dívidas pagas por cartão, o mês da fatura segue a configuração do cartão (`payment_month_offset`), cujo valor inicial no cadastro do cartão é `1`. Alterar o deslocamento do pagamento direto não muda a configuração do cartão.
