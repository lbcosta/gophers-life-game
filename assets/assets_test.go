package assets

import (
	"bytes"
	"errors"
	"io/fs"
	"testing"
)

// pngSignature são os oito bytes que abrem todo arquivo PNG.
var pngSignature = []byte("\x89PNG\r\n\x1a\n")

// TestFSOpensPathsWithTheAssetsPrefix garante que os caminhos a partir da raiz
// do repositório, como world.LoadSprites e hud.LoadPhotos os pedem, continuam
// abrindo os PNGs das Plantas e das Fotos de perfil (issue #14).
func TestFSOpensPathsWithTheAssetsPrefix(t *testing.T) {
	for _, name := range []string{
		"assets/plants/carrot.png",
		"assets/profiles/cacau_feliz.png",
	} {
		data, err := fs.ReadFile(FS(), name)
		if err != nil {
			t.Errorf("fs.ReadFile(FS(), %q): erro %v, esperado ler o PNG", name, err)
			continue
		}
		if !bytes.HasPrefix(data, pngSignature) {
			t.Errorf("fs.ReadFile(FS(), %q): esperado um PNG, mas o conteúdo não começa com a assinatura PNG", name)
		}
	}
}

// TestFSRejectsPathsWithoutTheAssetsPrefix garante que um caminho sem o prefixo
// "assets/" não abre nada, para um caminho errado num teste não passar sem
// ninguém perceber (issue #14, achado P7).
func TestFSRejectsPathsWithoutTheAssetsPrefix(t *testing.T) {
	for _, name := range []string{
		"plants/carrot.png",
		"profiles/cacau_feliz.png",
		"grass.png",
	} {
		_, err := fs.ReadFile(FS(), name)
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("fs.ReadFile(FS(), %q): erro %v, esperado um erro que satisfaça fs.ErrNotExist", name, err)
		}
	}
}
