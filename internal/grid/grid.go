// Package grid divide a tela em células quadradas e guarda, em cada uma, os
// objetos que estão nela. Serve para achar depressa o que está perto de um
// ponto: em vez de conferir a distância até todos os objetos, confere só os das
// células em volta dele.
//
// Com 1000 Gophers, por exemplo, testar a colisão de cada um com todos os
// outros são quase 1 milhão de contas por rodada; com a grade, cada um só olha
// as poucas dezenas que estão nas células vizinhas.
package grid

import (
	"math"
	"slices"
)

// Grid guarda objetos do tipo T, cada um na célula do seu ponto. Quem usa a
// grade guarda também o ponto de cada objeto e avisa (com Move ou Remove) quando
// ele muda ou sai: a grade não sabe onde os objetos estão, só em que célula
// foram postos.
type Grid[T comparable] struct {
	cell       float64 // lado de cada célula, em pixels
	cols, rows int
	cells      [][]T // a célula da coluna col e da linha row é cells[row*cols+col]
}

// New cria uma grade vazia que cobre o retângulo de (0, 0) a (width, height),
// com células de lado cell. Pontos fora do retângulo também podem ser
// guardados: ficam na célula da borda mais próxima.
//
// O lado da célula é um meio-termo: se for pequeno demais, cada busca visita
// muitas células; se for grande demais, cada célula traz muitos objetos longe do
// ponto procurado. Um bom valor é perto do raio das buscas mais comuns.
func New[T comparable](width, height, cell float64) *Grid[T] {
	cols := max(1, int(math.Ceil(width/cell)))
	rows := max(1, int(math.Ceil(height/cell)))
	return &Grid[T]{cell: cell, cols: cols, rows: rows, cells: make([][]T, cols*rows)}
}

// col e row devolvem a coluna e a linha da célula em que cai a coordenada x ou
// y, trazidas para dentro da grade.
func (g *Grid[T]) col(x float64) int { return clampIndex(x/g.cell, g.cols) }
func (g *Grid[T]) row(y float64) int { return clampIndex(y/g.cell, g.rows) }

func clampIndex(v float64, n int) int {
	// Limita ainda em float64: converter para int um valor enorme não dá um
	// resultado confiável.
	return int(min(max(math.Floor(v), 0), float64(n-1)))
}

// index devolve a posição, em cells, da célula do ponto (x, y).
func (g *Grid[T]) index(x, y float64) int {
	return g.row(y)*g.cols + g.col(x)
}

// Add guarda item na célula do ponto (x, y).
func (g *Grid[T]) Add(item T, x, y float64) {
	i := g.index(x, y)
	g.cells[i] = append(g.cells[i], item)
}

// Remove tira item, que tinha sido guardado no ponto (x, y). Se ele não estiver
// lá, não faz nada.
func (g *Grid[T]) Remove(item T, x, y float64) {
	i := g.index(x, y)
	cell := g.cells[i]
	j := slices.Index(cell, item)
	if j < 0 {
		return
	}
	// A ordem dentro da célula não importa: o último toma o lugar do que sai.
	last := len(cell) - 1
	cell[j] = cell[last]
	var zero T
	cell[last] = zero // para o coletor de lixo não ver uma referência que já saiu
	g.cells[i] = cell[:last]
}

// Move avisa que item, guardado no ponto (fromX, fromY), agora está em (toX,
// toY). Só mexe na grade se ele tiver mudado de célula.
func (g *Grid[T]) Move(item T, fromX, fromY, toX, toY float64) {
	if g.index(fromX, fromY) == g.index(toX, toY) {
		return
	}
	g.Remove(item, fromX, fromY)
	g.Add(item, toX, toY)
}

// Clear tira todos os objetos. A memória das células é mantida, para ser
// reaproveitada.
func (g *Grid[T]) Clear() {
	for i, cell := range g.cells {
		clear(cell)
		g.cells[i] = cell[:0]
	}
}

// AppendNear acrescenta a dst, e devolve, os objetos das células que tocam o
// quadrado de centro (x, y) e lado 2r. Entre eles estão todos os que estão a até
// r de (x, y), mas também alguns mais distantes: quem chama confere a distância.
//
// Para não alocar memória a cada busca, passe como dst o resultado da busca
// anterior esvaziado, como em buf = g.AppendNear(buf[:0], x, y, r).
func (g *Grid[T]) AppendNear(dst []T, x, y, r float64) []T {
	c0, c1 := g.col(x-r), g.col(x+r)
	for row := g.row(y - r); row <= g.row(y+r); row++ {
		for col := c0; col <= c1; col++ {
			dst = append(dst, g.cells[row*g.cols+col]...)
		}
	}
	return dst
}
