package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mpihlak/gosailing2/pkg/game"
)

func main() {
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Go Sailing!")

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
