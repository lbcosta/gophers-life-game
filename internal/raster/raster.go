// Package raster desenha imagens de formas simples (elipses e polígonos) em
// cores chapadas, com contorno e suavização, para os geradores de assets.
//
// As formas são pintadas numa imagem k vezes maior que a final e depois
// reduzidas fazendo a média de k×k pixels, o que dá o serrilhado suave. As
// coordenadas das formas são sempre as da imagem final.
package raster

import (
	"image"
	"image/color"
	"math"
)

// Canvas é a imagem em que as formas são pintadas. Pintar sobre o que já existe
// cobre o que estava embaixo.
type Canvas struct {
	size, k int
	big     int
	pix     []color.NRGBA
}

// New cria um Canvas que resulta numa imagem de size×size pixels, pintado com
// fator de supersampling k.
func New(size, k int) *Canvas {
	big := size * k
	return &Canvas{size: size, k: k, big: big, pix: make([]color.NRGBA, big*big)}
}

// RGB converte um valor 0xRRGGBB numa cor opaca.
func RGB(hex uint32) color.NRGBA {
	return color.NRGBA{uint8(hex >> 16), uint8(hex >> 8), uint8(hex), 255}
}

// Pt é um ponto, nas coordenadas da imagem final.
type Pt struct{ X, Y float64 }

func (c *Canvas) set(x, y int, col color.NRGBA) {
	if x >= 0 && y >= 0 && x < c.big && y < c.big {
		c.pix[y*c.big+x] = col
	}
}

// Ellipse pinta uma elipse de centro (cx, cy), raios rx e ry, girada deg graus.
func (c *Canvas) Ellipse(cx, cy, rx, ry, deg float64, col color.NRGBA) {
	k := float64(c.k)
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

// Poly pinta um polígono (regra par-ímpar).
func (c *Canvas) Poly(col color.NRGBA, pts ...Pt) {
	k := float64(c.k)
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range pts {
		minX, maxX = min(minX, p.X*k), max(maxX, p.X*k)
		minY, maxY = min(minY, p.Y*k), max(maxY, p.Y*k)
	}
	for y := int(minY); y <= int(maxY); y++ {
		for x := int(minX); x <= int(maxX); x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			inside := false
			for i, a := range pts {
				b := pts[(i+1)%len(pts)]
				ax, ay, bx, by := a.X*k, a.Y*k, b.X*k, b.Y*k
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

// Outline pinta, nos pontos transparentes, um contorno de espessura radius (em
// pixels da imagem final) ao redor de tudo o que já foi pintado.
func (c *Canvas) Outline(radius float64, col color.NRGBA) {
	r := int(math.Round(radius * float64(c.k)))
	edge := make([]bool, len(c.pix))
	for y := 0; y < c.big; y++ {
		for x := 0; x < c.big; x++ {
			if c.pix[y*c.big+x].A != 0 {
				continue
			}
		scan:
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					if dx*dx+dy*dy > r*r {
						continue
					}
					nx, ny := x+dx, y+dy
					if nx >= 0 && ny >= 0 && nx < c.big && ny < c.big && c.pix[ny*c.big+nx].A != 0 {
						edge[y*c.big+x] = true
						break scan
					}
				}
			}
		}
	}
	for i, e := range edge {
		if e {
			c.pix[i] = col
		}
	}
}

// Image reduz o Canvas ao tamanho final.
func (c *Canvas) Image() *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, c.size, c.size))
	k := c.k
	for y := 0; y < c.size; y++ {
		for x := 0; x < c.size; x++ {
			var r, g, b, n int
			for dy := 0; dy < k; dy++ {
				for dx := 0; dx < k; dx++ {
					p := c.pix[(y*k+dy)*c.big+x*k+dx]
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
