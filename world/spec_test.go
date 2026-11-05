package world

import (
	"testing"

	"gophers-life-game/internal/spectable"
)

// TestSpecParameterTablesMatchTheCode confere as tabelas de parâmetros da spec
// com as constantes do pacote. Cada parâmetro aponta para a expressão Go, e não
// para um número: para mudar um valor, basta mexer na tabela e na constante.
func TestSpecParameterTablesMatchTheCode(t *testing.T) {
	spectable.Check(t, "world", map[string]spectable.Params{
		"../docs/spec/simulacao.md": {
			"world.ticksPerSecond": ticksPerSecond,
		},
		"../docs/spec/gophers.md": {},
		"../docs/spec/plantas.md": {
			"world.kinds[Potato].satiety":         float64(kinds[Potato].satiety),
			"world.kinds[Carrot].satiety":         float64(kinds[Carrot].satiety),
			"world.kinds[Radish].satiety":         float64(kinds[Radish].satiety),
			"world.kinds[Grass].satiety":          float64(kinds[Grass].satiety),
			"world.kinds[Clover].satiety":         float64(kinds[Clover].satiety),
			"world.kinds[Alfalfa].satiety":        float64(kinds[Alfalfa].satiety),
			"world.maturityRanges[Growing-1].min": float64(maturityRanges[Growing-1].min),
			"world.maturityRanges[Growing-1].max": float64(maturityRanges[Growing-1].max),
			"world.maturityRanges[Ready-1].min":   float64(maturityRanges[Ready-1].min),
			"world.maturityRanges[Ready-1].max":   float64(maturityRanges[Ready-1].max),
			"world.plantRadius":                   plantRadius,
			"world.spawnEveryTicks":               spawnEveryTicks,
			"world.spawnAttempts":                 spawnAttempts,
			"world.initialAgeFactor":              initialAgeFactor,
		},
		"../docs/spec/interface.md": {},
	})
}
