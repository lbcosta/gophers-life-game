# Simulação

## Tempo e Tela

- **SIM-01** — A Simulação avança em Ticks, 60 por segundo, que é o padrão da Ebitengine. Toda a lógica (movimento, crescimento, fome, plantio, Log, Hover e Seleção) roda no Tick. Desenhar um Frame nunca muda a Simulação.
- **SIM-02** — O jogo abre em tela cheia, com o título "Gophers Life". A resolução lógica acompanha a Tela: um pixel lógico é um pixel da Tela.
- **SIM-03** — A tecla Esc fecha o jogo na hora, sem confirmação.
- **SIM-04** — A Tela é o limite do Mundo: Gophers e Plantas só existem dentro dela.

## Início e mudança de tamanho

- **SIM-05** — Enquanto o tamanho da Tela não é conhecido, nada anda e nada nasce, e só o gramado aparece.
- **SIM-06** — Quando o tamanho da Tela fica conhecido pela primeira vez, o Mundo planta as plantas iniciais (PLA-05) e a Colônia espalha os Gophers (GOP-09).
- **SIM-07** — Se a Tela mudar de tamanho depois disso, os Gophers são trazidos para dentro dela, as Plantas que ficaram fora somem (PLA-09), e nada nasce por causa da mudança.

## População

- **SIM-08** — A Colônia começa com `main.gopherCount` Gophers e comporta no máximo `gopher.MaxGophers`. Um pedido maior é cortado nesse limite, e um negativo vira zero.
- **SIM-09** — A meta do Mundo é ter `main.gopherCount × main.plantsPerGopher` Plantas. Ela é calculada uma vez, no início, e não acompanha os Gophers vivos.
- **SIM-10** — Não há nascimentos depois do início: a Colônia só diminui. Se todos morrerem, as Plantas continuam crescendo num gramado vazio.

## Ordem de um Tick

- **SIM-11** — Cada Tick acontece nesta ordem:
  1. No primeiro Tick, todas as imagens são preparadas (DES-05).
  2. Se o tamanho da Tela mudou, o Mundo e depois a Colônia recebem o tamanho novo.
  3. As Plantas envelhecem, e talvez nasça uma semente (PLA-06).
  4. Cada Gopher, na ordem da Colônia, anda e faz a Busca por comida. Os que morreram de fome saem da Colônia no fim desta etapa.
  5. Os Eventos do Tick viram Mensagens no Log, na ordem em que aconteceram (INT-13).
  6. O Log envelhece (INT-12).
  7. O Hover passa a ser o Gopher sob o mouse (INT-07).
  8. Um clique esquerdo muda a Seleção (INT-08).
- **SIM-12** — Os Gophers de um Tick são atualizados um depois do outro, e cada um já encontra os anteriores nas posições novas. Nada da Simulação roda em paralelo.

## Parâmetros

| Parâmetro | Valor | Significado |
|---|---|---|
| `main.gopherCount` | 60 | Gophers no início |
| `game.plantsPerGopher` | 2 | Plantas da meta do Mundo para cada Gopher inicial |
| `gopher.MaxGophers` | 1000 | Limite de Gophers da Colônia |
| `gopher.ticksPerSecond` | 60 | Ticks por segundo |
| `world.ticksPerSecond` | 60 | Ticks por segundo |
| `logs.ticksPerSecond` | 60 | Ticks por segundo |
