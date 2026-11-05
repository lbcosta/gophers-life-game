// Package spectable confere as tabelas de parâmetros da spec
// (docs/spec/README.md, "Tabelas de parâmetros") contra as constantes do
// código. Fica em internal/ porque só os testes dos pacotes usam.
package spectable

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Params liga o parâmetro (a expressão Go, como "gopher.hungryAt") ao valor.
type Params map[string]float64

const sectionHeading = "## Parâmetros"

// numberFormat é o formato do Valor na spec: só o número, com ponto decimal.
// O ParseFloat aceitaria também 9e1, 90., .5, +90, Inf e NaN, que a spec não
// promete aceitar.
var numberFormat = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

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

	// As mensagens citam só o nome do arquivo, como Diff, para não mudarem
	// com a pasta de onde o teste do pacote roda.
	path = filepath.Base(path)
	params := Params{}
	inSection, found := false, false
	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "## ") {
			inSection = isSectionTitle(line)
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
		if err != nil || !numberFormat.MatchString(value) {
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
	for _, name := range slices.Sorted(maps.Keys(spec)) {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		inCode, known := code[name]
		if !known {
			msgs = append(msgs, fmt.Sprintf("%s: %s: a tabela cita um parâmetro que o teste não conhece", file, name))
		} else if inCode != spec[name] {
			msgs = append(msgs, fmt.Sprintf("%s: %s: a spec diz %s, o código diz %s", file, name, format(spec[name]), format(inCode)))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(code)) {
		if _, ok := spec[name]; !ok {
			msgs = append(msgs, fmt.Sprintf("%s: %s: o parâmetro falta na tabela", file, name))
		}
	}
	return msgs
}

// isSectionTitle diz se a linha é o título "## Parâmetros". É o único lugar com
// esse critério, usado por Read e por hasSection, para os dois concordarem.
// Uma menção no meio do texto não conta.
func isSectionTitle(line string) bool {
	return strings.TrimSpace(line) == sectionHeading
}

// hasSection diz se o arquivo tem a seção "## Parâmetros" como título.
func hasSection(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if isSectionTitle(scanner.Text()) {
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
	// Os arquivos em ordem alfabética, para as falhas saírem sempre na mesma
	// ordem (o range de um map é aleatório).
	//
	// Os caminhos são comparados pela forma absoluta: Clean iguala
	// "./docs/spec/x.md" a "docs/spec/x.md", mas não iguala um caminho
	// absoluto a um relativo, e o mapa pode misturar as duas formas. As
	// pastas guardadas já são absolutas, então os arquivos listados saem
	// na mesma forma das chaves de read.
	dirs := map[string]bool{}
	read := map[string]bool{}
	for _, path := range slices.Sorted(maps.Keys(files)) {
		code := files[path]
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
	for _, dir := range slices.Sorted(maps.Keys(dirs)) {
		// ReadDir em vez de Glob: o Glob trata o caminho inteiro como padrão,
		// e um "[", "*", "?" ou "\" no caminho do clone o faria devolver uma
		// lista vazia, sem erro, e a conferência sumiria em silêncio.
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
				continue
			}
			path := filepath.Join(dir, entry.Name())
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
