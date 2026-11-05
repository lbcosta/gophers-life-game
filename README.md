# Gophers Life

Jogo 2D em Go, feito com a [Ebitengine](https://ebitengine.org/). Cinco Gophers,
cada um com cor e velocidade sorteadas, andam sozinhos sobre um gramado em
direções aleatórias, quicando nas bordas da tela e uns nos outros. Eles nascem
bebês, como bolinhas, e crescem em quatro fases: bebê, criança (uma bolinha
maior), adolescente (ganha olhos) e adulto (a forma completa, com orelhas e
nariz). A hit box acompanha o crescimento. A idade em que cada Gopher entra em
cada fase é sorteada dentro de um intervalo: criança entre 10 e 20 s,
adolescente entre 25 e 40 s e adulto entre 45 e 60 s.

Pelo gramado crescem seis alimentos (batata, cenoura, rabanete, capim, trevo e
alfafa). Cada planta passa por três fases: semente, crescendo e pronta. Passar
por cima de uma semente não faz nada, passar por uma planta crescendo a desfaz,
e passar por uma pronta é comê-la: batata, cenoura e rabanete dão +2 de
saciedade, e capim, trevo e alfafa dão +1. Todo Gopher nasce com saciedade 10,
perde 1 a cada 1000 passos (o primeiro ponto cai depois de um número sorteado
de passos, para que não percam todos juntos) e morre (some) ao chegar a zero. Cada Gopher também
nasce com um nome sorteado (nunca repetido), que aparece quando o mouse passa por cima dele. Clicar em um Gopher abre, do
lado direito da tela, um painel com as características dele (nome, cor,
foto, emoção, velocidade, saciedade, fase da vida, idade e tamanho); clicar no vazio fecha o painel.

No canto inferior direito, um log avisa quando um Gopher passa a ter fome
(saciedade 3 ou menos), em amarelo, e quando um Gopher morre de fome, em
vermelho. O log mostra as 8 mensagens mais recentes, e cada uma some aos poucos
depois de 10 s.

## Como executar

```bash
make run
```

No WSL, `make run` liga a variável `GALLIUM_DRIVER=d3d12`, que faz o OpenGL
usar a placa de vídeo. Sem ela, o Mesa desenha com a CPU (o driver `llvmpipe`):
o jogo fica bem mais lento, e o notebook esquenta. Fora do WSL, `make run` é o
mesmo que `go run .`. Para comparar com o desenho pela CPU, rode
`make run GALLIUM_DRIVER=llvmpipe`.

No VS Code, use **F5** (com depuração) ou **Ctrl+F5** (sem depuração), com a
extensão oficial Go instalada. Há duas configurações: **Gophers Life (GPU)**,
a padrão, e **Gophers Life (CPU)**, para comparar. As duas definem
`GALLIUM_DRIVER`, que só faz sentido no WSL.

O jogo abre em tela cheia. Pressione **Esc** para fechar.

`make help` lista os outros comandos.

## Testes

```bash
make test
```

Para medir o desempenho com muitos Gophers (1000 adultos se esbarrando, e um
mundo lotado de plantas):

```bash
make bench
```

## Estrutura

- `main.go`: janela, loop do jogo e fundo de grama.
- `Makefile`: atalhos para rodar (na GPU, no WSL), testar e medir.
- `gopher/`: os Gophers.
  - `gopher.go`: movimento, giro e nascimento de um Gopher.
  - `life.go`: fases da vida, maturidade sorteada e crescimento do corpo e da hit box.
  - `interaction.go`: colisões entre Gophers e as reações a elas.
  - `personality.go`: velocidades, paleta de cores e criação de Gophers sorteados.
  - `colony.go`: vários Gophers na mesma tela, a grade que diz quem está perto de quem e o hover do mouse.
  - `profile.go`: a fotografia das características de um Gopher, mostrada no painel.
  - `event.go`: os eventos que a colônia informa a cada tick (um Gopher passou a ter fome ou morreu de fome).
  - `names.go` e `names.txt`: os nomes que os Gophers podem receber, e de onde vieram.
  - `draw.go` e `ellipse.go`: desenho.
- `hud/`: a interface sobreposta ao jogo (o painel de perfil do Gopher).
- `logs/`: o log de mensagens do canto inferior direito da tela. Não conhece os Gophers: recebe só o texto e a gravidade de cada mensagem.
- `world/`: os objetos do mundo.
  - `plant.go`: tipos de alimento, fases de crescimento e maturidade sorteada.
  - `world.go`: plantio, crescimento e o que acontece quando um Gopher passa por cima.
  - `sprites.go`: carrega e pré-renderiza os sprites das plantas.
- `cmd/genprofiles/`: gera as fotos de perfil (uma por cor e emoção). Rode `go run ./cmd/genprofiles`; com `-sheet arquivo.png`, grava também uma folha com todas.
- `internal/raster/`: desenho de formas com contorno, usado pelos dois geradores.
- `internal/grid/`: grade espacial, que acha depressa o que está perto de um ponto (Gophers para colidir, plantas para comer).
- `internal/circle/`: círculos desenhados uma vez e depois só copiados, usados pelos Gophers e pelo painel.
- `internal/fonts/`: a fonte do jogo, carregada uma única vez para todos os textos.
- `cmd/genassets/`: gera os sprites das plantas. Rode `go run ./cmd/genassets` para refazê-los.
- `assets/grass.png`: textura de grama.
- `assets/profiles/`: fotos de perfil dos Gophers, 256x256 (7 cores × 5 emoções). Cada Gopher sorteia a emoção no nascimento, e o painel mostra a foto da cor e da emoção dele.
- `assets/plants/`: sprites das plantas (semente, broto e os seis alimentos).
