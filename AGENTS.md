# Instruções do projeto

## Versionamento e releases

Para qualquer operação ou orientação envolvendo Git, branches, commits, pull
requests, releases, tags, hotfixes ou governança do repositório, use
obrigatoriamente a skill `$git-release-flow`.

O fluxo completo e as decisões específicas deste projeto estão documentados em
[`Docs/git-release-flow.md`](Docs/git-release-flow.md). Em caso de divergência
com uma convenção Git genérica, prevalece esse fluxo específico do projeto.

Regras essenciais:

- `main` representa produção;
- `develop` representa integração e staging;
- alterações entram por pull request como fluxo padrão;
- `main` nunca recebe push direto: alterações entram somente por pull request;
- `develop` pode receber merge local e push direto exclusivamente quando o
  usuário autorizar de forma clara no prompt atual; essa exceção nunca deve
  ser inferida, reutilizada automaticamente ou aplicada a `main`;
- branches novas seguem `feature/*`, `fix/*`, `hotfix/*`, `chore/*`, `docs/*`,
  `refactor/*` ou, quando realmente necessário, `release/*`;
- commits e títulos de PR seguem Conventional Commits;
- releases de produção recebem tag anotada `vMAJOR.MINOR.PATCH`.

Não renomeie, apague nem reescreva branches ou commits já publicados para
corrigir convenções retroativamente. Registre a exceção e aplique o padrão a
partir do próximo trabalho.
