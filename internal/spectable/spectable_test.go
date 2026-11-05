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
	"fmt"
	"maps"
	"slices"
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

func TestReadAcceptsEveryPackageOfTheSpec(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": o parâmetro tem o pacote
	// na frente, e o pacote é main., game., gopher., world. ou logs.
	got, err := Read("testdata/all_prefixes.md")
	if err != nil {
		t.Fatalf("Read devolveu erro %v, esperado nenhum", err)
	}
	want := Params{
		"main.gopherCount":     60,
		"game.plantsPerGopher": 2,
		"gopher.hungryAt":      3,
		"world.spawnAttempts":  20,
		"logs.maxEntries":      8,
	}
	if !maps.Equal(got, want) {
		t.Errorf("Read = %v, esperado %v", got, want)
	}
}

func TestReadRejectsALineWithoutTheLeadingPipe(t *testing.T) {
	// No Markdown do GitHub, a barra do começo da linha é opcional: uma linha
	// sem ela ainda aparece na tabela. Dentro da seção ## Parâmetros, toda
	// linha não vazia que não começa com | é um erro, em vez de sumir da
	// conferência em silêncio (achado S7 do PR #10, aprovado na issue #2).
	tests := []struct {
		name, file string
	}{
		{"linha da tabela sem a barra do começo", "testdata/no_leading_pipe.md"},
		{"linha de texto dentro da seção", "testdata/text_in_section.md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Read(tt.file); err == nil {
				t.Errorf("Read(%q) aceitou uma linha sem | no começo dentro da seção, esperado um erro", tt.file)
			}
		})
	}
}

func TestReadRejectsAValueThatIsNotANumber(t *testing.T) {
	// A spec pede ponto decimal: "0,15" não é um número.
	if _, err := Read("testdata/not_a_number.md"); err == nil {
		t.Error("Read aceitou o valor 0,15, esperado um erro")
	}
}

func TestReadRejectsANumberOutsideTheSpecFormat(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": o Valor é "só o número
	// [...] com ponto decimal", ou seja, ^-?\d+(\.\d+)?$ (issue #26, S9).
	// As outras grafias que o Go entende como número ficam de fora. Em cada
	// arquivo, a linha 8 tem gopher.turnChance com o valor fora do formato, e
	// o erro precisa apontar o arquivo, a linha e o parâmetro.
	tests := []struct {
		value, file string
	}{
		{"9e1", "bad_number_exponent.md"},
		{"90.", "bad_number_trailing_point.md"},
		{".5", "bad_number_leading_point.md"},
		{"+90", "bad_number_plus_sign.md"},
		{"Inf", "bad_number_inf.md"},
		{"NaN", "bad_number_nan.md"},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			_, err := Read("testdata/" + tt.file)
			if err == nil {
				t.Fatalf("Read aceitou o valor %s, esperado um erro", tt.value)
			}
			for _, s := range []string{tt.file + ":8", "gopher.turnChance"} {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("Read devolveu o erro %q, esperado citar %s", err, s)
				}
			}
		})
	}
}

func TestReadAcceptsNumbersInTheSpecFormat(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": inteiros e decimais com
	// ponto continuam valendo (issue #26, S9).
	got, err := Read("testdata/valid_numbers.md")
	if err != nil {
		t.Fatalf("Read devolveu erro %v, esperado nenhum", err)
	}
	want := Params{
		"gopher.hungryAt":         3,
		"gopher.speedSteps[Slow]": 0.25,
		"gopher.ticksPerHunger":   1200,
	}
	if !maps.Equal(got, want) {
		t.Errorf("Read = %v, esperado %v", got, want)
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
		"game.plantsPerGopher":  2,
		"gopher.ticksPerSecond": 60,
		"world.ticksPerSecond":  60,
		"logs.ticksPerSecond":   60,
	}
	code := Params{"gopher.ticksPerSecond": 60}

	if got := Diff("simulacao.md", "gopher", spec, code); len(got) != 0 {
		t.Errorf("Diff = %q, esperado nenhuma divergência", got)
	}
}

