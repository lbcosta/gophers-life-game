// Package circle desenha círculos (discos e anéis) de borda suavizada copiando
// imagens prontas, em vez de usar o pacote vector a cada frame.
//
// A Ebitengine só consegue juntar desenhos vetoriais suavizados que vêm em
// sequência: cada DrawImage no meio deles (o sprite de um Gopher, por exemplo)
// a obriga a refazer, só para os círculos pendentes, todo o processo da
// suavização, que passa várias vezes por uma imagem auxiliar. Com um círculo
// por Gopher entre um sprite e outro, isso acontecia dezenas de vezes por frame
// e o jogo travava. Aqui cada círculo é desenhado com vector uma única vez, em
// branco, e depois só copiado com DrawImage, reduzido e pintado na cor pedida:
// a Ebitengine junta essas cópias com os sprites numa só chamada à placa de
// vídeo.
package circle

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// key identifica um círculo branco pré-desenhado: o raio, inteiro, e a
// espessura do traço, se for um anel (0 é um disco cheio).
type key struct {
	radius int
	stroke float32
}

// images guarda os círculos brancos já desenhados (ver whiteCircle).
var images = map[key]*ebiten.Image{}

// whiteCircle devolve o círculo branco descrito por k, desenhando-o se ainda não
// existir. A imagem tem uma folga em volta, para a suavização e o traço, e o
// centro do círculo é o centro dela.
func whiteCircle(k key) *ebiten.Image {
	if images[k] == nil {
		margin := int(math.Ceil(float64(k.stroke)/2)) + 1
		size := 2 * (k.radius + margin)
		img := ebiten.NewImage(size, size)
		center, radius := float32(size)/2, float32(k.radius)
		if k.stroke == 0 {
			vector.FillCircle(img, center, center, radius, color.White, true)
		} else {
			vector.StrokeCircle(img, center, center, radius, k.stroke, color.White, true)
		}
		images[k] = img
	}
	return images[k]
}

// draw pinta em dst um círculo de centro (cx, cy), raio r e cor clr: um disco
// cheio se stroke for 0, ou um anel com traço de espessura stroke.
func draw(dst *ebiten.Image, cx, cy, r, stroke float32, clr color.RGBA) {
	if r <= 0 {
		return
	}
	radius := int(math.Ceil(float64(r))) // o círculo nunca é ampliado, só reduzido um pouco
	img := whiteCircle(key{radius, stroke})
	half := float64(img.Bounds().Dx()) / 2
	scale := float64(r) / float64(radius)

	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-half, -half)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(cx), float64(cy))
	// O mesmo que ScaleWithColor(clr), mas sem converter clr para color.Color,
	// o que alocaria memória a cada círculo. color.RGBA já é pré-multiplicada,
	// como ColorScale espera.
	op.ColorScale.Scale(float32(clr.R)/0xff, float32(clr.G)/0xff, float32(clr.B)/0xff, float32(clr.A)/0xff)
	dst.DrawImage(img, op)
}

// Fill pinta em dst um disco de centro (cx, cy), raio r e cor clr.
func Fill(dst *ebiten.Image, cx, cy, r float32, clr color.RGBA) {
	draw(dst, cx, cy, r, 0, clr)
}

// Stroke pinta em dst um anel de centro (cx, cy), raio r e cor clr, com traço
// de espessura width.
func Stroke(dst *ebiten.Image, cx, cy, r, width float32, clr color.RGBA) {
	draw(dst, cx, cy, r, width, clr)
}

// Prerender desenha, uma única vez, os discos de todos os raios até maxRadius,
// para que o jogo não precise desenhá-los no meio da partida. Os anéis e os
// discos maiores são desenhados na primeira vez em que aparecem. Deve ser
// chamado com o jogo já rodando.
func Prerender(maxRadius float32) {
	for radius := 1; radius <= int(math.Ceil(float64(maxRadius))); radius++ {
		whiteCircle(key{radius: radius})
	}
}
