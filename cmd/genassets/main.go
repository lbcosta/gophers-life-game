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
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

const (
	size = 64 // lado do sprite final, em pixels
	k    = 4  // fator de supersampling
	big  = size * k

	outlineRadius = 1.5 * k // espessura do contorno, em pixels da imagem grande
)

var outlineColor = color.NRGBA{22, 32, 16, 255}

func rgb(hex uint32) color.NRGBA {
	return color.NRGBA{uint8(hex >> 16), uint8(hex >> 8), uint8(hex), 255}
}

type pt struct{ x, y float64 }

// canvas é a imagem grande em que as formas são pintadas. As coordenadas das
// formas são do sprite final (0 a 64); o canvas as multiplica por k.
type canvas struct{ pix [big * big]color.NRGBA }

func (c *canvas) set(x, y int, col color.NRGBA) {
	if x >= 0 && y >= 0 && x < big && y < big {
		c.pix[y*big+x] = col
	}
}

// ellipse pinta uma elipse de centro (cx, cy), raios rx e ry, girada deg graus.
func (c *canvas) ellipse(cx, cy, rx, ry, deg float64, col color.NRGBA) {
	cx, cy, rx, ry = cx*k, cy*k, rx*k, ry*k
	sin, cos := math.Sincos(deg * math.Pi / 180)
	reach := int(math.Ceil(max(rx, ry)))
	for y := int(cy) - reach; y <= int(cy)+reach; y++ {
		for x := int(cx) - reach; x <= int(cx)+reach; x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			u, v := dx*cos+dy*sin, -dx*sin+dy*cos
			if u*u/(rx*rx)+v*v/(ry*ry) <= 1 {
				c.set(x, y, col)
			}
		}
	}
}

// poly pinta um polígono (regra par-ímpar).
func (c *canvas) poly(col color.NRGBA, pts ...pt) {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range pts {
		minX, maxX = min(minX, p.x*k), max(maxX, p.x*k)
		minY, maxY = min(minY, p.y*k), max(maxY, p.y*k)
	}
	for y := int(minY); y <= int(maxY); y++ {
		for x := int(minX); x <= int(maxX); x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			inside := false
			for i, a := range pts {
				b := pts[(i+1)%len(pts)]
				ax, ay, bx, by := a.x*k, a.y*k, b.x*k, b.y*k
				if (ay > py) != (by > py) && px < (bx-ax)*(py-ay)/(by-ay)+ax {
					inside = !inside
				}
			}
			if inside {
				c.set(x, y, col)
			}
		}
	}
}

// outline pinta, atrás de tudo o que já foi desenhado, um contorno escuro
// ao redor da silhueta.
func (c *canvas) outline() {
	var edge [big * big]bool
	r := int(outlineRadius)
	for y := 0; y < big; y++ {
		for x := 0; x < big; x++ {
			if c.pix[y*big+x].A != 0 {
				continue
			}
		scan:
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					if dx*dx+dy*dy > r*r {
						continue
					}
					nx, ny := x+dx, y+dy
					if nx >= 0 && ny >= 0 && nx < big && ny < big && c.pix[ny*big+nx].A != 0 {
						edge[y*big+x] = true
						break scan
					}
				}
			}
		}
	}
	for i, e := range edge {
		if e {
			c.pix[i] = outlineColor
		}
	}
}

// image reduz o canvas para o tamanho final, fazendo a média de k×k pixels.
func (c *canvas) image() *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var r, g, b, n int
			for dy := 0; dy < k; dy++ {
				for dx := 0; dx < k; dx++ {
					p := c.pix[(y*k+dy)*big+x*k+dx]
					if p.A != 0 {
						r, g, b, n = r+int(p.R), g+int(p.G), b+int(p.B), n+1
					}
				}
			}
			if n > 0 {
				out.SetNRGBA(x, y, color.NRGBA{uint8(r / n), uint8(g / n), uint8(b / n), uint8(255 * n / (k * k))})
			}
		}
	}
	return out
}

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
	c.ellipse(cx-7, cy-7, 3, 10, -30, leafDark)
	c.ellipse(cx+7, cy-7, 3, 10, 30, leafDark)
	c.ellipse(cx, cy-9, 3.5, 11, 0, leafGreen)
}

