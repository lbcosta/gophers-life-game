package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"gophers-life-game/game"
)

// gopherCount é a quantidade de Gophers do jogo.
const gopherCount = 60

func main() {
	ebiten.SetFullscreen(true)
	ebiten.SetWindowTitle("Gophers Life")

	g, err := game.New(gopherCount)
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
