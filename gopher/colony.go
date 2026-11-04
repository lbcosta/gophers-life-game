package gopher

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Colony é um grupo de Gophers que dividem a mesma tela e colidem entre si.
//
// Ela cuida da ordem em que as coisas precisam acontecer: os Gophers só são
// espalhados quando o tamanho da tela é conhecido (no primeiro Resize), cada um
// nasce longe dos que já foram espalhados, e só então andam e aparecem na tela.
// Os que morrem de fome somem dela.
type Colony struct {
	gophers  []*Gopher
	spawned  bool    // os Gophers já foram espalhados pela tela
	hovered  *Gopher // o Gopher sob o mouse, ou nil; seu nome é mostrado por Draw
	selected *Gopher // o Gopher clicado, ou nil; fica marcado e pode ser acompanhado
}

// MaxGophers é o maior número de Gophers que uma colônia comporta.
const MaxGophers = 1000

// NewColony cria uma colônia de n Gophers, cada um com velocidade e cor
// sorteadas (ver NewGophers). Se n passar de MaxGophers, a colônia é criada com
// MaxGophers; se for negativo, sem nenhum Gopher.
func NewColony(n int) *Colony {
	n = min(max(n, 0), MaxGophers)
	return &Colony{gophers: NewGophers(n)}
}

// Resize informa o tamanho da tela a todos os Gophers. Na primeira chamada,
// também os espalha pela tela.
func (c *Colony) Resize(width, height float64) {
	for _, g := range c.gophers {
		g.Resize(width, height)
	}
	if !c.spawned {
		for i, g := range c.gophers {
			g.Spawn(c.gophers[:i]...) // sem nascer em cima dos que já foram espalhados
		}
		c.spawned = true
	}
}

// Forager é o que um Gopher encontra ao andar: dado o círculo que ele ocupa
// (centro e raio), devolve quanta saciedade ele ganha com o que comeu ali.
type Forager func(x, y, radius float64) int

// Update deve ser chamado uma vez por tick do jogo. Cada Gopher anda e, se
// forage não for nil, come o que encontra. Os que morreram de fome saem da
// colônia. Antes do primeiro Resize não faz nada.
func (c *Colony) Update(forage Forager) {
	if !c.spawned {
		return
	}
	for _, g := range c.gophers {
		g.Update(c.gophers...)
		if forage != nil && g.Alive() {
			g.Eat(forage(g.x, g.y, g.hitbox()))
		}
	}

	alive := c.gophers[:0]
	for _, g := range c.gophers {
		if g.Alive() {
			alive = append(alive, g)
		}
	}
	clear(c.gophers[len(alive):])
	c.gophers = alive
}

// At devolve o Gopher vivo que está sob o ponto (x, y), isto é, com o ponto
// dentro da hit box, ou nil se não houver nenhum. Se houver vários, devolve o
// que aparece por cima dos outros.
func (c *Colony) At(x, y float64) *Gopher {
	if !c.spawned {
		return nil
	}
	for i := len(c.gophers) - 1; i >= 0; i-- {
		g := c.gophers[i]
		if g.Alive() && math.Hypot(g.x-x, g.y-y) <= g.hitbox() {
			return g
		}
	}
	return nil
}

// Hover informa onde está o mouse. O Gopher que estiver embaixo dele passa a ter
// o nome mostrado por Draw, até o mouse sair de cima.
func (c *Colony) Hover(x, y float64) {
	c.hovered = c.At(x, y)
}

// Select informa um clique em (x, y). O Gopher que estiver ali passa a ser o
// selecionado; se não houver nenhum, a seleção é desfeita.
func (c *Colony) Select(x, y float64) {
	c.selected = c.At(x, y)
}

// Selected devolve o Gopher selecionado, ou nil. Se ele morreu depois de
// selecionado, continua sendo devolvido, para dar para ver o que aconteceu.
func (c *Colony) Selected() *Gopher {
	return c.selected
}

// Gophers devolve os Gophers vivos da colônia.
func (c *Colony) Gophers() []*Gopher {
	return c.gophers
}

// Draw pinta os Gophers em screen. Todas as sombras vêm antes de todos os
// Gophers, para que a sombra de um nunca cubra outro que esteja encostado.
// Antes do primeiro Resize não desenha nada, já que os Gophers ainda não foram
// espalhados.
func (c *Colony) Draw(screen *ebiten.Image) {
	if !c.spawned {
		return
	}
	for _, g := range c.gophers {
		g.drawShadow(screen)
	}
	for _, g := range c.gophers {
		g.drawFigure(screen)
	}
	if c.selected != nil && c.selected.Alive() {
		c.selected.drawSelection(screen)
	}
	// O nome vem por último, para nunca ficar atrás de outro Gopher.
	if c.hovered != nil && c.hovered.Alive() {
		c.hovered.drawName(screen)
	}
}
