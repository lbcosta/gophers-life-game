// Package spectable confere as tabelas de parâmetros da spec
// (docs/spec/README.md, "Tabelas de parâmetros") contra as constantes do
// código. Fica em internal/ porque só os testes dos pacotes usam.
package spectable

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Params liga o parâmetro (a expressão Go, como "gopher.hungryAt") ao valor.
type Params map[string]float64

const sectionHeading = "## Parâmetros"

// knownPrefixes são os pacotes que têm teste de spec (docs/spec/README.md). Um
// parâmetro com outro prefixo, ou sem prefixo, não seria conferido por teste
// nenhum, então Read o rejeita em vez de deixá-lo passar em silêncio.
var knownPrefixes = []string{"main.", "game.", "gopher.", "world.", "logs."}

// Read lê a tabela da seção "## Parâmetros" de um arquivo da spec. O formato é
// rígido de propósito: qualquer linha fora dele é um erro, para a spec e o
// teste não divergirem em silêncio.
func Read(path string) (Params, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	params := Params{}
	inSection, found := false, false
	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "## ") {
			inSection = line == sectionHeading
			found = found || inSection
			continue
		}
		if !inSection || line == "" {
			continue
		}
		// No Markdown do GitHub, a barra do começo é opcional: uma linha logo
		// abaixo da tabela, sem a barra, ainda aparece como linha dela. Por
		// isso toda linha não vazia da seção precisa começar com |, ou o
		// parâmetro sumiria da conferência em silêncio.
		if !strings.HasPrefix(line, "|") {
			return nil, fmt.Errorf("%s:%d: linha da seção %q sem | no começo: %q", path, n, sectionHeading, line)
		}
		if isSeparator(line) {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) < 2 {
			return nil, fmt.Errorf("%s:%d: linha da tabela sem valor", path, n)
		}
		name, value := strings.TrimSpace(cells[0]), strings.TrimSpace(cells[1])
		if name == "Parâmetro" {
			continue // cabeçalho
		}
		if len(name) < 3 || !strings.HasPrefix(name, "`") || !strings.HasSuffix(name, "`") {
			return nil, fmt.Errorf("%s:%d: o parâmetro %q precisa estar entre crases", path, n, name)
		}
		name = name[1 : len(name)-1]
		if !slices.ContainsFunc(knownPrefixes, func(p string) bool { return strings.HasPrefix(name, p) }) {
			return nil, fmt.Errorf("%s:%d: o parâmetro %s não começa com um pacote conhecido (%s)", path, n, name, strings.Join(knownPrefixes, ", "))
		}
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %s: o valor %q não é um número (use ponto decimal)", path, n, name, value)
		}
		if _, dup := params[name]; dup {
			return nil, fmt.Errorf("%s:%d: o parâmetro %s aparece mais de uma vez na tabela", path, n, name)
		}
		params[name] = v
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("%s: sem a seção %q", path, sectionHeading)
	}
	return params, nil
}

// isSeparator diz se a linha é o separador do cabeçalho (|---|:---:|). Olha a
// linha inteira, e não só a primeira célula: uma linha com o parâmetro vazio
// (| | 5 | x |) tem outros caracteres e precisa cair no erro de formato, em vez de
// passar por separador.
func isSeparator(line string) bool {
	return strings.Contains(line, "-") && strings.Trim(line, "|-: ") == ""
}

func format(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

// Diff compara a tabela de um arquivo com as constantes que o teste do pacote
// pkg conhece. Só olha os parâmetros do próprio pacote, porque um arquivo como
// simulacao.md mistura parâmetros de vários pacotes.
func Diff(file, pkg string, spec, code Params) []string {
	prefix := pkg + "."
	var msgs []string
	for _, name := range sortedKeys(spec) {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		want, known := code[name]
		if !known {
			msgs = append(msgs, fmt.Sprintf("%s: %s: a tabela cita um parâmetro que o teste não conhece", file, name))
		} else if want != spec[name] {
			msgs = append(msgs, fmt.Sprintf("%s: %s: a spec diz %s, o código diz %s", file, name, format(spec[name]), format(want)))
		}
	}
	for _, name := range sortedKeys(code) {
		if _, ok := spec[name]; !ok {
			msgs = append(msgs, fmt.Sprintf("%s: %s: o parâmetro falta na tabela", file, name))
		}
	}
	return msgs
}

func sortedKeys(p Params) []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// hasSection diz se o arquivo tem a seção "## Parâmetros" como título. Uma
// menção no meio do texto não conta, igual em Read.
func hasSection(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == sectionHeading {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// Check lê cada arquivo da spec e reporta as divergências do pacote pkg como
// falhas do teste. As chaves de files são caminhos; a mensagem usa só o nome
// do arquivo. Também reprova para todo .md das mesmas pastas que tem a seção
// "## Parâmetros" e não está em files: uma tabela que nenhum teste lê nunca
// seria conferida.
func Check(t testing.TB, pkg string, files map[string]Params) {
	t.Helper()
	// Os caminhos são comparados pela forma absoluta: Clean iguala
	// "./docs/spec/x.md" a "docs/spec/x.md", mas não iguala um caminho
	// absoluto a um relativo, e o mapa pode misturar as duas formas. As
	// pastas guardadas já são absolutas, então o Glob devolve os arquivos
	// na mesma forma das chaves de read.
	dirs := map[string]bool{}
	read := map[string]bool{}
	for path, code := range files {
		abs, err := filepath.Abs(path)
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		dirs[filepath.Dir(abs)] = true
		read[abs] = true
		spec, err := Read(path)
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		for _, msg := range Diff(filepath.Base(path), pkg, spec, code) {
			t.Error(msg)
		}
	}
	for _, dir := range sortedDirs(dirs) {
		entries, err := filepath.Glob(filepath.Join(dir, "*.md"))
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		for _, path := range entries {
			if read[path] {
				continue
			}
			has, err := hasSection(path)
			if err != nil {
				t.Errorf("%v", err)
			} else if has {
				t.Errorf("%s: tem a seção %q, mas o teste do pacote %s não a lê (acrescente o arquivo ao mapa de Check)", filepath.Base(path), sectionHeading, pkg)
			}
		}
	}
}

func sortedDirs(m map[string]bool) []string {
	dirs := make([]string, 0, len(m))
	for d := range m {
		dirs = append(dirs, d)
	}
	slices.Sort(dirs)
	return dirs
}
