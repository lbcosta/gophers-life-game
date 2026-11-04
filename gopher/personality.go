package gopher

import (
	"fmt"
	"image/color"
	"math/rand/v2"
)

// Speed é a velocidade de um Gopher: quantos pixels ele anda a cada passo. Os
// passos continuam acontecendo no mesmo ritmo (walkEveryTicks), então um Gopher
// mais rápido simplesmente anda mais a cada passo.
type Speed int

const (
	Slow Speed = iota
	Normal
	Fast
)

// Speeds lista as velocidades, da mais lenta à mais rápida.
var Speeds = []Speed{Slow, Normal, Fast}

// speedSteps e speedNames são, para cada Speed, o passo em pixels e o nome.
var (
	speedSteps = [...]float64{Slow: 2.5, Normal: 4, Fast: 6}
	speedNames = [...]string{Slow: "lento", Normal: "normal", Fast: "rápido"}
)

// step devolve quantos pixels um Gopher desta velocidade anda a cada passo.
// Entra em pânico se s não for uma das velocidades de Speeds.
func (s Speed) step() float64 {
	return speedSteps[s]
}

func (s Speed) String() string {
	if s < 0 || int(s) >= len(speedNames) {
		return fmt.Sprintf("Speed(%d)", int(s))
	}
	return speedNames[s]
}

// Palette é a paleta de cores dos Gophers: sete cores bem distintas entre si,
// cada uma com uma "personalidade", e que se destacam do gramado e do branco dos
// olhos (os testes conferem essas distâncias). Sem verde, para ninguém se
// camuflar na grama.
var Palette = []color.RGBA{
	{0x00, 0xAD, 0xD8, 255}, // Clássico: o azul do Go
	{0xD8, 0x1F, 0x3F, 255}, // Cereja: vermelho intenso
	{0xFF, 0x7A, 0x2F, 255}, // Tangerina: laranja vibrante
	{0xFF, 0xD2, 0x1F, 255}, // Ouro: amarelo brilhante
	{0xFF, 0x6F, 0xB5, 255}, // Chiclete: rosa doce
	{0x7B, 0x4F, 0xD9, 255}, // Uva: violeta
	{0xA9, 0x74, 0x4F, 255}, // Cacau: marrom terroso
}

// PaletteNames são os nomes das cores de Palette, na mesma ordem.
var PaletteNames = []string{"Clássico", "Cereja", "Tangerina", "Ouro", "Chiclete", "Uva", "Cacau"}

// PaletteSlugs são os nomes das cores de Palette em minúsculas e sem acento, na
// mesma ordem, como aparecem nos nomes dos arquivos das fotos de perfil.
var PaletteSlugs = []string{"classico", "cereja", "tangerina", "ouro", "chiclete", "uva", "cacau"}

// Emotion é a emoção de um Gopher, sorteada no nascimento. Aparece na foto de
// perfil, como o formato dos olhos.
type Emotion int

const (
	Neutral Emotion = iota
	Surprised
	Suspicious
	Sleepy
	Happy

	numEmotions = iota
)

// Emotions lista as emoções.
var Emotions = [numEmotions]Emotion{Neutral, Surprised, Suspicious, Sleepy, Happy}

// emotionNames e emotionSlugs são, para cada Emotion, o nome e o trecho do
// nome do arquivo da foto.
var (
	emotionNames = [numEmotions]string{Neutral: "neutro", Surprised: "surpreso", Suspicious: "desconfiado", Sleepy: "sonolento", Happy: "feliz"}
	emotionSlugs = emotionNames
)

func (e Emotion) String() string {
	if e < 0 || int(e) >= numEmotions {
		return fmt.Sprintf("Emotion(%d)", int(e))
	}
	return emotionNames[e]
}

// PhotoName devolve o nome do arquivo da foto de perfil de um Gopher de corpo
// na cor color e emoção emotion, ou "" se color não for da paleta. A foto é
// sempre da cor do Gopher: o nome é montado a partir da própria cor.
func PhotoName(color color.RGBA, emotion Emotion) string {
	for i, c := range Palette {
		if c == color {
			return PaletteSlugs[i] + "_" + emotionSlugs[emotion] + ".png"
		}
	}
	return ""
}

// NewGophers cria n Gophers, cada um com velocidade e cor sorteadas.
//
// A velocidade é sorteada para cada Gopher, de forma independente. A cor é
// sorteada sem repetição: as cores da paleta são embaralhadas e distribuídas em
// ordem, então os Gophers têm cores diferentes enquanto n não passar do tamanho
// da paleta. Passando disso, a paleta é embaralhada de novo a cada rodada.
//
// Os nomes são sorteados de Names sem repetição: dois Gophers nunca têm o
// mesmo nome (ver assignNames).
//
// A posição deles só deve ser sorteada, por Spawn, quando o tamanho da tela
// for conhecido. Colony já cuida disso.
func NewGophers(n int) []*Gopher {
	gophers := newGophers(n, rand.Perm, rand.IntN)
	assignNames(gophers, rand.Perm)
	return gophers
}

// assignNames dá a cada Gopher um nome diferente de Names, na ordem de uma
// permutação sorteada. perm devolve uma permutação de [0, n). Só se houver mais
// Gophers que nomes a permutação recomeça, e então os nomes se repetem.
func assignNames(gophers []*Gopher, perm func(n int) []int) {
	var order []int
	for i, g := range gophers {
		if i%len(Names) == 0 {
			order = perm(len(Names))
		}
		g.name = Names[order[i%len(Names)]]
	}
}

// newGophers é NewGophers com as fontes de sorteio informadas, para os testes.
// perm devolve uma permutação de [0, n) e intn um inteiro em [0, n).
func newGophers(n int, perm func(n int) []int, intn func(n int) int) []*Gopher {
	gophers := make([]*Gopher, n)
	var order []int
	for i := range gophers {
		if i%len(Palette) == 0 {
			order = perm(len(Palette))
		}
		speed := Speeds[intn(len(Speeds))]
		body := Palette[order[i%len(Palette)]]
		gophers[i] = NewGopher(speed, body)
	}
	return gophers
}
