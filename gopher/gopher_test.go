package gopher

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"
)

const eps = 1e-4

// newTestGopher cria um Gopher no meio de uma tela de width por height, que
// nunca gira por conta própria (chance sempre alta) e olha para heading.
// Os testes que precisam de giro trocam chance e intn.
func newTestGopher(width, height, heading float64) *Gopher {
	g := newGopher(
		func() float64 { return 0.99 }, // >= turnChance: não gira
		func(n int) int { return 0 },
	)
	g.Resize(width, height)
	g.x, g.y = width/2, height/2
	g.heading, g.facing = heading, heading
	return g
}

func near(a, b float64) bool { return math.Abs(a-b) < eps }

func (g *Gopher) insideScreen() bool {
	return g.fitsX(g.x) && g.fitsY(g.y)
}

func TestNewGopherDrawsInitialHeading(t *testing.T) {
	g := newGopher(
		func() float64 { return 0 },
		func(n int) int {
			if n != 360 {
				t.Errorf("intn(%d): esperado sorteio entre 360 direções", n)
			}
			return 123
		},
	)

	if g.heading != 123 || g.facing != 123 {
		t.Errorf("heading = %v, facing = %v, esperado 123 nos dois", g.heading, g.facing)
	}
	if g.x != hitboxRadius || g.y != hitboxRadius {
		t.Errorf("posição inicial (%v, %v), esperado o canto (%v, %v)", g.x, g.y, hitboxRadius, hitboxRadius)
	}
}

func TestMoveNeverLeavesScreen(t *testing.T) {
	// Sorteios com semente fixa, para que uma falha seja reproduzível.
	r := rand.New(rand.NewPCG(1, 2))
	g := newGopher(r.Float64, r.IntN)
	g.Resize(200, 150)

	for i := range 200_000 {
		g.move()
		if !g.insideScreen() {
			t.Fatalf("passo %d: hit box fora da tela, posição (%v, %v)", i, g.x, g.y)
		}
		if g.heading < 0 || g.heading >= 360 {
			t.Fatalf("passo %d: heading fora de [0, 360): %v", i, g.heading)
		}
	}
}

func TestMoveGoesStraightWhenNotTurning(t *testing.T) {
	tests := []struct {
		heading        float64
		wantDX, wantDY float64
	}{
		{0, step, 0},
		{90, 0, step},
		{180, -step, 0},
		{270, 0, -step},
		{45, step * math.Cos(math.Pi/4), step * math.Sin(math.Pi/4)},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("heading %v", tt.heading), func(t *testing.T) {
			g := newTestGopher(500, 500, tt.heading)
			x0, y0 := g.x, g.y

			g.move()

			if dx, dy := g.x-x0, g.y-y0; !near(dx, tt.wantDX) || !near(dy, tt.wantDY) {
				t.Errorf("andou (%v, %v), esperado (%v, %v)", dx, dy, tt.wantDX, tt.wantDY)
			}
			if g.heading != tt.heading {
				t.Errorf("heading mudou para %v sem sortear giro", g.heading)
			}
		})
	}
}

func TestMoveTurnsWhenChanceHits(t *testing.T) {
	g := newTestGopher(500, 500, 0)
	g.chance = func() float64 { return 0 } // < turnChance: gira
	g.intn = func(n int) int {
		if n != 2*maxTurn+1 {
			t.Errorf("intn(%d): esperado sorteio entre %d valores", n, 2*maxTurn+1)
		}
		return n - 1 // maior valor: giro de +maxTurn
	}

	g.move()

	if g.heading != maxTurn {
		t.Errorf("heading = %v, esperado %v", g.heading, maxTurn)
	}
}

func TestTurnCanBeNegativeAndWrapsAround(t *testing.T) {
	g := newTestGopher(500, 500, 10)
	g.chance = func() float64 { return 0 }
	g.intn = func(n int) int { return 0 } // menor valor: giro de -maxTurn

	g.move()

	if want := float64(360 + 10 - maxTurn); !near(g.heading, want) {
		t.Errorf("heading = %v, esperado %v", g.heading, want)
	}
}

