package gopher

import (
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Cores do Gopher.
var (
	bodyColor   = color.RGBA{0x00, 0xAD, 0xD8, 255} // azul clássico do Go (#00ADD8): corpo e orelhas
	eyeColor    = color.RGBA{255, 255, 255, 255}    // branco dos olhos
	pupilColor  = color.RGBA{0, 0, 0, 255}          // pupilas e nariz
	shadowColor = color.RGBA{0, 0, 0, 100}          // sombra: preto semitransparente
)

// Deslocamento da sombra em relação ao corpo, em pixels (para a direita e para baixo).
const (
	shadowOffsetX float32 = 2
	shadowOffsetY float32 = 3
)

// part é uma elipse do desenho. As medidas são em múltiplos do raio do corpo e
// o centro é relativo ao centro do corpo, então o desenho escala com bodyRadius.
type part struct {
	cx, cy   float32
	rx, ry   float32
	angle    float32 // radianos, sentido horário na tela
	color    color.Color
	mirrored bool // também desenha a cópia espelhada (cx e angle invertidos)
}

// parts descreve o Gopher, de trás para frente.
var parts = []part{
	{cx: 1.16, cy: -0.18, rx: 0.19, ry: 0.13, angle: 0.3, color: bodyColor, mirrored: true}, // orelhas
	{cx: 0, cy: 0, rx: 1, ry: 1, angle: 0, color: bodyColor},                                // corpo
	{cx: 0.46, cy: -0.9, rx: 0.33, ry: 0.2, angle: 0.3, color: eyeColor, mirrored: true},    // olhos
	{cx: 0.58, cy: -1.05, rx: 0.06, ry: 0.09, angle: 0, color: pupilColor, mirrored: true},  // pupilas
	{cx: 0, cy: -0.97, rx: 0.08, ry: 0.06, angle: 0, color: pupilColor},                     // nariz
}

// shapes é parts com as cópias espelhadas já expandidas, na ordem de desenho.
var shapes = expand(parts)

// drawingReach é a distância máxima do desenho ao centro do corpo, em múltiplos
// do raio do corpo. Calculada a partir de shapes, então acompanha qualquer
// mudança em parts.
var drawingReach = reach(shapes)

// bodyRadius é o raio do corpo, em pixels, escolhido para que o desenho inteiro
// caiba na hit box.
var bodyRadius = hitboxRadius / drawingReach

func expand(ps []part) []part {
	var out []part
	for _, p := range ps {
		out = append(out, p)
		if p.mirrored {
			m := p
			m.cx, m.angle = -p.cx, -p.angle
			out = append(out, m)
		}
	}
	return out
}

func reach(ps []part) float32 {
	var max float64
	for _, p := range ps {
		for _, pt := range ellipsePoints(p.cx, p.cy, p.rx, p.ry, p.angle) {
			max = math.Max(max, math.Hypot(float64(pt.x), float64(pt.y)))
		}
	}
	return float32(max)
}

// spriteSize é o lado, em pixels, da imagem do Gopher: a hit box inteira.
const spriteSize = int(2 * hitboxRadius)

// sprite é a imagem do Gopher, desenhada uma única vez (na primeira chamada) e
// reaproveitada por todos os Gophers e em todos os frames. A sombra cabe na hit
// box, então também faz parte da imagem.
var sprite = sync.OnceValue(renderSprite)

func renderSprite() *ebiten.Image {
	img := ebiten.NewImage(spriteSize, spriteSize)
	c := hitboxRadius // centro do corpo na imagem

	vector.FillCircle(img, c+shadowOffsetX, c+shadowOffsetY, bodyRadius, shadowColor, true)
	for _, s := range shapes {
		fillEllipse(img, c+s.cx*bodyRadius, c+s.cy*bodyRadius, s.rx*bodyRadius, s.ry*bodyRadius, s.angle, s.color)
	}
	return img
}

// Draw pinta o Gopher em screen, centrado na posição atual.
func (g *Gopher) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.x-hitboxRadius), float64(g.y-hitboxRadius))
	screen.DrawImage(sprite(), op)
}
