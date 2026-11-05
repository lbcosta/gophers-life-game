# Gophers Life

Simulação de vida artificial em Go com a Ebitengine. Gophers nascem, crescem, andam, comem plantas e morrem de fome num gramado, e o jogador só observa. O projeto também serve para o mantenedor aprender Go, então código claro, testado e com as decisões explicadas vale mais que código esperto.

## Fontes da verdade

- **Comportamento:** [docs/spec/](docs/spec/README.md). Cada regra tem um ID, como `GOP-07`, e issues, testes, commits e PRs citam esses IDs.
- **Vocabulário:** [GLOSSARY.md](GLOSSARY.md). Use os termos dele em código, testes, issues e documentação.
- **Decisões:** [docs/adr/](docs/adr/). Leia os ADRs da área antes de mexer nela.

Quando o código e a spec divergirem, aponte a divergência na issue ou no PR. Escolher um dos lados é decisão do mantenedor.

## Como o trabalho acontece

Toda mudança no repositório passa por quatro agentes de contexto isolado, definidos em `.claude/agents/`: **documentador**, **testador**, **programador** e **revisor** ([ADR-0003](docs/adr/0003-fluxo-de-quatro-agentes.md)). A sessão principal só orquestra, com três skills:

- `/pedido <descrição>`: o documentador transforma o pedido em issues com `needs-triage`.
- `/pipeline <issue>`: leva uma issue `ready-for-agent` até o merge no `develop`.
- `/release [versão]`: leva o `develop` para a `main`, com tag.

A sessão principal dispara agentes, cuida de rótulos e do git e repassa ponteiros. Código, testes e documentação ficam com os agentes. Perguntas, explicações e o Learning Mode continuam direto na conversa.

O mantenedor aprova as specs (trocando `needs-triage` por `ready-for-agent`), responde as escaladas (`needs-info`) e decide as releases. O modelo e o esforço de cada papel foram escolhidos por evidência ([docs/agents/modelos.md](docs/agents/modelos.md)).

O jogo abre em tela cheia, e o Claude não consegue olhar a tela nem clicar. Para ver o jogo ou medir FPS, use a skill `harness`.

## Comandos

- `make run`: roda o jogo (no WSL, na GPU).
- `make test`, `make vet` (go vet e gofmt) e `make bench` (desempenho com 1000 Gophers).
- `go run ./cmd/genassets` e `go run ./cmd/genprofiles`: refazem os desenhos das Plantas e as Fotos de perfil ([ADR-0002](docs/adr/0002-assets-gerados-por-codigo.md)).

## Ambiente

O jogo roda no WSL2, com GPU NVIDIA. Sem `GALLIUM_DRIVER=d3d12`, o Mesa desenha com a CPU (o driver llvmpipe), e qualquer medida de FPS passa a ser da CPU, não do jogo. O `make run` e o F5 do VS Code já definem essa variável. O harness não a define, mas, no WSL, avisa no stderr quando ela falta.

Os commits são assinados pelo 1Password, e a assinatura faz parte do commit. Depois de um tempo parado, a primeira assinatura costuma falhar com `1Password: failed to fill whole buffer`: repita o mesmo commit. Se falhar de novo, pare e avise o mantenedor.

## Arquitetura

- `main.go` abre a janela em tela cheia e roda o `game.Game`.
- `game/`: o jogo (Update, desenho e Gramado). É o único pacote que conhece `gopher`, `world`, `hud` e `logs` ao mesmo tempo, e o harness usa o mesmo código.
- `assets/`: os PNGs, embutidos no binário pelo próprio pacote.
- `cmd/harness`: roda o jogo num cenário por flags, para screenshots e medidas (skill `harness`).
- `gopher/`: os Gophers e a Colônia, com a lógica separada do desenho (`draw.go` e `ellipse.go`).
- `world/`: as Plantas. Não importa `gopher`: a Colônia recebe a Busca por comida como uma função (`Forager`).
- `hud/`: o Painel de perfil. Só lê um `gopher.Profile`.
- `logs/`: o Log. Recebe texto e Gravidade e não conhece os Gophers.
- `internal/grid` (vizinhança), `internal/circle` (círculos prontos), `internal/fonts` (fonte única) e `internal/raster` (formas para os geradores).

## Convenções de código

- Comentários e documentação em português, explicando o porquê. Identificadores em inglês. "Gopher" sempre com inicial maiúscula.
- Comentários quebram em até 80 colunas, contando o tab como 4. A régua vale para os comentários escritos ou mexidos a partir de agora: ninguém requebra um arquivo inteiro só por ela.
- `Update` muda o estado; `Draw` só desenha.
- O tempo é contado em Ticks, 60 por segundo. Segundos só aparecem na tela.
- A lógica usa `float64`, e o `float32` fica nas chamadas de desenho.
- Os sorteios são injetáveis: as structs guardam `chance func() float64` e `intn func(int) int`, e os testes passam os seus. Testes longos usam `rand.New(rand.NewPCG(...))`, com semente fixa.
- Enums são tabelas indexadas, terminando em `numX = iota`, em vez de `switch`. O `String()` devolve `Tipo(n)` para valores inválidos.
- Valores derivados são calculados no código, como `drawingReach`, e não escritos à mão.
- O desenho e o desempenho seguem [docs/spec/desempenho.md](docs/spec/desempenho.md): círculos via `internal/circle`, imagens preparadas no primeiro Tick, vizinhança via `internal/grid` e nenhuma alocação evitável por Frame.
- Nos testes, o nome é em inglês e descreve o comportamento (`TestColonyReportsHungerOnceWhenItStarts`), e as mensagens são em português (`"... esperado ..."`). `newGopher` cria adultos, com sorteios controlados.
- `make vet` e `make test` passam antes de todo commit.

## Git

- Git Flow:
  - a `main` só recebe releases, com tags semver;
  - o `develop` é a integração e o branch padrão;
  - cada ticket tem um branch `feature/<issue>-<slug>`, criado a partir do `develop` e mergeado por PR com merge commit.
- Commits e títulos de PR em inglês, no padrão Conventional Commits (`feat:`, `fix:`, `perf:`, `refactor:`, `docs:`, `test:`, `build:`, `chore:`). Issues e corpos de PR em português.
- Os agentes commitam, fazem push de branches de feature, abrem PRs e fazem merge no `develop` sem pedir. Ninguém faz push direto na `main` ou no `develop`, force-push ou reescrita de histórico compartilhado.
- As worktrees dos tickets ficam em `.claude/worktrees/`, que o git ignora.

## Agent skills

### Issue tracker

GitHub Issues de `lbcosta/gophers-life-game`, pelo `gh`. Ver `docs/agents/issue-tracker.md`.

### Triage labels

Os cinco rótulos padrão de triagem, mais os rótulos `etapa:*` do `/pipeline`. Ver `docs/agents/triage-labels.md`.

### Domain docs

Contexto único: `GLOSSARY.md` e `docs/adr/`, na raiz. Ver `docs/agents/domain.md`.
