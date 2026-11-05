# Rótulos

## Triagem

Os cinco papéis canônicos das skills e o rótulo de cada um neste repositório:

| Papel nas skills | Rótulo aqui | Significado |
|---|---|---|
| `needs-triage` | `needs-triage` | O documentador escreveu a issue, e o mantenedor ainda não aprovou a spec |
| `needs-info` | `needs-info` | O fluxo parou e espera uma resposta do mantenedor, com a pergunta escrita na issue |
| `ready-for-agent` | `ready-for-agent` | Spec aprovada: os agentes podem começar |
| `ready-for-human` | `ready-for-human` | Só o mantenedor resolve (gosto pessoal, ou algo fora do alcance dos agentes) |
| `wontfix` | `wontfix` | Não será feito |

Quem aprova é sempre o mantenedor: os agentes criam issues com `needs-triage` e nunca aplicam `ready-for-agent`.

## Etapas do `/pipeline`

Uma issue em andamento tem exatamente um rótulo de etapa, e quem troca esse rótulo é o orquestrador:

| Rótulo | Quem está trabalhando |
|---|---|
| `etapa:spec` | O documentador commita o delta da spec no branch |
| `etapa:testes` | O testador escreve os testes que falham |
| `etapa:código` | O programador faz os testes passarem e abre o PR |
| `etapa:docs` | O documentador confere a documentação contra o código |
| `etapa:revisão` | O revisor revisa o PR |

Uma issue aberta com `ready-for-agent` e sem rótulo de etapa ainda não começou. Uma issue fechada já teve o merge feito.
