package world

import (
	"math"
	"math/rand/v2"
)

const (
	// plantRadius é o raio, em pixels, da área que uma planta ocupa. As plantas
	// nascem inteiras dentro da tela e afastadas umas das outras.
	plantRadius = 18

	// spawnEveryTicks é de quantos em quantos ticks nasce uma nova semente, se
	// o mundo tiver menos plantas que o alvo. A 60 ticks por segundo, são 2
	// sementes por segundo.
	spawnEveryTicks = 30

	// spawnAttempts é quantas posições o mundo tenta para uma semente antes de
	// desistir dela (nesse tick).
	spawnAttempts = 20

	// initialAgeFactor, multiplicado pela idade de "pronta" de cada planta, dá
	// o limite superior da idade sorteada para as plantas iniciais. Com 1,5, uma
	// em cada três já começa pronta, e as outras estão espalhadas entre as fases.
	initialAgeFactor = 1.5
)

// World reúne as plantas e cuida de plantá-las, fazê-las crescer e entregá-las
// aos Gophers que passam por cima.
type World struct {
	plants        []*Plant
	target        int // quantidade de plantas que o mundo tenta manter
	width, height float64
	populated     bool // as plantas iniciais já foram plantadas
	ticks         int

	sprites *Sprites

	// Fontes de sorteio. São rand.Float64 e rand.IntN, mas os testes informam as suas.
	chance func() float64
	intn   func(n int) int
}

// New cria um mundo que mantém cerca de target plantas, desenhadas com sprites.
func New(target int, sprites *Sprites) *World {
	return &World{target: target, sprites: sprites, chance: rand.Float64, intn: rand.IntN}
}

// Plants devolve as plantas do mundo.
func (w *World) Plants() []*Plant { return w.plants }

// Resize informa o tamanho da tela. Na primeira chamada planta as plantas
// iniciais; nas seguintes, remove as que ficaram fora da tela.
func (w *World) Resize(width, height float64) {
	w.width, w.height = width, height
	if !w.populated {
		for range w.target {
			if p := w.newPlant(); p != nil {
				p.age = w.intn(int(float64(p.maturity[len(p.maturity)-1]) * initialAgeFactor))
				w.plants = append(w.plants, p)
			}
		}
		w.populated = true
		return
	}
	kept := w.plants[:0]
	for _, p := range w.plants {
		if w.inside(p.x, p.y) {
			kept = append(kept, p)
		}
	}
	clear(w.plants[len(kept):])
	w.plants = kept
}

func (w *World) inside(x, y float64) bool {
	return x >= plantRadius && x <= w.width-plantRadius && y >= plantRadius && y <= w.height-plantRadius
}

// newPlant cria uma semente de tipo sorteado numa posição livre e sorteada da
// tela. Devolve nil se, depois de spawnAttempts tentativas, não achou lugar.
func (w *World) newPlant() *Plant {
	for range spawnAttempts {
		x := plantRadius + w.chance()*(w.width-2*plantRadius)
		y := plantRadius + w.chance()*(w.height-2*plantRadius)
		if !w.inside(x, y) || w.crowded(x, y) {
			continue
		}
		return &Plant{kind: Kinds[w.intn(numKinds)], x: x, y: y, maturity: newMaturity(w.intn)}
	}
	return nil
}

// crowded informa se (x, y) está perto demais de alguma planta.
func (w *World) crowded(x, y float64) bool {
	for _, p := range w.plants {
		if math.Hypot(p.x-x, p.y-y) < 2*plantRadius {
			return true
		}
	}
	return false
}

// Update deve ser chamado uma vez por tick do jogo: as plantas crescem e, de
// tempos em tempos, nasce uma semente se faltarem plantas. Antes do primeiro
// Resize não faz nada.
func (w *World) Update() {
	if !w.populated {
		return
	}
	for _, p := range w.plants {
		p.age++
	}
	w.ticks++
	if w.ticks%spawnEveryTicks == 0 && len(w.plants) < w.target {
		if p := w.newPlant(); p != nil {
			w.plants = append(w.plants, p)
		}
	}
}

// Forage trata um Gopher que passa por cima das plantas: o que está com o
// centro dentro do círculo de centro (x, y) e raio radius. Semente não sofre
// nada; planta crescendo é desfeita; planta pronta é comida. Devolve a
// saciedade que o Gopher ganhou com o que comeu.
func (w *World) Forage(x, y, radius float64) (satiety int) {
	kept := w.plants[:0]
	for _, p := range w.plants {
		if math.Hypot(p.x-x, p.y-y) >= radius {
			kept = append(kept, p)
			continue
		}
		switch p.Stage() {
		case Seed:
			kept = append(kept, p)
		case Ready:
			satiety += p.kind.Satiety()
		}
		// Growing: pisoteada, some sem dar nada.
	}
	clear(w.plants[len(kept):])
	w.plants = kept
	return satiety
}
