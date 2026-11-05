---
name: pedido
description: Transforma um pedido de mudança no Gophers Life (feature, correção, ajuste de regra ou refactor) em issues no GitHub, escritas pelo agente documentador e marcadas com needs-triage para o mantenedor aprovar. Use quando o mantenedor pedir qualquer mudança no jogo ou no repositório, porque toda mudança começa por aqui.
---

# /pedido <descrição>

Você é o orquestrador. Quem transforma o pedido em spec e issues é o agente `documentador` ([ADR-0003](../../../docs/adr/0003-fluxo-de-quatro-agentes.md)).

1. Junte o pedido com as palavras do mantenedor: o texto passado à skill e as mensagens dele nesta conversa que definem o que ele quer. Copie essas palavras como estão, sem plano, opinião ou interpretação sua.
2. Prepare uma pasta com o `develop` atual:
   - rode `git fetch origin`;
   - se `.claude/worktrees/pedido` já existir, remova-a com `git worktree remove --force .claude/worktrees/pedido`;
   - crie-a: `git worktree add --detach .claude/worktrees/pedido origin/develop`;
   - entre nela com EnterWorktree (`path`).
3. Dispare o `documentador`:

   > Tarefa: pedido. Pasta <caminho absoluto da worktree>. Pedido do mantenedor: <as palavras dele>

   Encerre o turno e espere a notificação de que ele terminou.
4. Saia da worktree (ExitWorktree, `keep`) e remova-a: `git worktree remove .claude/worktrees/pedido`.
5. Relate ao mantenedor:
   - as issues criadas, com os links;
   - as perguntas que o documentador deixou nelas;
   - como aprovar: ler a issue e trocar o rótulo, com `gh issue edit N --add-label ready-for-agent --remove-label needs-triage`.

   Depois de aprovar, ele roda `/pipeline N`.

Pronto quando: as issues existem no GitHub com `needs-triage`, e o mantenedor sabe o que decidir em cada uma.
