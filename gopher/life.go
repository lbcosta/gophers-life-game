package gopher

import "fmt"

// Stage é a fase da vida de um Gopher. O corpo cresce o tempo todo, mas o que
// aparece muda a cada fase: bebê e criança são só uma bolinha (a da criança é
// maior), o adolescente ganha olhos e o adulto tem a forma completa.
type Stage int

const (
	Baby Stage = iota
	Child
	Teen
	Adult

	numStages = iota
)

var stageNames = [numStages]string{Baby: "bebê", Child: "criança", Teen: "adolescente", Adult: "adulto"}

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

// maturityRanges são, para cada fase a partir de Child, o intervalo de idades
// em que um Gopher pode entrar nela. Cada Gopher sorteia a sua idade dentro do
// intervalo (ver newMaturity), então amadurecem em ritmos diferentes. Os
// intervalos não se sobrepõem e são crescentes, o que garante que um Gopher
// sempre entra numa fase depois de ter entrado na anterior.
var maturityRanges = [numStages - 1]span{
	Child - 1: {10 * ticksPerSecond, 20 * ticksPerSecond},
	Teen - 1:  {25 * ticksPerSecond, 40 * ticksPerSecond},
	Adult - 1: {45 * ticksPerSecond, 60 * ticksPerSecond},
}

// maturity guarda a idade, em ticks, em que um Gopher entra em Child, Teen e
// Adult, nessa ordem. Todo Gopher nasce Baby.
type maturity [numStages - 1]int

// newMaturity sorteia a maturidade de um Gopher: uma idade em cada intervalo de
// maturityRanges. intn devolve um inteiro em [0, n).
func newMaturity(intn func(n int) int) maturity {
	var m maturity
	for i, r := range maturityRanges {
		m[i] = r.min + intn(r.max-r.min+1)
	}
	return m
}

// middleMaturity é a maturidade no meio de cada intervalo, sem sorteio.
func middleMaturity() maturity {
	var m maturity
	for i, r := range maturityRanges {
		m[i] = (r.min + r.max) / 2
	}
	return m
}

// Raio do corpo, em pixels, de um recém-nascido e de quem acaba de entrar em
// cada fase. Entre uma fase e a seguinte o corpo cresce em linha reta; o do
// adulto é bodyRadius.
const (
	babySize  float32 = 6
	childSize float32 = 11
	teenSize  float32 = 17
)

// Stage devolve a fase da vida do Gopher.
func (g *Gopher) Stage() Stage {
	s := Baby
	for _, entersAt := range g.maturity {
		if g.age >= entersAt {
			s++
		}
	}
	return s
}

// bodySize devolve o raio do corpo na idade atual: cresce desde o nascimento,
// passando pelo tamanho inicial de cada fase, até bodyRadius na idade adulta.
func (g *Gopher) bodySize() float32 {
	if g.Stage() == Adult {
		return bodyRadius
	}
	ages := [numStages]int{0, g.maturity[0], g.maturity[1], g.maturity[2]}
	sizes := [numStages]float32{babySize, childSize, teenSize, bodyRadius}

	s := g.Stage() // a fase em que o Gopher está é a que começa na idade ages[s]
	t := float32(g.age-ages[s]) / float32(ages[s+1]-ages[s])
	return sizes[s] + t*(sizes[s+1]-sizes[s])
}

// hitbox devolve o raio da hit box na idade atual: o do desenho da fase em que
// o Gopher está, já no tamanho que o corpo tem agora. O adulto tem hitboxRadius.
func (g *Gopher) hitbox() float64 {
	stage := g.Stage()
	if stage == Adult {
		return hitboxRadius
	}
	return float64(g.bodySize() * stageReach[stage])
}
