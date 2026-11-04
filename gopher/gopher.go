// Package gopher implementa o Gopher do jogo: um personagem que anda sozinho,
// quase sempre em linha reta, fazendo curvas ocasionais e quicando nas bordas
// da tela.
package gopher

import (
	"math"
	"math/rand/v2"
)

// step é a distância, em pixels, percorrida a cada passo.
const step float32 = 4

// walkEveryTicks é de quantos em quantos ticks o Gopher dá um passo.
// Com 60 ticks por segundo, 7 ticks equivalem a cerca de 8,6 passos por segundo.
const walkEveryTicks = 7

// hitboxRadius é o raio da hit box do Gopher: um círculo de 64px de diâmetro,
// centrado no corpo, que contém o desenho inteiro (corpo, orelhas e olhos).
const hitboxRadius float32 = 32

// Parâmetros da caminhada. Os ângulos são em graus, com 0º apontando para a
// direita e crescendo no sentido horário na tela (o eixo Y cresce para baixo).
const (
	// turnChance é a chance, a cada passo, de o Gopher girar. Nos demais passos
	// ele segue em frente.
	turnChance = 0.15

	// maxTurn é o maior giro, para cada lado, em um único passo. Com no máximo
	// 180º o sinal do giro coincide com o caminho mais curto, então o desenho
	// gira no mesmo sentido do giro sorteado.
	maxTurn = 90

	// turnSpeed é quanto o desenho gira por tick até alcançar a direção do movimento.
	turnSpeed = 6.0
)

// Gopher anda pela tela mantendo a hit box sempre inteira dentro dela.
// A posição é o centro do corpo.
type Gopher struct {
	x, y          float32
	width, height float32 // tamanho da tela

	heading float64 // direção do movimento, em graus [0, 360)
	facing  float64 // direção para a qual o desenho aponta; gira até alcançar heading

	// Fontes de sorteio. São rand.Float64 e rand.IntN, mas podem ser trocadas nos testes.
	chance func() float64  // número em [0, 1)
	intn   func(n int) int // inteiro em [0, n)
}

// NewGopher cria um Gopher no canto superior esquerdo de uma tela de
// width por height pixels, já com a hit box inteira dentro dela, olhando para
// uma direção aleatória.
func NewGopher(width, height float32) *Gopher {
	heading := float64(rand.IntN(360))
	return &Gopher{
		x:       hitboxRadius,
		y:       hitboxRadius,
		width:   width,
		height:  height,
		heading: heading,
		facing:  heading,
		chance:  rand.Float64,
		intn:    rand.IntN,
	}
}

// Position devolve o centro do corpo do Gopher.
func (g *Gopher) Position() (x, y float32) {
	return g.x, g.y
}

// Resize informa o novo tamanho da tela. Se o Gopher ficar fora dela, é trazido
// de volta para dentro.
func (g *Gopher) Resize(width, height float32) {
	g.width, g.height = width, height
	g.keepInside()
}

// Update deve ser chamado uma vez por tick do jogo, com o contador global de
// ticks. A cada walkEveryTicks ticks o Gopher dá um passo, e a cada tick o
// desenho gira um pouco em direção ao sentido do movimento.
func (g *Gopher) Update(ticks int) {
	if ticks%walkEveryTicks == 0 {
		g.move()
	}
	g.turnTowardsHeading()
}

// move dá um passo.
//
// Primeiro sorteia se o Gopher gira (com chance turnChance). Se girar, sorteia
// o giro em [-maxTurn, maxTurn] graus. Depois anda step pixels na direção
// resultante. Se o passo levaria a hit box para fora da tela, o Gopher quica
// na parede: a componente do movimento perpendicular a ela é invertida.
func (g *Gopher) move() {
	if g.chance() < turnChance {
		g.heading = normalize(g.heading + float64(g.intn(2*maxTurn+1)-maxTurn))
	}

	rad := g.heading * math.Pi / 180
	dx := step * float32(math.Cos(rad))
	dy := step * float32(math.Sin(rad))

	if !g.fitsX(g.x + dx) { // parede esquerda ou direita
		g.heading = normalize(180 - g.heading)
		dx = -dx
	}
	if !g.fitsY(g.y + dy) { // parede de cima ou de baixo
		g.heading = normalize(-g.heading)
		dy = -dy
	}

	g.x += dx
	g.y += dy
	g.keepInside()
}

// turnTowardsHeading gira facing até turnSpeed graus, pelo caminho mais curto,
// em direção a heading.
func (g *Gopher) turnTowardsHeading() {
	diff := shortestTurn(g.facing, g.heading)
	if math.Abs(diff) <= turnSpeed {
		g.facing = g.heading
		return
	}
	g.facing = normalize(g.facing + math.Copysign(turnSpeed, diff))
}

func (g *Gopher) fitsX(x float32) bool {
	return x-hitboxRadius >= 0 && x+hitboxRadius <= g.width
}

func (g *Gopher) fitsY(y float32) bool {
	return y-hitboxRadius >= 0 && y+hitboxRadius <= g.height
}

// keepInside traz a posição para dentro da tela, caso tenha saído.
func (g *Gopher) keepInside() {
	g.x = clamp(g.x, hitboxRadius, g.width-hitboxRadius)
	g.y = clamp(g.y, hitboxRadius, g.height-hitboxRadius)
}

// clamp limita v ao intervalo [lo, hi]. Se hi < lo, devolve lo.
func clamp(v, lo, hi float32) float32 {
	if hi < lo {
		return lo
	}
	return min(max(v, lo), hi)
}

// normalize leva um ângulo, em graus, para o intervalo [0, 360).
func normalize(deg float64) float64 {
	deg = math.Mod(deg, 360)
	if deg < 0 {
		deg += 360
	}
	return deg
}

// shortestTurn devolve o giro, em graus, de from até to pelo caminho mais curto:
// um valor em [-180, 180). Positivo é sentido horário na tela.
func shortestTurn(from, to float64) float64 {
	return normalize(to-from+180) - 180
}
