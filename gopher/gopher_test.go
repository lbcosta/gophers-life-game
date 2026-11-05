package gopher

import (
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

const eps = 1e-4

// collisionDistance é a distância entre os centros de dois Gophers adultos
// (os que newGopher cria) abaixo da qual as hit boxes se sobrepõem.
const collisionDistance = 2 * hitboxRadius

// step é o passo de um Gopher de velocidade Normal, que é a dos Gophers criados
// por newGopher.
var step = Normal.step()

// newTestGopher cria um Gopher no meio de uma tela de width por height, que
// nunca gira por conta própria (chance sempre alta) e olha para heading.
// Os testes que precisam de giro trocam chance e intn.
func newTestGopher(width, height, heading float64) *Gopher {
	g := newGopher(
		func() float64 { return 0.99 }, // >= turnChance: não gira
		func(n int) int { return 0 },
	)
	g.satiety = math.MaxInt // os testes de movimento não querem Gophers morrendo de fome
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
	if walkEveryTicks == 1 {
		t.Skip("com walkEveryTicks = 1 todos dão um passo a cada tick, e não há fases para defasar")
	}
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

func TestSpawnDesynchronizesSatietyLoss(t *testing.T) {
	// O ponto de partida sorteado decide em que passo cai o primeiro ponto de
	// saciedade: Gophers com sorteios diferentes não perdem juntos.
	for _, start := range []int{0, 1, stepsPerSatiety / 2, stepsPerSatiety - 1} {
		g := newGopher(func() float64 { return 0.99 }, func(n int) int { return start % n })
		g.Resize(500, 500)
		g.Spawn()

		steps := 0
		for g.satiety == initialSatiety && steps <= stepsPerSatiety {
			g.move()
			steps++
		}
		if want := stepsPerSatiety - start; steps != want {
			t.Errorf("partindo de %d: perdeu o primeiro ponto no passo %d, esperado %d", start, steps, want)
		}
	}
}

// twoGophers cria dois Gophers que nunca giram por conta própria, numa tela
// grande: a em (ax, ay) olhando para headingA, e b em (bx, by) parado no lugar
// em que o teste o deixa.
func twoGophers(ax, ay, headingA, bx, by float64) (a, b *Gopher) {
	a = newTestGopher(1000, 1000, headingA)
	a.x, a.y = ax, ay
	b = newTestGopher(1000, 1000, 0)
	b.x, b.y = bx, by
	return a, b
}

func TestCollisionBouncesLikeWall(t *testing.T) {
	// b está 60px à direita de a, que anda para a direita: as hit boxes (64px)
	// passariam a se sobrepor ainda mais, então a quica de volta para a esquerda.
	a, b := twoGophers(400, 500, 0, 460, 500)

	a.move(a, b)

	if !near(a.heading, 180) {
		t.Errorf("heading = %v, esperado 180", a.heading)
	}
	if !near(a.x, 400-step) || !near(a.y, 500) {
		t.Errorf("posição (%v, %v), esperado (%v, 500)", a.x, a.y, 400-step)
	}
	if b.x != 460 || b.y != 500 {
		t.Errorf("b foi para (%v, %v), mas só a reage", b.x, b.y)
	}
}

func TestCollisionReflectsAroundTheLineBetweenCenters(t *testing.T) {
	// b está em 45º de a (para baixo e para a direita) e a anda para a direita:
	// refletir em torno da normal (a -> b invertida) manda a para cima.
	d := 60 / math.Sqrt2
	a, b := twoGophers(400, 500, 0, 400+d, 500+d)
	before := math.Hypot(a.x-b.x, a.y-b.y)

	a.move(a, b)

	if !near(a.heading, 270) {
		t.Errorf("heading = %v, esperado 270 (para cima)", a.heading)
	}
	if after := math.Hypot(a.x-b.x, a.y-b.y); after <= before {
		t.Errorf("a deveria se afastar de b, mas a distância foi de %v para %v", before, after)
	}
	if moved := math.Hypot(a.x-400, a.y-500); !near(moved, step) {
		t.Errorf("a andou %v, esperado %v (a velocidade se conserva)", moved, step)
	}
}

func TestNoCollisionWhenMovingAway(t *testing.T) {
	// a já encosta em b, mas anda para o lado oposto: pode seguir.
	a, b := twoGophers(400, 500, 180, 460, 500)

	a.move(a, b)

	if a.heading != 180 {
		t.Errorf("heading mudou para %v sem colisão", a.heading)
	}
	if !near(a.x, 400-step) {
		t.Errorf("x = %v, esperado %v", a.x, 400-step)
	}
}

func TestNoCollisionWhenFarApart(t *testing.T) {
	// 100px de distância: depois do passo ainda sobram mais de 64px.
	a, b := twoGophers(400, 500, 0, 500, 500)

	a.move(a, b)

	if a.heading != 0 || !near(a.x, 400+step) {
		t.Errorf("heading = %v, x = %v: a deveria seguir em frente", a.heading, a.x)
	}
}

func TestGopherIgnoresItselfAsOther(t *testing.T) {
	a := newTestGopher(1000, 1000, 0)
	x0 := a.x

	a.move(a) // o próprio Gopher na lista, como o Game faz

	if a.heading != 0 || !near(a.x, x0+step) {
		t.Errorf("heading = %v, x = %v: a deveria andar sem reagir a si mesmo", a.heading, a.x)
	}
}

func TestOverlappingGophersCanSeparate(t *testing.T) {
	// Sobrepostos (distância 10), a anda para longe de b: não é colisão.
	a, b := twoGophers(400, 500, 180, 410, 500)

	a.move(a, b)

	if a.heading != 180 || !near(a.x, 400-step) {
		t.Errorf("heading = %v, x = %v: a deveria poder se afastar", a.heading, a.x)
	}
}

func TestCoincidentCentersDoNotCollide(t *testing.T) {
	// Sem direção de "parede" a considerar, o passo é livre (e não gera NaN).
	a, b := twoGophers(400, 500, 0, 400, 500)

	a.move(a, b)

	if math.IsNaN(a.x) || math.IsNaN(a.y) || !near(a.x, 400+step) {
		t.Errorf("posição (%v, %v), esperado (%v, 500)", a.x, a.y, 400+step)
	}
}

func TestDeadGophersAreNotObstacles(t *testing.T) {
	// Mesma situação de TestCollisionBouncesLikeWall, mas b morreu de fome neste
	// tick: ainda está na lista, mas já sumiu da tela, então a segue em frente.
	a, b := twoGophers(400, 500, 0, 460, 500)
	b.satiety = 0

	a.move(a, b)

	if a.heading != 0 || !near(a.x, 400+step) {
		t.Errorf("heading = %v, x = %v: a deveria passar pelo Gopher morto", a.heading, a.x)
	}
}

func TestReactionsNeverLengthenTheStep(t *testing.T) {
	// collisionRange conta com isso para saber quem está perto o bastante para
	// colidir. Uma reação que alongue o passo precisa aumentar o alcance também.
	for i, react := range reactions {
		for angle := 0; angle < 360; angle += 15 {
			sin, cos := math.Sincos(radians(float64(angle)))
			a, b := twoGophers(400, 500, 0, 400+60*cos, 500+60*sin)
			for heading := 0; heading < 360; heading += 15 {
				hs, hc := math.Sincos(radians(float64(heading)))
				dx, dy := react(a, b, step*hc, step*hs)
				if got := math.Hypot(dx, dy); got > step+eps {
					t.Fatalf("reação %d, outro a %dº, passo a %dº: o passo foi de %v para %v",
						i, angle, heading, step, got)
				}
			}
		}
	}
}

func TestAroundFindsEveryGopherInCollisionRange(t *testing.T) {
	// Muitos Gophers numa tela pequena, depois de andarem um pouco: a grade da
	// colônia precisa trazer, para cada um, todos os que ele pode tocar.
	c := newTestColony(300, 71, 72)
	c.Resize(800, 600)
	for range 10 * walkEveryTicks {
		c.Update(nil)
	}

	for i, g := range c.gophers {
		around := c.around(g)
		for j, o := range c.gophers {
			if math.Hypot(g.x-o.x, g.y-o.y) <= g.collisionRange() && !slices.Contains(around, o) {
				t.Fatalf("around(Gopher %d) não trouxe o Gopher %d, a %.1f dele (alcance %.1f)",
					i, j, math.Hypot(g.x-o.x, g.y-o.y), g.collisionRange())
			}
		}
	}
}

func TestInteractDrawsAmongReactions(t *testing.T) {
	a, b := twoGophers(400, 500, 0, 460, 500)
	a.intn = func(n int) int {
		if n != len(reactions) {
			t.Errorf("intn(%d): esperado sorteio entre as %d reações", n, len(reactions))
		}
		return 0
	}

	a.move(a, b)

	if !near(a.heading, 180) {
		t.Errorf("heading = %v, esperado 180 (reação asWall)", a.heading)
	}
}

func TestGophersKeepApart(t *testing.T) {
	// Vários Gophers, de todas as velocidades, andando muito tempo numa tela
	// pequena para se esbarrarem sempre. Eles podem se sobrepor um pouco por
	// causa dos passos e de empurrões das paredes, mas nunca de verdade.
	const w, h = 500, 400
	c := newTestColony(5, 5, 6)
	c.Resize(w, h)

	minDist := math.Inf(1)
	for tick := range 100_000 {
		c.Update(nil)
		for i, a := range c.gophers {
			if !a.insideScreen() {
				t.Fatalf("tick %d: hit box fora da tela, posição (%v, %v)", tick, a.x, a.y)
			}
			for _, b := range c.gophers[:i] {
				minDist = min(minDist, math.Hypot(a.x-b.x, a.y-b.y))
			}
		}
	}

	if limit := collisionDistance - 2*Fast.step(); minDist < limit {
		t.Errorf("menor distância entre dois Gophers: %v, esperado pelo menos %v", minDist, limit)
	}
	t.Logf("menor distância entre dois Gophers: %.1f (hit boxes se tocam em %v)", minDist, collisionDistance)
}

func TestSpawnAvoidsOverlappingOthers(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 8))
	var gophers []*Gopher
	for i := range 5 {
		g := newGopher(r.Float64, r.IntN)
		g.Resize(800, 600)
		g.Spawn(gophers...)
		for _, o := range gophers {
			if d := math.Hypot(g.x-o.x, g.y-o.y); d < collisionDistance {
				t.Fatalf("Gopher %d nasceu a %v de outro, sobreposto", i, d)
			}
		}
		gophers = append(gophers, g)
	}
}

