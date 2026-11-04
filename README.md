# Gophers Life

Jogo 2D em Go, feito com a [Ebitengine](https://ebitengine.org/). Um Gopher anda
sozinho, em direções aleatórias, sobre um gramado.

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
- `gopher/`: o Gopher (movimento, desenho e testes).
- `assets/grass.png`: textura de grama.