var sprites = map[string]func(c *canvas){
	"seed": func(c *canvas) {
		c.ellipse(32, 44, 17, 8, 0, soil)
		c.ellipse(30, 41, 9, 3, 0, soilLight)
		c.ellipse(32, 39, 4, 2.6, 20, rgb(0xEBDDA6))
	},
	"sprout": func(c *canvas) {
		c.ellipse(32, 48, 15, 6.5, 0, soil)
		c.ellipse(30, 46, 8, 2.5, 0, soilLight)
		c.poly(rgb(0x4DB03E), pt{31, 48}, pt{33, 48}, pt{33, 36}, pt{31, 36})
		c.ellipse(24, 36, 8.5, 3.6, -30, rgb(0x5CCB4B))
		c.ellipse(40, 36, 8.5, 3.6, 30, rgb(0x5CCB4B))
	},
	"potato": func(c *canvas) {
		c.ellipse(32, 38, 19, 14, -15, rgb(0xB98A55))
		c.ellipse(25, 32, 7, 3, -15, rgb(0xD2A877))
		c.ellipse(26, 37, 2.2, 1.6, 0, rgb(0x7A5530))
		c.ellipse(38, 42, 2.2, 1.6, 0, rgb(0x7A5530))
		c.ellipse(32, 45, 2, 1.4, 0, rgb(0x7A5530))
		c.ellipse(40, 33, 2, 1.4, 0, rgb(0x7A5530))
	},
	"carrot": func(c *canvas) {
		leaves(c, 32, 26)
		c.poly(rgb(0xF28A1D), pt{22, 26}, pt{42, 26}, pt{32, 58})
		c.ellipse(32, 26, 10, 4, 0, rgb(0xF9A23B))
		c.ellipse(30, 34, 4, 1.1, 0, rgb(0xD56E0E))
		c.ellipse(34, 42, 3, 1, 0, rgb(0xD56E0E))
	},
	"radish": func(c *canvas) {
		leaves(c, 32, 28)
		c.poly(rgb(0xF4E9EE), pt{29, 44}, pt{35, 44}, pt{32, 59})
		c.ellipse(32, 38, 14, 13, 0, rgb(0xE0356B))
		c.ellipse(27, 32, 4, 2.4, -30, rgb(0xF27AA0))
	},
	"grass": func(c *canvas) {
		c.poly(bladeMid, pt{22, 57}, pt{29, 57}, pt{11, 28})
		c.poly(bladeMid, pt{35, 57}, pt{42, 57}, pt{53, 30})
		c.poly(bladeLight, pt{25, 57}, pt{33, 57}, pt{20, 14})
		c.poly(bladeLight, pt{31, 57}, pt{39, 57}, pt{44, 18})
		c.poly(bladeLight, pt{28, 57}, pt{36, 57}, pt{32, 8})
	},
	"clover": func(c *canvas) {
		c.poly(leafDark, pt{30.5, 34}, pt{33.5, 34}, pt{33.5, 58}, pt{30.5, 58})
		light := rgb(0x7BE090)
		for _, l := range []struct{ x, y, deg float64 }{{32, 22, 90}, {22, 36, -30}, {42, 36, 30}} {
			c.ellipse(l.x, l.y, 10, 10, 0, rgb(0x3FBF62))
			c.ellipse(l.x, l.y, 5.5, 2, l.deg, light)
		}
	},
	"alfalfa": func(c *canvas) {
		c.poly(leafDark, pt{30.5, 58}, pt{33.5, 58}, pt{33.5, 16}, pt{30.5, 16})
		for _, y := range []float64{50, 41, 32, 23} {
			c.ellipse(24, y, 6.5, 3, -35, leafGreen)
			c.ellipse(40, y, 6.5, 3, 35, leafGreen)
		}
		purple := rgb(0xA67BE8)
		c.ellipse(32, 14, 5, 5, 0, rgb(0x8E5BD6))
		c.ellipse(27, 17, 3.6, 3.6, 0, purple)
		c.ellipse(37, 17, 3.6, 3.6, 0, purple)
		c.ellipse(32, 8, 3.6, 3.6, 0, purple)
	},
}

func main() {
	dir := filepath.Join("assets", "plants")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	for name, draw := range sprites {
		var c canvas
		draw(&c)
		c.outline()

		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(f, c.image()); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
