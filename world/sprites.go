package world

import (
	"fmt"
	"image"
	"image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

// spriteSize é o lado, em pixels, com que as plantas aparecem na tela.
const spriteSize = 2 * plantRadius

// Sprites são as imagens das plantas. LoadSprites só decodifica os arquivos;
// Prerender monta as imagens finais, uma única vez.
type Sprites struct {
	decoded  map[string]image.Image
	rendered map[string]*ebiten.Image
}

// spriteNames são os arquivos usados: os dois estágios iniciais e um por alimento.
func spriteNames() []string {
	names := []string{"seed", "sprout"}
	for _, k := range kinds {
		names = append(names, k.sprite)
	}
	return names
}

// LoadSprites lê de fsys, na pasta assets/plants, os PNGs de todas as plantas.
func LoadSprites(fsys fs.FS) (*Sprites, error) {
	s := &Sprites{decoded: map[string]image.Image{}}
	for _, name := range spriteNames() {
		f, err := fsys.Open("assets/plants/" + name + ".png")
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s.png: %w", name, err)
		}
		s.decoded[name] = img
	}
	return s, nil
}

// Prerender reduz cada sprite, uma vez, ao tamanho com que aparece na tela, para
// que o Draw só copie imagens prontas. Deve ser chamado com o jogo já rodando;
// chamar de novo não tem efeito.
func (s *Sprites) Prerender() {
	if s.rendered != nil {
		return
	}
	s.rendered = map[string]*ebiten.Image{}
	for name, img := range s.decoded {
		src := ebiten.NewImageFromImage(img)
		dst := ebiten.NewImage(spriteSize, spriteSize)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		scale := float64(spriteSize) / float64(src.Bounds().Dx())
		op.GeoM.Scale(scale, scale)
		dst.DrawImage(src, op)
		s.rendered[name] = dst
	}
}

// image devolve o sprite que representa p na fase em que está.
func (s *Sprites) image(p *Plant) *ebiten.Image {
	switch p.Stage() {
	case Seed:
		return s.rendered["seed"]
	case Growing:
		return s.rendered["sprout"]
	}
	return s.rendered[kinds[p.kind].sprite]
}

// Prerender prepara as imagens para o Draw. Ver Sprites.Prerender.
func (w *World) Prerender() {
	if w.sprites != nil {
		w.sprites.Prerender()
	}
}

// Draw pinta as plantas em screen. Antes de Prerender não desenha nada.
func (w *World) Draw(screen *ebiten.Image) {
	if w.sprites == nil || w.sprites.rendered == nil {
		return
	}
	for _, p := range w.plants {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x-spriteSize/2, p.y-spriteSize/2)
		screen.DrawImage(w.sprites.image(p), op)
	}
}
