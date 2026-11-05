package spectable

// Estes testes fixam a API do pacote de apoio que confere as tabelas de
// parâmetros da spec (docs/spec/README.md, "Tabelas de parâmetros"):
//
//	type Params map[string]float64 // parâmetro (a expressão Go, como "gopher.hungryAt") -> valor
//	func Read(path string) (Params, error)
//	func Diff(file, pkg string, spec, code Params) []string
//	func Check(t testing.TB, pkg string, files map[string]Params)
//
// Read lê a seção "## Parâmetros" de um arquivo da spec. Diff compara a tabela
// de um arquivo com as constantes que o teste de um pacote conhece e devolve
// uma mensagem por divergência. Check, usado pelo teste de cada pacote, lê cada
// arquivo e reporta as divergências do pacote como falhas do teste.

import (
	"maps"
	"strings"
	"testing"
)

func TestReadParsesTheParametersSection(t *testing.T) {
	got, err := Read("testdata/valid.md")
	if err != nil {
		t.Fatalf("Read devolveu erro %v, esperado nenhum", err)
	}
	// O parâmetro citado nas regras, antes da seção, não entra: só conta a
	// tabela de "## Parâmetros".
	want := Params{
		"gopher.stepsPerSatiety":             1000,
		"gopher.speedSteps[Slow]":            0.25,
		"gopher.maturityRanges[Child-1].min": 600,
		"world.initialAgeFactor":             1.5,
	}
	if !maps.Equal(got, want) {
		t.Errorf("Read = %v, esperado %v", got, want)
	}
}

func TestReadRejectsAValueThatIsNotANumber(t *testing.T) {
	// A spec pede ponto decimal: "0,15" não é um número.
	if _, err := Read("testdata/not_a_number.md"); err == nil {
		t.Error("Read aceitou o valor 0,15, esperado um erro")
	}
}

func TestReadRejectsAParameterWithoutBackticks(t *testing.T) {
	if _, err := Read("testdata/no_backticks.md"); err == nil {
		t.Error("Read aceitou um parâmetro sem crases, esperado um erro")
	}
}

func TestReadRejectsAFileWithoutTheParametersSection(t *testing.T) {
	if _, err := Read("testdata/no_section.md"); err == nil {
		t.Error("Read aceitou um arquivo sem a seção ## Parâmetros, esperado um erro")
	}
}

func TestDiffReportsAValueThatDiffers(t *testing.T) {
	tests := []struct {
		name       string
		spec, code Params
		want       string
	}{
		{
			name: "inteiro",
			spec: Params{"gopher.hungryAt": 3},
			code: Params{"gopher.hungryAt": 4},
			want: "gophers.md: gopher.hungryAt: a spec diz 3, o código diz 4",
		},
		{
			name: "decimal",
			spec: Params{"gopher.turnChance": 0.15},
			code: Params{"gopher.turnChance": 0.2},
			want: "gophers.md: gopher.turnChance: a spec diz 0.15, o código diz 0.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Diff("gophers.md", "gopher", tt.spec, tt.code)
			if len(got) != 1 || got[0] != tt.want {
				t.Errorf("Diff = %q, esperado [%q]", got, tt.want)
			}
		})
	}
}

func TestDiffReportsAParameterTheTestDoesNotKnow(t *testing.T) {
	spec := Params{"gopher.hungryAt": 3, "gopher.sleepChance": 0.1}
	code := Params{"gopher.hungryAt": 3}

	got := Diff("gophers.md", "gopher", spec, code)
	if len(got) != 1 {
		t.Fatalf("Diff = %q, esperada uma mensagem sobre gopher.sleepChance", got)
	}
	if !strings.Contains(got[0], "gophers.md") || !strings.Contains(got[0], "gopher.sleepChance") {
		t.Errorf("Diff = %q, esperado citar gophers.md e gopher.sleepChance", got[0])
	}
}

func TestDiffReportsAKnownParameterMissingFromTheTable(t *testing.T) {
	spec := Params{"gopher.hungryAt": 3}
	code := Params{"gopher.hungryAt": 3, "gopher.initialSatiety": 10}

	got := Diff("gophers.md", "gopher", spec, code)
	if len(got) != 1 {
		t.Fatalf("Diff = %q, esperada uma mensagem sobre gopher.initialSatiety", got)
	}
	if !strings.Contains(got[0], "gophers.md") || !strings.Contains(got[0], "gopher.initialSatiety") {
		t.Errorf("Diff = %q, esperado citar gophers.md e gopher.initialSatiety", got[0])
	}
}

func TestDiffIgnoresParametersOfOtherPackages(t *testing.T) {
	// simulacao.md mistura parâmetros de vários pacotes, e o teste de cada
	// pacote só conhece os seus.
	spec := Params{
		"main.gopherCount":      60,
		"gopher.ticksPerSecond": 60,
		"world.ticksPerSecond":  60,
		"logs.ticksPerSecond":   60,
	}
	code := Params{"gopher.ticksPerSecond": 60}

	if got := Diff("simulacao.md", "gopher", spec, code); len(got) != 0 {
		t.Errorf("Diff = %q, esperado nenhuma divergência", got)
	}
}
