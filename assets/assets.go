// Package assets embute no binário os desenhos do jogo (a grama, as Plantas e
// as Fotos de perfil), para o jogo e o harness funcionarem em qualquer pasta.
//
// O go:embed só alcança arquivos da pasta do próprio pacote, e por isso o
// embed mora aqui, ao lado dos PNGs, e não no main.go.
package assets

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed grass.png plants/*.png profiles/*.png
var files embed.FS

// Grass devolve o PNG da textura de grama.
func Grass() ([]byte, error) {
	return files.ReadFile("grass.png")
}

// FS devolve os assets com os caminhos a partir da raiz do repositório
// ("assets/plants/...", "assets/profiles/..."), que é como world.LoadSprites e
// hud.LoadPhotos os procuram, para funcionarem igual com os arquivos do disco
// nos testes.
func FS() fs.FS {
	return rooted{files}
}

// rooted tira o prefixo "assets/" dos caminhos pedidos e recusa os que não o
// têm, para um caminho errado não passar sem ninguém perceber.
type rooted struct{ fs.FS }

func (r rooted) Open(name string) (fs.File, error) {
	rest, ok := strings.CutPrefix(name, "assets/")
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return r.FS.Open(rest)
}
