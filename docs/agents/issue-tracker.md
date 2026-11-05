# Issue tracker: GitHub

Issues, specs de feature e tickets vivem no GitHub Issues de `lbcosta/gophers-life-game`. Use o `gh` para tudo: dentro do clone, ele acha o repositório sozinho. Corpos longos vão por arquivo (`--body-file`) ou heredoc.

## Issues

- **Criar:** `gh issue create --title "..." --body-file <arquivo> --label needs-triage`
- **Ler:** `gh issue view <n> --json number,title,body,labels,comments,state`
- **Listar:** `gh issue list --state open --label <rótulo> --json number,title,labels`
- **Sub-issue:** `gh issue create --parent <pai> ...`, ou depois `gh issue edit <pai> --add-sub-issue <filho>`.
- **Bloqueio:** `gh api --method POST repos/lbcosta/gophers-life-game/issues/<filho>/dependencies/blocked_by -F issue_id=<id>`, com o id numérico do bloqueador (`gh api repos/lbcosta/gophers-life-game/issues/<n> --jq .id`, que não é o `#número`). Se a API falhar, use uma linha `Bloqueada por: #a, #b` no topo do corpo.
- **Comentar:** `gh issue comment <n> --body-file <arquivo>`
- **Rótulos:** `gh issue edit <n> --add-label <x> --remove-label <y>`
- **Fechar:** pelo `Closes #N` do PR. À mão: `gh issue close <n> --comment "..."`

## Pull requests

- **Abrir:** `gh pr create --base develop --head feature/<n>-<slug> --title "<tipo>: ..." --body-file <arquivo>`
- **Ler:** `gh pr view <n> --comments`; diff: `gh pr diff <n>`
- **Comentar:** `gh pr comment <n> --body-file <arquivo>`
- **Merge:** `gh pr merge <n> --merge --delete-branch`

O GitHub usa a mesma numeração para issues e PRs, então `#42` pode ser um ou outro: tente `gh pr view 42` e, se falhar, `gh issue view 42`.

PRs como superfície de pedidos: **não**.

## Quando uma skill disser "publique no issue tracker"

Crie uma issue no GitHub, com o rótulo `needs-triage`.

## Quando uma skill disser "busque o ticket"

Leia a issue como em **Ler**.