func TestSpawnGivesUpWhenThereIsNoRoom(t *testing.T) {
	// Tela que só comporta um Gopher: o segundo nasce sobreposto, mas dentro dela.
	r := rand.New(rand.NewPCG(9, 10))
	a := newGopher(r.Float64, r.IntN)
	a.Resize(100, 100)
	a.Spawn()
	b := newGopher(r.Float64, r.IntN)
	b.Resize(100, 100)

	b.Spawn(a)

	if !b.insideScreen() {
		t.Errorf("hit box fora da tela, posição (%v, %v)", b.x, b.y)
	}
}

func TestSpeedsAreDistinctAndOrdered(t *testing.T) {
	slow, normal, fast := Slow.step(), Normal.step(), Fast.step()
	if !(slow < normal && normal < fast) {
		t.Errorf("passos %v, %v, %v: esperado lento < normal < rápido", slow, normal, fast)
	}
	if len(Speeds) != 3 {
		t.Errorf("len(Speeds) = %d, esperado 3", len(Speeds))
	}
}

func TestGopherWalksAtItsOwnSpeed(t *testing.T) {
	for _, speed := range Speeds {
		g := newTestGopher(1000, 1000, 0)
		g.stepSize = speed.step()
		x0 := g.x

		g.move()

		if !near(g.x-x0, speed.step()) {
			t.Errorf("%v: andou %v, esperado %v", speed, g.x-x0, speed.step())
		}
	}
}

func TestNewGopherUsesSpeedAndColor(t *testing.T) {
	g := NewGopher(Fast, Palette[3])

	if g.stepSize != Fast.step() || g.body != Palette[3] {
		t.Errorf("stepSize = %v, body = %v; esperado %v e %v", g.stepSize, g.body, Fast.step(), Palette[3])
	}
}

// lab converte uma cor sRGB para o espaço CIE L*a*b*, onde a distância
// euclidiana se aproxima da diferença percebida entre duas cores.
func lab(c color.RGBA) [3]float64 {
	lin := func(v uint8) float64 {
		f := float64(v) / 255
		if f <= 0.04045 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	r, g, b := lin(c.R), lin(c.G), lin(c.B)
	x := (0.4124*r + 0.3576*g + 0.1805*b) / 0.95047
	y := 0.2126*r + 0.7152*g + 0.0722*b
	z := (0.0193*r + 0.1192*g + 0.9505*b) / 1.08883
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Cbrt(t)
		}
		return 7.787*t + 16.0/116
	}
	fx, fy, fz := f(x), f(y), f(z)
	return [3]float64{116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)}
}

