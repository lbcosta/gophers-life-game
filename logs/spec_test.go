package logs

import (
	"testing"

	"gophers-life-game/internal/spectable"
)

// TestSpecParameterTablesMatchTheCode confere as tabelas de parâmetros da spec
// com as constantes do pacote. Cada parâmetro aponta para a expressão Go, e não
// para um número: para mudar um valor, basta mexer na tabela e na constante.
func TestSpecParameterTablesMatchTheCode(t *testing.T) {
	spectable.Check(t, "logs", map[string]spectable.Params{
		"../docs/spec/simulacao.md": {
			"logs.ticksPerSecond": ticksPerSecond,
		},
		"../docs/spec/gophers.md": {},
		"../docs/spec/plantas.md": {},
		"../docs/spec/interface.md": {
			"logs.maxEntries": maxEntries,
			"logs.lifetime":   lifetime,
			"logs.fadeTicks":  fadeTicks,
		},
	})
}
