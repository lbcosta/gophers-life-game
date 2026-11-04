# Gophers Life

Jogo 2D em Go, feito com a [Ebitengine](https://ebitengine.org/). Cinco Gophers,
cada um com cor e velocidade sorteadas, andam sozinhos sobre um gramado em
direções aleatórias, quicando nas bordas da tela e uns nos outros. Eles nascem
bebês, como bolinhas, e crescem em quatro fases: bebê, criança (uma bolinha
maior), adolescente (ganha olhos) e adulto (a forma completa, com orelhas e
nariz). A hit box acompanha o crescimento. A idade em que cada Gopher entra em
cada fase é sorteada dentro de um intervalo: criança entre 10 e 20 s,
adolescente entre 25 e 40 s e adulto entre 45 e 60 s.

## Como executar

```bash
go run .
```

No VS Code, use **F5** (com depuração) ou **Ctrl+F5** (sem depuração), com a
extensão oficial Go instalada.

O jogo abre em tela cheia. Pressione **Esc** para fechar.

## Testes

```bash
go test ./...
```

## Estrutura

- `main.go`: janela, loop do jogo e fundo de grama.
- `gopher/`: os Gophers.
  - `gopher.go`: movimento, giro e nascimento de um Gopher.
  - `life.go`: fases da vida, maturidade sorteada e crescimento do corpo e da hit box.
  - `interaction.go`: colisões entre Gophers e as reações a elas.
  - `personality.go`: velocidades, paleta de cores e criação de Gophers sorteados.
  - `colony.go`: vários Gophers na mesma tela.
  - `draw.go` e `ellipse.go`: desenho.
- `assets/grass.png`: textura de grama.