func colorDistance(a, b color.RGBA) float64 {
	la, lb := lab(a), lab(b)
	return math.Sqrt(math.Pow(la[0]-lb[0], 2) + math.Pow(la[1]-lb[1], 2) + math.Pow(la[2]-lb[2], 2))
}

func TestPaletteHasSevenDistinctColors(t *testing.T) {
	const minDistance = 40 // acima de ~30 as cores já são claramente diferentes

	if len(Palette) != 7 {
		t.Fatalf("len(Palette) = %d, esperado 7", len(Palette))
	}
	for i, a := range Palette {
		for j, b := range Palette[:i] {
			if d := colorDistance(a, b); d < minDistance {
				t.Errorf("cores %d e %d parecidas demais: distância %.1f < %v", i, j, d, minDistance)
			}
		}
	}
}

func TestPaletteStandsOutFromGrassAndEyes(t *testing.T) {
	const minDistance = 40
	grass := color.RGBA{0x3A, 0x5C, 0x28, 255} // verde da grama já escurecida
	white := eyeColor

	for i, c := range Palette {
		if d := colorDistance(c, grass); d < minDistance {
			t.Errorf("cor %d se confunde com a grama: distância %.1f < %v", i, d, minDistance)
		}
		if d := colorDistance(c, white); d < minDistance {
			t.Errorf("cor %d se confunde com o branco dos olhos: distância %.1f < %v", i, d, minDistance)
		}
	}
}

func TestNewGophersHaveDistinctColorsUpToPaletteSize(t *testing.T) {
	for range 200 {
		seen := map[color.RGBA]bool{}
		for _, g := range NewGophers(len(Palette)) {
			if seen[g.body] {
				t.Fatalf("cor %v repetida com n igual ao tamanho da paleta", g.body)
			}
			seen[g.body] = true
		}
	}
}

func TestNewGophersReshufflesPaletteEachRound(t *testing.T) {
	// Com o dobro de Gophers, cada cor aparece exatamente duas vezes.
	count := map[color.RGBA]int{}
	for _, g := range NewGophers(2 * len(Palette)) {
		count[g.body]++
	}
	for _, c := range Palette {
		if count[c] != 2 {
			t.Errorf("cor %v usada %d vezes, esperado 2", c, count[c])
		}
	}
}

func TestNewGophersColorsComeFromThePermutation(t *testing.T) {
	// Permutação fixa: o primeiro Gopher recebe Palette[3], o segundo Palette[0]...
	perm := func(n int) []int { return []int{3, 0, 6, 1, 5, 2, 4} }
	gophers := newGophers(3, perm, func(n int) int { return 0 })

	for i, want := range []int{3, 0, 6} {
		if gophers[i].body != Palette[want] {
			t.Errorf("Gopher %d: cor %v, esperado Palette[%d]", i, gophers[i].body, want)
		}
	}
}

func TestNewGophersDrawSpeedsAtRandom(t *testing.T) {
	draws := []int{2, 0, 1, 1, 0}
	i := 0
	intn := func(n int) int {
		if n != len(Speeds) {
			t.Errorf("intn(%d): esperado sorteio entre as %d velocidades", n, len(Speeds))
		}
		v := draws[i]
		i++
		return v
	}

	gophers := newGophers(len(draws), rand.Perm, intn)

	for k, d := range draws {
		if want := Speeds[d].step(); gophers[k].stepSize != want {
			t.Errorf("Gopher %d: passo %v, esperado %v", k, gophers[k].stepSize, want)
		}
	}
}

func TestNewGophersUseEverySpeed(t *testing.T) {
	r := rand.New(rand.NewPCG(11, 12))
	seen := map[float64]bool{}
	for _, g := range newGophers(100, r.Perm, r.IntN) {
		seen[g.stepSize] = true
	}
	if len(seen) != len(Speeds) {
		t.Errorf("%d velocidades diferentes em 100 Gophers, esperado %d", len(seen), len(Speeds))
	}
}

func TestSpeedString(t *testing.T) {
	tests := map[Speed]string{Slow: "lento", Normal: "normal", Fast: "rápido", Speed(7): "Speed(7)", Speed(-1): "Speed(-1)"}
	for speed, want := range tests {
		if got := speed.String(); got != want {
			t.Errorf("Speed(%d).String() = %q, esperado %q", int(speed), got, want)
		}
	}
}

// newTestColony cria uma colônia de n Gophers com sorteios de semente fixa,
// passando por todas as velocidades.
func newTestColony(n int, seed1, seed2 uint64) *Colony {
	r := rand.New(rand.NewPCG(seed1, seed2))
	c := &Colony{}
	for i := range n {
		g := newGopher(r.Float64, r.IntN)
		g.satiety = math.MaxInt
		g.stepSize = Speeds[i%len(Speeds)].step()
		c.gophers = append(c.gophers, g)
	}
	return c
}

func TestColonyDoesNothingBeforeResize(t *testing.T) {
	c := newTestColony(3, 13, 14)

	for range 2 * walkEveryTicks {
		c.Update(nil)
	}

	for i, g := range c.gophers {
		if g.x != hitboxRadius || g.y != hitboxRadius || g.ticks != 0 {
			t.Errorf("Gopher %d mudou antes do primeiro Resize: posição (%v, %v), ticks %d", i, g.x, g.y, g.ticks)
		}
	}
}

func TestColonySpawnsOnFirstResizeOnly(t *testing.T) {
	c := newTestColony(5, 15, 16)

	c.Resize(800, 600)

	type pos struct{ x, y float64 }
	spawned := make([]pos, len(c.gophers))
	for i, g := range c.gophers {
		if !g.insideScreen() {
			t.Errorf("Gopher %d nasceu fora da tela, em (%v, %v)", i, g.x, g.y)
		}
		for j, o := range c.gophers[:i] {
			if d := math.Hypot(g.x-o.x, g.y-o.y); d < collisionDistance {
				t.Errorf("Gophers %d e %d nasceram sobrepostos, a %v um do outro", i, j, d)
			}
		}
		spawned[i] = pos{g.x, g.y}
	}

	c.Resize(800, 600) // de novo: não deve espalhá-los outra vez

	for i, g := range c.gophers {
		if (pos{g.x, g.y}) != spawned[i] {
			t.Errorf("Gopher %d foi espalhado de novo: de %v para (%v, %v)", i, spawned[i], g.x, g.y)
		}
	}
}

