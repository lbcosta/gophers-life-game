// Package hud desenha a interface sobreposta ao jogo.
package hud

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gophers-life-game/gopher"
	"gophers-life-game/internal/circle"
	"gophers-life-game/internal/fonts"
)

// Medidas do painel de perfil, em pixels.
const (
	panelWidth   = 300
	panelHeight  = 410
	panelMargin  = 16 // distância do painel à borda direita e ao topo da tela
	panelPadding = 16

	photoSize = 120 // lado do espaço da foto de perfil
	rowHeight = 30
)

var (
	panelBack   = color.NRGBA{16, 20, 14, 225}
	panelBorder = color.NRGBA{255, 255, 255, 60}
	photoBack   = color.NRGBA{255, 255, 255, 22}
	textColor   = color.NRGBA{240, 240, 235, 255}
	dimColor    = color.NRGBA{240, 240, 235, 140}
	deadColor   = color.NRGBA{255, 110, 110, 255}
)

// PanelContains informa se o ponto (x, y) cai sobre o painel de perfil, numa
// tela de largura screenWidth.
func PanelContains(x, y, screenWidth int) bool {
	left := screenWidth - panelMargin - panelWidth
	return x >= left && x < left+panelWidth && y >= panelMargin && y < panelMargin+panelHeight
}

func drawText(screen *ebiten.Image, s string, size, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, s, fonts.Regular(size), op)
}

// DrawProfile desenha, do lado direito da tela, o painel com as características
// do Gopher descrito por p e a foto dele, tirada de photos (que pode ser nil).
func DrawProfile(screen *ebiten.Image, p gopher.Profile, photos *Photos) {
	screenW := screen.Bounds().Dx()
	left := float32(screenW - panelMargin - panelWidth)
	top := float32(panelMargin)

	// Os retângulos ficam em pixels inteiros e não precisam de suavização. Sem
	// ela, vector os desenha como imagens comuns, que a Ebitengine junta numa só
	// chamada à placa de vídeo; com ela, cada grupo de desenhos vetoriais
	// interrompido por um DrawImage custa várias passadas por uma imagem auxiliar.
	vector.FillRect(screen, left, top, panelWidth, panelHeight, panelBack, false)
	vector.StrokeRect(screen, left, top, panelWidth, panelHeight, 1, panelBorder, false)

	x := float64(left) + panelPadding
	y := float64(top) + panelPadding

	// Foto de perfil, sobre um fundo claro.
	vector.FillRect(screen, float32(x), float32(y), photoSize, photoSize, photoBack, false)
	if photo := photos.image(p.Photo); photo != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		screen.DrawImage(photo, op)
	}

	// Nome, ao lado da foto.
	textX := x + photoSize + 14
	drawText(screen, p.Name, 20, textX, y+4, textColor)
	if !p.Alive {
		drawText(screen, "Morreu", 15, textX, y+34, deadColor)
	}

	// Características, uma por linha, com o nome à esquerda e o valor à direita.
	y += photoSize + 18
	rows := []struct{ label, value string }{
		{"Cor", p.ColorName},
		{"Emoção", p.Emotion.String()},
		{"Velocidade", fmt.Sprintf("%v (%g px/passo)", p.Speed, p.StepSize)},
		{"Saciedade", fmt.Sprint(p.Satiety)},
		{"Fase da vida", p.Stage.String()},
		{"Idade", fmt.Sprintf("%.0f s", p.Age)},
		{"Próxima fase", nextStage(p)},
		{"Tamanho", fmt.Sprintf("raio %.0f px", p.Radius)},
	}
	for _, r := range rows {
		drawText(screen, r.label, 15, x, y, dimColor)
		valueW, _ := text.Measure(r.value, fonts.Regular(15), 0)
		valueX := float64(left) + panelWidth - panelPadding - valueW
		if r.label == "Cor" { // uma amostra da cor, à esquerda do nome dela
			circle.Fill(screen, float32(valueX)-14, float32(y)+9, 6, p.Color)
		}
		drawText(screen, r.value, 15, valueX, y, textColor)
		y += rowHeight
	}
}

// nextStage descreve quanto falta para a próxima fase da vida.
func nextStage(p gopher.Profile) string {
	if !p.HasNextStage || !p.Alive {
		return "—"
	}
	return fmt.Sprintf("%v em %.0f s", p.NextStage, p.NextStageIn)
}
