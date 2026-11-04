// Comando genprofiles desenha as fotos de perfil dos Gophers: um Gopher de
// frente, no centro da imagem, para cada cor da paleta e cada emoção. Rode a
// partir da raiz do projeto:
//
//	go run ./cmd/genprofiles
//
// As fotos são gravadas em assets/profiles, com o nome <cor>_<emoção>.png. Com
// -sheet arquivo.png, grava também uma folha com todas elas lado a lado.
package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"gophers-life-game/gopher"
	"gophers-life-game/internal/raster"
)

const (
	size = 256 // lado da foto, em pixels
	k    = 3   // fator de supersampling

	lineWidth = 1.7 // espessura do contorno, em pixels
)

var (
	ink   = color.NRGBA{34, 26, 22, 255}    // contorno e pupilas
	white = color.NRGBA{255, 255, 255, 255} // olhos e dentes
	tan   = raster.RGB(0xF5D2A5)            // patas e focinho
	black = color.NRGBA{10, 10, 10, 255}    // nariz, pupilas e dentro das orelhas
)

// O Gopher é desenhado nas medidas do desenho de referência (cerca de 1470 por
// 2000), e view as converte para a foto: reduz tudo à altura da foto e centra.
type view struct{ x0, y0, s float64 }

var v = view{x0: (size - 1469*0.126) / 2, y0: 2, s: 0.126}

func (v view) x(x float64) float64 { return v.x0 + x*v.s }
func (v view) y(y float64) float64 { return v.y0 + y*v.s }

// ellipse pinta uma elipse nas medidas do desenho de referência, com um
// contorno escuro ao redor se outline for verdadeiro.
func ellipse(c *raster.Canvas, cx, cy, rx, ry, deg float64, fill color.NRGBA, outline bool) {
	if outline {
		c.Ellipse(v.x(cx), v.y(cy), rx*v.s+lineWidth, ry*v.s+lineWidth, deg, ink)
	}
	c.Ellipse(v.x(cx), v.y(cy), rx*v.s, ry*v.s, deg, fill)
}

func poly(c *raster.Canvas, fill color.NRGBA, pts ...[2]float64) {
	out := make([]raster.Pt, len(pts))
	for i, p := range pts {
		out[i] = raster.Pt{X: v.x(p[0]), Y: v.y(p[1])}
	}
	c.Poly(fill, out...)
}

// eyeShape é um olho: centro, raios e giro, nas medidas do desenho de referência.
// Os dois olhos são diferentes, e é isso que dá o ar desajeitado do Gopher.
type eyeShape struct{ cx, cy, rx, ry, deg float64 }

var eyeShapes = [2]eyeShape{
	{473, 305, 188, 175, 5},
	{940, 283, 190, 178, -8},
}

// open desenha um olho aberto de e, com a pupila de tamanho pupil (1 é o do
// desenho de referência) deslocada de (dx, dy), em frações dos raios do olho.
func open(c *raster.Canvas, e eyeShape, scale, pupil, dx, dy float64) {
	ellipse(c, e.cx, e.cy, e.rx*scale, e.ry*scale, e.deg, white, true)
	px, py := e.cx+dx*e.rx, e.cy+dy*e.ry
	ellipse(c, px, py, 55*pupil, 60*pupil, 0, black, false)
	ellipse(c, px+23*pupil, py+12*pupil, 14*pupil, 14*pupil, 0, white, false)
}

// Emotion é uma emoção, mostrada só pelo formato dos olhos (sem sobrancelhas).
type Emotion struct {
	Kind gopher.Emotion
	eye  func(c *raster.Canvas, body color.NRGBA, e eyeShape)
}

// Emotions desenha cada emoção do jogo, na ordem de gopher.Emotions.
var Emotions = []Emotion{
	// Olhando para o lado, como no desenho de referência.
	{gopher.Neutral, func(c *raster.Canvas, _ color.NRGBA, e eyeShape) { open(c, e, 1, 1, -0.48, 0.06) }},
	{gopher.Surprised, func(c *raster.Canvas, _ color.NRGBA, e eyeShape) { open(c, e, 1.08, 0.5, 0, 0) }},
	{gopher.Suspicious, func(c *raster.Canvas, _ color.NRGBA, e eyeShape) { open(c, e, 1, 1, 0.5, 0.05) }},
	{gopher.Sleepy, func(c *raster.Canvas, body color.NRGBA, e eyeShape) {
		open(c, e, 1, 1, -0.2, 0.4)
		// Pálpebra: a cor do corpo cobre a metade de cima do olho.
		l, r, top, mid := e.cx-e.rx-18, e.cx+e.rx+18, e.cy-e.ry-24, e.cy-0.05*e.ry
		poly(c, body, [2]float64{l, top}, [2]float64{r, top}, [2]float64{r, mid}, [2]float64{l, mid})
		l, r = e.cx-e.rx-6, e.cx+e.rx+6
		poly(c, ink, [2]float64{l, mid - 6}, [2]float64{r, mid - 6}, [2]float64{r, mid + 14}, [2]float64{l, mid + 14})
	}},
	{gopher.Happy, func(c *raster.Canvas, body color.NRGBA, e eyeShape) {
		// Olho fechado em arco: um círculo escuro menos outro, mais baixo.
		ellipse(c, e.cx, e.cy+0.1*e.ry, 0.82*e.rx, 0.82*e.ry, 0, ink, false)
		ellipse(c, e.cx, e.cy+0.1*e.ry+0.34*e.ry, 0.82*e.rx, 0.82*e.ry, 0, body, false)
	}},
}

