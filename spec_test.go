package main

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
	spectable.Check(t, "main", map[string]spectable.Params{
		"docs/spec/simulacao.md": {
			"main.gopherCount": gopherCount,
		},
		"docs/spec/gophers.md":   {},
		"docs/spec/plantas.md":   {},
		"docs/spec/interface.md": {},
	})
}
