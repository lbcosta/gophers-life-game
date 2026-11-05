# Modelo e esforço de cada papel

Os valores ficam no frontmatter de cada agente, em `.claude/agents/`, e este arquivo guarda o porquê. Quem muda um, muda o outro. Os modelos estão fixados pelo ID completo, para que uma versão nova não entre sem passar por esta escolha.

| Papel | Modelo | Esforço | Por quê |
|---|---|---|---|
| documentador: delta da spec e issues | Opus 5.5 | `medium` | Análise e escrita. Um erro aqui se propaga para os testes, o código e a revisão |
| documentador: conferência e release | Opus 5.5 | `low` | Tarefa simples e verificável. O `/pipeline` pede esse esforço na chamada |
| testador | Opus 5.5 | `high` | Os testes são o oráculo do ticket, e nada além da revisão confere um teste errado |
| programador | Sonnet 5.5 | `medium`, `high` na nova tentativa | Tickets pequenos, com os testes como alvo, e o revisor confere tudo depois |
| revisor | Opus 5.5 | `high` | Último portão antes do merge automático: o modelo mais capaz disponível, em esforço alto. É diferente do modelo do programador, cujo código ele revisa |
| sessão principal (orquestrador) | o da sessão | `medium` sugerido | Trabalho procedural: rótulos, git e encaminhamento |

O Fable 5.1 seria a primeira escolha para o revisor, mas o mantenedor não tem acesso a ele (05/11/2026). Sem o Fable, o revisor usa o mesmo modelo do testador. O contexto isolado e o prompt próprio de cada papel atenuam esse viés.

## Evidências

Fontes: as páginas da Anthropic sobre [modelos](https://platform.claude.com/docs/en/about-claude/models/overview), [preços](https://platform.claude.com/docs/en/about-claude/pricing), [esforço](https://platform.claude.com/docs/en/build-with-claude/effort), [custo e inteligência](https://platform.claude.com/docs/en/about-claude/models/optimizing-for-cost-and-intelligence) e o [guia de migração](https://platform.claude.com/docs/en/about-claude/models/migration-guide), consultadas em 05/11/2026.

- **Preço por milhão de tokens** (entrada/saída): Fable 5.1, US$ 10/50; Opus 5.5, US$ 4/20; Sonnet 5.5, US$ 2/10; Haiku 5.5, US$ 0,10/0,50.
- **Regra geral de esforço:** no mínimo `high` para trabalho sensível à inteligência, `max` quando a correção importa mais que o custo, e `low` para subagentes ou tarefas simples.
- **Curvas medidas pela Anthropic:** em pesquisa e trabalho de conhecimento, `medium` igualou `high`, e `high` não trouxe ganho mensurável (documentador em `medium`). Em programação, cada degrau abaixo custou alguns pontos de acerto por um custo bem menor (testador em `high`).
- **Opus 5.5:**
  - em `medium`, o padrão dele, supera o Opus 5 em `high` em programação e em trabalho de conhecimento;
  - `xhigh` e `max` ficam reservados para ganhos medidos;
  - é um analista mais confiável, escreve com mais clareza e, em revisão de código, acha mais bugs com menos alarmes falsos.
- **Sonnet 5.5:**
  - é o mais forte em programação agêntica de vários passos num repositório real;
  - o ponto de partida recomendado para programação é `medium`;
  - em `low`, às vezes dá uma mudança por pronta sem rodar um teste, e por isso o programador não usa `low`.
- **Modelo menor com a ajuda de testes:**
  - num conjunto de tarefas de programação que os dois modelos mais fortes quase resolvem por inteiro, o de baixo, no esforço padrão, igualou o de cima por uma fração do custo;
  - rodar tudo num esforço menor e repetir só as falhas num esforço maior manteve o acerto por pouco mais da metade do custo. É a nova tentativa do programador em `high`.
- **Fable 5.1:** é o modelo mais capaz, e a revisão de código e a depuração estão entre os maiores ganhos dele. Em esforço baixo, costuma competir em custo por tarefa com o Opus e o Sonnet, com desempenho melhor. Hoje está fora do alcance do mantenedor.
- **Verificação independente:** a Anthropic relata que verificadores em subagentes de contexto novo tendem a superar a autocrítica.
- **Viés de autopreferência:** avaliadores LLM reconhecem e favorecem os próprios textos ([Panickssery, Bowman e Feng, NeurIPS 2024](https://arxiv.org/abs/2404.13076)). Com dois modelos capazes disponíveis, as diferenças ficaram onde pesam mais. O revisor usa um modelo diferente do programador, cujo código é o principal objeto da revisão. O testador também usa um modelo diferente do programador, para que os testes e o código não compartilhem a mesma leitura errada da spec.
- **Subagentes do Claude Code:** o frontmatter aceita `model` (apelido ou ID completo) e `effort`, e o `effort` da definição sobrepõe o da sessão ([documentação](https://code.claude.com/docs/en/sub-agents)).

## Quando revisar

- **Saiu um modelo novo:** refaça a escolha comparando o custo por ticket concluído, não o preço por token.
- **O revisor barra falhas que um teste deveria ter pegado:** suba o esforço do testador.
- **O programador costuma precisar de nova tentativa:** comece em `high` ou troque para o Opus 5.5 em `medium`.
- **O Fable 5.1 fica disponível:** ele é o candidato natural para o revisor, e com ele o revisor também deixa de usar o mesmo modelo do testador.
- **O custo ou o limite de uso pesa:** o primeiro passo é baixar o revisor para `medium`, o padrão do Opus 5.5.
