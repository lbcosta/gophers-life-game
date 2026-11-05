package world

import (
	"math/rand/v2"
	"slices"
	"testing"
)

// newTestWorld cria um mundo de 800x600 com sorteios de semente fixa.
func newTestWorld(target int) *World {
	r := rand.New(rand.NewPCG(1, 2))
	w := &World{target: target, chance: r.Float64, intn: r.IntN}
	w.Resize(800, 600)
	return w
}

// plenty é o quanto ainda come um Gopher que, nos testes que não tratam do teto
// da Saciedade, nunca fica cheio.
const plenty = 100

// plantAt põe em w uma planta do tipo kind na idade age, em (x, y).
func plantAt(w *World, kind Kind, x, y float64, age int) *Plant {
	p := &Plant{kind: kind, x: x, y: y, age: age, maturity: maturity{100, 200}}
	w.add(p)
	return p
}

func TestSatietyOfEachKind(t *testing.T) {
	want := map[Kind]int{Potato: 2, Carrot: 2, Radish: 2, Grass: 1, Clover: 1, Alfalfa: 1}
	if len(want) != len(Kinds) {
		t.Fatalf("%d alimentos no teste, %d no jogo", len(want), len(Kinds))
	}
	for k, s := range want {
		if k.Satiety() != s {
			t.Errorf("%v dá %d de saciedade, esperado %d", k, k.Satiety(), s)
		}
	}
}

func TestMaturityRangesAreDisjointAndIncreasing(t *testing.T) {
	prev := 0
	for i, r := range maturityRanges {
		if r.min <= prev || r.max < r.min {
			t.Errorf("intervalo %d = %+v: esperado começar depois de %d e não ser vazio", i, r, prev)
		}
		prev = r.max
	}
}

func TestNewMaturityStaysInRangesAndIncreases(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for range 1000 {
		m := newMaturity(r.IntN)
		for i, age := range m {
			if age < maturityRanges[i].min || age > maturityRanges[i].max {
				t.Fatalf("maturidade %v fora do intervalo %+v", m, maturityRanges[i])
			}
			if i > 0 && age <= m[i-1] {
				t.Fatalf("maturidade %v não é crescente", m)
			}
		}
	}
}

func TestStageChangesExactlyAtMaturity(t *testing.T) {
	p := &Plant{maturity: maturity{100, 200}}
	for _, c := range []struct {
		age   int
		stage Stage
	}{{0, Seed}, {99, Seed}, {100, Growing}, {199, Growing}, {200, Ready}, {9999, Ready}} {
		p.age = c.age
		if p.Stage() != c.stage {
			t.Errorf("idade %d: fase %v, esperado %v", c.age, p.Stage(), c.stage)
		}
	}
}

func TestResizePopulatesOnceInsideTheScreen(t *testing.T) {
	w := newTestWorld(50)
	if len(w.plants) != 50 {
		t.Fatalf("%d plantas, esperado 50", len(w.plants))
	}
	stages := map[Stage]int{}
	for i, p := range w.plants {
		if !w.inside(p.x, p.y) {
			t.Errorf("planta %d fora da tela, em (%v, %v)", i, p.x, p.y)
		}
		for _, o := range w.plants[:i] {
			if d := (p.x-o.x)*(p.x-o.x) + (p.y-o.y)*(p.y-o.y); d < 4*plantRadius*plantRadius {
				t.Errorf("planta %d encostada em outra", i)
			}
		}
		stages[p.Stage()]++
	}
	if len(stages) != numStages {
		t.Errorf("plantas iniciais em %d fases, esperado todas as %d: %v", len(stages), numStages, stages)
	}

	w.Resize(800, 600)
	if len(w.plants) != 50 {
		t.Errorf("segundo Resize mudou para %d plantas", len(w.plants))
	}
}

func TestResizeDropsPlantsOutsideTheScreen(t *testing.T) {
	w := newTestWorld(50)
	w.Resize(200, 150)
	for i, p := range w.plants {
		if !w.inside(p.x, p.y) {
			t.Errorf("planta %d ficou fora da tela, em (%v, %v)", i, p.x, p.y)
		}
	}
	if len(w.plants) >= 50 {
		t.Errorf("%d plantas depois de encolher a tela, esperado menos", len(w.plants))
	}
}

