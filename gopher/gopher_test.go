package gopher

import (
	"math"
	"testing"
)

const eps = 1e-4

// newTestGopher cria um Gopher no meio de uma tela de width por height, que
// nunca gira por conta própria (chance sempre alta) e olha para heading.
// Os testes que precisam de giro trocam chance e intn.
func newTestGopher(width, height float32, heading float64) *Gopher {
	g := NewGopher(width, height)
	g.x, g.y = width/2, height/2
	g.heading, g.facing = heading, heading
	g.chance = func() float64 { return 0.99 } // >= turnChance: não gira
	g.intn = func(n int) int { return 0 }
	return g
}

func near(a, b float64) bool { return math.Abs(a-b) < eps }

func (g *Gopher) insideScreen() bool {
	return g.fitsX(g.x) && g.fitsY(g.y)
}

func TestMoveNeverLeavesScreen(t *testing.T) {
	g := NewGopher(200, 150)
	for i := 0; i < 200_000; i++ {
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
		heading    float64
		wantDX     float64
		wantDY     float64
		wantHeadng float64
	}{
		{0, 4, 0, 0},
		{90, 0, 4, 90},
		{180, -4, 0, 180},
		{270, 0, -4, 270},
		{45, 4 * math.Cos(math.Pi/4), 4 * math.Sin(math.Pi/4), 45},
	}
	for _, tt := range tests {
		g := newTestGopher(500, 500, tt.heading)
		x0, y0 := g.x, g.y

		g.move()

		if !near(float64(g.x-x0), tt.wantDX) || !near(float64(g.y-y0), tt.wantDY) {
			t.Errorf("heading %v: andou (%v, %v), esperado (%v, %v)", tt.heading, g.x-x0, g.y-y0, tt.wantDX, tt.wantDY)
		}
		if g.heading != tt.wantHeadng {
			t.Errorf("heading %v: mudou para %v sem sortear giro", tt.heading, g.heading)
		}
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
		t.Errorf("heading = %v, esperado %v", g.heading, float64(maxTurn))
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
	if wantDX := -4 * math.Cos(30*math.Pi/180); !near(float64(g.x-x0), wantDX) {
		t.Errorf("dx = %v, esperado %v (volta para a esquerda)", g.x-x0, wantDX)
	}
	if wantDY := 4 * math.Sin(30*math.Pi/180); !near(float64(g.y-y0), wantDY) {
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
	if !near(float64(g.y-y0), 4) {
		t.Errorf("dy = %v, esperado 4 (volta para baixo)", g.y-y0)
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

	for ticks := 1; ticks < walkEveryTicks; ticks++ {
		g.Update(ticks)
		if g.x != x0 {
			t.Fatalf("tick %d: Gopher andou antes da hora", ticks)
		}
	}

	g.Update(walkEveryTicks)
	if g.x != x0+step {
		t.Errorf("tick %d: x = %v, esperado %v", walkEveryTicks, g.x, x0+step)
	}
}

func TestFacingTurnsTheShortWay(t *testing.T) {
	// De 350º para 10º são 20º no sentido horário, passando por 0º.
	g := newTestGopher(500, 500, 10)
	g.facing = 350

	g.Update(1)

	if want := normalize(350 + turnSpeed); !near(g.facing, want) {
		t.Errorf("facing = %v, esperado %v (pelo caminho curto)", g.facing, want)
	}
}

func TestFacingTurnsCounterclockwiseWhenShorter(t *testing.T) {
	g := newTestGopher(500, 500, 350)
	g.facing = 10

	g.Update(1)

	if want := normalize(10 - turnSpeed); !near(g.facing, want) {
		t.Errorf("facing = %v, esperado %v", g.facing, want)
	}
}

func TestFacingReachesHeadingAndStops(t *testing.T) {
	g := newTestGopher(500, 500, 100)
	g.facing = 0

	for i := 0; i < 100; i++ {
		g.Update(1)
	}

	if g.facing != g.heading {
		t.Errorf("facing = %v, esperado %v", g.facing, g.heading)
	}
}

func TestResizeBringsGopherBackInside(t *testing.T) {
	g := NewGopher(1000, 1000)
	g.x, g.y = 900, 800

	g.Resize(300, 200)

	if want := float32(300 - hitboxRadius); g.x != want {
		t.Errorf("x = %v, esperado %v", g.x, want)
	}
	if want := float32(200 - hitboxRadius); g.y != want {
		t.Errorf("y = %v, esperado %v", g.y, want)
	}
	if !g.insideScreen() {
		t.Error("hit box deveria estar dentro da tela após Resize")
	}
}

func TestResizeKeepsPositionWhenStillInside(t *testing.T) {
	g := NewGopher(1000, 1000)
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
			if d > float64(hitboxRadius)+1e-3 {
				t.Fatalf("parte %d passa da hit box: distância %v > %v", i, d, hitboxRadius)
			}
		}
	}

	shadowReach := float64(bodyRadius) + math.Hypot(float64(shadowOffsetX), float64(shadowOffsetY))
	if shadowReach > float64(hitboxRadius) {
		t.Errorf("sombra passa da hit box: alcance %v > %v", shadowReach, hitboxRadius)
	}
}

func TestMirroredPartsAreExpanded(t *testing.T) {
	want := 0
	for _, p := range parts {
		want++
		if p.mirrored {
			want++
		}
	}
	if len(shapes) != want {
		t.Errorf("len(shapes) = %d, esperado %d", len(shapes), want)
	}
}
