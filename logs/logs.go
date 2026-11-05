// Package logs mostra, no canto inferior direito da tela, as últimas mensagens
// do jogo: avisos do que acontece na simulação, como um Gopher passando fome.
// Cada mensagem fica um tempo na tela e depois some aos poucos.
package logs

import (
	"image/color"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gophers-life-game/internal/fonts"
)

// Level é a gravidade de uma mensagem, que define a cor dela.
type Level int

const (
	// Warning é algo que vai mal, mas ainda tem volta. Aparece em amarelo.
	Warning Level = iota

	// Critical é algo que não tem volta. Aparece em vermelho.
	Critical
)

// ticksPerSecond é quantos ticks o jogo tem por segundo.
const ticksPerSecond = 60

const (
	// maxEntries é quantas mensagens o log guarda. Quando chega uma nova e ele
	// está cheio, a mais antiga sai.
	maxEntries = 8

	// lifetime é quantos ticks uma mensagem fica no log.
	lifetime = 10 * ticksPerSecond

	// fadeTicks é o fim da vida de uma mensagem, em ticks, em que ela vai ficando
	// transparente até sumir.
	fadeTicks = ticksPerSecond
)

// Medidas, em pixels.
const (
	fontSize = 15
	margin   = 16 // distância das mensagens à borda direita e à de baixo da tela
	padding  = 6  // espaço entre o texto e a borda do fundo dele
	lineGap  = 4  // espaço entre duas mensagens
)

var (
	backColor   = color.NRGBA{16, 20, 14, 210}
	levelColors = [...]color.NRGBA{
		Warning:  {255, 205, 90, 255},
		Critical: {255, 110, 110, 255},
	}
)

// Log guarda as mensagens mais recentes e as desenha. O valor zero é um log
// vazio, pronto para uso.
type Log struct {
	entries []entry // da mais antiga para a mais nova
	ticks   int     // ticks passados desde a criação do log
}

type entry struct {
	text  string
	level Level
	at    int // valor de Log.ticks quando a mensagem chegou
}

// Add acrescenta ao log a mensagem text, de gravidade level. Se o log já
// estiver cheio, a mensagem mais antiga sai.
func (l *Log) Add(level Level, text string) {
	if len(l.entries) == maxEntries {
		l.entries = slices.Delete(l.entries, 0, 1)
	}
	l.entries = append(l.entries, entry{text: text, level: level, at: l.ticks})
}

// Update deve ser chamado uma vez por tick do jogo. Tira do log as mensagens
// que já estão nele há lifetime ticks.
func (l *Log) Update() {
	l.ticks++
	l.entries = slices.DeleteFunc(l.entries, func(e entry) bool {
		return l.ticks-e.at >= lifetime
	})
}

// alpha devolve a opacidade de e: 1 durante quase toda a vida dela, caindo até
// perto de zero nos últimos fadeTicks.
func (l *Log) alpha(e entry) float32 {
	left := lifetime - (l.ticks - e.at)
	return min(1, float32(left)/fadeTicks)
}

// Draw desenha as mensagens no canto inferior direito de screen, alinhadas à
// direita, com a mais nova embaixo. Cada uma tem um fundo escuro do tamanho do
// texto, para ser lida sobre a grama.
func (l *Log) Draw(screen *ebiten.Image) {
	face := fonts.Regular(fontSize)
	b := screen.Bounds()
	right := float32(b.Max.X - margin)
	bottom := float32(b.Max.Y - margin)

	for i := len(l.entries) - 1; i >= 0; i-- {
		e := l.entries[i]
		w, h := text.Measure(e.text, face, 0)
		// Em pixels inteiros, o texto fica nítido e o fundo não precisa de
		// suavização (ver hud.DrawProfile).
		boxW := float32(math.Ceil(w)) + 2*padding
		boxH := float32(math.Ceil(h)) + 2*padding
		x, y := right-boxW, bottom-boxH
		a := l.alpha(e)

		back := backColor
		back.A = uint8(float32(back.A) * a)
		vector.FillRect(screen, x, y, boxW, boxH, back, false)

		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(x+padding), float64(y+padding))
		op.ColorScale.ScaleWithColor(levelColors[e.level])
		op.ColorScale.ScaleAlpha(a)
		text.Draw(screen, e.text, face, op)

		bottom = y - lineGap
	}
}
