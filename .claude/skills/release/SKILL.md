---
name: release
description: Faz uma release do Gophers Life pelo Git Flow, levando o develop para a main com tag semver e GitHub Release, ou um hotfix a partir da main. Use só quando o mantenedor pedir uma release ou um hotfix.
---

# /release [versão]

Você é o orquestrador ([ADR-0003](../../../docs/adr/0003-fluxo-de-quatro-agentes.md)). Releases e hotfixes só acontecem quando o mantenedor pede.

## Release

1. Defina a versão. Se o mantenedor não disse qual, calcule a partir da última tag (`git describe --tags --abbrev=0`) e dos commits desde ela:
   - se há algum `feat:`, suba o número do meio (de 0.1.0 para 0.2.0);
   - se não há, suba o último (de 0.1.0 para 0.1.1).

   Confirme a versão com o mantenedor antes de seguir.
2. Prepare o branch e a pasta:
   - `git fetch origin`
   - `git worktree add .claude/worktrees/release-<versão> -b release/<versão> origin/develop`
   - `git -C .claude/worktrees/release-<versão> push -u origin release/<versão>`
   - entre na worktree com EnterWorktree (`path`).
3. Dispare o `documentador` com `effort: low`:

   > Tarefa: release <versão>. Branch release/<versão>, pasta <caminho absoluto>.

4. Rode `make vet` e `make test` na worktree e confira que passam.
5. Saia da worktree (ExitWorktree, `keep`) e remova-a com `git worktree remove`.
6. Leve o branch para a `main`:
   - abra o PR: `gh pr create --base main --head release/<versão> --title "chore(release): v<versão>"`, com o corpo listando os PRs que entraram;
   - faça o merge: `gh pr merge <pr> --merge`.
7. Crie a tag e a release:
   - `git fetch origin`
   - `git tag -a v<versão> origin/main -m "v<versão>"`
   - `git push origin v<versão>`
   - `gh release create v<versão> --generate-notes --title "v<versão>"`
8. Leve o branch de volta para o `develop`:
   - abra o PR `release/<versão>` → `develop` e faça o merge com `--merge`;
   - apague o branch: `git push origin --delete release/<versão>`.
9. Relate ao mantenedor a versão, o link da release e o que entrou.

## Hotfix

Para um bug urgente na `main`:

1. Abra a issue pelo `/pedido` e espere o mantenedor aprovar.
2. Rode o `/pipeline` com duas trocas:
   - o branch é `hotfix/<versão>`, criado a partir de `origin/main`;
   - o PR aponta para a `main`.
3. A versão sobe o último número. Depois do merge, faça os passos 7 e 8 da release, com `hotfix/<versão>` no lugar de `release/<versão>`.