func TestMoveBouncesOnVerticalWalls(t *testing.T) {
	// Encostado na parede direita, olhando para a direita e um pouco para baixo.
	g := newTestGopher(500, 500, 30)
	g.x = 500 - hitboxRadius
	x0, y0 := g.x, g.y

	g.move()

	if !near(g.heading, 150) {
		t.Errorf("heading = %v, esperado 150", g.heading)
	}
	if wantDX := -step * math.Cos(radians(30)); !near(g.x-x0, wantDX) {
		t.Errorf("dx = %v, esperado %v (volta para a esquerda)", g.x-x0, wantDX)
	}
	if wantDY := step * math.Sin(radians(30)); !near(g.y-y0, wantDY) {
		t.Errorf("dy = %v, esperado %v (continua descendo)", g.y-y0, wantDY)
	}
}

func TestMoveBouncesOnHorizontalWalls(t *testing.T) {
	// Encostado no topo, olhando para cima.
	g := newTestGopher(500, 500, 270)
	g.y = hitboxRadius
	y0 := g.y

	g.move()

	if !near(g.heading, 90) {
		t.Errorf("heading = %v, esperado 90", g.heading)
	}
	if !near(g.y-y0, step) {
		t.Errorf("dy = %v, esperado %v (volta para baixo)", g.y-y0, step)
	}
}

func TestMoveBouncesInCorner(t *testing.T) {
	g := newTestGopher(500, 500, 45) // olhando para o canto inferior direito
	g.x, g.y = 500-hitboxRadius, 500-hitboxRadius

	g.move()

	if !near(g.heading, 225) {
		t.Errorf("heading = %v, esperado 225 (de volta pelo mesmo caminho)", g.heading)
	}
	if !g.insideScreen() {
		t.Errorf("hit box fora da tela em (%v, %v)", g.x, g.y)
	}
}

func TestMoveStaysInsideOnScreenTooSmall(t *testing.T) {
	// Tela do tamanho exato da hit box: não há para onde ir.
	g := newTestGopher(2*hitboxRadius, 2*hitboxRadius, 0)
	g.x, g.y = hitboxRadius, hitboxRadius

	g.move()

	if g.x != hitboxRadius || g.y != hitboxRadius {
		t.Errorf("Gopher deveria ficar parado, mas foi para (%v, %v)", g.x, g.y)
	}
}

func TestUpdateWalksOnlyEveryWalkEveryTicks(t *testing.T) {
	g := newTestGopher(500, 500, 0)
	x0 := g.x

	for tick := 1; tick < walkEveryTicks; tick++ {
		g.Update()
		if g.x != x0 {
			t.Fatalf("tick %d: Gopher andou antes da hora", tick)
		}
	}

	g.Update()
	if g.x != x0+step {
		t.Errorf("tick %d: x = %v, esperado %v", walkEveryTicks, g.x, x0+step)
	}
}

func TestFacingTurnsTheShortWay(t *testing.T) {
	// De 350º para 10º são 20º no sentido horário, passando por 0º.
	g := newTestGopher(500, 500, 10)
	g.facing = 350

	g.Update()

	if want := normalize(350 + turnSpeed); !near(g.facing, want) {
		t.Errorf("facing = %v, esperado %v (pelo caminho curto)", g.facing, want)
	}
}

func TestFacingTurnsCounterclockwiseWhenShorter(t *testing.T) {
	g := newTestGopher(500, 500, 350)
	g.facing = 10

	g.Update()

	if want := normalize(10 - turnSpeed); !near(g.facing, want) {
		t.Errorf("facing = %v, esperado %v", g.facing, want)
	}
}

func TestFacingReachesHeadingAndStops(t *testing.T) {
	g := newTestGopher(500, 500, 100)
	g.facing = 0

	for range 100 {
		g.Update()
	}

	if g.facing != g.heading {
		t.Errorf("facing = %v, esperado %v", g.facing, g.heading)
	}
}

func TestResizeBringsGopherBackInside(t *testing.T) {
	g := newTestGopher(1000, 1000, 0)
	g.x, g.y = 900, 800

	g.Resize(300, 200)

	if want := 300.0 - hitboxRadius; g.x != want {
		t.Errorf("x = %v, esperado %v", g.x, want)
	}
	if want := 200.0 - hitboxRadius; g.y != want {
		t.Errorf("y = %v, esperado %v", g.y, want)
	}
	if !g.insideScreen() {
		t.Error("hit box deveria estar dentro da tela após Resize")
	}
}

