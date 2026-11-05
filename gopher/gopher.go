// Package gopher implementa os Gophers do jogo: personagens que andam sozinhos,
// quase sempre em linha reta, fazendo curvas ocasionais e quicando nas bordas
// da tela e uns nos outros. Cada Gopher tem velocidade e cor próprias, e
// Colony reúne vários deles na mesma tela.
package gopher

import (
	"image/color"
	"math"
	"math/rand/v2"
)

// walkEveryTicks é de quantos em quantos ticks o Gopher dá um passo.
// Com 60 ticks por segundo, 7 ticks equivalem a cerca de 8,6 passos por segundo.
const walkEveryTicks = 1

// hitboxRadius é o raio da hit box do Gopher adulto: um círculo de 64px de
// diâmetro, centrado no corpo, que contém o desenho inteiro (corpo, orelhas e
// olhos). Antes de adulto a hit box é menor (ver hitbox).
const hitboxRadius = 32

// Saciedade. O Gopher nasce com initialSatiety e perde 1 ponto a cada
// stepsPerSatiety passos. Quando chega a zero, ele morre. Comer devolve pontos
// (ver Eat).
const (
	initialSatiety  = 10
	stepsPerSatiety = 1000
)

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
// nos outros Gophers. A posição é o centro do corpo. Ele nasce bebê e cresce
// com a idade (ver Stage), e a hit box acompanha o crescimento.
type Gopher struct {
	x, y          float64
	width, height float64  // tamanho da tela
	ticks         int      // controla o ritmo dos passos; Spawn o defasa, então não serve de idade
	age           int      // ticks vividos
	maturity      maturity // idades em que entra em cada fase da vida
	satiety       int      // quando chega a zero, o Gopher morre
	steps         int      // passos dados; a cada stepsPerSatiety, perde 1 de saciedade

	speed    Speed      // velocidade sorteada no nascimento
	stepSize float64    // pixels por passo; vem da Speed
	body     color.RGBA // cor do corpo e das orelhas
	name     string     // sorteado de Names no nascimento
	emotion  Emotion    // sorteada no nascimento

	heading float64 // direção do movimento, em graus [0, 360)
	facing  float64 // direção para a qual o desenho aponta; gira até alcançar heading

	// Fontes de sorteio. São rand.Float64 e rand.IntN, mas os testes informam
	// as suas por newGopher ou trocam depois.
	chance func() float64  // número em [0, 1)
	intn   func(n int) int // inteiro em [0, n)
}

// NewGopher cria um Gopher bebê de velocidade speed e corpo da cor body, no
// canto superior esquerdo da tela, olhando para uma direção aleatória, com um
// nome sorteado de Names. A idade
// em que ele passa para cada fase da vida é sorteada (ver newMaturity). O
// tamanho da tela é informado depois, por Resize.
func NewGopher(speed Speed, body color.RGBA) *Gopher {
	g := newGopher(rand.Float64, rand.IntN)
	g.maturity = newMaturity(g.intn)
	g.age = 0
	g.name = Names[g.intn(len(Names))]
	g.emotion = Emotion(g.intn(numEmotions))
	g.speed = speed
	g.stepSize = speed.step()
	g.body = body
	return g
}

// newGopher é NewGopher com as fontes de sorteio informadas, para que os
// testes controlem também o sorteio da direção inicial. O Gopher criado tem
// velocidade Normal, a primeira cor da paleta e maturidade no meio dos
// intervalos, e já é adulto, que é o que a maioria dos testes precisa.
func newGopher(chance func() float64, intn func(n int) int) *Gopher {
	heading := float64(intn(360))
	g := &Gopher{
		x:        hitboxRadius,
		y:        hitboxRadius,
		maturity: middleMaturity(),
		satiety:  initialSatiety,
		speed:    Normal,
		stepSize: Normal.step(),
		body:     Palette[0],
		heading:  heading,
		facing:   heading,
		chance:   chance,
		intn:     intn,
	}
	g.age = g.maturity[len(g.maturity)-1]
	return g
}

// Name devolve o nome do Gopher.
func (g *Gopher) Name() string {
	return g.name
}

