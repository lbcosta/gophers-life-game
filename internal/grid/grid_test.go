package grid

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

// item é um objeto de teste: um ponto com um número para identificá-lo.
type item struct {
	id   int
	x, y float64
}

func TestAppendNearFindsEverythingWithinTheRadius(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	const width, height = 800, 600
	g := New[*item](width, height, 50)

	// Alguns pontos ficam fora do retângulo, de propósito.
	var items []*item
	for i := range 500 {
		it := &item{id: i, x: r.Float64()*(width+200) - 100, y: r.Float64()*(height+200) - 100}
		items = append(items, it)
		g.Add(it, it.x, it.y)
	}

	var found []*item
	for range 200 {
		x, y := r.Float64()*(width+200)-100, r.Float64()*(height+200)-100
		radius := r.Float64() * 120
		found = g.AppendNear(found[:0], x, y, radius)
		for _, it := range items {
			if math.Hypot(it.x-x, it.y-y) <= radius && !slices.Contains(found, it) {
				t.Fatalf("busca em (%.1f, %.1f) com raio %.1f não trouxe o item %d, em (%.1f, %.1f)",
					x, y, radius, it.id, it.x, it.y)
			}
		}
		if len(found) != len(unique(found)) {
			t.Fatalf("busca em (%.1f, %.1f) trouxe itens repetidos", x, y)
		}
	}
}

func unique(items []*item) []*item {
	out := slices.Clone(items)
	slices.SortFunc(out, func(a, b *item) int { return a.id - b.id })
	return slices.Compact(out)
}

func TestAppendNearSkipsFarCells(t *testing.T) {
	g := New[*item](1000, 1000, 100)
	near, far := &item{id: 1, x: 150, y: 150}, &item{id: 2, x: 850, y: 850}
	g.Add(near, near.x, near.y)
	g.Add(far, far.x, far.y)

	got := g.AppendNear(nil, 160, 160, 30)

	if !slices.Equal(got, []*item{near}) {
		t.Errorf("busca trouxe %v, esperado só o item perto", got)
	}
}

func TestAppendNearKeepsWhatIsAlreadyInDst(t *testing.T) {
	g := New[*item](100, 100, 10)
	a, b := &item{id: 1}, &item{id: 2}
	g.Add(b, 5, 5)

	got := g.AppendNear([]*item{a}, 5, 5, 1)

	if !slices.Equal(got, []*item{a, b}) {
		t.Errorf("AppendNear devolveu %v, esperado o que já estava em dst e depois o encontrado", got)
	}
}

func TestRemoveTakesOnlyThatItem(t *testing.T) {
	g := New[*item](100, 100, 50)
	a, b := &item{id: 1}, &item{id: 2}
	g.Add(a, 10, 10)
	g.Add(b, 20, 20) // mesma célula

	g.Remove(a, 10, 10)
	g.Remove(a, 10, 10) // de novo: não está mais lá, não faz nada

	if got := g.AppendNear(nil, 10, 10, 1); !slices.Equal(got, []*item{b}) {
		t.Errorf("depois de remover a, a célula tem %v, esperado só b", got)
	}
}

func TestMoveChangesTheCell(t *testing.T) {
	g := New[*item](300, 100, 100)
	a := &item{id: 1}
	g.Add(a, 50, 50)

	g.Move(a, 50, 50, 60, 50) // mesma célula
	g.Move(a, 60, 50, 250, 50)

	if got := g.AppendNear(nil, 50, 50, 1); len(got) != 0 {
		t.Errorf("a célula antiga ainda tem %v", got)
	}
	if got := g.AppendNear(nil, 250, 50, 1); !slices.Equal(got, []*item{a}) {
		t.Errorf("a célula nova tem %v, esperado a", got)
	}
}

func TestClearEmptiesEveryCell(t *testing.T) {
	g := New[*item](100, 100, 10)
	for i := range 20 {
		g.Add(&item{id: i}, float64(i*5), float64(i*5))
	}

	g.Clear()

	if got := g.AppendNear(nil, 50, 50, 100); len(got) != 0 {
		t.Errorf("depois de Clear a grade ainda tem %d itens", len(got))
	}
}

func TestPointsFarOutsideGoToTheBorder(t *testing.T) {
	g := New[*item](100, 100, 10)
	a := &item{id: 1}
	g.Add(a, 1e300, -1e300) // canto de cima à direita

	if got := g.AppendNear(nil, 99, 0, 1); !slices.Equal(got, []*item{a}) {
		t.Errorf("a célula do canto tem %v, esperado o item de fora da tela", got)
	}
}
