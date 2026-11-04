package hud

import (
	"fmt"
	"image"
	"image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"

	"gophers-life-game/gopher"
)

// Photos são as fotos de perfil dos Gophers. LoadPhotos só decodifica os
// arquivos; Prerender monta as imagens finais, uma única vez.
type Photos struct {
	decoded  map[string]image.Image
	rendered map[string]*ebiten.Image
}

// photoNames são os nomes de todas as fotos: uma para cada cor da paleta e cada
// emoção.
func photoNames() []string {
	var names []string
	for _, color := range gopher.Palette {
		for _, e := range gopher.Emotions {
			names = append(names, gopher.PhotoName(color, e))
		}
	}
	return names
}

// LoadPhotos lê de fsys, na pasta assets/profiles, as fotos de todas as cores e
// emoções.
func LoadPhotos(fsys fs.FS) (*Photos, error) {
	p := &Photos{decoded: map[string]image.Image{}}
	for _, name := range photoNames() {
		f, err := fsys.Open("assets/profiles/" + name)
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		p.decoded[name] = img
	}
	return p, nil
}

// Prerender reduz cada foto, uma vez, ao tamanho do espaço dela no painel. Deve
// ser chamado com o jogo já rodando; chamar de novo não tem efeito. Depois dele,
// as fotos no tamanho original não são mais necessárias e são liberadas.
func (p *Photos) Prerender() {
	if p.rendered != nil {
		return
	}
	p.rendered = map[string]*ebiten.Image{}
	for name, img := range p.decoded {
		src := ebiten.NewImageFromImage(img)
		dst := ebiten.NewImage(photoSize, photoSize)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		scale := float64(photoSize) / float64(src.Bounds().Dx())
		op.GeoM.Scale(scale, scale)
		dst.DrawImage(src, op)
		src.Deallocate() // libera já a memória de vídeo, sem esperar o coletor de lixo
		p.rendered[name] = dst
	}
	p.decoded = nil
}

// image devolve a foto name pronta para desenhar, ou nil se não houver.
func (p *Photos) image(name string) *ebiten.Image {
	if p == nil {
		return nil
	}
	return p.rendered[name]
}
