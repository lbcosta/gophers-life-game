package spectable

// Estes testes fixam o comportamento do pacote de apoio que confere as tabelas
// de parâmetros da spec (docs/spec/README.md, "Tabelas de parâmetros"). As
// assinaturas ficam no próprio pacote: aqui só vale o que cada função faz.
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
	"os"
	"path/filepath"
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

// mount monta, numa pasta temporária chamada name, só os arquivos de files,
// que liga o nome de cada arquivo na pasta ao exemplo de testdata/ copiado
// para ele. Check também reprova as tabelas da mesma pasta que ficaram fora
// do mapa (docs/spec/README.md, "Tabelas de parâmetros"), e testdata/ tem
// várias: por isso os testes de Check que precisam de uma pasta só com certos
// arquivos a montam na hora, em vez de guardar cópias dos exemplos (P-3 da
// issue #41). Devolve o caminho da pasta relativo à pasta do pacote, onde o
// teste roda, e o absoluto.
func mount(t *testing.T, name string, files map[string]string) (rel, abs string) {
	t.Helper()
	abs = filepath.Join(t.TempDir(), name)
	if err := os.Mkdir(abs, 0o755); err != nil {
		t.Fatalf("os.Mkdir(%q) devolveu erro %v", abs, err)
	}
	for file, example := range files {
		data, err := os.ReadFile(filepath.Join("testdata", example))
		if err != nil {
			t.Fatalf("os.ReadFile(%q) devolveu erro %v", example, err)
		}
		if err := os.WriteFile(filepath.Join(abs, file), data, 0o644); err != nil {
			t.Fatalf("os.WriteFile(%q) devolveu erro %v", file, err)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd devolveu erro %v", err)
	}
	rel, err = filepath.Rel(wd, abs)
	if err != nil {
		t.Fatalf("filepath.Rel(%q, %q) devolveu erro %v", wd, abs, err)
	}
	return rel, abs
}

func TestCheckPassesATableThatMatchesTheCode(t *testing.T) {
	// Controle do recordingTB: uma tabela igual ao código não reprova nada.
	// O arquivo fica sozinho numa pasta porque Check também reprova as
	// tabelas da mesma pasta que ficaram fora do mapa (docs/spec/README.md,
	// "Tabelas de parâmetros"), e testdata/ tem várias.
	dir, _ := mount(t, "covered", map[string]string{"valid.md": "valid.md"})
	rec := &recordingTB{TB: t}
	Check(rec, "gopher", map[string]Params{
		filepath.Join(dir, "valid.md"): {
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
	// da issue #25). Na pasta montada, cited.md e forgotten.md têm a seção, e
	// o mapa só traz cited.md.
	dir, _ := mount(t, "uncovered", map[string]string{
		"cited.md":     "hungry_at.md",
		"forgotten.md": "turn_chance.md",
	})
	for _, pkg := range []string{"main", "game", "gopher", "world", "logs"} {
		// cited.md só tem gopher.hungryAt, que o teste do pacote gopher
		// conhece: a tabela lida está certa, e o único defeito é a outra.
		code := Params{}
		if pkg == "gopher" {
			code["gopher.hungryAt"] = 3
		}
		rec := &recordingTB{TB: t}
		Check(rec, pkg, map[string]Params{filepath.Join(dir, "cited.md"): code})
		if !rec.cites("forgotten.md") {
			t.Errorf("Check(%q) sem forgotten.md no mapa: falhas %q, esperada uma falha citando forgotten.md", pkg, rec.failures)
		}
	}
}

func TestCheckFailsOnATableNoTestReadsInAFolderWithBracketsInItsName(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": uma tabela que nenhum
	// teste lê faz o teste falhar, seja qual for o nome da pasta. Colchetes
	// podem aparecer no caminho do clone, e a conferência não pode sumir em
	// silêncio por causa deles (achado S-5 da revisão do PR #35). Na pasta
	// montada, uncovered[brackets]/, cited.md e forgotten.md têm a seção, e o
	// mapa só traz cited.md, pelo caminho relativo ou pelo absoluto.
	rel, abs := mount(t, "uncovered[brackets]", map[string]string{
		"cited.md":     "hungry_at.md",
		"forgotten.md": "turn_chance.md",
	})
	tests := []struct {
		name, key string
	}{
		{"caminho relativo", filepath.Join(rel, "cited.md")},
		{"caminho absoluto", filepath.Join(abs, "cited.md")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// cited.md só tem gopher.hungryAt, que o teste do pacote gopher
			// conhece: a tabela lida está certa, e o único defeito é a outra.
			rec := &recordingTB{TB: t}
			Check(rec, "gopher", map[string]Params{tt.key: {"gopher.hungryAt": 3}})
			if !rec.cites("forgotten.md") {
				t.Errorf("Check com a chave %q, sem forgotten.md no mapa: falhas %q, esperada uma falha citando forgotten.md", tt.key, rec.failures)
			}
			// cited.md está no mapa e a tabela dele bate com o código: os
			// colchetes não podem fazer Check reprová-lo como uma tabela que
			// nenhum teste lê (T-1 da issue #41).
			if rec.cites("cited.md") {
				t.Errorf("Check com a chave %q: falhas %q, esperado nenhuma falha citando cited.md", tt.key, rec.failures)
			}
		})
	}
}

func TestCheckAcceptsAMapKeyThatIsNotACleanPath(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": a regra reprova a tabela
	// que nenhum teste lê, não a que está no mapa. Um caminho escrito de outro
	// jeito, como "./pasta/valid.md", aponta para o mesmo arquivo e conta
	// como lido (achado S-1 da revisão do PR #35). A pasta montada só tem
	// valid.md, então nada mais pode reprovar o teste.
	dir, _ := mount(t, "covered", map[string]string{"valid.md": "valid.md"})
	tests := []struct {
		name, file string
	}{
		{"com ./ na frente", "./" + dir + "/valid.md"},
		{"com barra dupla", dir + "//valid.md"},
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

func TestCheckAcceptsAbsoluteAndRelativeMapKeysTogether(t *testing.T) {
	// docs/spec/README.md, "Tabelas de parâmetros": a regra reprova a tabela
	// que nenhum teste lê, não a que está no mapa. Um caminho absoluto aponta
	// para o mesmo arquivo que o relativo e conta como lido, mesmo quando o
	// mapa mistura as duas formas (achado S-4 da revisão do PR #35). Na pasta
	// montada, cited.md e forgotten.md têm a seção e os dois estão no mapa,
	// então nada pode reprovar o teste.
	rel, abs := mount(t, "uncovered", map[string]string{
		"cited.md":     "hungry_at.md",
		"forgotten.md": "turn_chance.md",
	})
	citedCode := Params{"gopher.hungryAt": 3}
	forgottenCode := Params{"gopher.turnChance": 0.15}

	tests := []struct {
		name  string
		files map[string]Params
	}{
		{"cited.md relativo e forgotten.md absoluto", map[string]Params{
			filepath.Join(rel, "cited.md"):     citedCode,
			filepath.Join(abs, "forgotten.md"): forgottenCode,
		}},
		{"cited.md absoluto e forgotten.md relativo", map[string]Params{
			filepath.Join(abs, "cited.md"):     citedCode,
			filepath.Join(rel, "forgotten.md"): forgottenCode,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &recordingTB{TB: t}
			Check(rec, "gopher", tt.files)
			if rec.Failed() {
				t.Errorf("Check com as chaves %q: falhas %q, esperado nenhuma falha", slices.Sorted(maps.Keys(tt.files)), rec.failures)
			}
		})
	}
}

func TestCheckIgnoresAFileWithoutTheParametersSection(t *testing.T) {
	// A regra vale para os arquivos com a seção ## Parâmetros. Um arquivo que
	// só cita a seção no meio do texto, como o docs/spec/README.md, não tem
	// tabela para ler e não reprova o teste. Na pasta montada, valid.md está
	// no mapa e notes.md é esse arquivo.
	dir, _ := mount(t, "prose_mention", map[string]string{
		"valid.md": "valid.md",
		"notes.md": "prose_mention.md",
	})
	rec := &recordingTB{TB: t}
	Check(rec, "gopher", map[string]Params{
		filepath.Join(dir, "valid.md"): {
			"gopher.stepsPerSatiety":             1000,
			"gopher.speedSteps[Slow]":            0.25,
			"gopher.maturityRanges[Child-1].min": 600,
		},
	})
	if rec.Failed() {
		t.Errorf("Check reprovou um arquivo sem a seção ## Parâmetros: %q, esperado nenhuma falha", rec.failures)
	}
}

func TestCheckReportsFailuresInAlphabeticalOrderOfFiles(t *testing.T) {
	// Issue #26, P2: as falhas saem na mesma ordem em toda execução, com os
	// arquivos em ordem alfabética. Os dois arquivos dizem gopher.hungryAt 3,
	// e o código diz 4, então cada um gera uma falha. Uma única chamada
	// poderia acertar a ordem por sorte, por isso o teste repete a conferência.
	// Os arquivos ficam sozinhos numa pasta montada porque Check também
	// reprova as tabelas da mesma pasta que ficaram fora do mapa.
	dir, _ := mount(t, "ordered", map[string]string{
		"a_diverges.md": "hungry_at.md",
		"b_diverges.md": "hungry_at.md",
	})
	files := map[string]Params{
		filepath.Join(dir, "b_diverges.md"): {"gopher.hungryAt": 4},
		filepath.Join(dir, "a_diverges.md"): {"gopher.hungryAt": 4},
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
	// a_diverges.md, a spec diz gopher.hungryAt 3, e o código diz 4. Os dois
	// ficam sozinhos numa pasta montada porque Check também reprova as tabelas
	// da mesma pasta que ficaram fora do mapa. A pasta é passada por dois
	// caminhos diferentes, como docs/spec/ e ../docs/spec/ nos pacotes: o
	// relativo e um que sai da pasta do pacote e volta para ela.
	rel, _ := mount(t, "named", map[string]string{
		"not_a_number.md": "not_a_number.md",
		"a_diverges.md":   "hungry_at.md",
	})
	want := []string{"a_diverges.md: gopher.hungryAt: ", "not_a_number.md:8: "}
	for _, dir := range []string{rel + "/", "../spectable/" + rel + "/"} {
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
