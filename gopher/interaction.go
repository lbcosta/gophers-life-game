package gopher

import "math"

// touchDistance é a distância entre os centros de g e other abaixo da qual as
// hit boxes se sobrepõem. Muda conforme eles crescem.
func (g *Gopher) touchDistance(other *Gopher) float64 {
	return g.hitbox() + other.hitbox()
}

// collisionRange é a maior distância entre o centro de g e o de um Gopher que o
// próximo passo de g pode tocar. O passo leva g a até stepSize de onde está, e
// ali ele toca quem estiver a menos de touchDistance, que é no máximo g.hitbox()
// + hitboxRadius (a hit box de um adulto).
func (g *Gopher) collisionRange() float64 {
	return g.hitbox() + hitboxRadius + g.stepSize
}

// neighbors diz com quem um Gopher pode esbarrar. Colony responde consultando a
// grade em que guarda os Gophers (ver Colony.around); everyone, uma lista
// simples, responde com todos.
type neighbors interface {
	// around devolve os Gophers que g pode tocar no próximo passo: todos os que
	// estão a até g.collisionRange() dele. Pode trazer também o próprio g e
	// outros mais distantes, que move e overlapsAny descartam.
	around(g *Gopher) []*Gopher
}

// everyone é uma lista de Gophers em que qualquer um pode esbarrar em qualquer
// outro.
type everyone []*Gopher

func (e everyone) around(*Gopher) []*Gopher { return e }

// reaction é um possível desfecho de uma colisão. Recebe o Gopher g, aquele com
// quem ele colide (other) e o deslocamento (dx, dy) que g faria neste passo, e
// devolve o deslocamento que g fará de fato.
type reaction func(g, other *Gopher, dx, dy float64) (float64, float64)

// reactions são os desfechos que interact sorteia, todos com a mesma chance.
// Para acrescentar um comportamento, basta escrever a função e listá-la aqui. Ela
// pode mudar a direção do passo e encurtá-lo, mas não alongá-lo (ver
// collisionRange); os testes conferem isso.
var reactions = []reaction{asWall}

// interact sorteia como g reage a uma colisão com other e devolve o novo
// deslocamento de g.
func (g *Gopher) interact(other *Gopher, dx, dy float64) (float64, float64) {
	return reactions[g.intn(len(reactions))](g, other, dx, dy)
}

// asWall trata other como uma parede: g quica nela, ou seja, a componente do
// deslocamento na direção de other é invertida e a outra é mantida. A "parede" é
// perpendicular à reta que liga os centros.
func asWall(g, other *Gopher, dx, dy float64) (float64, float64) {
	// Normal da parede: unitária, de other para g. collides garante que os
	// centros não coincidem, então a distância não é zero.
	nx, ny := g.x-other.x, g.y-other.y
	dist := math.Hypot(nx, ny)
	nx, ny = nx/dist, ny/dist

	dot := dx*nx + dy*ny
	return dx - 2*dot*nx, dy - 2*dot*ny
}

// collides informa se o passo (dx, dy) de g o faria colidir com other: as hit
// boxes passariam a se sobrepor, e o passo aproxima os dois. Se os Gophers já
// estiverem sobrepostos, passos que os afastam são permitidos, para que eles
// consigam se separar. Centros coincidentes não colidem, pois não há uma
// direção de "parede" a considerar.
func (g *Gopher) collides(other *Gopher, dx, dy float64) bool {
	current := math.Hypot(g.x-other.x, g.y-other.y)
	next := math.Hypot(g.x+dx-other.x, g.y+dy-other.y)
	return current > 0 && next < g.touchDistance(other) && next < current
}
