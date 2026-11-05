---
name: pipeline
description: Leva uma issue aprovada (ready-for-agent) do Gophers Life até o merge no develop, chamando os agentes documentador, testador, programador e revisor na ordem e cuidando de rótulos, branch, worktree e PR. Use quando o mantenedor pedir para rodar, continuar ou retomar o fluxo de uma issue, por exemplo "/pipeline 4" ou "roda a issue 4".
---

# /pipeline <issue>

Você é o orquestrador. Você dispara os agentes, troca os rótulos, cuida do git e repassa ponteiros. Código, testes, documentação e opinião sobre o conteúdo ficam com os agentes ([ADR-0003](../../../docs/adr/0003-fluxo-de-quatro-agentes.md)).

Cada agente começa sem contexto e recebe só a mensagem do passo, preenchida com ponteiros. Dispare um agente por vez e encerre o turno com uma linha de status (ex.: "#4 em etapa:testes, testador trabalhando"). Ao receber a notificação de que ele terminou, leia o relatório e siga para o próximo passo.

Use o `effort` do frontmatter de cada agente, a não ser que o passo peça outro.

## 0. Preparar

1. `gh issue view N --json number,title,state,labels`. Siga só se a issue estiver aberta e com `ready-for-agent`. Se estiver com `needs-triage` ou `needs-info`, pare e diga ao mantenedor o que falta.
2. Defina os nomes, com o slug feito do título em minúsculas, sem acentos e com hífens:
   - `<id>`: `N-<slug>`;
   - branch: `feature/<id>`;
   - worktree: `.claude/worktrees/<id>`;
   - `<pasta>`: o caminho absoluto da worktree.
3. Retomada: se a issue já tem um rótulo `etapa:*`, recomece por essa etapa. Confira `git worktree list` e `git branch -a`, e reaproveite o branch, a worktree e o PR que já existirem.
4. Se a worktree não existe:
   - rode `git fetch origin`;
   - crie a worktree: `git worktree add .claude/worktrees/<id> -b feature/<id> origin/develop`. Se o branch já existir, crie sem o `-b`;
   - publique o branch: `git -C .claude/worktrees/<id> push -u origin feature/<id>`.
5. Entre na worktree com EnterWorktree (`path`). Os agentes trabalham nela, e a pasta do mantenedor fica intocada.

Ao trocar de etapa, sempre use `gh issue edit N --add-label <nova> --remove-label <anterior>`.

## 1. Spec (`etapa:spec`)

Se a issue não tem delta de spec (um refactor puro, por exemplo), pule para o passo 2. Senão, dispare o `documentador`:

> Tarefa: spec. Issue #N. Branch feature/<id>, pasta <pasta>.

## 2. Testes (`etapa:testes`)

Dispare o `testador`:

> Issue #N. Branch feature/<id>, pasta <pasta>.

Se ele reportar que a spec é ambígua, escale.

## 3. Código (`etapa:código`)

Dispare o `programador`:

> Issue #N. Branch feature/<id>, pasta <pasta>.

- Se ele contestar um teste, devolva ao `testador` com o trecho exato do relatório: "Issue #N. Branch ..., pasta .... Teste contestado pelo programador: <trecho>". Depois, dispare o programador de novo.
- Se ele terminar sem os testes verdes, dispare-o mais uma vez com `effort: high`.
- Se ainda assim não ficarem verdes, escale.

## 4. Conferência (`etapa:docs`)

Dispare o `documentador` com `effort: low`:

> Tarefa: conferência. Issue #N, PR #P. Branch feature/<id>, pasta <pasta>.

## 5. Revisão (`etapa:revisão`)

Dispare o `revisor`:

> PR #P, issue #N. Branch feature/<id>, pasta <pasta>.

Depois, leia a primeira linha do último comentário do revisor no PR.

- `Veredito: aprovado`: siga para o passo 6.
- `Veredito: mudanças pedidas`:
  1. Comente na issue o número da rodada ("revisão 2 de 3").
  2. Mande cada achado bloqueante ao dono dele, com o link do comentário:
     - spec e documentação: o `documentador`, com a tarefa conferência;
     - testes: o `testador`;
     - código: o `programador`, com `effort: high`.
  3. Troque o rótulo para a etapa do dono e, quando ele terminar, volte para o passo 5.
  4. Se a terceira revisão também pedir mudanças, escale.

## 6. Merge

1. Na worktree, rode `make vet` e `make test` e confira que passam. Confira também se o PR pode ser mergeado: `gh pr view P --json mergeable`.
2. Saia da worktree (ExitWorktree, `keep`) e remova-a: `git worktree remove .claude/worktrees/<id>`.
3. `gh pr merge P --merge --delete-branch`
4. Tire o rótulo de etapa. Confira se a issue fechou pelo `Closes #N`. Se não fechou, feche-a com o link do PR.
5. Rode `git fetch origin --prune` e não mexa em mais nada na pasta do mantenedor.
6. Relate ao mantenedor:
   - o que entrou, com o link do PR;
   - os achados não bloqueantes do revisor, sugerindo quais viram issue;
   - os follow-ups que os agentes anotaram.

## Escalar

1. Pare o fluxo: aplique `needs-info` e mantenha o rótulo de etapa.
2. Comente na issue a pergunta exata, com as palavras do agente que travou, e o que cada resposta implicaria.
3. Saia da worktree (ExitWorktree, `keep`) e avise o mantenedor.

Para retomar depois que ele responder: `/pipeline N`.

## Sempre

- As mensagens aos agentes levam só ponteiros: issue, PR, branch, pasta e, numa devolução, o trecho exato que a motivou, com o link. Elas nunca levam um resumo da issue, a sua opinião ou o relatório de um agente para outro.
- Ninguém faz push direto na `main` ou no `develop`, force-push ou reescrita de histórico.
- Um ticket por vez.