func TestUpdateGrowsPlants(t *testing.T) {
	w := newTestWorld(0)
	p := plantAt(w, Potato, 100, 100, 0)
	for range 100 {
		w.Update()
	}
	if p.age != 100 || p.Stage() != Growing {
		t.Errorf("idade %d, fase %v; esperado 100 e %v", p.age, p.Stage(), Growing)
	}
}

func TestUpdateReplantsUpToTarget(t *testing.T) {
	w := newTestWorld(10)
	for _, p := range slices.Clone(w.plants[3:]) {
		w.remove(p)
	}
	for range 100 * spawnEveryTicks {
		w.Update()
	}
	if len(w.plants) != 10 {
		t.Errorf("%d plantas, esperado voltar ao alvo de 10", len(w.plants))
	}
}

func TestWorldDoesNothingBeforeResize(t *testing.T) {
	w := New(10, nil)
	w.Update()
	if len(w.plants) != 0 || w.ticks != 0 {
		t.Errorf("mundo mudou antes do primeiro Resize: %d plantas, %d ticks", len(w.plants), w.ticks)
	}
}

func TestForageIgnoresSeeds(t *testing.T) {
	w := newTestWorld(0)
	plantAt(w, Potato, 100, 100, 0)
	if got := w.Forage(100, 100, 30, plenty); got != 0 || len(w.plants) != 1 {
		t.Errorf("ganhou %d com %d plantas, esperado 0 e a semente intacta", got, len(w.plants))
	}
}

func TestForageTramplesGrowingPlants(t *testing.T) {
	w := newTestWorld(0)
	plantAt(w, Carrot, 100, 100, 100)
	if got := w.Forage(100, 100, 30, plenty); got != 0 || len(w.plants) != 0 {
		t.Errorf("ganhou %d com %d plantas, esperado 0 e a planta pisoteada", got, len(w.plants))
	}
}

func TestForageEatsReadyPlants(t *testing.T) {
	for _, k := range Kinds {
		w := newTestWorld(0)
		plantAt(w, k, 100, 100, 200)
		if got := w.Forage(110, 100, 30, plenty); got != k.Satiety() || len(w.plants) != 0 {
			t.Errorf("%v: ganhou %d com %d plantas, esperado %d e a planta comida", k, got, len(w.plants), k.Satiety())
		}
	}
}

func TestForageOnlyTouchesPlantsUnderTheGopher(t *testing.T) {
	w := newTestWorld(0)
	plantAt(w, Potato, 100, 100, 200) // sob o Gopher
	plantAt(w, Potato, 300, 100, 200) // longe
	plantAt(w, Grass, 105, 100, 100)  // sob o Gopher, crescendo
	plantAt(w, Clover, 95, 100, 0)    // sob o Gopher, semente
	if got := w.Forage(100, 100, 30, plenty); got != 2 {
		t.Errorf("ganhou %d, esperado 2 (só a batata sob o Gopher)", got)
	}
	if len(w.plants) != 2 {
		t.Errorf("%d plantas, esperado 2 (a longe e a semente)", len(w.plants))
	}
}

func TestForageTakesWhatItRemovesOffTheGrid(t *testing.T) {
	w := newTestWorld(0)
	plantAt(w, Potato, 100, 100, 200)

	if got := w.Forage(100, 100, 30, plenty); got != Potato.Satiety() {
		t.Fatalf("ganhou %d, esperado %d", got, Potato.Satiety())
	}

	// A planta comida não pode continuar na grade: nem ser comida de novo, nem
	// impedir que outra nasça no lugar dela.
	if got := w.Forage(100, 100, 30, plenty); got != 0 {
		t.Errorf("comeu %d de novo da mesma planta", got)
	}
	if w.crowded(100, 100) {
		t.Error("o lugar da planta comida continua ocupado")
	}
}

// Os testes abaixo cobrem o teto da Saciedade do lado do Mundo. O último
// argumento de Forage é a Folga do Gopher (quanto falta para a Saciedade dele
// chegar ao teto): 0 para um Gopher Cheio.

