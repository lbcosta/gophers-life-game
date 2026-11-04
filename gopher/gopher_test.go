package gopher

import (
	"math"
	"testing"
)

// newTestGopher cria um Gopher cujo sorteio sempre devolve pick (limitado à
// quantidade de direções válidas).
func newTestGopher(width, height float32, pick int) *Gopher {
	g := NewGopher(width, height)
	g.intn = func(n int) int { return min(pick, n-1) }
	return g
}

func (g *Gopher) insideScreen() bool {
	return g.fits(g.x, g.y)
}

func TestMoveNeverLeavesScreen(t *testing.T) {
	g := NewGopher(200, 150)
	for i := 0; i < 100_000; i++ {
		g.move()
		if !g.insideScreen() {
			t.Fatalf("passo %d: hit box fora da tela, posição (%v, %v)", i, g.x, g.y)
		}
	}
}

func TestMoveOnlyPicksValidDirections(t *testing.T) {
	// No canto superior esquerdo só dá para ir para a direita ou para baixo.
	g := newTestGopher(500, 500, 0)
	g.move()
	if g.x != hitboxRadius+step || g.y != hitboxRadius {
		t.Errorf("primeira direção válida deveria ser a direita, mas foi para (%v, %v)", g.x, g.y)
	}

	g = newTestGopher(500, 500, 1)
	g.move()
	if g.x != hitboxRadius || g.y != hitboxRadius+step {
		t.Errorf("segunda direção válida deveria ser para baixo, mas foi para (%v, %v)", g.x, g.y)
	}
}

func TestMoveStaysStillWithoutValidDirection(t *testing.T) {
	// Tela do tamanho exato da hit box: nenhum passo cabe.
	g := newTestGopher(2*hitboxRadius, 2*hitboxRadius, 0)
	g.move()
	if g.x != hitboxRadius || g.y != hitboxRadius {
		t.Errorf("Gopher deveria ficar parado, mas foi para (%v, %v)", g.x, g.y)
	}
}

func TestUpdateWalksOnlyEveryWalkEveryTicks(t *testing.T) {
	g := newTestGopher(500, 500, 0)
	x0, y0 := g.Position()

	for ticks := 1; ticks < walkEveryTicks; ticks++ {
		g.Update(ticks)
		if x, y := g.Position(); x != x0 || y != y0 {
			t.Fatalf("tick %d: Gopher andou antes da hora, foi para (%v, %v)", ticks, x, y)
		}
	}

	g.Update(walkEveryTicks)
	if x, y := g.Position(); x == x0 && y == y0 {
		t.Errorf("tick %d: Gopher deveria ter andado", walkEveryTicks)
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

func TestDrawingFitsInHitbox(t *testing.T) {
	const eps = 1e-3

	for i, s := range shapes {
		for _, p := range ellipsePoints(s.cx, s.cy, s.rx, s.ry, s.angle) {
			d := math.Hypot(float64(p.x*bodyRadius), float64(p.y*bodyRadius))
			if d > float64(hitboxRadius)+eps {
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
