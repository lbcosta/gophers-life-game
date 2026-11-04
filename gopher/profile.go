package gopher

import (
	"fmt"
	"image/color"
)

// Profile é uma fotografia das características de um Gopher, para serem
// mostradas ao jogador. Não acompanha o Gopher: peça outra quando precisar de
// valores atualizados.
type Profile struct {
	Name      string
	Color     color.RGBA
	ColorName string
	Emotion   Emotion
	Photo     string // nome do arquivo da foto de perfil (ver PhotoName)
	Speed     Speed
	StepSize  float64 // pixels por passo
	Satiety   int
	Alive     bool
	Stage     Stage
	Age       float64 // segundos de vida
	Radius    float64 // raio da hit box, em pixels

	// NextStage é a fase que vem depois da atual, e NextStageIn quantos segundos
	// faltam para ela. HasNextStage é falso para o adulto, que não muda mais.
	HasNextStage bool
	NextStage    Stage
	NextStageIn  float64
}

// colorName devolve o nome da cor de Palette igual a c, ou o código hexadecimal
// dela se não for uma cor da paleta.
func colorName(c color.RGBA) string {
	for i, p := range Palette {
		if p == c {
			return PaletteNames[i]
		}
	}
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// Profile devolve as características atuais do Gopher.
func (g *Gopher) Profile() Profile {
	p := Profile{
		Name:      g.name,
		Color:     g.body,
		ColorName: colorName(g.body),
		Emotion:   g.emotion,
		Photo:     PhotoName(g.body, g.emotion),
		Speed:     g.speed,
		StepSize:  g.stepSize,
		Satiety:   max(g.satiety, 0),
		Alive:     g.Alive(),
		Stage:     g.Stage(),
		Age:       float64(g.age) / ticksPerSecond,
		Radius:    g.hitbox(),
	}
	if p.Stage < Adult {
		p.HasNextStage = true
		p.NextStage = p.Stage + 1
		p.NextStageIn = float64(g.maturity[p.Stage]-g.age) / ticksPerSecond
	}
	return p
}
