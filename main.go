package main

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"gophers-life-game/gopher"
	"gophers-life-game/world"
)

// grassBrightness escurece a grama para os Gophers se destacarem.
// 1 é o brilho original da textura; valores menores deixam mais escuro.
const grassBrightness = 0.6

// gopherCount é a quantidade de Gophers do jogo.
const gopherCount = 60

// plantsPerGopher é quantas plantas o mundo tenta manter para cada Gopher.
const plantsPerGopher = 2

// plantsFS tem os sprites das plantas (assets/plants), embutidos no binário.
//
//go:embed assets/plants/*.png
var plantsFS embed.FS

// grassPNG é a textura de grama, embutida no binário para o jogo
// funcionar independentemente do diretório em que for executado.
//
//go:embed assets/grass.png
var grassPNG []byte

type Game struct {
	gophers *gopher.Colony
	world   *world.World

	prerendered bool // os sprites já foram preparados (só dá depois que o jogo roda)

	grass      *ebiten.Image // ladrilho da textura de grama
	background *ebiten.Image // grama já ladrilhada e escurecida, do tamanho da tela

	screen        image.Point // tamanho da tela, informado por Layout
	screenChanged bool        // screen mudou e os Gophers ainda não foram avisados
}

// loadGrass decodifica a textura de grama embutida.
func loadGrass() (*ebiten.Image, error) {
	img, err := png.Decode(bytes.NewReader(grassPNG))
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

	// Os sprites só podem ser desenhados com o jogo rodando, então são preparados
	// uma única vez, no primeiro tick.
	if !g.prerendered {
		gopher.Prerender()
		g.world.Prerender()
		g.prerendered = true
	}

	// O tamanho da tela só é conhecido depois da primeira chamada a Layout.
	if g.screen.X == 0 || g.screen.Y == 0 {
		return nil
	}

	if g.screenChanged {
		g.world.Resize(float64(g.screen.X), float64(g.screen.Y))
		g.gophers.Resize(float64(g.screen.X), float64(g.screen.Y))
		g.screenChanged = false
	}
	g.world.Update()
	g.gophers.Update(g.world.Forage)

	return nil
}

// Draw roda a cada frame: só desenha.
func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)
	g.world.Draw(screen)
	g.gophers.Draw(screen)
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

func main() {
	ebiten.SetFullscreen(true)
	ebiten.SetWindowTitle("Gophers Life")

	grass, err := loadGrass()
	if err != nil {
		log.Fatal(err)
	}

	sprites, err := world.LoadSprites(plantsFS)
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		gophers: gopher.NewColony(gopherCount),
		world:   world.New(gopherCount*plantsPerGopher, sprites),
		grass:   grass,
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
