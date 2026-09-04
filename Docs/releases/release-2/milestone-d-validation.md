# Validação do Marco D — Vínculo e alocação

Data: 03/09/2026

## Resultado

O Marco D foi implementado e integrado localmente em `develop`. A API agora
permite alterar a forma de pagamento por vigência e registrar movimentações
excepcionais de ocorrências com histórico auditável.

## Evidências

- `POST` e `GET` de histórico de métodos de pagamento autenticados;
- períodos anteriores são encerrados na mesma transação da nova vigência;
- ausência de período explícito mantém o fallback `direct`;
- renda, cartão de outro usuário, cartão arquivado e sobreposição são
  rejeitados;
- movimentação deriva a origem atual, impede o mesmo destino e mantém eventos
  append-only;
- duas movimentações sucessivas usam a última alocação como nova origem;
- testes de compilação dos módulos T08/T09 e HTTP aprovados.

Nenhum ambiente Supabase Cloud foi alterado.
