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
- o agente pode criar branches, preparar staging, criar commits, realizar
  merges locais, criar tags e executar pushes sem solicitar autorização a
  cada operação, desde que respeite este fluxo e conclua os checks aplicáveis;
- todo gerenciamento do repositório e do versionamento pelo agente deve usar
  exclusivamente o executável `git` pela linha de comando local;
- o agente não deve usar navegador, interface gráfica, `gh`, APIs, MCPs ou
  conectores do provedor Git para administrar o repositório;
- como pull requests e proteções de branch não fazem parte do protocolo Git,
  o agente apenas prepara e publica as branches necessárias; essas operações
  do provedor ficam a cargo do mantenedor ou de automação externa;
- branches novas seguem `feature/*`, `fix/*`, `hotfix/*`, `chore/*`, `docs/*`,
  `refactor/*` ou, quando realmente necessário, `release/*`;
- commits e títulos de PR seguem Conventional Commits;
- releases de produção recebem tag anotada `vMAJOR.MINOR.PATCH`.

Não renomeie, apague nem reescreva branches ou commits já publicados para
corrigir convenções retroativamente. Registre a exceção e aplique o padrão a
partir do próximo trabalho.
