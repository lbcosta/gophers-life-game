package gopher

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Cores fixas do Gopher. A cor do corpo e das orelhas é de cada Gopher (ver Palette).
var (
	eyeColor       = color.RGBA{255, 255, 255, 255} // branco dos olhos
	pupilColor     = color.RGBA{0, 0, 0, 255}       // pupilas e nariz
	highlightColor = color.RGBA{255, 255, 255, 255} // brilho nas pupilas
	shadowColor    = color.RGBA{0, 0, 0, 100}       // sombra: preto semitransparente
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
	angle    float64 // graus, sentido horário na tela (mesma convenção de heading)
	fill     fill
	mirrored bool // também desenha a cópia espelhada (cx e angle invertidos)
}

// fill é o papel de cor de uma parte. A cor de fato é resolvida ao desenhar, já
// que o corpo e as orelhas têm a cor de cada Gopher.
type fill int

const (
	fillBody fill = iota
	fillEye
	fillPupil
	fillHighlight
)

// color devolve a cor do papel f para um Gopher de corpo body.
func (f fill) color(body color.RGBA) color.Color {
	switch f {
	case fillEye:
		return eyeColor
	case fillPupil:
		return pupilColor
	case fillHighlight:
		return highlightColor
	}
	return body
}

// parts descreve o Gopher, de trás para frente.
var parts = []part{
	{cx: 1.16, cy: -0.18, rx: 0.19, ry: 0.13, angle: 17, fill: fillBody, mirrored: true}, // orelhas
	{cx: 0, cy: 0, rx: 1, ry: 1, angle: 0, fill: fillBody},                               // corpo
	{cx: 0.46, cy: -0.9, rx: 0.33, ry: 0.2, angle: 17, fill: fillEye, mirrored: true},    // olhos
	// Pupilas redondas e grandes (cerca de 45% do olho), na borda do olho e vazando
	// cerca de 1px por ela, para dar o efeito de olhos esbugalhados.
	{cx: 0.582, cy: -0.956, rx: 0.15, ry: 0.15, angle: 0, fill: fillPupil, mirrored: true},
	// Brilhos nas pupilas. Não são espelhados: a luz vem sempre de cima e da
	// esquerda, o oposto do lado da sombra.
	{cx: -0.632, cy: -1.016, rx: 0.05, ry: 0.05, angle: 0, fill: fillHighlight},
	{cx: 0.532, cy: -1.016, rx: 0.05, ry: 0.05, angle: 0, fill: fillHighlight},
	{cx: 0, cy: -0.97, rx: 0.08, ry: 0.06, angle: 0, fill: fillPupil}, // nariz
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
	var farthest float64
	for _, p := range ps {
		for _, pt := range ellipsePoints(p.cx, p.cy, p.rx, p.ry, p.angle) {
			farthest = max(farthest, math.Hypot(float64(pt.x), float64(pt.y)))
		}
	}
	return float32(farthest)
}

// spriteSize é o lado, em pixels, da imagem do Gopher: a hit box inteira.
const spriteSize = 2 * hitboxRadius

// spriteCenter é a coordenada, nos dois eixos, do centro do corpo no sprite.
const spriteCenter = spriteSize / 2

// spriteHeading é a direção para a qual o sprite aponta ao ser desenhado: para
// cima, que na convenção dos ângulos (0º à direita, sentido horário) é -90º.
const spriteHeading = -90

// renderSprite desenha o Gopher (sem a sombra, que não gira junto) com o corpo
// na cor body.
func renderSprite(body color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(spriteSize, spriteSize)
	const c = spriteCenter // centro do corpo na imagem

	for _, s := range shapes {
		fillEllipse(img, c+s.cx*bodyRadius, c+s.cy*bodyRadius, s.rx*bodyRadius, s.ry*bodyRadius, s.angle, s.fill.color(body))
	}
	return img
}

// Draw pinta o Gopher em screen, centrado na posição atual e apontando para
// facing. A sombra fica sempre no mesmo lado, já que a luz não gira com ele.
func (g *Gopher) Draw(screen *ebiten.Image) {
	vector.FillCircle(screen, float32(g.x)+shadowOffsetX, float32(g.y)+shadowOffsetY, bodyRadius, shadowColor, true)

	// O sprite é desenhado na primeira vez em que é preciso, já na cor do Gopher.
	if g.sprite == nil {
		g.sprite = renderSprite(g.body)
	}

	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-spriteCenter, -spriteCenter) // gira em torno do centro do corpo
	op.GeoM.Rotate(radians(g.facing - spriteHeading))
	op.GeoM.Translate(g.x, g.y)
	screen.DrawImage(g.sprite, op)
}
