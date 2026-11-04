// Package gopher implementa os Gophers do jogo: personagens que andam sozinhos,
// quase sempre em linha reta, fazendo curvas ocasionais e quicando nas bordas
// da tela e uns nos outros. Cada Gopher tem velocidade e cor próprias, e
// Colony reúne vários deles na mesma tela.
package gopher

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// walkEveryTicks é de quantos em quantos ticks o Gopher dá um passo.
// Com 60 ticks por segundo, 7 ticks equivalem a cerca de 8,6 passos por segundo.
const walkEveryTicks = 7

// hitboxRadius é o raio da hit box do Gopher: um círculo de 64px de diâmetro,
// centrado no corpo, que contém o desenho inteiro (corpo, orelhas e olhos).
const hitboxRadius = 32

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

// Gopher anda pela tela mantendo a hit box sempre inteira dentro dela e quica
// nos outros Gophers. A posição é o centro do corpo.
type Gopher struct {
	x, y          float64
	width, height float64 // tamanho da tela
	ticks         int     // ticks desde a criação; controla o ritmo dos passos

	stepSize float64    // pixels por passo; vem da Speed
	body     color.RGBA // cor do corpo e das orelhas
	sprite   *ebiten.Image

	heading float64 // direção do movimento, em graus [0, 360)
	facing  float64 // direção para a qual o desenho aponta; gira até alcançar heading

	// Fontes de sorteio. São rand.Float64 e rand.IntN, mas os testes informam
	// as suas por newGopher ou trocam depois.
	chance func() float64  // número em [0, 1)
	intn   func(n int) int // inteiro em [0, n)
}

// NewGopher cria um Gopher de velocidade speed e corpo da cor body, no canto
// superior esquerdo da tela, olhando para uma direção aleatória. O tamanho da
// tela é informado depois, por Resize.
func NewGopher(speed Speed, body color.RGBA) *Gopher {
	g := newGopher(rand.Float64, rand.IntN)
	g.stepSize = speed.step()
	g.body = body
	return g
}

// newGopher é NewGopher com as fontes de sorteio informadas, para que os
// testes controlem também o sorteio da direção inicial. O Gopher criado tem
// velocidade Normal e a primeira cor da paleta.
func newGopher(chance func() float64, intn func(n int) int) *Gopher {
	heading := float64(intn(360))
	return &Gopher{
		x:        hitboxRadius,
		y:        hitboxRadius,
		stepSize: Normal.step(),
		body:     Palette[0],
		heading:  heading,
		facing:   heading,
		chance:   chance,
		intn:     intn,
	}
}

// Position devolve o centro do corpo do Gopher.
func (g *Gopher) Position() (x, y float64) {
	return g.x, g.y
}

// Resize informa o tamanho da tela. Se o Gopher ficar fora dela, é trazido de
// volta para dentro. Chamar com o mesmo tamanho não tem efeito.
func (g *Gopher) Resize(width, height float64) {
	g.width, g.height = width, height
	g.keepInside()
}

// spawnAttempts é quantas posições Spawn tenta antes de aceitar uma que se
// sobreponha a outro Gopher.
const spawnAttempts = 100

// Spawn põe o Gopher em uma posição aleatória da tela, com a hit box inteira
// dentro dela e, se possível, sem se sobrepor aos Gophers em others. Defasa
// também o ritmo dos passos, para que vários Gophers não andem em sincronia.
// Deve ser chamado depois de Resize, que informa o tamanho da tela.
func (g *Gopher) Spawn(others ...*Gopher) {
	for range spawnAttempts {
		g.x = hitboxRadius + g.chance()*(g.width-2*hitboxRadius)
		g.y = hitboxRadius + g.chance()*(g.height-2*hitboxRadius)
		g.keepInside() // necessário quando a tela é pequena demais
		if !g.overlapsAny(others) {
			break
		}
	}
	g.ticks = g.intn(walkEveryTicks)
}

// overlapsAny informa se a hit box de g se sobrepõe à de algum Gopher em others.
func (g *Gopher) overlapsAny(others []*Gopher) bool {
	for _, o := range others {
		if o != g && math.Hypot(g.x-o.x, g.y-o.y) < collisionDistance {
			return true
		}
	}
	return false
}

// Update deve ser chamado uma vez por tick do jogo. A cada walkEveryTicks
// ticks o Gopher dá um passo, e a cada tick o desenho gira um pouco em direção
// ao sentido do movimento. others são os Gophers com quem ele pode colidir; o
// próprio Gopher pode estar entre eles, e é ignorado.
func (g *Gopher) Update(others ...*Gopher) {
	g.ticks++
	if g.ticks%walkEveryTicks == 0 {
		g.move(others...)
	}
	g.turnTowardsHeading()
}

// move dá um passo.
//
// Primeiro sorteia se o Gopher gira (com chance turnChance). Se girar, sorteia
// o giro em [-maxTurn, maxTurn] graus. Depois anda stepSize pixels na direção
// resultante.
//
// Se o passo levaria a hit box de g a colidir com a de outro Gopher, g reage
// como sorteado por interact. Se levaria a hit box para fora da tela, o Gopher
// quica na parede: a componente do movimento perpendicular a ela é invertida.
func (g *Gopher) move(others ...*Gopher) {
	if g.chance() < turnChance {
		g.heading = normalize(g.heading + float64(g.intn(2*maxTurn+1)-maxTurn))
	}

	sin, cos := math.Sincos(radians(g.heading))
	dx, dy := g.stepSize*cos, g.stepSize*sin

	for _, o := range others {
		if o != g && g.collides(o, dx, dy) {
			dx, dy = g.interact(o, dx, dy)
			g.heading = normalize(degrees(math.Atan2(dy, dx)))
		}
	}

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
	g.keepInside() // necessário quando a tela é pequena demais para quicar
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

func (g *Gopher) fitsX(x float64) bool {
	return x-hitboxRadius >= 0 && x+hitboxRadius <= g.width
}

func (g *Gopher) fitsY(y float64) bool {
	return y-hitboxRadius >= 0 && y+hitboxRadius <= g.height
}

// keepInside traz a posição para dentro da tela, caso tenha saído.
func (g *Gopher) keepInside() {
	g.x = clamp(g.x, hitboxRadius, g.width-hitboxRadius)
	g.y = clamp(g.y, hitboxRadius, g.height-hitboxRadius)
}

// clamp limita v ao intervalo [lo, hi]. Se hi < lo, devolve lo.
func clamp(v, lo, hi float64) float64 {
	if hi < lo {
		return lo
	}
	return min(max(v, lo), hi)
}

// radians converte um ângulo de graus para radianos.
func radians(deg float64) float64 {
	return deg * math.Pi / 180
}

// degrees converte um ângulo de radianos para graus.
func degrees(rad float64) float64 {
	return rad * 180 / math.Pi
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
