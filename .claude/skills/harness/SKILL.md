---
name: harness
description: Ver e medir o Gophers Life sem olhar a tela. Roda o jogo numa janela com um cenário (N Gophers, avançar ticks, selecionar, hover), salva um screenshot e mede FPS, TPS, ms por tick e alocações. Use para conferir um critério visual, medir desempenho antes e depois de uma mudança, ou investigar lentidão e aquecimento.
---

# Harness

O jogo abre em tela cheia, e o Claude não consegue olhar a tela nem clicar. O harness (`cmd/harness`) roda o mesmo jogo num cenário controlado e devolve uma imagem ou números.

Se `cmd/harness` não existir no seu branch, diga isso no relatório e meça só com `make bench`.

## Use a GPU

No WSL, rode sempre com `GALLIUM_DRIVER=d3d12` (o `make` não tem alvo para o harness, então a variável é sua; sem ela, o harness imprime um aviso no stderr). Sem essa variável, o Mesa desenha com a CPU (o driver llvmpipe), e o FPS medido é o da CPU, não o do jogo. Uma rodada com o aviso no stderr não entra na comparação (DES-08).

## Contrato

`go run ./cmd/harness [flags]` abre uma janela de 1600×900, fora da tela cheia e sem vsync, com o mesmo Update e o mesmo desenho do jogo (o pacote `game`, que o `main.go` também usa).

| Flag | Efeito |
|---|---|
| `-n N` | Número de Gophers (padrão 60) |
| `-advance T` | Avança T ticks antes de abrir a janela. Com 3700, todos já são adultos |
| `-select` | Seleciona um Gopher, o que abre o Painel de perfil |
| `-hover` | Marca outro Gopher como Hover, a cada Tick, enquanto ele anda (sem mexer no mouse) |
| `-shot arquivo.png` | Salva o frame 30 e sai |
| `-dur 8s` | Descarta o primeiro segundo, mede por 8s e imprime uma linha: FPS, TPS, ms de update por tick, ms de draw por frame, alocações por frame e Gophers vivos. O ms de update conta só o Tick do jogo, sem o trabalho do próprio harness. |
| `-cpuprofile arquivo` | Grava um perfil de CPU, para ler com `go tool pprof -top` |

`-shot` e `-dur` não podem vir juntos: o harness recusa a combinação, explica o motivo no stderr e sai com código 2, sem abrir a janela.

Compilado com `-tags ebitenginedebug`, a Ebitengine imprime os comandos de desenho de cada frame, o que serve para contar as chamadas à placa de vídeo.

O ms de draw é o tempo da CPU para emitir os comandos de desenho; o FPS é o que mede o frame inteiro. As alocações por frame contam tudo o que o processo aloca, inclusive a Ebitengine.

Screenshots e perfis vão para o scratchpad ou para `/tmp`, nunca para o repositório.

## Receitas

- **Critério visual:** `GALLIUM_DRIVER=d3d12 go run ./cmd/harness -n 40 -advance 3700 -select -shot <arquivo>`. Abra a imagem com Read e amplie os recortes que precisar conferir.
- **Desempenho antes e depois:** `GALLIUM_DRIVER=d3d12 go run ./cmd/harness -n 1000 -advance 3700 -dur 8s` e `make bench`, pelo menos três rodadas sem a mudança e três com ela, na mesma sessão. Todas as rodadas vão para o PR (DES-07): o FPS varia entre rodadas, e a referência de DES-01 é só uma ordem de grandeza.
