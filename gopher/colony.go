package gopher

import "github.com/hajimehoshi/ebiten/v2"

// Colony é um grupo de Gophers que dividem a mesma tela e colidem entre si.
//
// Ela cuida da ordem em que as coisas precisam acontecer: os Gophers só são
// espalhados quando o tamanho da tela é conhecido (no primeiro Resize), cada um
// nasce longe dos que já foram espalhados, e só então andam e aparecem na tela.
type Colony struct {
	gophers []*Gopher
	spawned bool // os Gophers já foram espalhados pela tela
}

// NewColony cria uma colônia de n Gophers, cada um com velocidade e cor
// sorteadas (ver NewGophers).
func NewColony(n int) *Colony {
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

// Update deve ser chamado uma vez por tick do jogo. Antes do primeiro Resize
// não faz nada.
func (c *Colony) Update() {
	if !c.spawned {
		return
	}
	for _, g := range c.gophers {
		g.Update(c.gophers...)
	}
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
}
