# Gophers Life

Simulação de vida artificial em Go, feita com a [Ebitengine](https://ebitengine.org/). Gophers com cor, nome e velocidade sorteados nascem bebês, crescem, andam pelo gramado esbarrando uns nos outros, comem as plantas que crescem nele e morrem de fome. Você só observa:

- passe o mouse sobre um Gopher para ver o nome dele;
- clique num Gopher para acompanhá-lo no painel de perfil;
- leia os avisos de fome e de morte no log do canto da tela.

As regras completas estão na [spec](docs/spec/README.md), e os termos usados nelas, no [glossário](GLOSSARY.md).

## Como executar

```bash
make run
```

No WSL, `make run` define `GALLIUM_DRIVER=d3d12`, que faz o OpenGL usar a placa de vídeo. Sem essa variável, o Mesa desenha com a CPU (o driver `llvmpipe`): o jogo fica bem mais lento, e o notebook esquenta. Fora do WSL, `make run` é o mesmo que `go run .`. Para comparar com o desenho pela CPU, rode `make run GALLIUM_DRIVER=llvmpipe`.

No VS Code, com a extensão oficial Go instalada, use **F5** (com depuração) ou **Ctrl+F5** (sem depuração). Há duas configurações:

- **Gophers Life (GPU)**, a padrão;
- **Gophers Life (CPU)**, para comparar.

O jogo abre em tela cheia, e **Esc** fecha. `make help` lista os outros comandos.

### Ver e medir sem olhar a tela

`go run ./cmd/harness` roda o mesmo jogo numa janela de 1600×900, sem vsync, e tira screenshots ou mede FPS, TPS, ms por tick e alocações. Por exemplo:

```bash
GALLIUM_DRIVER=d3d12 go run ./cmd/harness -n 40 -advance 3700 -select -shot /tmp/jogo.png
GALLIUM_DRIVER=d3d12 go run ./cmd/harness -n 1000 -advance 3700 -dur 8s
```

`go run ./cmd/harness -h` lista as flags, e a skill [harness](.claude/skills/harness/SKILL.md) explica cada uma.

## Testes

```bash
make test    # testes
make vet     # go vet e gofmt
make bench   # desempenho com 1000 Gophers e um mundo lotado de plantas
```

## Estrutura

- `main.go`: abre a janela em tela cheia e roda o jogo.
- `game/`: o jogo (Update, desenho e fundo de grama). É onde os pacotes se ligam, e o harness usa o mesmo.
- `assets/`: os PNGs embutidos no binário.
- `gopher/`: os Gophers e a Colônia (movimento, colisão, fases da vida, saciedade, eventos e desenho).
- `world/`: as plantas (alimentos, crescimento, plantio e busca por comida).
- `hud/`: o painel de perfil e as fotos.
- `logs/`: o log de mensagens do canto inferior direito.
- `internal/`: a grade espacial, os círculos prontos, a fonte e o desenho de formas usado pelos geradores.
- `cmd/harness`: roda o jogo num cenário por flags, para screenshots e medidas.
- `cmd/genassets` e `cmd/genprofiles`: geram os desenhos das plantas e as fotos de perfil, que ficam em `assets/`.
- `docs/spec/`: o comportamento do jogo, regra por regra.
- `docs/adr/`: as decisões de arquitetura e o motivo de cada uma.
- `docs/agents/`: como os agentes do Claude Code trabalham neste repositório.
- `CLAUDE.md` e `.claude/`: as instruções, os agentes e as skills do Claude Code.

## Como o projeto evolui

Toda mudança é feita por quatro agentes do Claude Code: o documentador, o testador, o programador e o revisor. Eles se coordenam por issues e PRs no GitHub, seguindo o Git Flow. Veja o [CLAUDE.md](CLAUDE.md) e o [ADR-0003](docs/adr/0003-fluxo-de-quatro-agentes.md).
