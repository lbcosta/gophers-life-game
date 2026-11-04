// Package gopher implementa o Gopher do jogo: um personagem que anda sozinho,
// em direções aleatórias, sem nunca sair da tela.
package gopher

import "math/rand/v2"

// step é a distância, em pixels, percorrida a cada passo.
const step float32 = 4

// walkEveryTicks é de quantos em quantos ticks o Gopher dá um passo.
// Com 60 ticks por segundo, 7 ticks equivalem a cerca de 8,6 passos por segundo.
const walkEveryTicks = 7

// hitboxRadius é o raio da hit box do Gopher: um círculo de 64px de diâmetro,
// centrado no corpo, que contém o desenho inteiro (corpo, orelhas e olhos).
const hitboxRadius float32 = 32

// direction é um passo unitário em x e y.
type direction struct{ dx, dy float32 }

var directions = [...]direction{
	{-1, 0}, // esquerda
	{1, 0},  // direita
	{0, -1}, // cima
	{0, 1},  // baixo
}

// Gopher anda pela tela mantendo a hit box sempre inteira dentro dela.
// A posição é o centro do corpo.
type Gopher struct {
	x, y          float32
	width, height float32 // tamanho da tela

	// intn sorteia um inteiro em [0, n). É rand.IntN, mas pode ser trocado nos testes.
	intn func(n int) int
}

// NewGopher cria um Gopher no canto superior esquerdo de uma tela de
// width por height pixels, já com a hit box inteira dentro dela.
func NewGopher(width, height float32) *Gopher {
	return &Gopher{
		x:      hitboxRadius,
		y:      hitboxRadius,
		width:  width,
		height: height,
		intn:   rand.IntN,
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
	g.x = clamp(g.x, hitboxRadius, width-hitboxRadius)
	g.y = clamp(g.y, hitboxRadius, height-hitboxRadius)
}

// Update deve ser chamado uma vez por tick do jogo, com o contador global de
// ticks. A cada walkEveryTicks ticks o Gopher dá um passo em uma direção
// aleatória entre as válidas.
func (g *Gopher) Update(ticks int) {
	if ticks%walkEveryTicks == 0 {
		g.move()
	}
}

// move dá um passo em uma direção válida sorteada. Uma direção é válida se a
// hit box continuar inteira dentro da tela após o passo. Se nenhuma for válida,
// o Gopher fica parado.
func (g *Gopher) move() {
	valid := make([]direction, 0, len(directions))
	for _, d := range directions {
		if g.fits(g.x+d.dx*step, g.y+d.dy*step) {
			valid = append(valid, d)
		}
	}
	if len(valid) == 0 {
		return
	}

	d := valid[g.intn(len(valid))]
	g.x += d.dx * step
	g.y += d.dy * step
}

// fits informa se a hit box, centrada em (x, y), cabe inteira na tela.
func (g *Gopher) fits(x, y float32) bool {
	return x-hitboxRadius >= 0 && x+hitboxRadius <= g.width &&
		y-hitboxRadius >= 0 && y+hitboxRadius <= g.height
}

// clamp limita v ao intervalo [lo, hi]. Se hi < lo, devolve lo.
func clamp(v, lo, hi float32) float32 {
	if hi < lo {
		return lo
	}
	return min(max(v, lo), hi)
}
