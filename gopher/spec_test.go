package gopher

import (
	"testing"

	"gophers-life-game/internal/spectable"
)

// TestSpecParameterTablesMatchTheCode confere as tabelas de parâmetros da spec
// com as constantes do pacote. Cada parâmetro aponta para a expressão Go, e não
// para um número: para mudar um valor, basta mexer na tabela e na constante.
//
// Os arquivos sem parâmetros deste pacote entram no mapa com {} de propósito:
// assim um parâmetro do pacote posto no arquivo errado da spec falha como
// desconhecido, em vez de passar sem ninguém conferir.
func TestSpecParameterTablesMatchTheCode(t *testing.T) {
	spectable.Check(t, "gopher", map[string]spectable.Params{
		"../docs/spec/simulacao.md": {
			"gopher.MaxGophers":     MaxGophers,
			"gopher.ticksPerSecond": ticksPerSecond,
		},
		"../docs/spec/gophers.md": {
			"gopher.initialSatiety":              initialSatiety,
			"gopher.maxSatiety":                  maxSatiety,
			"gopher.stepsPerSatiety":             stepsPerSatiety,
			"gopher.hungryAt":                    hungryAt,
			"gopher.walkEveryTicks":              walkEveryTicks,
			"gopher.speedSteps[Slow]":            speedSteps[Slow],
			"gopher.speedSteps[Normal]":          speedSteps[Normal],
			"gopher.speedSteps[Fast]":            speedSteps[Fast],
			"gopher.turnChance":                  turnChance,
			"gopher.maxTurn":                     maxTurn,
			"gopher.turnSpeed":                   turnSpeed,
			"gopher.spawnAttempts":               spawnAttempts,
			"gopher.hitboxRadius":                hitboxRadius,
			"gopher.babySize":                    float64(babySize),
			"gopher.childSize":                   float64(childSize),
			"gopher.teenSize":                    float64(teenSize),
			"gopher.maturityRanges[Child-1].min": float64(maturityRanges[Child-1].min),
			"gopher.maturityRanges[Child-1].max": float64(maturityRanges[Child-1].max),
			"gopher.maturityRanges[Teen-1].min":  float64(maturityRanges[Teen-1].min),
			"gopher.maturityRanges[Teen-1].max":  float64(maturityRanges[Teen-1].max),
			"gopher.maturityRanges[Adult-1].min": float64(maturityRanges[Adult-1].min),
			"gopher.maturityRanges[Adult-1].max": float64(maturityRanges[Adult-1].max),
		},
		"../docs/spec/plantas.md":   {},
		"../docs/spec/interface.md": {},
	})
}