func TestHighlightsStayInsidePupilsAndTowardsTheLight(t *testing.T) {
	if len(shinyCenters) == 0 {
		t.Fatal("nenhuma parte com brilho")
	}
	for _, s := range shapes {
		if !s.shiny {
			continue
		}
		if s.rx != s.ry {
			t.Fatalf("parte com brilho não é circular: rx = %v, ry = %v", s.rx, s.ry)
		}
		// O deslocamento é o mesmo em qualquer direção, então basta conferir uma vez.
		if d := math.Hypot(float64(highlightOffset.x), float64(highlightOffset.y)); d+highlightRadius > float64(s.rx) {
			t.Errorf("brilho passa da pupila: %v + %v > %v", d, highlightRadius, s.rx)
		}
	}

	for facing := 0.0; facing < 360; facing += 15 {
		sin, cos := math.Sincos(radians(facing - spriteHeading))
		for i, c := range shinyCenters {
			h := highlightCenter(c, facing)
			pupil := c.rotate(float32(sin), float32(cos))
			// A luz vem de cima e da esquerda, qualquer que seja a direção do Gopher.
			if h.x >= pupil.x || h.y >= pupil.y {
				t.Errorf("facing %v: brilho %d em %+v não está acima e à esquerda da pupila %+v", facing, i, h, pupil)
			}
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
	r := rand.New(rand.NewPCG(21, 22))
	seen := map[maturity]bool{}
	for range 1000 {
		m := newMaturity(r.IntN)
		seen[m] = true
		for i, age := range m {
			if age < maturityRanges[i].min || age > maturityRanges[i].max {
				t.Fatalf("maturidade %v fora do intervalo %+v na fase %d", m, maturityRanges[i], i)
			}
			if i > 0 && age <= m[i-1] {
				t.Fatalf("maturidade %v não é crescente", m)
			}
		}
	}
	if len(seen) < 100 {
		t.Errorf("só %d maturidades diferentes em 1000 sorteios", len(seen))
	}
}

func TestNewGopherIsBornBabyWithDrawnMaturity(t *testing.T) {
	g := NewGopher(Normal, Palette[0])
	if g.Stage() != Baby || g.age != 0 {
		t.Errorf("fase %v com idade %d, esperado bebê recém-nascido", g.Stage(), g.age)
	}
	if g.hitbox() >= hitboxRadius {
		t.Errorf("hit box do bebê = %v, esperado menor que %v", g.hitbox(), hitboxRadius)
	}
}

func TestStageChangesExactlyAtMaturity(t *testing.T) {
	g := newTestGopher(800, 600, 0)
	g.age = 0
	g.maturity = maturity{100, 200, 300}

	want := []struct {
		age   int
		stage Stage
	}{{0, Baby}, {99, Baby}, {100, Child}, {199, Child}, {200, Teen}, {299, Teen}, {300, Adult}, {5000, Adult}}
	for _, w := range want {
		g.age = w.age
		if got := g.Stage(); got != w.stage {
			t.Errorf("idade %d: fase %v, esperado %v", w.age, got, w.stage)
		}
	}
}

func TestUpdateAgesTheGopher(t *testing.T) {
	g := newTestGopher(800, 600, 0)
	g.age = 0
	for range 10 {
		g.Update()
	}
	if g.age != 10 {
		t.Errorf("age = %d, esperado 10", g.age)
	}
}

func TestSpawnDoesNotAgeTheGopher(t *testing.T) {
	g := newTestGopher(800, 600, 0)
	g.age = 0
	g.intn = func(n int) int { return n - 1 } // defasagem máxima dos passos
	g.Spawn()

	if g.age != 0 {
		t.Errorf("age = %d depois de Spawn, esperado 0", g.age)
	}
}

func TestBodyAndHitboxGrowWithAge(t *testing.T) {
	g := newTestGopher(800, 600, 0)
	g.maturity = maturity{100, 200, 300}

	g.age = 0
	if g.bodySize() != babySize {
		t.Errorf("corpo ao nascer = %v, esperado %v", g.bodySize(), babySize)
	}
	// No começo de cada fase o corpo tem o tamanho dela.
	for age, want := range map[int]float32{100: childSize, 200: teenSize, 300: bodyRadius, 1000: bodyRadius} {
		g.age = age
		if got := g.bodySize(); math.Abs(float64(got-want)) > 1e-3 {
			t.Errorf("idade %d: corpo %v, esperado %v", age, got, want)
		}
	}

	// O corpo nunca encolhe, e a hit box nunca é menor que o corpo nem maior que a do adulto.
	prev := float32(0)
	for age := 0; age <= 400; age++ {
		g.age = age
		if g.bodySize() < prev {
			t.Fatalf("idade %d: corpo encolheu de %v para %v", age, prev, g.bodySize())
		}
		prev = g.bodySize()
		if h := g.hitbox(); h < float64(g.bodySize())-1e-3 || h > hitboxRadius+1e-3 {
			t.Fatalf("idade %d: hit box %v fora de [%v, %v]", age, h, g.bodySize(), hitboxRadius)
		}
	}

	g.age = 300
	if g.hitbox() != hitboxRadius {
		t.Errorf("hit box do adulto = %v, esperado %v", g.hitbox(), hitboxRadius)
	}
}

func TestHitboxContainsTheDrawnFigureAtEveryStage(t *testing.T) {
	for stage := Baby; stage <= Adult; stage++ {
		var visible []part
		for _, p := range shapes {
			if p.from <= stage {
				visible = append(visible, p)
			}
		}
		if got := stageReach[stage]; got != reach(visible) {
			t.Errorf("fase %v: alcance %v, esperado %v", stage, got, reach(visible))
		}
	}
	if stageReach[Baby] != 1 || stageReach[Child] != 1 {
		t.Errorf("bebê e criança são só o corpo: alcance %v e %v, esperado 1", stageReach[Baby], stageReach[Child])
	}
	if !(stageReach[Teen] > 1 && stageReach[Teen] < stageReach[Adult]) {
		t.Errorf("alcance do adolescente %v, esperado entre 1 e %v", stageReach[Teen], stageReach[Adult])
	}
}

func TestGrowingGopherStaysInsideTheScreen(t *testing.T) {
	g := newTestGopher(300, 300, 0)
	g.age = 0
	g.maturity = maturity{50, 100, 150}
	g.x, g.y = 7, 293 // encostado no canto, onde um corpo maior não cabe
	g.keepInside()

	for range 200 {
		g.Update()
		if !g.insideScreen() {
			t.Fatalf("idade %d: hit box %v fora da tela, posição (%v, %v)", g.age, g.hitbox(), g.x, g.y)
		}
	}
}

func TestBabiesFitWhereAdultsDoNot(t *testing.T) {
	c := &Colony{gophers: NewGophers(5)}
	c.Resize(300, 200) // cinco adultos de 64px não cabem folgados aqui
	for i, g := range c.gophers {
		if !g.insideScreen() {
			t.Errorf("Gopher %d nasceu fora da tela, em (%v, %v)", i, g.x, g.y)
		}
	}
}

func TestGopherIsBornWithSatietyAndLosesOnePerThousandSteps(t *testing.T) {
	g := newTestGopher(800, 600, 0)
	g.satiety = initialSatiety
	if initialSatiety != 10 || stepsPerSatiety != 1000 {
		t.Fatalf("initialSatiety = %d e stepsPerSatiety = %d, esperado 10 e 1000", initialSatiety, stepsPerSatiety)
	}

	for step := 1; step <= 3*stepsPerSatiety; step++ {
		g.move()
		if want := initialSatiety - step/stepsPerSatiety; g.satiety != want {
			t.Fatalf("passo %d: saciedade %d, esperado %d", step, g.satiety, want)
		}
	}
}

func TestGopherStarvesAndStopsMoving(t *testing.T) {
	g := newTestGopher(800, 600, 0)
	g.satiety = initialSatiety

	for range initialSatiety * stepsPerSatiety * walkEveryTicks {
		g.Update()
	}
	if g.Alive() || g.satiety != 0 {
		t.Fatalf("saciedade %d, vivo = %v; esperado morto com 0", g.satiety, g.Alive())
	}

	x, y := g.Position()
	for range 10 * walkEveryTicks {
		g.Update()
	}
	if nx, ny := g.Position(); nx != x || ny != y || g.satiety != 0 {
		t.Errorf("Gopher morto mudou: posição (%v, %v) -> (%v, %v), saciedade %d", x, y, nx, ny, g.satiety)
	}
}

func TestEatAddsSatietyButNotToTheDead(t *testing.T) {
	g := newTestGopher(800, 600, 0)
	g.satiety = 5
	g.Eat(2)
	if g.satiety != 7 {
		t.Errorf("saciedade %d depois de comer 2, esperado 7", g.satiety)
	}

	g.satiety = 0
	g.Eat(2)
	if g.Alive() {
		t.Error("Gopher morto voltou a viver ao comer")
	}
}

func TestColonyRemovesStarvedGophersAndFeedsTheRest(t *testing.T) {
	c := newTestColony(3, 31, 32)
	c.Resize(800, 600)
	starved, fed := c.gophers[0], c.gophers[1]

	// Os dois estão a um passo de perder 1 de saciedade. O primeiro, com só 1,
	// morre; o segundo tem 2 e come mais 2.
	starved.satiety, starved.steps = 1, stepsPerSatiety-1
	fed.satiety, fed.steps = 2, stepsPerSatiety-1
	forage := func(x, y, r float64, room int) int {
		if x == fed.x && y == fed.y {
			return 2
		}
		return 0
	}
	for range walkEveryTicks {
		c.Update(forage)
	}

	if len(c.Gophers()) != 2 {
		t.Fatalf("%d Gophers vivos, esperado 2", len(c.Gophers()))
	}
	for _, g := range c.Gophers() {
		if g == starved {
			t.Error("o Gopher que morreu de fome continua na colônia")
		}
	}
	if fed.satiety < 3 {
		t.Errorf("o que comeu ficou com saciedade %d, esperado pelo menos 3 (2 - 1 + 2)", fed.satiety)
	}
}

// lonelyGopher cria uma colônia de um Gopher só, numa tela de 800x600, com a
// saciedade informada e longe de perder um ponto no próximo passo.
func lonelyGopher(seed1, seed2 uint64, satiety int) (*Colony, *Gopher) {
	c := newTestColony(1, seed1, seed2)
	c.Resize(800, 600)
	g := c.gophers[0]
	g.satiety, g.steps = satiety, 0
	return c, g
}

// rememberRoom devolve um Forager que não dá comida nenhuma e anota, em rooms,
// a Folga do Gopher a cada Busca por comida.
func rememberRoom(rooms *[]int) Forager {
	return func(x, y, r float64, room int) int {
		*rooms = append(*rooms, room)
		return 0
	}
}

func TestColonyForagesWithWhatTheGopherStillEats(t *testing.T) {
	// GOP-28: o teto é 10, e a Busca por comida fica sabendo quanto falta para
	// ele. GOP-34: o Gopher cheio também busca (com 0), porque ainda pisoteia as
	// Plantas crescendo.
	for _, c := range []struct{ satiety, room int }{{10, 0}, {9, 1}, {4, 6}} {
		col, _ := lonelyGopher(41, 42, c.satiety)
		var rooms []int
		col.Update(rememberRoom(&rooms))
		if want := []int{c.room}; !slices.Equal(rooms, want) {
			t.Errorf("saciedade %d: buscas com %v, esperado %v", c.satiety, rooms, want)
		}
	}
}

func TestEatingNeverFillsSatietyAboveTheCap(t *testing.T) {
	// GOP-28 e GOP-29: comer soma à Saciedade sem passar do teto 10, e o que
	// passa se perde.
	for _, c := range []struct{ satiety, eaten, want int }{
		{9, 2, 10},  // a batata enche o Gopher, e 1 se perde
		{10, 2, 10}, // cheio continua com 10, mesmo que a busca dê comida
		{7, 2, 9},   // abaixo do teto, soma tudo
	} {
		col, g := lonelyGopher(43, 44, c.satiety)
		col.Update(func(x, y, r float64, room int) int { return c.eaten })
		if g.Satiety() != c.want {
			t.Errorf("com %d, comendo %d: saciedade %d, esperado %d", c.satiety, c.eaten, g.Satiety(), c.want)
		}
	}
}

func TestRoomIsWhatIsLeftUntilTheCap(t *testing.T) {
	// GOP-28: a Folga é quanto falta para a Saciedade chegar ao teto 10, e é
	// zero quando o Gopher está Cheio. Saciedade acima do teto só existe nos
	// testes (como a dos helpers, math.MaxInt), e a Folga continua zero.
	for _, c := range []struct{ satiety, room int }{
		{9, 1},
		{10, 0},
		{math.MaxInt, 0},
	} {
		g := newTestGopher(800, 600, 0)
		g.satiety = c.satiety
		if got := g.Room(); got != c.room {
			t.Errorf("com saciedade %d: Folga %d, esperado %d", c.satiety, got, c.room)
		}
	}
}

func TestEatNeverLowersSatiety(t *testing.T) {
	// GOP-28 e GOP-29: comer só limita o que soma ao teto 10; nunca tira
	// Saciedade. Um Gopher acima do teto (só possível em teste) continua com o
	// que tinha, coma 0 ou 2.
	for _, c := range []struct{ satiety, eaten, want int }{
		{math.MaxInt, 0, math.MaxInt},
		{math.MaxInt, 2, math.MaxInt},
		{9, 2, 10}, // a batata enche o Gopher, e 1 se perde
	} {
		g := newTestGopher(800, 600, 0)
		g.satiety = c.satiety
		g.Eat(c.eaten)
		if g.Satiety() != c.want {
			t.Errorf("com %d, comendo %d: saciedade %d, esperado %d", c.satiety, c.eaten, g.Satiety(), c.want)
		}
	}
}

func TestColonyForagesWithTheGophersRoom(t *testing.T) {
	// GOP-28 e GOP-34: a Colônia passa para a Busca por comida a Folga que o
	// próprio Gopher calcula, inclusive para um Gopher acima do teto (Folga 0).
	for _, satiety := range []int{10, 9, 4, math.MaxInt} {
		col, g := lonelyGopher(45, 46, satiety)
		want := []int{g.Room()}
		var rooms []int
		col.Update(rememberRoom(&rooms))
		if !slices.Equal(rooms, want) {
			t.Errorf("saciedade %d: buscas com %v, esperado %v (a Folga do Gopher)", satiety, rooms, want)
		}
	}
}

func TestNewColonyGophersAreBornFull(t *testing.T) {
	// GOP-01 e GOP-28: todo Gopher nasce com Saciedade 10, que é o teto, então
	// nasce cheio.
	c := NewColony(1)
	c.Resize(800, 600)
	g := c.Gophers()[0]
	g.steps = 0 // longe de perder um ponto no próximo passo
	if g.Satiety() != 10 {
		t.Fatalf("nasceu com saciedade %d, esperado 10", g.Satiety())
	}

	var rooms []int
	c.Update(rememberRoom(&rooms))
	if want := []int{0}; !slices.Equal(rooms, want) {
		t.Errorf("recém-nascido buscou comida com %v, esperado %v (cheio)", rooms, want)
	}
}

func TestHungryFromThresholdDownWhileAlive(t *testing.T) {
	g := newTestGopher(800, 600, 0)
	for satiety, want := range map[int]bool{hungryAt + 1: false, hungryAt: true, 1: true, 0: false} {
		g.satiety = satiety
		if g.Hungry() != want {
			t.Errorf("saciedade %d: Hungry() = %v, esperado %v", satiety, g.Hungry(), want)
		}
	}
}

// stepUntilSatietyDrops faz a colônia andar um passo, deixando antes cada
// Gopher em gophers a um passo de perder 1 de saciedade, e devolve os eventos.
func stepUntilSatietyDrops(c *Colony, forage Forager, gophers ...*Gopher) []Event {
	for _, g := range gophers {
		g.steps = stepsPerSatiety - 1
	}
	var events []Event
	for range walkEveryTicks {
		events = append(events, c.Update(forage)...)
	}
	return events
}

func TestColonyReportsHungerOnceWhenItStarts(t *testing.T) {
	c := newTestColony(3, 33, 34)
	c.Resize(800, 600)
	g := c.gophers[1]
	g.name = "Ana"
	g.satiety = hungryAt + 1

	got := stepUntilSatietyDrops(c, nil, g)
	if want := []Event{{Kind: BecameHungry, Name: "Ana"}}; !slices.Equal(got, want) {
		t.Fatalf("ao chegar a %d de saciedade: eventos %v, esperado %v", hungryAt, got, want)
	}

	// Continuar com fome, ou com mais fome, não é novidade.
	if got := stepUntilSatietyDrops(c, nil, g); len(got) != 0 {
		t.Errorf("ao cair de %d para %d: eventos %v, esperado nenhum", hungryAt, hungryAt-1, got)
	}
}

func TestColonyReportsHungerAgainAfterEatingEnough(t *testing.T) {
	// Um Gopher só, porque feed alimenta todos, e os outros de newTestColony já
	// têm saciedade math.MaxInt.
	c := newTestColony(1, 35, 36)
	c.Resize(800, 600)
	g := c.gophers[0]
	g.name = "Bia"
	g.satiety = hungryAt

	// Come 2 (e perde 1): sai da fome, sem evento.
	feed := func(x, y, r float64, room int) int { return 2 }
	if got := stepUntilSatietyDrops(c, feed, g); len(got) != 0 || g.Hungry() {
		t.Fatalf("ao comer: eventos %v e saciedade %d, esperado nenhum evento e sem fome", got, g.satiety)
	}

	got := stepUntilSatietyDrops(c, nil, g)
	if want := []Event{{Kind: BecameHungry, Name: "Bia"}}; !slices.Equal(got, want) {
		t.Errorf("ao voltar a ter fome: eventos %v, esperado %v", got, want)
	}
}

func TestColonyReportsStarvationOnly(t *testing.T) {
	c := newTestColony(3, 37, 38)
	c.Resize(800, 600)
	a, b := c.gophers[0], c.gophers[2]
	a.name, b.name = "Caio", "Davi"
	a.satiety, b.satiety = 1, hungryAt+1

	got := stepUntilSatietyDrops(c, nil, a, b)
	want := []Event{{Kind: Starved, Name: "Caio"}, {Kind: BecameHungry, Name: "Davi"}}
	if !slices.Equal(got, want) {
		t.Errorf("eventos %v, esperado %v (quem morre não passa a ter fome, e na ordem da colônia)", got, want)
	}
}

func TestColonyReportsNothingBeforeResize(t *testing.T) {
	c := newTestColony(1, 39, 40)
	c.gophers[0].satiety = 1
	if got := stepUntilSatietyDrops(c, nil, c.gophers[0]); len(got) != 0 {
		t.Errorf("eventos %v antes do primeiro Resize, esperado nenhum", got)
	}
}

func TestEventString(t *testing.T) {
	cases := []struct {
		e    Event
		want string
	}{
		{Event{Kind: BecameHungry, Name: "Ana"}, "Ana está passando fome"},
		{Event{Kind: Starved, Name: "Ana"}, "Ana morreu de fome"},
		{Event{Kind: 99, Name: "Ana"}, "Ana: EventKind(99)"},
		{Event{Kind: EventKind(-1), Name: "Ana"}, "Ana: EventKind(-1)"},
	}
	for _, c := range cases {
		if got := c.e.String(); got != c.want {
			t.Errorf("%#v.String() = %q, esperado %q", c.e, got, c.want)
		}
	}
}

// Os tipos válidos devolvem a frase sem o nome, como o Speed devolve "lento"
// (issue #8, pergunta 1); os inválidos, dos dois lados da tabela, devolvem
// EventKind(n), como pede a convenção de enums.
func TestEventKindString(t *testing.T) {
	cases := []struct {
		kind EventKind
		want string
	}{
		{BecameHungry, "está passando fome"},
		{Starved, "morreu de fome"},
		{EventKind(-1), "EventKind(-1)"},
		{numEventKinds, "EventKind(2)"},
	}
	for _, c := range cases {
		if got := c.kind.String(); got != c.want {
			t.Errorf("EventKind(%d).String() = %q, esperado %q", int(c.kind), got, c.want)
		}
	}
}

func TestParseNamesSkipsCommentsAndBlankLines(t *testing.T) {
	got := parseNames("# comentário\n\nAna\n  Bia  \n#Outro\nCaio\n")
	want := []string{"Ana", "Bia", "Caio"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("parseNames = %v, esperado %v", got, want)
	}
}

func TestNamesFileHasNoRepeatsAndNoneOfTheHundredMostPopular(t *testing.T) {
	// Os 100 nomes mais populares do Brasil segundo os censos do IBGE, sem acentos.
	popular := "Maria Jose Ana Joao Antonio Francisco Carlos Paulo Pedro Lucas Luiz Marcos Luis Gabriel " +
		"Rafael Francisca Daniel Marcelo Bruno Eduardo Felipe Raimundo Rodrigo Antonia Manoel " +
		"Mateus Andre Adriana Juliana Fernando Marcia Fabio Leonardo Gustavo Fernanda Patricia " +
		"Guilherme Aline Leandro Tiago Sandra Anderson Camila Ricardo Marcio Amanda Jorge Bruna " +
		"Sebastiao Jessica Alexandre Roberto Leticia Julia Edson Luciana Diego Vitor Vanessa Sergio " +
		"Mariana Gabriela Vera Vitoria Larissa Claudia Beatriz Luana Rita Sonia Claudio Matheus " +
		"Renata Thiago Eliane Josefa Simone Geraldo Adriano Luciano Julio Natalia Renato Cristiane " +
		"Carla Debora Alex Rosangela Jaqueline Rosa Vinicius Daniela Aparecida Marlene Terezinha " +
		"Raimunda Rogerio Samuel Andreia Fabiana"
	banned := map[string]bool{}
	for _, n := range strings.Fields(popular) {
		banned[n] = true
	}
	if len(banned) != 100 {
		t.Fatalf("%d nomes populares no teste, esperado 100", len(banned))
	}

	if len(Names) == 0 {
		t.Fatal("nenhum nome em names.txt")
	}
	seen := map[string]bool{}
	for _, n := range Names {
		if banned[n] {
			t.Errorf("%q está entre os 100 mais populares e devia ter sido excluído", n)
		}
		if seen[n] {
			t.Errorf("%q repetido em names.txt", n)
		}
		seen[n] = true
	}
}

func TestNewGopherDrawsNameFromNames(t *testing.T) {
	inNames := map[string]bool{}
	for _, n := range Names {
		inNames[n] = true
	}
	different := map[string]bool{}
	for range 50 {
		g := NewGopher(Normal, Palette[0])
		if !inNames[g.Name()] {
			t.Fatalf("nome %q não está em Names", g.Name())
		}
		different[g.Name()] = true
	}
	if len(different) < 10 {
		t.Errorf("só %d nomes diferentes em 50 Gophers", len(different))
	}
}

func TestAtFindsTheGopherUnderThePoint(t *testing.T) {
	c := newTestColony(2, 41, 42)
	if c.At(0, 0) != nil {
		t.Error("At devolveu um Gopher antes do primeiro Resize")
	}
	c.Resize(800, 600)
	a, b := c.gophers[0], c.gophers[1]
	a.x, a.y, b.x, b.y = 200, 200, 500, 300

	if got := c.At(a.x+hitboxRadius-1, a.y); got != a {
		t.Errorf("At dentro da hit box de a devolveu %v", got)
	}
	if got := c.At(b.x, b.y); got != b {
		t.Errorf("At no centro de b devolveu %v", got)
	}
	if got := c.At(a.x+hitboxRadius+1, a.y); got != nil {
		t.Errorf("At fora das hit boxes devolveu %v", got)
	}

	// Sobrepostos: vence o que é desenhado por cima, o último.
	b.x, b.y = a.x+10, a.y
	if got := c.At(a.x+5, a.y); got != b {
		t.Errorf("At sobreposto devolveu %v, esperado o de cima (b)", got)
	}

	// Morto não é encontrado.
	b.satiety = 0
	if got := c.At(a.x+5, a.y); got != a {
		t.Errorf("At devolveu %v, esperado a (b está morto)", got)
	}
}

func TestHoverTracksTheGopherUnderTheMouse(t *testing.T) {
	c := newTestColony(2, 43, 44)
	c.Resize(800, 600)
	a := c.gophers[0]
	a.x, a.y = 300, 300

	c.Hover(300, 300)
	if c.hovered != a {
		t.Errorf("hovered = %v, esperado a", c.hovered)
	}
	c.Hover(0, 0)
	if c.hovered != nil {
		t.Errorf("hovered = %v depois de o mouse sair, esperado nil", c.hovered)
	}
}

func TestNewGophersHaveDistinctNames(t *testing.T) {
	for range 20 {
		seen := map[string]bool{}
		for _, g := range NewGophers(50) {
			if g.Name() == "" || seen[g.Name()] {
				t.Fatalf("nome %q vazio ou repetido", g.Name())
			}
			seen[g.Name()] = true
		}
	}
}

func TestAssignNamesUsesThePermutationAndReshufflesWhenNamesRunOut(t *testing.T) {
	var asked []int
	perm := func(n int) []int {
		asked = append(asked, n)
		out := make([]int, n)
		for i := range out {
			out[i] = n - 1 - i // do último nome para o primeiro
		}
		return out
	}

	gophers := make([]*Gopher, len(Names)+2)
	for i := range gophers {
		gophers[i] = newGopher(rand.Float64, rand.IntN)
	}
	assignNames(gophers, perm)

	if gophers[0].name != Names[len(Names)-1] || gophers[1].name != Names[len(Names)-2] {
		t.Errorf("nomes %q e %q, esperado os dois últimos de Names, em ordem", gophers[0].name, gophers[1].name)
	}
	if len(asked) != 2 || asked[0] != len(Names) || asked[1] != len(Names) {
		t.Errorf("perm chamada com %v, esperado duas vezes com %d", asked, len(Names))
	}
	if last := gophers[len(Names)]; last.name != Names[len(Names)-1] {
		t.Errorf("depois de esgotar os nomes, o próximo foi %q, esperado recomeçar a permutação", last.name)
	}
}

func TestNewColonyCapsTheNumberOfGophers(t *testing.T) {
	for _, c := range []struct{ n, want int }{
		{-3, 0}, {0, 0}, {5, 5}, {MaxGophers, MaxGophers}, {MaxGophers + 1, MaxGophers}, {5000, MaxGophers},
	} {
		if got := len(NewColony(c.n).Gophers()); got != c.want {
			t.Errorf("NewColony(%d) tem %d Gophers, esperado %d", c.n, got, c.want)
		}
	}
}

func TestPaletteNamesMatchPalette(t *testing.T) {
	if len(PaletteNames) != len(Palette) {
		t.Fatalf("%d nomes de cor para %d cores", len(PaletteNames), len(Palette))
	}
	for i, c := range Palette {
		if colorName(c) != PaletteNames[i] {
			t.Errorf("colorName(Palette[%d]) = %q, esperado %q", i, colorName(c), PaletteNames[i])
		}
	}
	if got := colorName(color.RGBA{1, 2, 255, 255}); got != "#0102FF" {
		t.Errorf("cor fora da paleta: %q, esperado #0102FF", got)
	}
}

func TestProfileDescribesTheGopher(t *testing.T) {
	g := NewGopher(Fast, Palette[3])
	g.maturity = maturity{600, 1200, 1800}
	g.age = 300
	g.satiety = 7

	p := g.Profile()
	if p.Name != g.name || p.Color != Palette[3] || p.ColorName != PaletteNames[3] {
		t.Errorf("nome/cor errados: %+v", p)
	}
	if p.Speed != Fast || p.StepSize != Fast.step() || p.Satiety != 7 || !p.Alive {
		t.Errorf("velocidade/saciedade/vida erradas: %+v", p)
	}
	if p.Stage != Baby || p.Age != 5 || p.Radius != g.hitbox() {
		t.Errorf("fase/idade/raio errados: %+v", p)
	}
	if !p.HasNextStage || p.NextStage != Child || p.NextStageIn != 5 {
		t.Errorf("próxima fase errada: %+v, esperado criança em 5 s", p)
	}

	g.age = 5000
	if p := g.Profile(); p.Stage != Adult || p.HasNextStage {
		t.Errorf("adulto tem próxima fase: %+v", p)
	}

	g.satiety = 0
	if p := g.Profile(); p.Alive || p.Satiety != 0 {
		t.Errorf("morto aparece vivo ou com saciedade: %+v", p)
	}
}

func TestSelectPicksClearsAndKeepsTheDead(t *testing.T) {
	c := newTestColony(2, 51, 52)
	c.Resize(800, 600)
	a := c.gophers[0]
	a.x, a.y = 300, 300
	c.gophers[1].x, c.gophers[1].y = 600, 100

	c.Select(300, 300)
	if c.Selected() != a {
		t.Fatalf("Selected = %v, esperado a", c.Selected())
	}
	c.Select(700, 500) // clique no vazio
	if c.Selected() != nil {
		t.Errorf("Selected = %v depois de clicar no vazio, esperado nil", c.Selected())
	}

	c.Select(300, 300)
	a.satiety = 1
	a.steps = stepsPerSatiety - 1
	for range walkEveryTicks {
		c.Update(nil)
	}
	if a.Alive() {
		t.Fatal("o Gopher devia ter morrido")
	}
	if c.Selected() != a || c.Selected().Profile().Alive {
		t.Error("o selecionado que morreu devia continuar selecionado, como morto")
	}
}

func TestPaletteSlugsMatchPalette(t *testing.T) {
	if len(PaletteSlugs) != len(Palette) {
		t.Fatalf("%d slugs para %d cores", len(PaletteSlugs), len(Palette))
	}
}

func TestPhotoNameIsAlwaysOfTheGophersOwnColor(t *testing.T) {
	seen := map[string]bool{}
	for i, c := range Palette {
		for _, e := range Emotions {
			name := PhotoName(c, e)
			if !strings.HasPrefix(name, PaletteSlugs[i]+"_") || !strings.HasSuffix(name, ".png") {
				t.Errorf("PhotoName(%s, %v) = %q, esperado começar com %q", PaletteNames[i], e, name, PaletteSlugs[i]+"_")
			}
			if seen[name] {
				t.Errorf("nome de foto repetido: %q", name)
			}
			seen[name] = true
		}
	}
	if got := PhotoName(color.RGBA{1, 2, 3, 255}, Happy); got != "" {
		t.Errorf("cor fora da paleta tem foto %q, esperado nenhuma", got)
	}
}

func TestNewGopherDrawsEmotionAtBirth(t *testing.T) {
	seen := map[Emotion]bool{}
	for range 500 {
		g := NewGopher(Normal, Palette[2])
		seen[g.emotion] = true
		p := g.Profile()
		if p.Emotion != g.emotion || p.Photo != PhotoName(Palette[2], g.emotion) {
			t.Fatalf("perfil %+v não combina com a emoção %v e a cor do Gopher", p, g.emotion)
		}
	}
	if len(seen) != numEmotions {
		t.Errorf("sorteadas %d emoções em 500 Gophers, esperado todas as %d", len(seen), numEmotions)
	}
}

// BenchmarkColonyUpdate mede um tick de uma colônia lotada: MaxGophers adultos,
// mais do que cabe sem sobreposição numa tela de 1920x1080, esbarrando uns nos
// outros o tempo todo. Rode com: go test -run '^$' -bench Colony ./gopher
func BenchmarkColonyUpdate(b *testing.B) {
	c := newTestColony(MaxGophers, 61, 62)
	c.Resize(1920, 1080)
	for b.Loop() {
		c.Update(nil)
	}
}
