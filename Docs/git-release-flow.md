# Fluxo Git e releases

## Objetivo

Este projeto adota o `$git-release-flow` como padrão obrigatório de
versionamento. O fluxo separa integração, homologação e produção sem manter
branches longas para cada funcionalidade.

## Branches permanentes

| Branch | Responsabilidade | Destino de deploy |
| --- | --- | --- |
| `main` | código aprovado que representa produção | produção |
| `develop` | integração das próximas entregas | staging |

`main` não aceita push direto. Toda alteração de produção deve entrar por pull
request com os checks aplicáveis aprovados.

`develop` usa pull request como fluxo padrão. Excepcionalmente, o agente pode
realizar merge local e push direto para `develop` quando o usuário autorizar
isso de forma explícita no prompt atual. Essa autorização é pontual: nunca é
inferida, reaproveitada em outra operação ou aplicada a `main`.

## Branches de trabalho

Toda branch usa letras minúsculas e descrição em kebab-case:

- `feature/<ticket>-<descricao>` para funcionalidades e trabalho técnico
  normal;
- `fix/<ticket>-<descricao>` para correções que seguem o ciclo normal;
- `hotfix/<ticket>-<descricao>` para correções urgentes originadas de `main`;
- `chore/<ticket>-<descricao>` para manutenção, dependências e CI;
- `docs/<ticket>-<descricao>` para documentação;
- `refactor/<ticket>-<descricao>` para reestruturação sem mudança intencional
  de comportamento;
- `release/<versao>` somente quando uma estabilização paralela for necessária.

Quando não houver ticket, use o identificador da task da release, por exemplo:

```text
feature/r1-t05-saving-periods
fix/r1-t14-month-validation
docs/r1-release-notes
```

Branches normais partem de `develop`. Hotfixes partem de `main` e precisam ser
sincronizados de volta em `develop` depois da publicação.

## Commits

Use Conventional Commits:

```text
<tipo>(<escopo>)?!: <descrição imperativa>
```

Tipos aceitos:

- `feat`, `fix`, `hotfix`, `refactor`, `perf`, `test`;
- `docs`, `chore`, `build`, `ci`, `revert`.

Exemplos para este projeto:

```text
feat(database): add saving periods migration
fix(planning): reject months outside plan horizon
test(database): cover overlapping financial periods
docs(release): document release 1 acceptance criteria
```

O assunto deve ser imperativo, conciso, sem ponto final e preferencialmente com
até 72 caracteres. Um commit deve representar uma única mudança lógica.

Breaking changes usam `!` e o rodapé `BREAKING CHANGE:` com impacto, migração e
rollback documentados.

## Pull requests

Fluxo normal:

1. atualizar `develop`;
2. criar uma branch curta a partir de `develop`;
3. criar commits coerentes;
4. abrir PR para `develop`;
5. validar CI, revisão, migrations e critérios de aceite;
6. realizar squash merge, mantendo o título do PR como Conventional Commit.

Quando houver autorização explícita para a exceção de `develop`, o agente deve
antes atualizar a referência remota, validar a ausência de conflitos, executar
os checks aplicáveis e registrar no handoff que o merge local foi feito sob
essa autorização. A exceção não elimina as validações técnicas.

PRs de release partem de `develop` para `main`. Devem informar impacto, riscos,
evidências de teste, mudanças de banco ou configuração e rollback.

Política planejada:

- um approval para PR em `develop`;
- dois approvals ou aprovação do owner para PR em `main`;
- conversas resolvidas e ausência de conflitos;
- checks obrigatórios no commit mais recente;
- squash merge como estratégia padrão.

Enquanto houver somente um mantenedor, o approval pode ser dispensado pela
configuração do GitHub, mas PR e checks continuam obrigatórios. A exigência de
approval deve ser ativada quando houver outro revisor disponível.

## Checks mínimos

Antes de mergear:

```bash
go fmt ./...
go vet ./...
go test ./...
supabase db reset --local --yes
supabase test db supabase/tests --local
supabase db lint --local --level warning
```

Além disso, a CI deve incluir build, análise de segurança e validação do
OpenAPI. Migrations precisam ser compatíveis com o caminho de deploy e conter
nota de rollback no PR.

## Releases

Produção usa Semantic Versioning e tags anotadas:

```text
vMAJOR.MINOR.PATCH
```

- `MAJOR`: mudança incompatível de API ou comportamento;
- `MINOR`: funcionalidade compatível;
- `PATCH`: correção compatível ou segurança.

O commit ou artefato validado em staging deve ser promovido para produção sem
ser reconstruído ou alterado manualmente. Uma tag publicada nunca deve ser
movida ou reutilizada.

## Bootstrap deste repositório

A branch `feat/release-1-foundation` foi publicada antes da adoção deste fluxo e
permanece como exceção histórica. Ela não será renomeada, apagada ou reescrita.

Para implantar o fluxo pela primeira vez:

1. revisar e promover o baseline atual para `main`;
2. criar `develop` a partir desse mesmo baseline aprovado;
3. proteger `main` e `develop` no GitHub;
4. iniciar a próxima task em `feature/r1-t05-saving-periods`, partindo de
   `develop`;
5. aplicar o fluxo padrão a partir desse ponto.

Essa etapa de bootstrap é uma exceção histórica. Depois dela, `main` não aceita
push direto e `develop` segue a política de exceção explícita documentada acima.
