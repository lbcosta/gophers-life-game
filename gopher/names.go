package gopher

import (
	_ "embed"
	"strings"
)

// namesFile é a lista de nomes que os Gophers podem receber (ver names.txt, que
// também diz de onde ela veio), embutida no binário.
//
//go:embed names.txt
var namesFile string

// Names é a lista de nomes, na ordem do arquivo.
var Names = parseNames(namesFile)

// parseNames devolve os nomes de text: um por linha, ignorando linhas vazias e
// comentários (as que começam com #).
func parseNames(text string) []string {
	var names []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			names = append(names, line)
		}
	}
	return names
}