// recordingTB guarda as falhas que Check reportaria, em vez de reprovar este
// teste. Assim dá para ver se o teste de um pacote seria reprovado sem fixar
// se a linha errada cai em Read ou em Diff.
type recordingTB struct {
	testing.TB
	failures []string
}

func (r *recordingTB) Helper()           {}
func (r *recordingTB) Fail()             { r.failures = append(r.failures, "Fail") }
func (r *recordingTB) FailNow()          { r.Fail() }
func (r *recordingTB) Failed() bool      { return len(r.failures) > 0 }
func (r *recordingTB) Error(args ...any) { r.failures = append(r.failures, fmt.Sprint(args...)) }
func (r *recordingTB) Errorf(f string, a ...any) {
	r.failures = append(r.failures, fmt.Sprintf(f, a...))
}
func (r *recordingTB) Fatal(args ...any)          { r.Error(args...) }
func (r *recordingTB) Fatalf(f string, a ...any)  { r.Errorf(f, a...) }
func (r *recordingTB) Log(args ...any)            {}
func (r *recordingTB) Logf(f string, args ...any) {}

// cites diz se alguma falha registrada menciona s.
func (r *recordingTB) cites(s string) bool {
	for _, f := range r.failures {
		if strings.Contains(f, s) {
			return true
		}
	}
	return false
}

func TestCheckPassesATableThatMatchesTheCode(t *testing.T) {
	// Controle do recordingTB: uma tabela igual ao código não reprova nada.
	rec := &recordingTB{TB: t}
	Check(rec, "gopher", map[string]Params{
		"testdata/valid.md": {
			"gopher.stepsPerSatiety":             1000,
			"gopher.speedSteps[Slow]":            0.25,
			"gopher.maturityRanges[Child-1].min": 600,
		},
	})
	if rec.Failed() {
		t.Errorf("Check reprovou uma tabela igual ao código: %q, esperado nenhuma falha", rec.failures)
	}
}

func TestCheckFailsOnAParameterWithoutAKnownPackage(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": o parâmetro tem o pacote
	// na frente, e o pacote é main., game., gopher., world. ou logs. Uma
	// linha fora disso (um erro de digitação, outro pacote ou nenhum pacote)
	// não é de nenhum dos cinco testes, então precisa reprovar o teste de
	// qualquer um deles, em vez de sumir em silêncio.
	tests := []struct {
		name, file, param string
	}{
		{"prefixo com erro de digitação", "testdata/unknown_prefix_typo.md", "gophr.hungryAt"},
		{"prefixo de outro pacote", "testdata/unknown_prefix_package.md", "hud.panelWidth"},
		{"sem prefixo", "testdata/no_prefix.md", "hungryAt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, pkg := range []string{"main", "game", "gopher", "world", "logs"} {
				// O arquivo também tem gopher.hungryAt, que o teste do
				// pacote gopher conhece: a única linha errada é tt.param.
				code := Params{}
				if pkg == "gopher" {
					code["gopher.hungryAt"] = 3
				}
				rec := &recordingTB{TB: t}
				Check(rec, pkg, map[string]Params{tt.file: code})
				// Conferir que a falha cita tt.param garante que o teste
				// reprova por causa dessa linha, e não por outro defeito do
				// arquivo de exemplo.
				if !rec.cites(tt.param) {
					t.Errorf("Check(%q) com a linha %s: falhas %q, esperada uma falha citando %s", pkg, tt.param, rec.failures, tt.param)
				}
			}
		})
	}
}

func TestCheckFailsOnARepeatedParameter(t *testing.T) {
	// Critério de aceite 1 da issue #2: o teste falha quando um valor da
	// tabela difere da constante. Com gopher.hungryAt duas vezes, uma com o
	// valor do código (3) e outra com 99, a tabela mostra um valor diferente
	// do código, seja qual for a linha que aparece primeiro.
	tests := []struct {
		name, file string
	}{
		{"valor errado na primeira linha", "testdata/duplicate_wrong_first.md"},
		{"valor errado na última linha", "testdata/duplicate_wrong_last.md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &recordingTB{TB: t}
			Check(rec, "gopher", map[string]Params{
				tt.file: {"gopher.hungryAt": 3},
			})
			if !rec.cites("gopher.hungryAt") {
				t.Errorf("Check com gopher.hungryAt repetido: falhas %q, esperada uma falha citando gopher.hungryAt", rec.failures)
			}
		})
	}
}

