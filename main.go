package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"gophers-life-game/gopher"
)

// grassBrightness escurece a grama para o Gopher se destacar.
// 1 é o brilho original da textura; valores menores deixam mais escuro.
const grassBrightness = 0.6

// grassPNG é a textura de grama, embutida no binário para o jogo
// funcionar independentemente do diretório em que for executado.
//
//go:embed assets/grass.png
var grassPNG []byte

type Game struct {
	gopher *gopher.Gopher

	grass      *ebiten.Image // ladrilho da textura de grama
	background *ebiten.Image // grama já ladrilhada e escurecida, do tamanho da tela

	width, height int // tamanho da tela, informado por Layout
	ticks         int
}

// loadGrass decodifica a textura de grama embutida.
func loadGrass() (*ebiten.Image, error) {
	img, err := png.Decode(bytes.NewReader(grassPNG))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// tileBackground repete tile (em ladrilhos) até cobrir size, escurecendo-o.
func tileBackground(tile *ebiten.Image, size image.Point) *ebiten.Image {
	bg := ebiten.NewImage(size.X, size.Y)
	tileSize := tile.Bounds().Size()

	for y := 0; y < size.Y; y += tileSize.Y {
		for x := 0; x < size.X; x += tileSize.X {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			op.ColorScale.Scale(grassBrightness, grassBrightness, grassBrightness, 1)
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
		g.background = tileBackground(g.grass, size)
	}
	screen.DrawImage(g.background, nil)
}

// Update roda a cada tick (60x por segundo): lógica e input.
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	g.ticks++

	// O tamanho da tela só é conhecido depois da primeira chamada a Layout.
	if g.width == 0 || g.height == 0 {
		return nil
	}
	w, h := float32(g.width), float32(g.height)

	if g.gopher == nil {
		g.gopher = gopher.NewGopher(w, h)
	} else {
		g.gopher.Resize(w, h)
	}
	g.gopher.Update(g.ticks)

	return nil
}

// Draw roda a cada frame: só desenha.
func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)

	// Pode não existir ainda se Draw rodar antes do primeiro Update completo.
	if g.gopher != nil {
		g.gopher.Draw(screen)
	}
}

// Layout define a resolução lógica: acompanha o tamanho da tela. Só registra
// esse tamanho para o Update usar.
func (g *Game) Layout(w, h int) (int, int) {
	g.width, g.height = w, h
	return w, h
}

func main() {
	ebiten.SetFullscreen(true)
	ebiten.SetWindowTitle("Gophers Life")

	grass, err := loadGrass()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(&Game{grass: grass}); err != nil {
		log.Fatal(err)
	}
}
