package hud

import (
	"image/color"
	"os"
	"testing"

	"gophers-life-game/gopher"
)

func TestEveryPhotoExistsAndHasTheColorOfItsName(t *testing.T) {
	photos, err := LoadPhotos(os.DirFS(".."))
	if err != nil {
		t.Fatal(err)
	}
	if want := len(gopher.Palette) * len(gopher.Emotions); len(photos.decoded) != want {
		t.Fatalf("%d fotos, esperado %d", len(photos.decoded), want)
	}

	for _, body := range gopher.Palette {
		for _, e := range gopher.Emotions {
			name := gopher.PhotoName(body, e)
			// O meio da barriga é só corpo, na cor do Gopher.
			got := color.NRGBAModel.Convert(photos.decoded[name].At(128, 170)).(color.NRGBA)
			if want := (color.NRGBA{body.R, body.G, body.B, 255}); got != want {
				t.Errorf("%s: cor do corpo %v, esperado %v", name, got, want)
			}
		}
	}
}

func TestPanelContains(t *testing.T) {
	const w = 1000
	if !PanelContains(w-panelMargin-1, panelMargin, w) {
		t.Error("canto do painel fora do painel")
	}
	if PanelContains(w-panelMargin-panelWidth-1, 100, w) || PanelContains(500, 100, w) {
		t.Error("ponto à esquerda do painel dentro dele")
	}
	if PanelContains(w-100, panelMargin+panelHeight, w) {
		t.Error("ponto abaixo do painel dentro dele")
	}
}
