package gopher

import "fmt"

// EventKind é o tipo de um Event.
type EventKind int

const (
	// Starving: o Gopher passou a ter fome, isto é, a saciedade dele caiu para
	// hungryAt ou menos (ver Hungry).
	Starving EventKind = iota

	// Starved: o Gopher morreu de fome.
	Starved
)

// Event é algo que aconteceu com um Gopher e que vale contar ao jogador.
// Colony.Update devolve os eventos de cada tick.
type Event struct {
	Kind EventKind
	Name string // nome do Gopher
}

// String descreve o evento numa frase, como "Ana morreu de fome".
func (e Event) String() string {
	switch e.Kind {
	case Starving:
		return e.Name + " está passando fome"
	case Starved:
		return e.Name + " morreu de fome"
	}
	return fmt.Sprintf("%s: EventKind(%d)", e.Name, int(e.Kind))
}
