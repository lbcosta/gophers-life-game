package gopher

import "fmt"

// EventKind é o tipo de um Event.
type EventKind int

const (
	// BecameHungry: o Gopher passou a ter fome, isto é, a saciedade dele caiu para
	// hungryAt ou menos (ver Hungry).
	BecameHungry EventKind = iota

	// Starved: o Gopher morreu de fome.
	Starved

	numEventKinds = iota
)

// eventKindPhrases guarda a frase de cada EventKind, sem o nome do Gopher.
var eventKindPhrases = [numEventKinds]string{
	BecameHungry: "está passando fome",
	Starved:      "morreu de fome",
}

func (k EventKind) valid() bool { return k >= 0 && int(k) < numEventKinds }

// String devolve a frase do tipo, sem o nome ("morreu de fome"), ou
// "EventKind(n)" se o valor for inválido.
func (k EventKind) String() string {
	if !k.valid() {
		return fmt.Sprintf("EventKind(%d)", int(k))
	}
	return eventKindPhrases[k]
}

// Event é algo que aconteceu com um Gopher e que vale contar ao jogador.
// Colony.Update devolve os eventos de cada tick.
type Event struct {
	Kind EventKind
	Name string // nome do Gopher
}

// String descreve o evento numa frase, como "Ana morreu de fome". Para um tipo
// inválido, devolve "Ana: EventKind(99)".
func (e Event) String() string {
	if !e.Kind.valid() {
		return e.Name + ": " + e.Kind.String()
	}
	return e.Name + " " + e.Kind.String()
}