func TestCheckFailsOnAnEmptyParameterCell(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": o Parâmetro é a expressão
	// Go entre crases, com o pacote na frente. Uma linha com a célula de
	// parâmetro vazia não é de nenhum pacote, então precisa reprovar o teste
	// de qualquer um dos cinco, em vez de sumir em silêncio.
	for _, pkg := range []string{"main", "game", "gopher", "world", "logs"} {
		// O arquivo também tem gopher.hungryAt, que o teste do pacote
		// gopher conhece: a única linha errada é a de parâmetro vazio.
		code := Params{}
		if pkg == "gopher" {
			code["gopher.hungryAt"] = 3
		}
		rec := &recordingTB{TB: t}
		Check(rec, pkg, map[string]Params{"testdata/empty_parameter.md": code})
		if !rec.Failed() {
			t.Errorf("Check(%q) aceitou uma linha com a célula de parâmetro vazia, esperado reprovar o teste", pkg)
		}
	}
}

func TestCheckReportsFailuresInAlphabeticalOrderOfFiles(t *testing.T) {
	// Issue #26, P2: as falhas saem na mesma ordem em toda execução, com os
	// arquivos em ordem alfabética. Os dois arquivos dizem gopher.hungryAt 3,
	// e o código diz 4, então cada um gera uma falha. Uma única chamada
	// poderia acertar a ordem por sorte, por isso o teste repete a conferência.
	files := map[string]Params{
		"testdata/b_diverges.md": {"gopher.hungryAt": 4},
		"testdata/a_diverges.md": {"gopher.hungryAt": 4},
	}
	want := []string{"a_diverges.md: ", "b_diverges.md: "}
	for range 50 {
		rec := &recordingTB{TB: t}
		Check(rec, "gopher", files)
		if len(rec.failures) != len(want) {
			t.Fatalf("Check registrou as falhas %q, esperada uma por arquivo", rec.failures)
		}
		for i, prefix := range want {
			if !strings.HasPrefix(rec.failures[i], prefix) {
				t.Fatalf("Check registrou as falhas %q, esperada a de a_diverges.md antes da de b_diverges.md", rec.failures)
			}
		}
	}
}

func TestCheckCitesTheSpecFileByNameInEveryFailure(t *testing.T) {
	// Issue #26, P6 (resposta do mantenedor: só o nome do arquivo, como Diff
	// já fazia). O erro de formato (Read) e a divergência (Diff) citam o
	// arquivo do mesmo jeito, seja qual for o caminho que o teste do pacote
	// passou: o pacote main passa docs/spec/..., e os outros, ../docs/spec/...
	// Em not_a_number.md, o valor fora do formato está na linha 8; em
	// a_diverges.md, a spec diz gopher.hungryAt 3, e o código diz 4.
	want := []string{"a_diverges.md: gopher.hungryAt: ", "not_a_number.md:8: "}
	for _, dir := range []string{"testdata/", "../spectable/testdata/"} {
		t.Run(dir, func(t *testing.T) {
			rec := &recordingTB{TB: t}
			Check(rec, "gopher", map[string]Params{
				dir + "not_a_number.md": {"gopher.hungryAt": 3},
				dir + "a_diverges.md":   {"gopher.hungryAt": 4},
			})
			if len(rec.failures) != len(want) {
				t.Fatalf("Check registrou as falhas %q, esperada uma por arquivo", rec.failures)
			}
			for _, prefix := range want {
				if !slices.ContainsFunc(rec.failures, func(f string) bool { return strings.HasPrefix(f, prefix) }) {
					t.Errorf("Check registrou as falhas %q, esperada uma começando com %q", rec.failures, prefix)
				}
			}
		})
	}
}