func TestForageLeavesSeedsToAFullGopher(t *testing.T) {
	// GOP-34 e PLA-08: a semente não sofre nada, e um Gopher Cheio (Folga 0)
	// também não a afeta. Ele não ganha nada, e a semente continua no Mundo.
	w := newTestWorld(0)
	seed := plantAt(w, Potato, 100, 100, 0)
	if got := w.Forage(100, 100, 30, 0); got != 0 || !slices.Contains(w.plants, seed) {
		t.Errorf("Gopher Cheio ganhou %d com %d plantas, esperado 0 e a semente no Mundo", got, len(w.plants))
	}
}

func TestForageLeavesReadyPlantsToAFullGopher(t *testing.T) {
	// GOP-34: um Gopher cheio não come, e a Planta pronta continua no Mundo.
	w := newTestWorld(0)
	p := plantAt(w, Potato, 100, 100, 200)
	if got := w.Forage(100, 100, 30, 0); got != 0 || !slices.Contains(w.plants, p) {
		t.Errorf("Gopher cheio ganhou %d com %d plantas, esperado 0 e a batata pronta no Mundo", got, len(w.plants))
	}
}

func TestFullGopherStillTramplesGrowingPlants(t *testing.T) {
	// GOP-34 e PLA-08: cheio ou não, a Planta crescendo é pisoteada.
	w := newTestWorld(0)
	plantAt(w, Carrot, 100, 100, 100)
	if got := w.Forage(100, 100, 30, 0); got != 0 || len(w.plants) != 0 {
		t.Errorf("Gopher cheio ganhou %d com %d plantas, esperado 0 e a planta crescendo pisoteada", got, len(w.plants))
	}
}

func TestForageFeedsAGopherShortOfFullWithTheWholePlant(t *testing.T) {
	// GOP-28, GOP-29 e PLA-08: um Gopher com 9 (falta 1 para o teto 10) come a
	// batata, que some e dá os 2 do Alimento. O que passa do teto se perde no
	// Gopher, não no Mundo.
	w := newTestWorld(0)
	plantAt(w, Potato, 100, 100, 200)
	if got := w.Forage(100, 100, 30, 1); got != 2 || len(w.plants) != 0 {
		t.Errorf("faltando 1: ganhou %d com %d plantas, esperado 2 e a batata comida", got, len(w.plants))
	}
}

func TestForageEatsReadyPlantsOneAtATimeUntilFull(t *testing.T) {
	// GOP-35: o Gopher come as Plantas prontas uma de cada vez enquanto não está
	// cheio, e as que sobram continuam no Mundo. Com duas batatas (2 cada):
	// faltando 1, a primeira já enche; faltando 3, a primeira deixa faltar 1, e
	// ele come a segunda também.
	for _, c := range []struct {
		room, want, left int
	}{
		{room: 1, want: 2, left: 1},
		{room: 3, want: 4, left: 0},
	} {
		w := newTestWorld(0)
		plantAt(w, Potato, 95, 100, 200)
		plantAt(w, Potato, 105, 100, 200)
		got := w.Forage(100, 100, 30, c.room)
		if got != c.want || len(w.plants) != c.left {
			t.Errorf("faltando %d: ganhou %d com %d plantas, esperado %d com %d", c.room, got, len(w.plants), c.want, c.left)
			continue
		}
		for _, p := range w.plants {
			if p.Stage() != Ready {
				t.Errorf("faltando %d: sobrou uma planta %v, esperado a batata pronta", c.room, p.Stage())
			}
		}
	}
}

// BenchmarkForage mede uma busca por comida num mundo lotado de sementes
// (que Forage não tira do lugar), numa tela de 1920x1080. O jogo faz uma busca
// por Gopher em todo tick. Rode com: go test -run '^$' -bench Forage ./world
func BenchmarkForage(b *testing.B) {
	r := rand.New(rand.NewPCG(3, 4))
	w := &World{target: 2000, chance: r.Float64, intn: r.IntN}
	w.Resize(1920, 1080)
	for _, p := range w.plants {
		p.age = 0
	}
	for b.Loop() {
		w.Forage(r.Float64()*1920, r.Float64()*1080, 32, plenty)
	}
}
