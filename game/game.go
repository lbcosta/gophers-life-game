// Package game liga a Colônia, o Mundo, o Painel de perfil e o Log num jogo da
// Ebitengine. Fica fora do main para o jogo e o harness (cmd/harness) rodarem
// exatamente o mesmo Update e o mesmo desenho.
package game

import (
	"bytes"
	"image"
	"image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"gophers-life-game/assets"
	"gophers-life-game/gopher"
	"gophers-life-game/hud"
	"gophers-life-game/logs"
	"gophers-life-game/world"
)

// grassBrightness escurece a grama para os Gophers se destacarem.
// 1 é o brilho original da textura; valores menores deixam mais escuro.
const grassBrightness = 0.6

// plantsPerGopher é quantas plantas o mundo tenta manter para cada Gopher.
const plantsPerGopher = 2

// Game é o jogo inteiro. Use New para criar.
type Game struct {
	gophers *gopher.Colony
	world   *world.World
	photos  *hud.Photos
	log     logs.Log // mensagens no canto inferior direito da tela

	prerendered bool // os sprites já foram preparados (só dá depois que o jogo roda)

	grass      *ebiten.Image // ladrilho da textura de grama
	background *ebiten.Image // grama já ladrilhada e escurecida, do tamanho da tela

	screen        image.Point // tamanho da tela, informado por Layout
	screenChanged bool        // screen mudou e os Gophers ainda não foram avisados
}

// New cria o jogo com count Gophers e as Plantas que eles pedem.
func New(count int) (*Game, error) {
	grass, err := loadGrass()
	if err != nil {
		return nil, err
	}
	sprites, err := world.LoadSprites(assets.FS())
	if err != nil {
		return nil, err
	}
	photos, err := hud.LoadPhotos(assets.FS())
	if err != nil {
		return nil, err
	}
	return &Game{
		photos:  photos,
		gophers: gopher.NewColony(count),
		world:   world.New(count*plantsPerGopher, sprites),
		grass:   grass,
	}, nil
}

// loadGrass decodifica a textura de grama embutida.
func loadGrass() (*ebiten.Image, error) {
	data, err := assets.Grass()
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// tileBackground repete tile (em ladrilhos) até cobrir size, multiplicando as
// cores por brightness.
func tileBackground(tile *ebiten.Image, size image.Point, brightness float32) *ebiten.Image {
	bg := ebiten.NewImage(size.X, size.Y)
	tileSize := tile.Bounds().Size()

	for y := 0; y < size.Y; y += tileSize.Y {
		for x := 0; x < size.X; x += tileSize.X {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			op.ColorScale.Scale(brightness, brightness, brightness, 1)
			bg.DrawImage(tile, op)
		}
	}
	return bg
}

// drawBackground pinta a grama. O fundo é montado uma vez e só é refeito se o
// tamanho da tela mudar.
func (g *Game) drawBackground(screen *ebiten.Image) {
	size := screen.Bounds().Size()
	if g.background == nil || g.background.Bounds().Size() != size {
		if g.background != nil {
			g.background.Deallocate()
		}
		g.background = tileBackground(g.grass, size, grassBrightness)
	}
	screen.DrawImage(g.background, nil)
}

// Update roda a cada tick (60x por segundo): lógica e input.
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	g.Simulate()

	// O tamanho da tela só é conhecido depois da primeira chamada a Layout.
	if g.screen.X == 0 || g.screen.Y == 0 {
		return nil
	}

	mx, my := ebiten.CursorPosition()
	g.gophers.Hover(float64(mx), float64(my))

	// Um clique escolhe o Gopher sob o mouse (ou desfaz a escolha, se não houver
	// nenhum). Cliques sobre o painel de perfil não contam.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		onPanel := g.gophers.Selected() != nil && hud.PanelContains(mx, my, g.screen.X)
		if !onPanel {
			g.gophers.Select(float64(mx), float64(my))
		}
	}

	return nil
}

// Simulate avança a simulação em um Tick, sem ler o mouse nem o teclado: o
// Mundo, a Colônia e o Log. O Update chama este método, e o harness também, para
// avançar Ticks sem input.
func (g *Game) Simulate() {
	// Os sprites só podem ser desenhados com o jogo rodando, então são preparados
	// uma única vez, no primeiro tick.
	if !g.prerendered {
		gopher.Prerender()
		g.world.Prerender()
		g.photos.Prerender()
		g.prerendered = true
	}

	// O tamanho da tela só é conhecido depois da primeira chamada a Layout.
	if g.screen.X == 0 || g.screen.Y == 0 {
		return
	}

	if g.screenChanged {
		g.world.Resize(float64(g.screen.X), float64(g.screen.Y))
		g.gophers.Resize(float64(g.screen.X), float64(g.screen.Y))
		g.screenChanged = false
	}
	g.world.Update()
	for _, e := range g.gophers.Update(g.world.Forage) {
		g.log.Add(logLevel(e.Kind), e.String())
	}
	g.log.Update()
}

// Colony devolve a Colônia, para o harness escolher e apontar Gophers.
func (g *Game) Colony() *gopher.Colony {
	return g.gophers
}

// Draw roda a cada frame: só desenha.
func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)
	g.world.Draw(screen)
	g.gophers.Draw(screen)
	if selected := g.gophers.Selected(); selected != nil {
		hud.DrawProfile(screen, selected.Profile(), g.photos)
	}
	g.log.Draw(screen)
}

// logLevel diz com que gravidade cada tipo de evento dos Gophers aparece no log:
// a fome ainda tem volta, a morte não.
func logLevel(k gopher.EventKind) logs.Level {
	if k == gopher.Starved {
		return logs.Critical
	}
	return logs.Warning
}

// Layout define a resolução lógica: acompanha o tamanho da tela. Só registra
// esse tamanho, e se ele mudou, para o Update usar.
func (g *Game) Layout(w, h int) (int, int) {
	if size := image.Pt(w, h); size != g.screen {
		g.screen = size
		g.screenChanged = true
	}
	return w, h
}
