# Spec do Gophers Life

Esta pasta descreve o comportamento atual do jogo e é a fonte da verdade sobre ele. Se o código discordar da spec, um dos dois está errado: a divergência é apontada e resolvida numa issue, nunca em silêncio.

## Arquivos

| Arquivo | Prefixo | Assunto |
|---|---|---|
| [visao.md](visao.md) | — | O que o jogo é, o que está fora de escopo e as ideias futuras |
| [simulacao.md](simulacao.md) | `SIM` | Tempo, Tela, população inicial e a ordem das coisas em cada Tick |
| [gophers.md](gophers.md) | `GOP` | Nascimento, movimento, Colisão, fases da vida, Saciedade e Eventos |
| [plantas.md](plantas.md) | `PLA` | Alimentos, fases da planta, plantio e Busca por comida |
| [interface.md](interface.md) | `INT` | O que aparece na Tela: desenho, Hover, Seleção, Painel de perfil e Log |
| [desempenho.md](desempenho.md) | `DES` | Metas de desempenho e as regras de desenho que as garantem |

Os termos com inicial maiúscula (Tick, Passo, Saciedade…) estão definidos no [glossário](../../GLOSSARY.md).

## Regras e IDs

- Cada regra tem um ID estável: o prefixo do arquivo e um número, como `GOP-07`. Issues, testes, commits e PRs citam as regras pelo ID.
- Uma regra nova recebe o próximo número livre do arquivo, mesmo quando entra no meio de uma seção.
- Um ID nunca é reaproveitado. Uma regra que deixa de valer continua no arquivo, riscada e com a issue que a removeu: `~~GOP-07~~ (removida em #12)`.
- As regras descrevem comportamento observável, não implementação: o texto delas não cita arquivos nem funções. A ligação com o código fica nas tabelas de parâmetros.

## Tabelas de parâmetros

Os arquivos com regras terminam numa seção `## Parâmetros` com os números que o comportamento usa. Um teste compara cada tabela com as constantes do código e falha apontando a diferença, então o formato é rígido:

| Parâmetro | Valor | Significado |
|---|---|---|
| `gopher.stepsPerSatiety` | 1000 | Passos para perder 1 ponto de Saciedade |

- **Parâmetro:** a expressão Go da constante, entre crases, com o pacote na frente (`main.`, `gopher.`, `world.` ou `logs.`). Elemento de tabela usa índice e campo usa ponto, como no código: `gopher.speedSteps[Fast]`, `gopher.maturityRanges[Child-1].min`.
- **Valor:** só o número, na unidade do código (ticks, pixels, graus ou pontos), com ponto decimal (`0.25`).
- **Significado:** texto livre, que pode trazer a conversão para segundos.

Quem muda uma constante à mão, enquanto joga, atualiza a tabela no mesmo commit (ou pede ao documentador).

## Como a spec muda

Toda mudança de comportamento passa pelo fluxo de agentes ([ADR-0003](../adr/0003-fluxo-de-quatro-agentes.md)). O documentador descreve o delta da spec na issue, o mantenedor aprova, e o delta entra no repositório no mesmo PR que o código e os testes.
