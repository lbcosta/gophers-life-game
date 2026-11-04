// Package world controla os objetos do mundo em que os Gophers vivem. Por
// enquanto são as plantas: alimentos que nascem em lugares aleatórios, crescem
// em três fases e podem ser comidos quando estão prontos.
package world

import "fmt"

// Kind é o tipo de alimento de uma planta.
type Kind int

const (
	Potato Kind = iota
	Carrot
	Radish
	Grass
	Clover
	Alfalfa

	numKinds = iota
)

// Kinds lista todos os tipos de alimento.
var Kinds = [numKinds]Kind{Potato, Carrot, Radish, Grass, Clover, Alfalfa}

var kinds = [numKinds]struct {
	name    string
	sprite  string // nome do arquivo em assets/plants, sem a extensão
	satiety int
}{
	Potato:  {"batata", "potato", 2},
	Carrot:  {"cenoura", "carrot", 2},
	Radish:  {"rabanete", "radish", 2},
	Grass:   {"capim", "grass", 1},
	Clover:  {"trevo", "clover", 1},
	Alfalfa: {"alfafa", "alfalfa", 1},
}

// Satiety é quanta saciedade um Gopher ganha ao comer este alimento.
func (k Kind) Satiety() int { return kinds[k].satiety }

func (k Kind) String() string {
	if k < 0 || int(k) >= numKinds {
		return fmt.Sprintf("Kind(%d)", int(k))
	}
	return kinds[k].name
}

// Stage é a fase de crescimento de uma planta.
type Stage int

const (
	Seed    Stage = iota // semente plantada
	Growing              // crescendo
	Ready                // pronta para comer

	numStages = iota
)

var stageNames = [numStages]string{Seed: "semente", Growing: "crescendo", Ready: "pronta"}

func (s Stage) String() string {
	if s < 0 || int(s) >= numStages {
		return fmt.Sprintf("Stage(%d)", int(s))
	}
	return stageNames[s]
}

// ticksPerSecond é quantos ticks o jogo tem por segundo.
const ticksPerSecond = 60

// span é um intervalo de idades, em ticks, com os dois extremos incluídos.
type span struct{ min, max int }

// maturityRanges são, para cada fase a partir de Growing, o intervalo de
// idades em que uma planta pode entrar nela. É a mesma mecânica dos Gophers:
// cada planta sorteia a sua idade dentro do intervalo, então amadurecem em
// ritmos diferentes. Os intervalos não se sobrepõem e são crescentes, o que
// garante que uma planta sempre passa pelas fases em ordem.
var maturityRanges = [numStages - 1]span{
	Growing - 1: {16 * ticksPerSecond, 30 * ticksPerSecond},
	Ready - 1:   {40 * ticksPerSecond, 70 * ticksPerSecond},
}

// maturity guarda a idade, em ticks, em que uma planta entra em Growing e em
// Ready, nessa ordem. Toda planta nasce Seed.
type maturity [numStages - 1]int

// newMaturity sorteia a maturidade de uma planta. intn devolve um inteiro em
// [0, n).
func newMaturity(intn func(n int) int) maturity {
	var m maturity
	for i, r := range maturityRanges {
		m[i] = r.min + intn(r.max-r.min+1)
	}
	return m
}

// Plant é um alimento plantado no mundo. A posição é o centro do desenho.
type Plant struct {
	kind     Kind
	x, y     float64
	age      int // ticks desde que foi plantada
	maturity maturity
}

// Kind devolve o tipo de alimento.
func (p *Plant) Kind() Kind { return p.kind }

// Position devolve o centro da planta.
func (p *Plant) Position() (x, y float64) { return p.x, p.y }

// Stage devolve a fase de crescimento da planta.
func (p *Plant) Stage() Stage {
	s := Seed
	for _, entersAt := range p.maturity {
		if p.age >= entersAt {
			s++
		}
	}
	return s
}