func TestResizeKeepsPositionWhenStillInside(t *testing.T) {
	g := newTestGopher(1000, 1000, 0)
	g.x, g.y = 100, 120

	g.Resize(800, 600)

	if g.x != 100 || g.y != 120 {
		t.Errorf("posição mudou para (%v, %v) sem necessidade", g.x, g.y)
	}
}

func TestShortestTurn(t *testing.T) {
	tests := []struct{ from, to, want float64 }{
		{0, 90, 90},
		{90, 0, -90},
		{350, 10, 20},
		{10, 350, -20},
		{0, 180, -180}, // meia-volta: por convenção, anti-horário
		{45, 45, 0},
	}
	for _, tt := range tests {
		if got := shortestTurn(tt.from, tt.to); !near(got, tt.want) {
			t.Errorf("shortestTurn(%v, %v) = %v, esperado %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestDrawingFitsInHitbox(t *testing.T) {
	for i, s := range shapes {
		for _, p := range ellipsePoints(s.cx, s.cy, s.rx, s.ry, s.angle) {
			d := math.Hypot(float64(p.x*bodyRadius), float64(p.y*bodyRadius))
			if d > hitboxRadius+1e-3 {
				t.Fatalf("parte %d passa da hit box: distância %v > %v", i, d, hitboxRadius)
			}
		}
	}

	shadowReach := float64(bodyRadius) + math.Hypot(float64(shadowOffsetX), float64(shadowOffsetY))
	if shadowReach > hitboxRadius {
		t.Errorf("sombra passa da hit box: alcance %v > %v", shadowReach, hitboxRadius)
	}
}

func TestMirroredPartsAreExpanded(t *testing.T) {
	i := 0
	for _, p := range parts {
		if shapes[i] != p {
			t.Fatalf("shapes[%d] = %+v, esperado a parte original %+v", i, shapes[i], p)
		}
		i++
		if !p.mirrored {
			continue
		}
		m := shapes[i]
		if m.cx != -p.cx || m.angle != -p.angle || m.cy != p.cy || m.rx != p.rx || m.ry != p.ry {
			t.Errorf("shapes[%d] = %+v, esperado o espelho de %+v", i, m, p)
		}
		i++
	}
	if i != len(shapes) {
		t.Errorf("len(shapes) = %d, esperado %d", len(shapes), i)
	}
}

func TestSpawnPlacesGopherInsideScreen(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	g := newGopher(r.Float64, r.IntN)
	g.Resize(300, 200)

	for i := range 10_000 {
		g.Spawn()
		if !g.insideScreen() {
			t.Fatalf("sorteio %d: hit box fora da tela, posição (%v, %v)", i, g.x, g.y)
		}
	}
}

func TestSpawnUsesDrawnValues(t *testing.T) {
	g := newGopher(
		func() float64 { return 0.5 }, // meio da faixa válida, nos dois eixos
		func(n int) int { return 3 },
	)
	g.Resize(300, 200)

	g.Spawn()

	if g.x != 150 || g.y != 100 {
		t.Errorf("posição (%v, %v), esperado o centro (150, 100)", g.x, g.y)
	}
	if g.ticks != 3 {
		t.Errorf("ticks = %d, esperado a fase sorteada, 3", g.ticks)
	}
}

func TestSpawnOnScreenTooSmallStaysInside(t *testing.T) {
	g := newGopher(func() float64 { return 0.5 }, func(n int) int { return 0 })
	g.Resize(hitboxRadius, hitboxRadius) // menor que a hit box

	g.Spawn()

	if g.x != hitboxRadius || g.y != hitboxRadius {
		t.Errorf("posição (%v, %v), esperado o canto (%v, %v)", g.x, g.y, hitboxRadius, hitboxRadius)
	}
}

func TestSpawnDesynchronizesSteps(t *testing.T) {
	// Dois Gophers com fases diferentes não dão o passo no mesmo tick.
	stepTick := func(phase int) int {
		g := newGopher(func() float64 { return 0.99 }, func(n int) int { return phase })
		g.Resize(500, 500)
		g.Spawn()
		x0, y0 := g.x, g.y
		for tick := 1; tick <= walkEveryTicks; tick++ {
			g.Update()
			if g.x != x0 || g.y != y0 {
				return tick
			}
		}
		return -1
	}

	if a, b := stepTick(0), stepTick(3); a == b || a < 0 || b < 0 {
		t.Errorf("fases diferentes deveriam dar o passo em ticks diferentes, mas foram %d e %d", a, b)
	}
}
