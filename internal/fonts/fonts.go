// Package fonts guarda as fontes do jogo. Cada uma é carregada uma única vez e
// compartilhada por todos os pacotes que escrevem na tela: a Ebitengine guarda,
// em cada fonte carregada, as letras que já desenhou, então carregar a mesma
// fonte duas vezes duplicaria esse trabalho e essa memória.
package fonts

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	regular *text.GoTextFaceSource
	faces   = map[float64]*text.GoTextFace{} // uma por tamanho
)

// Regular devolve a fonte Go Regular com size pixels de altura. Ela é criada na
// primeira vez em que é pedida; as chamadas seguintes devolvem a mesma.
func Regular(size float64) *text.GoTextFace {
	if regular == nil {
		src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
		if err != nil {
			panic(err) // a fonte está embutida no programa: não deve falhar
		}
		regular = src
	}
	if faces[size] == nil {
		faces[size] = &text.GoTextFace{Source: regular, Size: size}
	}
	return faces[size]
}