// tooth pinta um dente: um retângulo de bordas de baixo arredondadas, com contorno.
func tooth(c *raster.Canvas, cx, hw, top, bottom float64) {
	r := hw
	for _, outline := range []bool{true, false} {
		fill, g := white, 0.0
		if outline {
			fill, g = ink, lineWidth/v.s
		}
		poly(c, fill, [2]float64{cx - hw - g, top - g}, [2]float64{cx + hw + g, top - g},
			[2]float64{cx + hw + g, bottom - r}, [2]float64{cx - hw - g, bottom - r})
		c.Ellipse(v.x(cx), v.y(bottom-r), (hw+g)*v.s, (r+g)*v.s, 0, fill)
	}
}

// render desenha o Gopher de frente, de corpo body, com a emoção e.
func render(body color.NRGBA, e Emotion) *image.NRGBA {
	c := raster.New(size, k)

	// Atrás do corpo: braços, pés e orelhas.
	ellipse(c, 85, 1050, 95, 55, -12, tan, true)
	ellipse(c, 1385, 1045, 95, 55, 12, tan, true)
	ellipse(c, 290, 1895, 140, 68, -33, tan, true)
	ellipse(c, 1180, 1895, 140, 68, 33, tan, true)
	ellipse(c, 185, 250, 160, 125, -30, body, true)
	ellipse(c, 158, 240, 55, 42, 20, black, false)
	ellipse(c, 1280, 220, 130, 128, 20, body, true)
	ellipse(c, 1262, 205, 40, 32, 20, black, false)

	// Corpo: um retângulo alto de cantos arredondados, com os de cima mais
	// abertos (a cabeça é larga e achatada) e os de baixo mais fechados.
	for _, outline := range []bool{true, false} {
		fill, g := body, 0.0
		if outline {
			fill, g = ink, lineWidth/v.s
		}
		const left, right, topY, botY = 117, 1341, 0, 1965
		const topRx, topRy, botRx, botRy = 480, 450, 520, 430
		ellipse(c, left+topRx, topY+topRy, topRx+g, topRy+g, 0, fill, false)
		ellipse(c, right-topRx, topY+topRy, topRx+g, topRy+g, 0, fill, false)
		ellipse(c, left+botRx, botY-botRy, botRx+g, botRy+g, 0, fill, false)
		ellipse(c, right-botRx, botY-botRy, botRx+g, botRy+g, 0, fill, false)
		poly(c, fill, [2]float64{left - g, topY + topRy}, [2]float64{right + g, topY + topRy},
			[2]float64{right + g, botY - botRy}, [2]float64{left - g, botY - botRy})
		poly(c, fill, [2]float64{left + topRx, topY - g}, [2]float64{right - topRx, topY - g},
			[2]float64{right - topRx, topY + topRy}, [2]float64{left + topRx, topY + topRy})
		poly(c, fill, [2]float64{left + botRx, botY - botRy}, [2]float64{right - botRx, botY - botRy},
			[2]float64{right - botRx, botY + g}, [2]float64{left + botRx, botY + g})
	}

	// Olhos, grandes e desencontrados.
	e.eye(c, body, eyeShapes[0])
	e.eye(c, body, eyeShapes[1])

	// Dentes, focinho e nariz.
	tooth(c, 673, 41, 590, 715)
	tooth(c, 766, 39, 585, 718)
	ellipse(c, 730, 545, 152, 64, -4, tan, true)
	ellipse(c, 724, 457, 82, 52, -3, black, false)

	return c.Image()
}

func main() {
	sheetPath := flag.String("sheet", "", "grava também uma folha com todas as fotos nesse arquivo")
	flag.Parse()

	dir := filepath.Join("assets", "profiles")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}

	var sheet *image.NRGBA
	if *sheetPath != "" {
		sheet = image.NewNRGBA(image.Rect(0, 0, size*len(Emotions), size*len(gopher.Palette)))
		draw.Draw(sheet, sheet.Bounds(), image.NewUniform(color.NRGBA{40, 80, 30, 255}), image.Point{}, draw.Src)
	}

	for row, rgba := range gopher.Palette {
		body := color.NRGBA{rgba.R, rgba.G, rgba.B, 255}
		for col, e := range Emotions {
			img := render(body, e)
			write(filepath.Join(dir, gopher.PhotoName(rgba, e.Kind)), img)
			if sheet != nil {
				draw.Draw(sheet, image.Rect(col*size, row*size, (col+1)*size, (row+1)*size), img, image.Point{}, draw.Over)
			}
		}
	}
	if sheet != nil {
		write(*sheetPath, sheet)
	}
}

func write(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
