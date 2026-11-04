package gopher

import (
	"fmt"
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
	mirrored bool  // também desenha a cópia espelhada (cx e angle invertidos)
	from     Stage // primeira fase em que a parte aparece; o zero (Baby) é o corpo
	shiny    bool  // recebe um brilho (ver highlightCenters); só para partes circulares
}

// fill é o papel de cor de uma parte. A cor de fato é resolvida ao desenhar, já
// que o corpo e as orelhas têm a cor de cada Gopher.
type fill int

const (
	fillBody fill = iota
	fillEye
	fillPupil
)

// color devolve a cor do papel f para um Gopher de corpo body.
func (f fill) color(body color.RGBA) color.Color {
	switch f {
	case fillBody:
		return body
	case fillEye:
		return eyeColor
	case fillPupil:
		return pupilColor
	}
	panic(fmt.Sprintf("gopher: papel de cor desconhecido: %d", f))
}

// parts descreve o Gopher, de trás para frente.
var parts = []part{
	{cx: 1.16, cy: -0.18, rx: 0.19, ry: 0.13, angle: 17, fill: fillBody, from: Adult, mirrored: true}, // orelhas
	{cx: 0, cy: 0, rx: 1, ry: 1, angle: 0, fill: fillBody},                                            // corpo
	{cx: 0.46, cy: -0.9, rx: 0.33, ry: 0.2, angle: 17, fill: fillEye, from: Teen, mirrored: true},     // olhos
	// Pupilas redondas e grandes (cerca de 45% do olho), na borda do olho e vazando
	// cerca de 1px por ela, para dar o efeito de olhos esbugalhados. Cada uma
	// recebe um brilho, desenhado à parte (ver highlightCenters).
	{cx: 0.582, cy: -0.956, rx: 0.15, ry: 0.15, angle: 0, fill: fillPupil, from: Teen, mirrored: true, shiny: true},
	{cx: 0, cy: -0.97, rx: 0.08, ry: 0.06, angle: 0, fill: fillPupil, from: Adult}, // nariz
}

// shapes é parts com as cópias espelhadas já expandidas, na ordem de desenho.
var shapes = expand(parts)

// drawingReach é a distância máxima do desenho ao centro do corpo, em múltiplos
// do raio do corpo. Calculada a partir de shapes, então acompanha qualquer
// mudança em parts.
var drawingReach = reach(shapes)

// stageReach é, para cada fase, a distância máxima do desenho ao centro do
// corpo, em múltiplos do raio do corpo: só conta as partes que já apareceram.
var stageReach = reachByStage(shapes)

func reachByStage(ps []part) [numStages]float32 {
	var out [numStages]float32
	for stage := range out {
		var visible []part
		for _, p := range ps {
			if p.from <= Stage(stage) {
				visible = append(visible, p)
			}
		}
		out[stage] = reach(visible)
	}
	return out
}

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

// Brilho das pupilas. A luz vem sempre de cima e da esquerda, o oposto do lado
// da sombra, então o brilho não pode ir no sprite, que gira com o Gopher: ele é
// desenhado a cada frame, sobre as pupilas já giradas, deslocado sempre para o
// mesmo lado da tela. As medidas são em múltiplos do raio do corpo.
const highlightRadius = 0.05

var highlightOffset = point{x: -0.05, y: -0.06}

// shinyCenters são os centros das partes com brilho, no sprite (apontando para
// spriteHeading) e em relação ao centro do corpo.
var shinyCenters = centersOfShiny(shapes)

func centersOfShiny(ps []part) []point {
	var out []point
	for _, p := range ps {
		if p.shiny {
			out = append(out, point{p.cx, p.cy})
		}
	}
	return out
}

// highlightCenters devolve o centro do brilho de cada pupila de um Gopher que
// aponta para facing, em relação ao centro do corpo e em múltiplos do raio do
// corpo. A pupila gira com o Gopher; o deslocamento do brilho, não.
func highlightCenters(facing float64) []point {
	sin, cos := math.Sincos(radians(facing - spriteHeading))
	out := make([]point, len(shinyCenters))
	for i, c := range shinyCenters {
		p := c.rotate(float32(sin), float32(cos))
		out[i] = point{p.x + highlightOffset.x, p.y + highlightOffset.y}
	}
	return out
}

// renderSprite desenha o Gopher (sem a sombra e o brilho, que não giram junto)
// com o corpo na cor body, só com as partes que já aparecem em stage.
func renderSprite(body color.RGBA, stage Stage) *ebiten.Image {
	img := ebiten.NewImage(spriteSize, spriteSize)
	for _, s := range shapes {
		if s.from > stage {
			continue
		}
		fillEllipse(img, spriteCenter+s.cx*bodyRadius, spriteCenter+s.cy*bodyRadius,
			s.rx*bodyRadius, s.ry*bodyRadius, s.angle, s.fill.color(body))
	}
	return img
}

// Draw pinta o Gopher em screen: a sombra e, por cima, o próprio Gopher. Para
// desenhar vários Gophers sem que a sombra de um cubra outro, use Colony.
func (g *Gopher) Draw(screen *ebiten.Image) {
	g.drawShadow(screen)
	g.drawFigure(screen)
}

// drawShadow pinta a sombra do Gopher. Ela fica sempre no mesmo lado, já que a
// luz não gira com ele.
func (g *Gopher) drawShadow(screen *ebiten.Image) {
	vector.FillCircle(screen, float32(g.x)+shadowOffsetX, float32(g.y)+shadowOffsetY, g.bodySize(), shadowColor, true)
}

// drawFigure pinta o Gopher centrado na posição atual. Bebês e crianças são uma
// bolinha da cor do corpo. A partir da adolescência é o desenho da fase,
// reduzido ao tamanho do corpo e apontando para facing, com o brilho nas
// pupilas.
func (g *Gopher) drawFigure(screen *ebiten.Image) {
	stage, size := g.Stage(), g.bodySize()
	if stage < Teen {
		vector.FillCircle(screen, float32(g.x), float32(g.y), size, g.body, true)
		return
	}

	// O sprite de cada fase é desenhado na primeira vez em que é preciso, já na
	// cor do Gopher.
	if g.sprites[stage] == nil {
		g.sprites[stage] = renderSprite(g.body, stage)
	}

	scale := float64(size / bodyRadius)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-spriteCenter, -spriteCenter) // gira em torno do centro do corpo
	op.GeoM.Scale(scale, scale)
	op.GeoM.Rotate(radians(g.facing - spriteHeading))
	op.GeoM.Translate(g.x, g.y)
	screen.DrawImage(g.sprites[stage], op)

	for _, h := range highlightCenters(g.facing) {
		vector.FillCircle(screen, float32(g.x)+h.x*size, float32(g.y)+h.y*size,
			highlightRadius*size, highlightColor, true)
	}
}