// Alive informa se o Gopher está vivo, isto é, se ainda tem saciedade.
func (g *Gopher) Alive() bool {
	return g.satiety > 0
}

// Satiety devolve a saciedade atual do Gopher.
func (g *Gopher) Satiety() int {
	return g.satiety
}

// Eat devolve ao Gopher n pontos de saciedade. Um Gopher morto não come.
func (g *Gopher) Eat(n int) {
	if g.Alive() {
		g.satiety += n
	}
}

// Radius devolve o raio da hit box do Gopher na idade atual.
func (g *Gopher) Radius() float64 {
	return g.hitbox()
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
	g.spawn(everyone(others))
}

// spawn é Spawn, mas pergunta a n quem está perto de cada posição sorteada, em
// vez de comparar com todos os Gophers.
func (g *Gopher) spawn(n neighbors) {
	for range spawnAttempts {
		r := g.hitbox()
		g.x = r + g.chance()*(g.width-2*r)
		g.y = r + g.chance()*(g.height-2*r)
		g.keepInside() // necessário quando a tela é pequena demais
		if !g.overlapsAny(n.around(g)) {
			break
		}
	}
	g.ticks = g.intn(walkEveryTicks)
}

// overlapsAny informa se a hit box de g se sobrepõe à de algum Gopher em others.
func (g *Gopher) overlapsAny(others []*Gopher) bool {
	for _, o := range others {
		if o != g && math.Hypot(g.x-o.x, g.y-o.y) < g.touchDistance(o) {
			return true
		}
	}
	return false
}

// Update deve ser chamado uma vez por tick do jogo. A cada walkEveryTicks
// ticks o Gopher dá um passo, e a cada tick o desenho gira um pouco em direção
// ao sentido do movimento. Um Gopher morto não faz nada. others são os Gophers com quem ele pode colidir; o
// próprio Gopher pode estar entre eles, e é ignorado.
func (g *Gopher) Update(others ...*Gopher) {
	g.update(everyone(others))
}

// update é Update, mas pergunta a n com quem o Gopher pode colidir, e só quando
// ele vai dar um passo, que é quando precisa saber.
func (g *Gopher) update(n neighbors) {
	if !g.Alive() {
		return
	}
	g.age++
	g.ticks++
	g.keepInside() // a hit box cresce, e o Gopher não pode ficar com ela fora da tela
	if g.ticks%walkEveryTicks == 0 {
		g.move(n.around(g)...)
	}
	g.turnTowardsHeading()
}

// move dá um passo.
//
// Primeiro sorteia se o Gopher gira (com chance turnChance). Se girar, sorteia
// o giro em [-maxTurn, maxTurn] graus. Depois anda stepSize pixels na direção
// resultante.
//
// Se o passo levaria a hit box de g a colidir com a de outro Gopher vivo, g
// reage como sorteado por interact. Os mortos não contam: quem morre de fome no
// meio de um tick só sai da colônia no fim dele, mas já não aparece na tela. Se
// o passo levaria a hit box para fora da tela, o Gopher quica na parede: a
// componente do movimento perpendicular a ela é invertida.
func (g *Gopher) move(others ...*Gopher) {
	if g.chance() < turnChance {
		g.heading = normalize(g.heading + float64(g.intn(2*maxTurn+1)-maxTurn))
	}

	sin, cos := math.Sincos(radians(g.heading))
	dx, dy := g.stepSize*cos, g.stepSize*sin

	for _, o := range others {
		if o != g && o.Alive() && g.collides(o, dx, dy) {
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

	g.steps++
	if g.steps%stepsPerSatiety == 0 {
		g.satiety--
	}
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
	return x-g.hitbox() >= 0 && x+g.hitbox() <= g.width
}

func (g *Gopher) fitsY(y float64) bool {
	return y-g.hitbox() >= 0 && y+g.hitbox() <= g.height
}

// keepInside traz a posição para dentro da tela, caso tenha saído.
func (g *Gopher) keepInside() {
	r := g.hitbox()
	g.x = clamp(g.x, r, g.width-r)
	g.y = clamp(g.y, r, g.height-r)
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
