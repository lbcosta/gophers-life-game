// Comando genassets desenha os sprites das plantas do jogo e os grava em
// assets/plants. Rode a partir da raiz do projeto:
//
//	go run ./cmd/genassets
//
// Os desenhos usam o mesmo estilo dos Gophers: formas simples (elipses e
// polígonos) em cores chapadas. Cada sprite é desenhado em escala 4x, ganha um
// contorno escuro e é reduzido para 64x64, o que dá o serrilhado suave.
package main

import (
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"gophers-life-game/internal/raster"
)

const (
	size = 64 // lado do sprite final, em pixels
	k    = 4  // fator de supersampling

	outlineRadius = 1.5 // espessura do contorno, em pixels
)

var outlineColor = color.NRGBA{22, 32, 16, 255}

type canvas = raster.Canvas

var rgb = raster.RGB

func pt(x, y float64) raster.Pt { return raster.Pt{X: x, Y: y} }

var (
	leafGreen  = rgb(0x4FB83C)
	leafDark   = rgb(0x3A9A2E)
	bladeLight = rgb(0x8BDD52)
	bladeMid   = rgb(0x62C23E)
	soil       = rgb(0x6B4A2E)
	soilLight  = rgb(0x8A6240)
)

// leaves desenha as folhas de uma hortaliça, saindo do ponto (cx, cy).
func leaves(c *canvas, cx, cy float64) {
	c.Ellipse(cx-7, cy-7, 3, 10, -30, leafDark)
	c.Ellipse(cx+7, cy-7, 3, 10, 30, leafDark)
	c.Ellipse(cx, cy-9, 3.5, 11, 0, leafGreen)
}

var sprites = map[string]func(c *canvas){
	"seed": func(c *canvas) {
		c.Ellipse(32, 44, 17, 8, 0, soil)
		c.Ellipse(30, 41, 9, 3, 0, soilLight)
		c.Ellipse(32, 39, 4, 2.6, 20, rgb(0xEBDDA6))
	},
	"sprout": func(c *canvas) {
		c.Ellipse(32, 48, 15, 6.5, 0, soil)
		c.Ellipse(30, 46, 8, 2.5, 0, soilLight)
		c.Poly(rgb(0x4DB03E), pt(31, 48), pt(33, 48), pt(33, 36), pt(31, 36))
		c.Ellipse(24, 36, 8.5, 3.6, -30, rgb(0x5CCB4B))
		c.Ellipse(40, 36, 8.5, 3.6, 30, rgb(0x5CCB4B))
	},
	"potato": func(c *canvas) {
		c.Ellipse(32, 38, 19, 14, -15, rgb(0xB98A55))
		c.Ellipse(25, 32, 7, 3, -15, rgb(0xD2A877))
		c.Ellipse(26, 37, 2.2, 1.6, 0, rgb(0x7A5530))
		c.Ellipse(38, 42, 2.2, 1.6, 0, rgb(0x7A5530))
		c.Ellipse(32, 45, 2, 1.4, 0, rgb(0x7A5530))
		c.Ellipse(40, 33, 2, 1.4, 0, rgb(0x7A5530))
	},
	"carrot": func(c *canvas) {
		leaves(c, 32, 26)
		c.Poly(rgb(0xF28A1D), pt(22, 26), pt(42, 26), pt(32, 58))
		c.Ellipse(32, 26, 10, 4, 0, rgb(0xF9A23B))
		c.Ellipse(30, 34, 4, 1.1, 0, rgb(0xD56E0E))
		c.Ellipse(34, 42, 3, 1, 0, rgb(0xD56E0E))
	},
	"radish": func(c *canvas) {
		leaves(c, 32, 28)
		c.Poly(rgb(0xF4E9EE), pt(29, 44), pt(35, 44), pt(32, 59))
		c.Ellipse(32, 38, 14, 13, 0, rgb(0xE0356B))
		c.Ellipse(27, 32, 4, 2.4, -30, rgb(0xF27AA0))
	},
	"grass": func(c *canvas) {
		c.Poly(bladeMid, pt(22, 57), pt(29, 57), pt(11, 28))
		c.Poly(bladeMid, pt(35, 57), pt(42, 57), pt(53, 30))
		c.Poly(bladeLight, pt(25, 57), pt(33, 57), pt(20, 14))
		c.Poly(bladeLight, pt(31, 57), pt(39, 57), pt(44, 18))
		c.Poly(bladeLight, pt(28, 57), pt(36, 57), pt(32, 8))
	},
	"clover": func(c *canvas) {
		c.Poly(leafDark, pt(30.5, 34), pt(33.5, 34), pt(33.5, 58), pt(30.5, 58))
		light := rgb(0x7BE090)
		for _, l := range []struct{ x, y, deg float64 }{{32, 22, 90}, {22, 36, -30}, {42, 36, 30}} {
			c.Ellipse(l.x, l.y, 10, 10, 0, rgb(0x3FBF62))
			c.Ellipse(l.x, l.y, 5.5, 2, l.deg, light)
		}
	},
	"alfalfa": func(c *canvas) {
		c.Poly(leafDark, pt(30.5, 58), pt(33.5, 58), pt(33.5, 16), pt(30.5, 16))
		for _, y := range []float64{50, 41, 32, 23} {
			c.Ellipse(24, y, 6.5, 3, -35, leafGreen)
			c.Ellipse(40, y, 6.5, 3, 35, leafGreen)
		}
		purple := rgb(0xA67BE8)
		c.Ellipse(32, 14, 5, 5, 0, rgb(0x8E5BD6))
		c.Ellipse(27, 17, 3.6, 3.6, 0, purple)
		c.Ellipse(37, 17, 3.6, 3.6, 0, purple)
		c.Ellipse(32, 8, 3.6, 3.6, 0, purple)
	},
}

func main() {
	dir := filepath.Join("assets", "plants")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	for name, draw := range sprites {
		c := raster.New(size, k)
		draw(c)
		c.Outline(outlineRadius, outlineColor)

		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(f, c.Image()); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
