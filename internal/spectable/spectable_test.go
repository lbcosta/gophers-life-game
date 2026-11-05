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
// arquivo e reporta as divergências do pacote como falhas do teste, e também
// reprova cada arquivo da mesma pasta com a seção "## Parâmetros" que ficou
// fora do mapa.

import (
	"fmt"
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
	//
	// O erro precisa citar o arquivo e a linha sem barra (achado S10 da
	// issue #25). Só olhar se houve erro deixaria passar um erro vindo de
	// outra linha do arquivo de exemplo. O trecho conferido é "arquivo:linha",
	// sem a pasta, para não depender de como o caminho aparece (#26).
	tests := []struct {
		name, file, line string
	}{
		{"linha da tabela sem a barra do começo", "testdata/no_leading_pipe.md", "no_leading_pipe.md:8"},
		{"linha de texto dentro da seção", "testdata/text_in_section.md", "text_in_section.md:9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Read(tt.file)
			if err == nil {
				t.Fatalf("Read(%q) aceitou uma linha sem | no começo dentro da seção, esperado um erro", tt.file)
			}
			if !strings.Contains(err.Error(), tt.line) {
				t.Errorf("Read(%q) devolveu o erro %q, esperado um erro citando %s", tt.file, err, tt.line)
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

	// A mensagem inteira, como em TestDiffReportsAValueThatDiffers (achado S5
	// da issue #25).
	want := "gophers.md: gopher.sleepChance: a tabela cita um parâmetro que o teste não conhece"
	got := Diff("gophers.md", "gopher", spec, code)
	if len(got) != 1 || got[0] != want {
		t.Errorf("Diff = %q, esperado [%q]", got, want)
	}
}

func TestDiffReportsAKnownParameterMissingFromTheTable(t *testing.T) {
	spec := Params{"gopher.hungryAt": 3}
	code := Params{"gopher.hungryAt": 3, "gopher.initialSatiety": 10}

	// A mensagem inteira, como em TestDiffReportsAValueThatDiffers (achado S5
	// da issue #25).
	want := "gophers.md: gopher.initialSatiety: o parâmetro falta na tabela"
	got := Diff("gophers.md", "gopher", spec, code)
	if len(got) != 1 || got[0] != want {
		t.Errorf("Diff = %q, esperado [%q]", got, want)
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
	// O arquivo fica sozinho numa pasta porque Check também reprova as
	// tabelas da mesma pasta que ficaram fora do mapa (docs/spec/README.md,
	// "Tabelas de parâmetros"), e testdata/ tem várias.
	rec := &recordingTB{TB: t}
	Check(rec, "gopher", map[string]Params{
		"testdata/covered/valid.md": {
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
		// A falha precisa citar a linha 8, a da célula vazia (achado S8 da
		// issue #25): só olhar se o teste reprovou deixaria passar uma falha
		// vinda de outro defeito do arquivo de exemplo. O trecho conferido é
		// "arquivo:linha", sem a pasta, para não depender de como o caminho
		// aparece (#26).
		if !rec.cites("empty_parameter.md:8") {
			t.Errorf("Check(%q) com a célula de parâmetro vazia: falhas %q, esperada uma falha citando empty_parameter.md:8", pkg, rec.failures)
		}
	}
}

func TestCheckFailsOnATableNoTestReads(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": todo arquivo de
	// docs/spec/ com a seção ## Parâmetros precisa estar no teste de cada
	// pacote, e uma tabela que nenhum teste lê faz o teste falhar (achado S6
	// da issue #25). Em testdata/uncovered/, cited.md e forgotten.md têm a
	// seção, e o mapa só traz cited.md.
	for _, pkg := range []string{"main", "game", "gopher", "world", "logs"} {
		// cited.md só tem gopher.hungryAt, que o teste do pacote gopher
		// conhece: a tabela lida está certa, e o único defeito é a outra.
		code := Params{}
		if pkg == "gopher" {
			code["gopher.hungryAt"] = 3
		}
		rec := &recordingTB{TB: t}
		Check(rec, pkg, map[string]Params{"testdata/uncovered/cited.md": code})
		if !rec.cites("forgotten.md") {
			t.Errorf("Check(%q) sem forgotten.md no mapa: falhas %q, esperada uma falha citando forgotten.md", pkg, rec.failures)
		}
	}
}

func TestCheckAcceptsAMapKeyThatIsNotACleanPath(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": a regra reprova a tabela
	// que nenhum teste lê, não a que está no mapa. Um caminho escrito de outro
	// jeito, como "./testdata/covered/valid.md", aponta para o mesmo arquivo
	// e conta como lido (achado S-1 da revisão do PR #35). testdata/covered/
	// só tem valid.md, então nada mais pode reprovar o teste.
	tests := []struct {
		name, file string
	}{
		{"com ./ na frente", "./testdata/covered/valid.md"},
		{"com barra dupla", "testdata//covered/valid.md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &recordingTB{TB: t}
			Check(rec, "gopher", map[string]Params{
				tt.file: {
					"gopher.stepsPerSatiety":             1000,
					"gopher.speedSteps[Slow]":            0.25,
					"gopher.maturityRanges[Child-1].min": 600,
				},
			})
			if rec.Failed() {
				t.Errorf("Check com a chave %q: falhas %q, esperado nenhuma falha", tt.file, rec.failures)
			}
		})
	}
}

func TestCheckIgnoresAFileWithoutTheParametersSection(t *testing.T) {
	// A regra vale para os arquivos com a seção ## Parâmetros. Um arquivo que
	// só cita a seção no meio do texto, como o docs/spec/README.md, não tem
	// tabela para ler e não reprova o teste. Em testdata/prose_mention/,
	// valid.md está no mapa e notes.md é esse arquivo.
	rec := &recordingTB{TB: t}
	Check(rec, "gopher", map[string]Params{
		"testdata/prose_mention/valid.md": {
			"gopher.stepsPerSatiety":             1000,
			"gopher.speedSteps[Slow]":            0.25,
			"gopher.maturityRanges[Child-1].min": 600,
		},
	})
	if rec.Failed() {
		t.Errorf("Check reprovou um arquivo sem a seção ## Parâmetros: %q, esperado nenhuma falha", rec.failures)
	}
}
